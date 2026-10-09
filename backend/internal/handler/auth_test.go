package handler

import (
	"bytes"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/stevin/referral-tracker/backend/internal/auth"
	"github.com/stevin/referral-tracker/backend/internal/middleware"
	"github.com/stevin/referral-tracker/backend/internal/repository"
)

func setupAuthTestRouter(t *testing.T) *gin.Engine {
	t.Helper()
	url := os.Getenv("TEST_DATABASE_URL")
	if url == "" {
		t.Skip("TEST_DATABASE_URL not set, skipping integration test")
	}

	ctx := t.Context()
	pool, err := repository.ConnectDB(ctx, url)
	if err != nil {
		t.Fatalf("connecting to test DB: %v", err)
	}
	t.Cleanup(func() { pool.Close() })

	// Clean sessions
	_, _ = pool.Exec(ctx, `DELETE FROM sessions`)

	hash, _ := auth.HashPassword("testpass")
	sessionRepo := repository.NewSessionRepository(pool)
	authSvc := auth.NewService("admin", hash, 24*time.Hour, sessionRepo)
	authHandler := NewAuthHandler(authSvc, false)

	logger := slog.New(slog.DiscardHandler)
	reqLogger := middleware.NewRequestLogger(logger)
	return NewRouter(RouterDeps{
		Logger:      reqLogger,
		AuthHandler: authHandler,
		Mode:        "development",
	})
}

func loginJSON(username, password string) *bytes.Reader {
	body, _ := json.Marshal(map[string]string{"username": username, "password": password})
	return bytes.NewReader(body)
}

func TestAuthLogin_Success(t *testing.T) {
	r := setupAuthTestRouter(t)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", loginJSON("admin", "testpass"))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("status = %d, want %d, body: %s", w.Code, http.StatusOK, w.Body.String())
	}

	// Should set session and csrf cookies
	cookies := w.Result().Cookies()
	var hasSession, hasCSRF bool
	for _, c := range cookies {
		if c.Name == "session" {
			hasSession = true
			if !c.HttpOnly {
				t.Error("session cookie should be HttpOnly")
			}
			if c.SameSite != http.SameSiteStrictMode {
				t.Error("session cookie should be SameSite=Strict")
			}
		}
		if c.Name == "csrf_token" {
			hasCSRF = true
			if c.HttpOnly {
				t.Error("csrf_token cookie should NOT be HttpOnly")
			}
		}
	}
	if !hasSession {
		t.Error("response should set session cookie")
	}
	if !hasCSRF {
		t.Error("response should set csrf_token cookie")
	}
}

func TestAuthLogin_WrongPassword(t *testing.T) {
	r := setupAuthTestRouter(t)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", loginJSON("admin", "wrongpass"))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want %d", w.Code, http.StatusUnauthorized)
	}

	var body ErrorResponse
	json.NewDecoder(w.Body).Decode(&body)
	if body.Error.Code != "INVALID_CREDENTIALS" {
		t.Errorf("error.code = %q, want INVALID_CREDENTIALS", body.Error.Code)
	}
}

func TestAuthLogin_WrongUsername(t *testing.T) {
	r := setupAuthTestRouter(t)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", loginJSON("wrong", "testpass"))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want %d", w.Code, http.StatusUnauthorized)
	}
}

func TestAuthLogin_MissingFields(t *testing.T) {
	r := setupAuthTestRouter(t)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", bytes.NewReader([]byte(`{}`)))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want %d", w.Code, http.StatusBadRequest)
	}
}

// loginAndGetCookie performs a login and returns the session cookie.
func loginAndGetCookie(t *testing.T, r *gin.Engine) *http.Cookie {
	t.Helper()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", loginJSON("admin", "testpass"))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	for _, c := range w.Result().Cookies() {
		if c.Name == "session" {
			return c
		}
	}
	t.Fatal("no session cookie in login response")
	return nil
}

func TestAuthMe_WithValidSession(t *testing.T) {
	r := setupAuthTestRouter(t)
	cookie := loginAndGetCookie(t, r)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/auth/me", nil)
	req.AddCookie(cookie)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("status = %d, want %d, body: %s", w.Code, http.StatusOK, w.Body.String())
	}

	var body map[string]any
	json.NewDecoder(w.Body).Decode(&body)
	if body["authenticated"] != true {
		t.Errorf("authenticated = %v, want true", body["authenticated"])
	}
}

func TestAuthMe_WithoutSession(t *testing.T) {
	r := setupAuthTestRouter(t)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/auth/me", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want %d", w.Code, http.StatusUnauthorized)
	}
}

func TestAuthLogout(t *testing.T) {
	r := setupAuthTestRouter(t)
	cookie := loginAndGetCookie(t, r)

	// Logout
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/logout", nil)
	req.AddCookie(cookie)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("logout status = %d, want %d", w.Code, http.StatusOK)
	}

	// Session should be cleared — check that cookies are set to expire
	for _, c := range w.Result().Cookies() {
		if c.Name == "session" && c.MaxAge > 0 {
			t.Error("session cookie should be cleared after logout")
		}
	}

	// /me should now return 401
	req2 := httptest.NewRequest(http.MethodGet, "/api/v1/auth/me", nil)
	req2.AddCookie(cookie) // old cookie
	w2 := httptest.NewRecorder()
	r.ServeHTTP(w2, req2)

	if w2.Code != http.StatusUnauthorized {
		t.Errorf("me after logout status = %d, want %d", w2.Code, http.StatusUnauthorized)
	}
}

func TestAuthLogout_WithoutSession(t *testing.T) {
	r := setupAuthTestRouter(t)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/logout", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want %d", w.Code, http.StatusUnauthorized)
	}
}
