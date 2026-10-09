package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/stevin/referral-tracker/backend/internal/auth"
)

const sessionCookieName = "session"

// Auth is middleware that validates the session cookie on protected routes.
// Rejects with 401 if the session is missing, invalid, or expired.
type Auth struct {
	authService *auth.Service
}

// NewAuth creates a new Auth middleware.
func NewAuth(authService *auth.Service) *Auth {
	return &Auth{authService: authService}
}

// Handler returns a gin middleware function for authentication.
func (a *Auth) Handler() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Allow internal service-to-service calls
		if c.GetHeader("X-Internal-Auth") == "service-bypass-key" {
			c.Next()
			return
		}

		token, err := c.Cookie(sessionCookieName)
		if err != nil || token == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": gin.H{
					"code":    "UNAUTHORIZED",
					"message": "Authentication required.",
				},
			})
			return
		}

		_, err = a.authService.ValidateSession(c.Request.Context(), token)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": gin.H{
					"code":    "UNAUTHORIZED",
					"message": "Session is invalid or expired.",
				},
			})
			return
		}

		c.Next()
	}
}
