package service

import (
	"context"
	"errors"
	"slices"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"

	"logmesh/internal/model"
)

var ErrLogNotFound = errors.New("log not found")

type LogService interface {
	Ingest(ctx context.Context, req model.IngestLogRequest) (model.LogEvent, error)
	Search(ctx context.Context, query model.SearchLogsQuery) (model.SearchLogsResult, error)
	GetByID(ctx context.Context, id string) (model.LogEvent, error)
}

type InMemoryLogService struct {
	mu       sync.RWMutex
	logs     []model.LogEvent
	maxItems int
}

func NewInMemoryLogService(maxItems int) *InMemoryLogService {
	return &InMemoryLogService{
		logs:     make([]model.LogEvent, 0),
		maxItems: maxItems,
	}
}

func (s *InMemoryLogService) Ingest(_ context.Context, req model.IngestLogRequest) (model.LogEvent, error) {
	now := time.Now().UTC()
	timestamp := now
	if req.Timestamp != nil {
		timestamp = req.Timestamp.UTC()
	}

	event := model.LogEvent{
		ID:          uuid.NewString(),
		Timestamp:   timestamp,
		Service:     strings.TrimSpace(req.Service),
		Environment: normalizeEnvironment(req.Environment),
		Level:       normalizeLevel(req.Level),
		Message:     strings.TrimSpace(req.Message),
		Host:        strings.TrimSpace(req.Host),
		TraceID:     strings.TrimSpace(req.TraceID),
		Metadata:    maskSensitiveMetadata(req.Metadata),
		ReceivedAt:  now,
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	s.logs = append(s.logs, event)
	if s.maxItems > 0 && len(s.logs) > s.maxItems {
		s.logs = slices.Clone(s.logs[len(s.logs)-s.maxItems:])
	}

	return event, nil
}

func (s *InMemoryLogService) Search(_ context.Context, query model.SearchLogsQuery) (model.SearchLogsResult, error) {
	if query.Limit <= 0 {
		query.Limit = 100
	}
	if query.Limit > 500 {
		query.Limit = 500
	}
	if query.Offset < 0 {
		query.Offset = 0
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	matches := make([]model.LogEvent, 0, len(s.logs))
	for i := len(s.logs) - 1; i >= 0; i-- {
		event := s.logs[i]
		if matchesQuery(event, query) {
			matches = append(matches, event)
		}
	}

	total := len(matches)
	start := min(query.Offset, total)
	end := min(start+query.Limit, total)

	return model.SearchLogsResult{
		Logs:   slices.Clone(matches[start:end]),
		Total:  total,
		Limit:  query.Limit,
		Offset: query.Offset,
	}, nil
}

func (s *InMemoryLogService) GetByID(_ context.Context, id string) (model.LogEvent, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	for i := len(s.logs) - 1; i >= 0; i-- {
		if s.logs[i].ID == id {
			return s.logs[i], nil
		}
	}

	return model.LogEvent{}, ErrLogNotFound
}

func matchesQuery(event model.LogEvent, query model.SearchLogsQuery) bool {
	if query.Service != "" && event.Service != query.Service {
		return false
	}
	if query.Environment != "" && event.Environment != query.Environment {
		return false
	}
	if query.Level != "" && event.Level != query.Level {
		return false
	}
	if query.TraceID != "" && event.TraceID != query.TraceID {
		return false
	}
	if query.Host != "" && event.Host != query.Host {
		return false
	}
	if query.From != nil && event.Timestamp.Before(query.From.UTC()) {
		return false
	}
	if query.To != nil && event.Timestamp.After(query.To.UTC()) {
		return false
	}
	if query.Search != "" && !strings.Contains(strings.ToLower(event.Message), strings.ToLower(query.Search)) {
		return false
	}
	return true
}

func normalizeLevel(level model.LogLevel) model.LogLevel {
	if level == "" {
		return model.LevelInfo
	}
	return model.LogLevel(strings.ToUpper(strings.TrimSpace(string(level))))
}

func normalizeEnvironment(environment string) string {
	environment = strings.TrimSpace(environment)
	if environment == "" {
		return "development"
	}
	return environment
}

func maskSensitiveMetadata(metadata map[string]interface{}) map[string]interface{} {
	if len(metadata) == 0 {
		return nil
	}

	masked := make(map[string]interface{}, len(metadata))
	for key, value := range metadata {
		if isSensitiveKey(key) {
			masked[key] = "[REDACTED]"
			continue
		}

		if nested, ok := value.(map[string]interface{}); ok {
			masked[key] = maskSensitiveMetadata(nested)
			continue
		}

		masked[key] = value
	}
	return masked
}

func isSensitiveKey(key string) bool {
	normalized := strings.ToLower(key)
	sensitiveFragments := []string{"password", "token", "secret", "authorization", "credit_card", "apikey", "api_key"}
	for _, fragment := range sensitiveFragments {
		if strings.Contains(normalized, fragment) {
			return true
		}
	}
	return false
}
