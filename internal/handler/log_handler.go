package handler

import (
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"logmesh/internal/kafka"
	"logmesh/internal/metrics"
	"logmesh/internal/model"
	"logmesh/internal/service"
)

type LogHandler struct {
	logs     service.LogService
	hub      *service.EventHub
	producer kafka.Producer
}

func NewLogHandler(logs service.LogService, hub *service.EventHub, producer kafka.Producer) *LogHandler {
	return &LogHandler{logs: logs, hub: hub, producer: producer}
}

func (h *LogHandler) Ingest(c *gin.Context) {
	var req model.IngestLogRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "request body must be valid JSON"})
		return
	}

	if err := normalizeAndValidateLogRequest(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	applyProjectScope(c, &req)

	event, err := h.logs.Ingest(c.Request.Context(), req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to ingest log"})
		return
	}

	if h.hub != nil {
		h.hub.Publish(event)
	}
	if h.producer != nil {
		_ = h.producer.Publish(c.Request.Context(), event)
	}
	metrics.CountIngest(string(event.Level), event.Service)

	c.JSON(http.StatusAccepted, event)
}

func normalizeAndValidateLogRequest(req *model.IngestLogRequest) error {
	req.Service = strings.TrimSpace(req.Service)
	req.Message = strings.TrimSpace(req.Message)
	req.Level = model.LogLevel(strings.ToUpper(strings.TrimSpace(string(req.Level))))

	if req.Service == "" {
		return errors.New("service is required")
	}
	if req.Message == "" {
		return errors.New("message is required")
	}
	if req.Level == "" {
		req.Level = model.LevelInfo
	}
	if !model.IsValidLogLevel(req.Level) {
		return errors.New("level must be one of TRACE, DEBUG, INFO, WARN, ERROR, FATAL")
	}
	return nil
}

func (h *LogHandler) Search(c *gin.Context) {
	query, err := parseSearchQuery(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if query.ProjectID == "" {
		query.ProjectID = c.GetString("project_id")
	}

	result, err := h.logs.Search(c.Request.Context(), query)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to search logs"})
		return
	}

	c.JSON(http.StatusOK, result)
}

func applyProjectScope(c *gin.Context, req *model.IngestLogRequest) {
	if req.ProjectID == "" {
		req.ProjectID = c.GetString("project_id")
	}
}

func (h *LogHandler) GetByID(c *gin.Context) {
	event, err := h.logs.GetByID(c.Request.Context(), c.Param("id"))
	if errors.Is(err, service.ErrLogNotFound) {
		c.JSON(http.StatusNotFound, gin.H{"error": "log not found"})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to load log"})
		return
	}

	c.JSON(http.StatusOK, event)
}

func parseSearchQuery(c *gin.Context) (model.SearchLogsQuery, error) {
	level := model.LogLevel(strings.ToUpper(strings.TrimSpace(c.Query("level"))))
	if level != "" && !model.IsValidLogLevel(level) {
		return model.SearchLogsQuery{}, errors.New("level must be one of TRACE, DEBUG, INFO, WARN, ERROR, FATAL")
	}

	from, err := parseOptionalTime(c.Query("from"))
	if err != nil {
		return model.SearchLogsQuery{}, errors.New("from must be an RFC3339 timestamp")
	}

	to, err := parseOptionalTime(c.Query("to"))
	if err != nil {
		return model.SearchLogsQuery{}, errors.New("to must be an RFC3339 timestamp")
	}

	limit, err := parseOptionalInt(c.Query("limit"), 100)
	if err != nil {
		return model.SearchLogsQuery{}, errors.New("limit must be a positive integer")
	}

	offset, err := parseOptionalInt(c.Query("offset"), 0)
	if err != nil {
		return model.SearchLogsQuery{}, errors.New("offset must be a positive integer")
	}

	return model.SearchLogsQuery{
		Service:     strings.TrimSpace(c.Query("service")),
		ProjectID:   strings.TrimSpace(c.Query("project_id")),
		Environment: strings.TrimSpace(c.Query("environment")),
		Level:       level,
		Search:      strings.TrimSpace(c.Query("search")),
		TraceID:     strings.TrimSpace(c.Query("trace_id")),
		Host:        strings.TrimSpace(c.Query("host")),
		From:        from,
		To:          to,
		Limit:       limit,
		Offset:      offset,
	}, nil
}

func parseOptionalTime(value string) (*time.Time, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return nil, nil
	}

	parsed, err := time.Parse(time.RFC3339, value)
	if err != nil {
		return nil, err
	}
	return &parsed, nil
}

func parseOptionalInt(value string, fallback int) (int, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return fallback, nil
	}

	parsed, err := strconv.Atoi(value)
	if err != nil || parsed < 0 {
		return 0, errors.New("invalid integer")
	}
	return parsed, nil
}
