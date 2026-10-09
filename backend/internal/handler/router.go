package handler

import (
	"github.com/gin-gonic/gin"

	"github.com/stevin/referral-tracker/backend/internal/middleware"
)

// RouterDeps holds all dependencies needed to set up routes.
type RouterDeps struct {
	Logger      *middleware.RequestLogger
	AuthHandler *AuthHandler
	Mode        string // "development" or "production"
}

// NewRouter sets up the gin router with all routes and middleware.
func NewRouter(deps RouterDeps) *gin.Engine {
	if deps.Mode != "development" {
		gin.SetMode(gin.ReleaseMode)
	}

	r := gin.New()
	r.HandleMethodNotAllowed = true

	// Global middleware
	r.Use(gin.Recovery())
	r.Use(deps.Logger.Handler())

	// API routes
	v1 := r.Group("/api/v1")
	{
		v1.GET("/health", HealthCheck)

		// Auth routes (no auth middleware required)
		authGroup := v1.Group("/auth")
		{
			authGroup.POST("/login", deps.AuthHandler.Login)
			authGroup.POST("/logout", deps.AuthHandler.Logout)
			authGroup.GET("/me", deps.AuthHandler.Me)
		}

		// Protected routes will be added in later tasks with auth middleware.
	}

	// Handle 404 and 405
	r.NoRoute(NotFoundHandler)
	r.NoMethod(MethodNotAllowedHandler)

	return r
}
