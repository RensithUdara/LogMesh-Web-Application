package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"logmesh/internal/metrics"
	"logmesh/internal/model"
)

func (h *LogHandler) BulkIngest(c *gin.Context) {
	var req model.BulkIngestLogRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "request body must be valid JSON"})
		return
	}
	if len(req.Logs) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "logs must contain at least one item"})
		return
	}
	if len(req.Logs) > 500 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "bulk ingest accepts at most 500 logs"})
		return
	}

	accepted := make([]model.LogEvent, 0, len(req.Logs))
	for index := range req.Logs {
		item := req.Logs[index]
		if err := normalizeAndValidateLogRequest(&item); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error(), "index": index})
			return
		}

		event, err := h.logs.Ingest(c.Request.Context(), item)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to ingest bulk logs"})
			return
		}
		if h.hub != nil {
			h.hub.Publish(event)
		}
		if h.producer != nil {
			_ = h.producer.Publish(c.Request.Context(), event)
		}
		metrics.CountIngest(string(event.Level), event.Service)
		accepted = append(accepted, event)
	}

	c.JSON(http.StatusAccepted, model.BulkIngestLogResponse{
		Accepted: len(accepted),
		Logs:     accepted,
	})
}
