package collector

import (
	"errors"
	"regexp"
	"strings"
	"time"

	"logmesh/internal/model"
)

var textLogPattern = regexp.MustCompile(`^(\d{4}-\d{2}-\d{2}[ T]\d{2}:\d{2}:\d{2}(?:Z)?)\s+(TRACE|DEBUG|INFO|WARN|ERROR|FATAL)\s+(.+)$`)

func ParseTextLine(req model.ParseTextLogRequest) (model.IngestLogRequest, error) {
	line := strings.TrimSpace(req.Line)
	if line == "" {
		return model.IngestLogRequest{}, errors.New("line is required")
	}

	parsed := model.IngestLogRequest{
		Service:     req.Service,
		Environment: req.Environment,
		Level:       model.LevelInfo,
		Message:     line,
		Host:        req.Host,
		TraceID:     req.TraceID,
	}

	matches := textLogPattern.FindStringSubmatch(line)
	if len(matches) != 4 {
		return parsed, nil
	}

	timestamp, err := parseTextTimestamp(matches[1])
	if err == nil {
		parsed.Timestamp = &timestamp
	}
	parsed.Level = model.LogLevel(matches[2])
	parsed.Message = strings.TrimSpace(matches[3])

	return parsed, nil
}

func parseTextTimestamp(value string) (time.Time, error) {
	value = strings.Replace(value, " ", "T", 1)
	if strings.HasSuffix(value, "Z") {
		return time.Parse(time.RFC3339, value)
	}
	return time.ParseInLocation("2006-01-02T15:04:05", value, time.UTC)
}
