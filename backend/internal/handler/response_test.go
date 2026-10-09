package handler

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestJSONError(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/", nil)

	JSONError(c, 400, "VALIDATION_ERROR", "Title is required.")

	resp := w.Result()
	defer resp.Body.Close()

	if resp.StatusCode != 400 {
		t.Errorf("status = %d, want 400", resp.StatusCode)
	}

	var body ErrorResponse
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		t.Fatalf("decoding: %v", err)
	}
	if body.Error.Code != "VALIDATION_ERROR" {
		t.Errorf("error.code = %q, want %q", body.Error.Code, "VALIDATION_ERROR")
	}
	if body.Error.Message != "Title is required." {
		t.Errorf("error.message = %q, want %q", body.Error.Message, "Title is required.")
	}
}

func TestAbortWithError(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/", nil)

	AbortWithError(c, 401, "UNAUTHORIZED", "Authentication required.")

	if !c.IsAborted() {
		t.Error("expected context to be aborted")
	}

	resp := w.Result()
	defer resp.Body.Close()

	if resp.StatusCode != 401 {
		t.Errorf("status = %d, want 401", resp.StatusCode)
	}

	var body ErrorResponse
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		t.Fatalf("decoding: %v", err)
	}
	if body.Error.Code != "UNAUTHORIZED" {
		t.Errorf("error.code = %q, want %q", body.Error.Code, "UNAUTHORIZED")
	}
}
