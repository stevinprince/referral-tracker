package handler

import (
	"log/slog"

	"github.com/gin-gonic/gin"
	"github.com/stevin/referral-tracker/backend/internal/middleware"
)

func init() {
	gin.SetMode(gin.TestMode)
}

// newTestRouter creates a router for testing with a no-op logger.
func newTestRouter() *gin.Engine {
	logger := slog.New(slog.DiscardHandler)
	reqLogger := middleware.NewRequestLogger(logger)
	return NewRouter(reqLogger, "development")
}
