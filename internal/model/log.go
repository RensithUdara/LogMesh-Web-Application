package model

import "time"

type LogLevel string

const (
	LevelTrace LogLevel = "TRACE"
	LevelDebug LogLevel = "DEBUG"
	LevelInfo  LogLevel = "INFO"
	LevelWarn  LogLevel = "WARN"
	LevelError LogLevel = "ERROR"
	LevelFatal LogLevel = "FATAL"
)

type LogEvent struct {
	ID          string                 `json:"id"`
	Timestamp   time.Time              `json:"timestamp"`
	Service     string                 `json:"service"`
	Environment string                 `json:"environment"`
	Level       LogLevel               `json:"level"`
	Message     string                 `json:"message"`
	Host        string                 `json:"host,omitempty"`
	TraceID     string                 `json:"trace_id,omitempty"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
	ReceivedAt  time.Time              `json:"received_at"`
}

type IngestLogRequest struct {
	Timestamp   *time.Time             `json:"timestamp"`
	Service     string                 `json:"service"`
	Environment string                 `json:"environment"`
	Level       LogLevel               `json:"level"`
	Message     string                 `json:"message"`
	Host        string                 `json:"host"`
	TraceID     string                 `json:"trace_id"`
	Metadata    map[string]interface{} `json:"metadata"`
}

type SearchLogsQuery struct {
	Service     string
	Environment string
	Level       LogLevel
	Search      string
	TraceID     string
	Host        string
	From        *time.Time
	To          *time.Time
	Limit       int
	Offset      int
}

type SearchLogsResult struct {
	Logs   []LogEvent `json:"logs"`
	Total  int        `json:"total"`
	Limit  int        `json:"limit"`
	Offset int        `json:"offset"`
}

func IsValidLogLevel(level LogLevel) bool {
	switch level {
	case LevelTrace, LevelDebug, LevelInfo, LevelWarn, LevelError, LevelFatal:
		return true
	default:
		return false
	}
}
