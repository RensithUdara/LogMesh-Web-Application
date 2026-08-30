package handler

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	"logmesh/internal/model"
	"logmesh/internal/service"
)

type APIKeyHandler struct {
	keys service.APIKeyService
}

func NewAPIKeyHandler(keys service.APIKeyService) *APIKeyHandler {
	return &APIKeyHandler{keys: keys}
}

func (h *APIKeyHandler) Create(c *gin.Context) {
	var req model.CreateAPIKeyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "request body must be valid JSON"})
		return
	}

	key, err := h.keys.Create(c.Request.Context(), req.Name)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create api key"})
		return
	}

	c.JSON(http.StatusCreated, key)
}

func (h *APIKeyHandler) List(c *gin.Context) {
	keys, err := h.keys.List(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list api keys"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"api_keys": keys})
}

func (h *APIKeyHandler) Revoke(c *gin.Context) {
	if err := h.keys.Revoke(c.Request.Context(), c.Param("id")); errors.Is(err, service.ErrAPIKeyNotFound) {
		c.JSON(http.StatusNotFound, gin.H{"error": "api key not found"})
		return
	} else if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to revoke api key"})
		return
	}

	c.Status(http.StatusNoContent)
}
