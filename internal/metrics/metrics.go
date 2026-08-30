package metrics

import (
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
)

var (
	HTTPRequests = prometheus.NewCounterVec(
		prometheus.CounterOpts{Name: "logmesh_http_requests_total", Help: "Total HTTP requests."},
		[]string{"method", "path", "status"},
	)
	HTTPRequestDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{Name: "logmesh_http_request_duration_seconds", Help: "HTTP request duration in seconds."},
		[]string{"method", "path"},
	)
	IngestedLogs = prometheus.NewCounterVec(
		prometheus.CounterOpts{Name: "logmesh_ingested_logs_total", Help: "Total ingested logs."},
		[]string{"level", "service"},
	)
)

func Register() {
	prometheus.MustRegister(HTTPRequests, HTTPRequestDuration, IngestedLogs)
}

func Middleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		c.Next()

		path := c.FullPath()
		if path == "" {
			path = c.Request.URL.Path
		}
		status := strconv.Itoa(c.Writer.Status())

		HTTPRequests.WithLabelValues(c.Request.Method, path, status).Inc()
		HTTPRequestDuration.WithLabelValues(c.Request.Method, path).Observe(time.Since(start).Seconds())
	}
}

func CountIngest(level, service string) {
	IngestedLogs.WithLabelValues(level, service).Inc()
}
