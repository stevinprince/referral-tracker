package middleware

import (
	"log/slog"
	"time"

	"github.com/gin-gonic/gin"
)

// RequestLogger is a middleware that logs HTTP requests using slog.
type RequestLogger struct {
	Logger *slog.Logger
}

// NewRequestLogger creates a new RequestLogger.
func NewRequestLogger(logger *slog.Logger) *RequestLogger {
	return &RequestLogger{Logger: logger}
}

// Handler returns a gin middleware that logs each request's method, path, status, and duration.
func (rl *RequestLogger) Handler() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		c.Next()

		rl.Logger.Info("request",
			"method", c.Request.Method,
			"path", c.Request.URL.Path,
			"status", c.Writer.Status(),
			"duration_ms", time.Since(start).Milliseconds(),
			"remote", c.ClientIP(),
		)
	}
}
