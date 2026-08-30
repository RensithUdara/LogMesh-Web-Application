package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"logmesh/internal/collector"
	"logmesh/internal/model"
)

func (h *LogHandler) ParseAndIngest(c *gin.Context) {
	var req model.ParseTextLogRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "request body must be valid JSON"})
		return
	}

	ingestReq, err := collector.ParseTextLine(req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := normalizeAndValidateLogRequest(&ingestReq); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	event, err := h.logs.Ingest(c.Request.Context(), ingestReq)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to ingest parsed log"})
		return
	}
	if h.hub != nil {
		h.hub.Publish(event)
	}

	c.JSON(http.StatusAccepted, event)
}
