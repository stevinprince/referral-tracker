package handler

import (
	"log/slog"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/stevin/referral-tracker/backend/internal/auth"
	"github.com/stevin/referral-tracker/backend/internal/middleware"
	"github.com/stevin/referral-tracker/backend/internal/repository"
)

func init() {
	gin.SetMode(gin.TestMode)
}

// newStubAuthHandler creates a non-nil AuthHandler for tests that don't exercise auth.
func newStubAuthHandler() *AuthHandler {
	return &AuthHandler{
		authService: auth.NewService("testuser", "", 24*time.Hour, repository.NewSessionRepository(nil)),
		secure:      false,
	}
}

// newTestRouter creates a router for testing with a no-op logger and stub auth.
func newTestRouter() *gin.Engine {
	logger := slog.New(slog.DiscardHandler)
	reqLogger := middleware.NewRequestLogger(logger)
	stubAuth := newStubAuthHandler()

	return NewRouter(RouterDeps{
		Logger:         reqLogger,
		AuthHandler:    stubAuth,
		AuthMiddleware: middleware.NewAuth(stubAuth.authService),
		LoginLimiter:   middleware.NewRateLimiter(100, 100),
		ExtractLimiter: middleware.NewRateLimiter(100, 100),
		Mode:           "development",
	})
}
