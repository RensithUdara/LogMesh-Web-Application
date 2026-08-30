package config

import (
	"log/slog"
	"os"
	"strconv"
	"strings"
)

type Config struct {
	Environment   string
	HTTPAddr      string
	LogLevel      slog.Level
	MaxStoredLogs int
}

func Load() Config {
	return Config{
		Environment:   envString("LOGMESH_ENV", "development"),
		HTTPAddr:      envString("LOGMESH_HTTP_ADDR", ":8080"),
		LogLevel:      parseLogLevel(envString("LOGMESH_LOG_LEVEL", "info")),
		MaxStoredLogs: envInt("LOGMESH_MAX_STORED_LOGS", 10000),
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
