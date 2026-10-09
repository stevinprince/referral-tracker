package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestRateLimiter_ByIP_AllowsWithinLimit(t *testing.T) {
	rl := NewRateLimiter(10, 10) // 10 per second, burst 10
	r := gin.New()
	r.Use(rl.ByIP())
	r.GET("/test", func(c *gin.Context) { c.JSON(200, gin.H{"ok": true}) })

	// First few requests should succeed
	for i := 0; i < 5; i++ {
		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("request %d: status = %d, want %d", i, w.Code, http.StatusOK)
		}
	}
}

func TestRateLimiter_ByIP_RejectsOverLimit(t *testing.T) {
	rl := NewRateLimiter(0.001, 2) // Very low rate, burst of 2
	r := gin.New()
	r.Use(rl.ByIP())
	r.GET("/test", func(c *gin.Context) { c.JSON(200, gin.H{"ok": true}) })

	// Exhaust the burst
	for i := 0; i < 2; i++ {
		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		if w.Code != http.StatusOK {
			t.Fatalf("burst request %d should succeed, got %d", i, w.Code)
		}
	}

	// Next request should be rate-limited
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusTooManyRequests {
		t.Errorf("status = %d, want %d", w.Code, http.StatusTooManyRequests)
	}
}

func TestRateLimiter_BySession_AllowsWithoutSession(t *testing.T) {
	rl := NewRateLimiter(0.001, 1) // Very restrictive
	r := gin.New()
	r.Use(rl.BySession())
	r.GET("/test", func(c *gin.Context) { c.JSON(200, gin.H{"ok": true}) })

	// Without a session cookie, rate limiting is skipped
	for i := 0; i < 5; i++ {
		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("request %d without session: status = %d, want %d", i, w.Code, http.StatusOK)
		}
	}
}

func TestRateLimiter_BySession_RejectsOverLimit(t *testing.T) {
	rl := NewRateLimiter(0.001, 1) // Very low rate, burst of 1
	r := gin.New()
	r.Use(rl.BySession())
	r.GET("/test", func(c *gin.Context) { c.JSON(200, gin.H{"ok": true}) })

	sessionCookie := &http.Cookie{Name: "session", Value: "test-session-id"}

	// First request succeeds (uses burst)
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.AddCookie(sessionCookie)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("first request should succeed, got %d", w.Code)
	}

	// Second request exceeds limit
	req2 := httptest.NewRequest(http.MethodGet, "/test", nil)
	req2.AddCookie(sessionCookie)
	w2 := httptest.NewRecorder()
	r.ServeHTTP(w2, req2)

	if w2.Code != http.StatusTooManyRequests {
		t.Errorf("status = %d, want %d", w2.Code, http.StatusTooManyRequests)
	}
}

func TestRateLimiter_ByIP_DifferentIPsIndependent(t *testing.T) {
	rl := NewRateLimiter(0.001, 1)
	r := gin.New()
	r.Use(rl.ByIP())
	r.GET("/test", func(c *gin.Context) { c.JSON(200, gin.H{"ok": true}) })

	// Exhaust limit for 10.0.0.1
	req1 := httptest.NewRequest(http.MethodGet, "/test", nil)
	req1.RemoteAddr = "10.0.0.1:1234"
	w1 := httptest.NewRecorder()
	r.ServeHTTP(w1, req1)

	// 10.0.0.1 should now be limited
	req1b := httptest.NewRequest(http.MethodGet, "/test", nil)
	req1b.RemoteAddr = "10.0.0.1:1234"
	w1b := httptest.NewRecorder()
	r.ServeHTTP(w1b, req1b)
	if w1b.Code != http.StatusTooManyRequests {
		t.Errorf("10.0.0.1 second request: status = %d, want 429", w1b.Code)
	}

	// 10.0.0.2 should still be allowed (different IP, independent limiter)
	req2 := httptest.NewRequest(http.MethodGet, "/test", nil)
	req2.RemoteAddr = "10.0.0.2:1234"
	w2 := httptest.NewRecorder()
	r.ServeHTTP(w2, req2)
	if w2.Code != http.StatusOK {
		t.Errorf("10.0.0.2 first request: status = %d, want 200", w2.Code)
	}
}
