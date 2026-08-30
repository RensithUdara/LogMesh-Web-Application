package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"logmesh/internal/service"
)

func APIKeyAuth(keys service.APIKeyService, required bool) gin.HandlerFunc {
	return func(c *gin.Context) {
		value := strings.TrimSpace(c.GetHeader("X-API-Key"))
		if value == "" {
			value = bearerToken(c.GetHeader("Authorization"))
		}

		if value == "" {
			if required {
				c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "api key is required"})
				return
			}
			c.Next()
			return
		}

		key, ok := keys.Authenticate(c.Request.Context(), value)
		if !ok {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "api key is invalid or revoked"})
			return
		}

		c.Set("api_key_id", key.ID)
		c.Next()
	}
}

func bearerToken(header string) string {
	const prefix = "Bearer "
	if strings.HasPrefix(header, prefix) {
		return strings.TrimSpace(strings.TrimPrefix(header, prefix))
	}
	return ""
}
