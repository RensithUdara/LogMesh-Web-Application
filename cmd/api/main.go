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

	"logmesh/internal/config"
	"logmesh/internal/handler"
	"logmesh/internal/middleware"
	"logmesh/internal/service"
)

func main() {
	cfg := config.Load()

	if cfg.Environment == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: cfg.LogLevel,
	}))

	logService := service.NewInMemoryLogService(cfg.MaxStoredLogs)
	logHandler := handler.NewLogHandler(logService)

	router := gin.New()
	router.Use(gin.Recovery(), middleware.RequestLogger(logger))

	router.GET("/healthz", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":      "ok",
			"service":     "logmesh-api",
			"environment": cfg.Environment,
		})
	})

	v1 := router.Group("/v1")
	{
		v1.POST("/logs", logHandler.Ingest)
		v1.GET("/logs", logHandler.Search)
		v1.GET("/logs/:id", logHandler.GetByID)
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
