package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"logmesh/internal/service"
)

func JWTAuth(auth *service.AuthService) gin.HandlerFunc {
	return func(c *gin.Context) {
		tokenValue := bearerToken(c.GetHeader("Authorization"))
		if tokenValue == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "bearer token is required"})
			return
		}

		claims, err := auth.ParseToken(tokenValue)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "bearer token is invalid"})
			return
		}

		if projectID, ok := claims["project_id"].(string); ok && strings.TrimSpace(projectID) != "" {
			c.Set("project_id", projectID)
		}
		c.Next()
	}
}

func OptionalJWTProjectScope(auth *service.AuthService) gin.HandlerFunc {
	return func(c *gin.Context) {
		tokenValue := bearerToken(c.GetHeader("Authorization"))
		if tokenValue == "" {
			c.Next()
			return
		}

		claims, err := auth.ParseToken(tokenValue)
		if err == nil {
			if projectID, ok := claims["project_id"].(string); ok && strings.TrimSpace(projectID) != "" {
				c.Set("project_id", projectID)
			}
		}

		c.Next()
	}
}
