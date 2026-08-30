package middleware

import (
	"context"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)

func RedisRateLimiter(client *redis.Client, maxRequests int, window time.Duration) gin.HandlerFunc {
	return func(c *gin.Context) {
		if client == nil || maxRequests <= 0 || window <= 0 {
			c.Next()
			return
		}

		ctx := context.Background()
		key := "logmesh:rate:" + c.ClientIP()

		count, err := client.Incr(ctx, key).Result()
		if err != nil {
			c.Next()
			return
		}
		if count == 1 {
			_ = client.Expire(ctx, key, window).Err()
		}

		if count > int64(maxRequests) {
			ttl := client.TTL(ctx, key).Val()
			retryAfter := int(ttl.Seconds())
			if retryAfter < 1 {
				retryAfter = 1
			}
			c.Header("Retry-After", strconv.Itoa(retryAfter))
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{"error": "rate limit exceeded"})
			return
		}

		c.Next()
	}
}

func NewRedisClient(redisURL string) *redis.Client {
	if redisURL == "" {
		return nil
	}

	options, err := redis.ParseURL(redisURL)
	if err != nil {
		return nil
	}
	return redis.NewClient(options)
}
