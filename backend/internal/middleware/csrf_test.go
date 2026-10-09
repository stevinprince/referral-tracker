package middleware

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func csrfTestRouter() *gin.Engine {
	r := gin.New()
	r.Use(CSRF())
	r.POST("/test", func(c *gin.Context) { c.JSON(200, gin.H{"ok": true}) })
	r.GET("/test", func(c *gin.Context) { c.JSON(200, gin.H{"ok": true}) })
	r.PUT("/test", func(c *gin.Context) { c.JSON(200, gin.H{"ok": true}) })
	r.DELETE("/test", func(c *gin.Context) { c.JSON(200, gin.H{"ok": true}) })
	r.PATCH("/test", func(c *gin.Context) { c.JSON(200, gin.H{"ok": true}) })
	return r
}

func TestCSRF_GET_Skipped(t *testing.T) {
	r := csrfTestRouter()

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("GET should skip CSRF check, got status %d", w.Code)
	}
}

func TestCSRF_POST_MissingCookie(t *testing.T) {
	r := csrfTestRouter()

	req := httptest.NewRequest(http.MethodPost, "/test", nil)
	req.Header.Set("X-CSRF-Token", "some-token")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusForbidden {
		t.Errorf("status = %d, want %d", w.Code, http.StatusForbidden)
	}

	var body map[string]any
	json.NewDecoder(w.Body).Decode(&body)
	errObj := body["error"].(map[string]any)
	if errObj["code"] != "CSRF_ERROR" {
		t.Errorf("error.code = %v, want CSRF_ERROR", errObj["code"])
	}
}

func TestCSRF_POST_MissingHeader(t *testing.T) {
	r := csrfTestRouter()

	req := httptest.NewRequest(http.MethodPost, "/test", nil)
	req.AddCookie(&http.Cookie{Name: "csrf_token", Value: "token123"})
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusForbidden {
		t.Errorf("status = %d, want %d", w.Code, http.StatusForbidden)
	}
}

func TestCSRF_POST_Mismatch(t *testing.T) {
	r := csrfTestRouter()

	req := httptest.NewRequest(http.MethodPost, "/test", nil)
	req.AddCookie(&http.Cookie{Name: "csrf_token", Value: "token-cookie"})
	req.Header.Set("X-CSRF-Token", "token-header-different")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusForbidden {
		t.Errorf("status = %d, want %d", w.Code, http.StatusForbidden)
	}
}

func TestCSRF_POST_Valid(t *testing.T) {
	r := csrfTestRouter()

	req := httptest.NewRequest(http.MethodPost, "/test", nil)
	req.AddCookie(&http.Cookie{Name: "csrf_token", Value: "matching-token"})
	req.Header.Set("X-CSRF-Token", "matching-token")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", w.Code, http.StatusOK)
	}
}

func TestCSRF_AllStateMethods(t *testing.T) {
	r := csrfTestRouter()
	methods := []string{http.MethodPut, http.MethodPatch, http.MethodDelete}

	for _, method := range methods {
		t.Run(method+"_rejected_without_csrf", func(t *testing.T) {
			req := httptest.NewRequest(method, "/test", nil)
			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)

			if w.Code != http.StatusForbidden {
				t.Errorf("%s status = %d, want %d", method, w.Code, http.StatusForbidden)
			}
		})

		t.Run(method+"_passes_with_csrf", func(t *testing.T) {
			req := httptest.NewRequest(method, "/test", nil)
			req.AddCookie(&http.Cookie{Name: "csrf_token", Value: "tok"})
			req.Header.Set("X-CSRF-Token", "tok")
			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)

			if w.Code != http.StatusOK {
				t.Errorf("%s status = %d, want %d", method, w.Code, http.StatusOK)
			}
		})
	}
}
