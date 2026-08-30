package handler

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"

	"logmesh/internal/service"
)

type StreamHandler struct {
	hub *service.EventHub
}

func NewStreamHandler(hub *service.EventHub) *StreamHandler {
	return &StreamHandler{hub: hub}
}

func (h *StreamHandler) Logs(c *gin.Context) {
	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")
	c.Header("X-Accel-Buffering", "no")
	c.Status(http.StatusOK)

	flusher, ok := c.Writer.(http.Flusher)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "streaming unsupported"})
		return
	}

	events := h.hub.Subscribe()
	defer h.hub.Unsubscribe(events)

	fmt.Fprint(c.Writer, ": connected\n\n")
	flusher.Flush()

	for {
		select {
		case <-c.Request.Context().Done():
			return
		case event := <-events:
			payload, err := json.Marshal(event)
			if err != nil {
				continue
			}
			fmt.Fprintf(c.Writer, "event: log\ndata: %s\n\n", payload)
			flusher.Flush()
		}
	}
}
