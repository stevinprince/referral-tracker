package handler

import (
	"github.com/gin-gonic/gin"

	"github.com/stevin/referral-tracker/backend/internal/middleware"
)

// NewRouter sets up the gin router with all routes and middleware.
func NewRouter(logger *middleware.RequestLogger, mode string) *gin.Engine {
	if mode != "development" {
		gin.SetMode(gin.ReleaseMode)
	}

	r := gin.New()
	r.SetTrustedProxies(nil)
	r.HandleMethodNotAllowed = true

	// Global middleware
	r.Use(gin.Recovery())
	r.Use(logger.Handler())

	// API routes
	v1 := r.Group("/api/v1")
	{
		v1.GET("/health", HealthCheck)

		// Auth, jobs, extract, companies routes will be added in later tasks.
	}

	// Handle 404 and 405 for API routes
	r.NoRoute(NotFoundHandler)
	r.NoMethod(MethodNotAllowedHandler)

	return r
}
