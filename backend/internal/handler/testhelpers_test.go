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

// testAuthService creates an auth service backed by a mock session store for
// handler tests that don't need a real DB. It uses an in-memory map behind
// the repository interface, but since we can't easily mock pgxpool, handler
// tests that exercise login/logout/me use the full router with a real DB
// (integration tests). For the health/response unit tests we just need a
// non-nil AuthHandler.
func newStubAuthHandler() *AuthHandler {
	// We can't create a real SessionRepository without a DB pool, but
	// we need a non-nil AuthHandler to construct the router for non-auth tests.
	// Auth-specific handler tests are in auth_test.go with a real DB.
	return &AuthHandler{
		authService: auth.NewService("testuser", "", 24*time.Hour, repository.NewSessionRepository(nil)),
		secure:      false,
	}
}

// newTestRouter creates a router for testing with a no-op logger and stub auth.
func newTestRouter() *gin.Engine {
	logger := slog.New(slog.DiscardHandler)
	reqLogger := middleware.NewRequestLogger(logger)
	return NewRouter(RouterDeps{
		Logger:      reqLogger,
		AuthHandler: newStubAuthHandler(),
		Mode:        "development",
	})
}
