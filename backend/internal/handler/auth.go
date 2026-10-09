package handler

import (
	"encoding/hex"
	"errors"
	"fmt"
	"math/rand"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/stevin/referral-tracker/backend/internal/auth"
)

const (
	sessionCookieName = "session"
	csrfCookieName    = "csrf_token"
)

// AuthHandler handles authentication endpoints.
type AuthHandler struct {
	authService *auth.Service
	secure      bool // true in production (sets Secure flag on cookies)
}

// NewAuthHandler creates a new AuthHandler.
func NewAuthHandler(authService *auth.Service, secure bool) *AuthHandler {
	return &AuthHandler{
		authService: authService,
		secure:      secure,
	}
}

type loginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// Login handles POST /api/v1/auth/login.
func (h *AuthHandler) Login(c *gin.Context) {
	var req loginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		JSONError(c, http.StatusBadRequest, "VALIDATION_ERROR", "Username and password are required.")
		return
	}

	session, err := h.authService.Login(c.Request.Context(), req.Username, req.Password)
	if err != nil {
		if errors.Is(err, auth.ErrUserNotFound) {
			JSONError(c, http.StatusUnauthorized, "USER_NOT_FOUND", "User not found.")
			return
		}
		if errors.Is(err, auth.ErrInvalidCredentials) {
			JSONError(c, http.StatusUnauthorized, "INVALID_CREDENTIALS", "Invalid password.")
			return
		}
		JSONError(c, http.StatusInternalServerError, "INTERNAL_ERROR", fmt.Sprintf("Login failed: %v", err))
		return
	}

	// Set session cookie
	h.setSessionCookie(c, session.ID, int(session.ExpiresAt.Sub(session.CreatedAt).Seconds()))

	// Set CSRF cookie
	csrfToken, err := generateCSRFToken()
	if err != nil {
		JSONError(c, http.StatusInternalServerError, "INTERNAL_ERROR", "An unexpected error occurred.")
		return
	}
	h.setCSRFCookie(c, csrfToken, int(session.ExpiresAt.Sub(session.CreatedAt).Seconds()))

	c.JSON(http.StatusOK, gin.H{
		"message":  "Login successful.",
		"username": req.Username,
	})
}

// Logout handles POST /api/v1/auth/logout.
func (h *AuthHandler) Logout(c *gin.Context) {
	token, err := c.Cookie(sessionCookieName)
	if err != nil || token == "" {
		JSONError(c, http.StatusUnauthorized, "UNAUTHORIZED", "Not authenticated.")
		return
	}

	if err := h.authService.Logout(c.Request.Context(), token); err != nil {
		JSONError(c, http.StatusInternalServerError, "INTERNAL_ERROR", "An unexpected error occurred.")
		return
	}

	// Clear cookies
	h.clearSessionCookie(c)
	h.clearCSRFCookie(c)

	c.JSON(http.StatusOK, gin.H{"message": "Logged out."})
}

// Me handles GET /api/v1/auth/me.
func (h *AuthHandler) Me(c *gin.Context) {
	token, err := c.Cookie(sessionCookieName)
	if err != nil || token == "" {
		JSONError(c, http.StatusUnauthorized, "UNAUTHORIZED", "Not authenticated.")
		return
	}

	_, err = h.authService.ValidateSession(c.Request.Context(), token)
	if err != nil {
		h.clearSessionCookie(c)
		h.clearCSRFCookie(c)
		JSONError(c, http.StatusUnauthorized, "UNAUTHORIZED", "Session is invalid or expired.")
		return
	}

	c.JSON(http.StatusOK, gin.H{"authenticated": true})
}

func (h *AuthHandler) setSessionCookie(c *gin.Context, value string, maxAge int) {
	c.SetSameSite(http.SameSiteStrictMode)
	c.SetCookie(sessionCookieName, value, maxAge, "/", "", h.secure, false)
}

func (h *AuthHandler) clearSessionCookie(c *gin.Context) {
	c.SetSameSite(http.SameSiteStrictMode)
	c.SetCookie(sessionCookieName, "", -1, "/", "", h.secure, true)
}

func (h *AuthHandler) setCSRFCookie(c *gin.Context, value string, maxAge int) {
	c.SetSameSite(http.SameSiteStrictMode)
	c.SetCookie(csrfCookieName, value, maxAge, "/", "", h.secure, false)
}

func (h *AuthHandler) clearCSRFCookie(c *gin.Context) {
	c.SetSameSite(http.SameSiteStrictMode)
	c.SetCookie(csrfCookieName, "", -1, "/", "", h.secure, false)
}

func generateCSRFToken() (string, error) {
	b := make([]byte, 32)
	rand.Read(b)
	return hex.EncodeToString(b), nil
}
