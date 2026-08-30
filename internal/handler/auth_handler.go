package handler

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	"logmesh/internal/model"
	"logmesh/internal/service"
)

type AuthHandler struct {
	auth *service.AuthService
}

func NewAuthHandler(auth *service.AuthService) *AuthHandler {
	return &AuthHandler{auth: auth}
}

func (h *AuthHandler) Register(c *gin.Context) {
	var req model.AuthRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "request body must be valid JSON"})
		return
	}

	response, err := h.auth.Register(c.Request.Context(), req.Name, req.Email, req.Password)
	if errors.Is(err, service.ErrInvalidCredentials) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "name and email are required and password must be at least 8 characters"})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to register user"})
		return
	}

	c.JSON(http.StatusCreated, response)
}

func (h *AuthHandler) Login(c *gin.Context) {
	var req model.AuthRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "request body must be valid JSON"})
		return
	}

	response, err := h.auth.Login(c.Request.Context(), req.Email, req.Password)
	if errors.Is(err, service.ErrInvalidCredentials) {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid email or password"})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to login"})
		return
	}

	c.JSON(http.StatusOK, response)
}
