package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"logmesh/internal/service"
)

type AnalyticsHandler struct {
	analytics *service.AnalyticsService
}

func NewAnalyticsHandler(analytics *service.AnalyticsService) *AnalyticsHandler {
	return &AnalyticsHandler{analytics: analytics}
}

func (h *AnalyticsHandler) Summary(c *gin.Context) {
	summary, err := h.analytics.Summary(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to build analytics"})
		return
	}

	c.JSON(http.StatusOK, summary)
}

func (h *AnalyticsHandler) Sources(c *gin.Context) {
	sources, err := h.analytics.Sources(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to load sources"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"sources": sources})
}
