package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/stevin/referral-tracker/backend/internal/auth"
)

const SessionCookieName = "session"
const session_cookie_name = SessionCookieName

// Auth is middleware that validates the session cookie on protected routes.
// Rejects with 401 if the session is missing, invalid, or expired.
type Auth struct {
	auth_service *auth.Service
}

// NewAuth creates a new Auth middleware.
func NewAuth(authService *auth.Service) *Auth {
	return &Auth{auth_service: authService}
}

// Handler returns a gin middleware function for authentication.
func (a *Auth) Handler() gin.HandlerFunc {
	return func(c *gin.Context) {
		token, err := c.Cookie(session_cookie_name)
		if err != nil || token == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": gin.H{
					"code":    "UNAUTHORIZED",
					"message": "Authentication required.",
				},
			})
			return
		}

		_, err = a.auth_service.ValidateSession(c.Request.Context(), token)
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
