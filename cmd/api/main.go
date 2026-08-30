package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/redis/go-redis/v9"

	"logmesh/internal/config"
	"logmesh/internal/handler"
	"logmesh/internal/kafka"
	"logmesh/internal/metrics"
	"logmesh/internal/middleware"
	"logmesh/internal/repository"
	"logmesh/internal/service"
)

func main() {
	config.LoadDotEnv(".env")
	cfg := config.Load()

	if cfg.Environment == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: cfg.LogLevel,
	}))
	metrics.Register()

	logService := service.NewInMemoryLogService(cfg.MaxStoredLogs)
	var keyService service.APIKeyService = service.NewInMemoryAPIKeyService()
	authService := service.NewAuthService(cfg.JWTSecret)
	eventHub := service.NewEventHub()
	logProducer := kafka.NewProducer(cfg.KafkaBrokers, cfg.KafkaLogsTopic)
	defer logProducer.Close()
	redisClient := middleware.NewRedisClient(cfg.RedisURL)
	if redisClient != nil {
		defer redisClient.Close()
	}
	postgresPool, err := repository.NewPostgresPool(context.Background(), cfg.PostgresURL)
	if err != nil {
		logger.Error("postgres setup failed", "error", err)
	} else if postgresPool != nil {
		defer postgresPool.Close()
		if err := repository.EnsureMetadataSchema(context.Background(), postgresPool); err != nil {
			logger.Error("postgres schema setup failed", "error", err)
		} else {
			keyService = service.NewPostgresAPIKeyService(postgresPool)
		}
	}

	logHandler := handler.NewLogHandler(logService, eventHub, logProducer)
	analyticsHandler := handler.NewAnalyticsHandler(service.NewAnalyticsService(logService))
	keyHandler := handler.NewAPIKeyHandler(keyService)
	authHandler := handler.NewAuthHandler(authService)
	streamHandler := handler.NewStreamHandler(eventHub)
	exportHandler := handler.NewExportHandler(service.NewExportService(logService))
	runtimeHandler := handler.NewRuntimeHandler(service.NewRuntimeService(logService))

	router := gin.New()
	router.Use(
		gin.Recovery(),
		middleware.CORS(),
		middleware.ProjectScope(),
		middleware.OptionalJWTProjectScope(authService),
		rateLimiterMiddleware(redisClient, cfg),
		metrics.Middleware(),
		middleware.RequestLogger(logger),
	)

	router.GET("/healthz", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":      "ok",
			"service":     "logmesh-api",
			"environment": cfg.Environment,
		})
	})
	router.GET("/metrics", gin.WrapH(promhttp.Handler()))

	v1 := router.Group("/v1")
	{
		v1.POST("/auth/register", authHandler.Register)
		v1.POST("/auth/login", authHandler.Login)
		v1.POST("/logs", middleware.APIKeyAuth(keyService, cfg.RequireAPIKey), logHandler.Ingest)
		v1.POST("/logs/bulk", middleware.APIKeyAuth(keyService, cfg.RequireAPIKey), logHandler.BulkIngest)
		v1.POST("/logs/parse", middleware.APIKeyAuth(keyService, cfg.RequireAPIKey), logHandler.ParseAndIngest)
		v1.GET("/logs/export", exportHandler.LogsCSV)
		v1.GET("/logs", logHandler.Search)
		v1.GET("/logs/:id", logHandler.GetByID)
		v1.GET("/stream/logs", streamHandler.Logs)
		v1.GET("/analytics", analyticsHandler.Summary)
		v1.GET("/sources", analyticsHandler.Sources)
		v1.GET("/runtime", runtimeHandler.Stats)
		v1.GET("/api-keys", keyHandler.List)
		v1.POST("/api-keys", keyHandler.Create)
		v1.DELETE("/api-keys/:id", keyHandler.Revoke)
	}

	server := &http.Server{
		Addr:              cfg.HTTPAddr,
		Handler:           router,
		ReadHeaderTimeout: 5 * time.Second,
	}

	go func() {
		logger.Info("starting api", "addr", cfg.HTTPAddr)
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Error("api failed", "error", err)
			os.Exit(1)
		}
	}()

	shutdownCtx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	<-shutdownCtx.Done()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		logger.Error("graceful shutdown failed", "error", err)
		os.Exit(1)
	}

	logger.Info("api stopped")
}

func rateLimiterMiddleware(redisClient *redis.Client, cfg config.Config) gin.HandlerFunc {
	window := time.Duration(cfg.RateLimitWindow) * time.Second
	if redisClient != nil {
		return middleware.RedisRateLimiter(redisClient, cfg.RateLimitRequests, window)
	}
	return middleware.RateLimiter(cfg.RateLimitRequests, window)
}
