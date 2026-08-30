package config

import (
	"log/slog"
	"os"
	"strconv"
	"strings"
)

type Config struct {
	Environment       string
	HTTPAddr          string
	LogLevel          slog.Level
	MaxStoredLogs     int
	RequireAPIKey     bool
	RateLimitRequests int
	RateLimitWindow   int
}

func Load() Config {
	return Config{
		Environment:       envString("LOGMESH_ENV", "development"),
		HTTPAddr:          envString("LOGMESH_HTTP_ADDR", ":8081"),
		LogLevel:          parseLogLevel(envString("LOGMESH_LOG_LEVEL", "info")),
		MaxStoredLogs:     envInt("LOGMESH_MAX_STORED_LOGS", 10000),
		RequireAPIKey:     envBool("LOGMESH_REQUIRE_API_KEY", false),
		RateLimitRequests: envInt("LOGMESH_RATE_LIMIT_REQUESTS", 120),
		RateLimitWindow:   envInt("LOGMESH_RATE_LIMIT_WINDOW_SECONDS", 60),
	}
}

func envString(key, fallback string) string {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback
	}
	return value
}

func envInt(key string, fallback int) int {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback
	}

	parsed, err := strconv.Atoi(value)
	if err != nil || parsed <= 0 {
		return fallback
	}
	return parsed
}

func envBool(key string, fallback bool) bool {
	value := strings.ToLower(strings.TrimSpace(os.Getenv(key)))
	if value == "" {
		return fallback
	}
	return value == "true" || value == "1" || value == "yes"
}

func parseLogLevel(value string) slog.Level {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "debug":
		return slog.LevelDebug
	case "warn", "warning":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}
