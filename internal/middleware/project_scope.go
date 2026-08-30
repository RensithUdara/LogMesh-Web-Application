package middleware

import (
	"strings"

	"github.com/gin-gonic/gin"
)

func ProjectScope() gin.HandlerFunc {
	return func(c *gin.Context) {
		projectID := strings.TrimSpace(c.GetHeader("X-Project-ID"))
		if projectID != "" {
			c.Set("project_id", projectID)
		}
		c.Next()
	}
}
