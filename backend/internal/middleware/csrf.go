package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

const (
	csrfCookieName = "csrf_token"
	csrfHeaderName = "X-CSRF-Token"
)

// CSRF is middleware that validates the CSRF token on state-changing requests.
// Uses the double-submit cookie pattern: the X-CSRF-Token header must match
// the csrf_token cookie value.
func CSRF() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Only check state-changing methods
		switch c.Request.Method {
		case http.MethodPost, http.MethodPut, http.MethodPatch, http.MethodDelete:
			// proceed to check
		default:
			c.Next()
			return
		}

		cookieToken, err := c.Cookie(csrfCookieName)
		if err != nil || cookieToken == "" {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
				"error": gin.H{
					"code":    "CSRF_ERROR",
					"message": "CSRF token cookie is missing.",
				},
			})
			return
		}

		headerToken := c.GetHeader(csrfHeaderName)
		if headerToken == "" {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
				"error": gin.H{
					"code":    "CSRF_ERROR",
					"message": "X-CSRF-Token header is required.",
				},
			})
			return
		}

		if headerToken != cookieToken {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
				"error": gin.H{
					"code":    "CSRF_ERROR",
					"message": "CSRF token mismatch.",
				},
			})
			return
		}

		c.Next()
	}
}
