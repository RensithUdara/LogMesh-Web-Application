package middleware

import (
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

type rateLimitBucket struct {
	count     int
	resetTime time.Time
}

func RateLimiter(maxRequests int, window time.Duration) gin.HandlerFunc {
	var mu sync.Mutex
	buckets := map[string]rateLimitBucket{}

	return func(c *gin.Context) {
		if maxRequests <= 0 || window <= 0 {
			c.Next()
			return
		}

		key := c.ClientIP()
		now := time.Now()

		mu.Lock()
		bucket := buckets[key]
		if bucket.resetTime.IsZero() || now.After(bucket.resetTime) {
			bucket = rateLimitBucket{resetTime: now.Add(window)}
		}

		if bucket.count >= maxRequests {
			retryAfter := int(time.Until(bucket.resetTime).Seconds())
			if retryAfter < 1 {
				retryAfter = 1
			}
			mu.Unlock()

			c.Header("Retry-After", strconv.Itoa(retryAfter))
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{"error": "rate limit exceeded"})
			return
		}

		bucket.count++
		buckets[key] = bucket
		mu.Unlock()

		c.Next()
	}
}
