package api

import (
	"strconv"
	"time"

	"github.com/gofiber/adaptor/v2"
	"github.com/gofiber/fiber/v2"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	httpRequestDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name: "http_request_duration_seconds",
			Help: "HTTP request duration in seconds",
			// 20 buckets: 0.5ms to 200ms
			Buckets: []float64{
				0.0005, 0.001, 0.0015, 0.002, 0.003, 0.004, 0.005, 0.007, 0.01,
				0.015, 0.02, 0.03, 0.04, 0.05, 0.07, 0.1, 0.12, 0.15, 0.18, 0.2,
			},
		},
		[]string{"method", "path", "status"},
	)

	httpRequestsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_requests_total",
			Help: "Total number of HTTP requests",
		},
		[]string{"method", "path", "status"},
	)
)

func init() {
	prometheus.MustRegister(httpRequestDuration)
	prometheus.MustRegister(httpRequestsTotal)
}

// MetricsMiddleware records request duration and count
func MetricsMiddleware() fiber.Handler {
	return func(c *fiber.Ctx) error {
		start := time.Now()

		err := c.Next()

		duration := time.Since(start).Seconds()
		status := strconv.Itoa(c.Response().StatusCode())
		path := normalizePath(c.Route().Path)
		method := c.Method()

		httpRequestDuration.WithLabelValues(method, path, status).Observe(duration)
		httpRequestsTotal.WithLabelValues(method, path, status).Inc()

		return err
	}
}

// normalizePath normalizes path to avoid high cardinality
func normalizePath(path string) string {
	if path == "" {
		return "unknown"
	}
	return path
}

// MetricsHandler returns prometheus metrics endpoint handler
func MetricsHandler() fiber.Handler {
	return adaptor.HTTPHandler(promhttp.Handler())
}
