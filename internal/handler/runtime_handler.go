package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"logmesh/internal/service"
)

type RuntimeHandler struct {
	runtime *service.RuntimeService
}

func NewRuntimeHandler(runtime *service.RuntimeService) *RuntimeHandler {
	return &RuntimeHandler{runtime: runtime}
}

func (h *RuntimeHandler) Stats(c *gin.Context) {
	stats, err := h.runtime.Stats(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to load runtime stats"})
		return
	}

	c.JSON(http.StatusOK, stats)
}
