package auth

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/stevin/referral-tracker/backend/internal/repository"
)

func testDBPool(t *testing.T) *repository.TestPool {
	t.Helper()
	url := os.Getenv("TEST_DATABASE_URL")
	if url == "" {
		t.Skip("TEST_DATABASE_URL not set, skipping integration test")
	}

	ctx := context.Background()
	pool, err := repository.ConnectDB(ctx, url)
	if err != nil {
		t.Fatalf("connecting to test DB: %v", err)
	}

	// Clean sessions table before each test
	_, err = pool.Exec(ctx, `DELETE FROM sessions`)
	if err != nil {
		pool.Close()
		t.Fatalf("cleaning sessions table: %v", err)
	}

	t.Cleanup(func() { pool.Close() })
	return &repository.TestPool{Pool: pool}
}

func TestLogin_Success(t *testing.T) {
	tp := testDBPool(t)
	ctx := context.Background()

	hash, _ := HashPassword("testpass")
	sessionRepo := repository.NewSessionRepository(tp.Pool)
	svc := NewService("admin", hash, 24*time.Hour, sessionRepo)

	session, err := svc.Login(ctx, "admin", "testpass")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if session == nil {
		t.Fatal("session should not be nil")
	}
	if session.ID == "" {
		t.Error("session ID should not be empty")
	}
	if session.ExpiresAt.Before(time.Now()) {
		t.Error("session should not be already expired")
	}
}

func TestLogin_WrongUsername(t *testing.T) {
	tp := testDBPool(t)
	ctx := context.Background()

	hash, _ := HashPassword("testpass")
	sessionRepo := repository.NewSessionRepository(tp.Pool)
	svc := NewService("admin", hash, 24*time.Hour, sessionRepo)

	_, err := svc.Login(ctx, "wronguser", "testpass")
	if err != ErrInvalidCredentials {
		t.Errorf("err = %v, want ErrInvalidCredentials", err)
	}
}

func TestLogin_WrongPassword(t *testing.T) {
	tp := testDBPool(t)
	ctx := context.Background()

	hash, _ := HashPassword("testpass")
	sessionRepo := repository.NewSessionRepository(tp.Pool)
	svc := NewService("admin", hash, 24*time.Hour, sessionRepo)

	_, err := svc.Login(ctx, "admin", "wrongpass")
	if err != ErrInvalidCredentials {
		t.Errorf("err = %v, want ErrInvalidCredentials", err)
	}
}

func TestValidateSession_Valid(t *testing.T) {
	tp := testDBPool(t)
	ctx := context.Background()

	hash, _ := HashPassword("testpass")
	sessionRepo := repository.NewSessionRepository(tp.Pool)
	svc := NewService("admin", hash, 24*time.Hour, sessionRepo)

	session, _ := svc.Login(ctx, "admin", "testpass")

	validated, err := svc.ValidateSession(ctx, session.ID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if validated.ID != session.ID {
		t.Errorf("validated session ID = %q, want %q", validated.ID, session.ID)
	}
}

func TestValidateSession_NotFound(t *testing.T) {
	tp := testDBPool(t)
	ctx := context.Background()

	hash, _ := HashPassword("testpass")
	sessionRepo := repository.NewSessionRepository(tp.Pool)
	svc := NewService("admin", hash, 24*time.Hour, sessionRepo)

	_, err := svc.ValidateSession(ctx, "nonexistent-token")
	if err != ErrSessionNotFound {
		t.Errorf("err = %v, want ErrSessionNotFound", err)
	}
}

func TestValidateSession_EmptyToken(t *testing.T) {
	tp := testDBPool(t)
	ctx := context.Background()

	hash, _ := HashPassword("testpass")
	sessionRepo := repository.NewSessionRepository(tp.Pool)
	svc := NewService("admin", hash, 24*time.Hour, sessionRepo)

	_, err := svc.ValidateSession(ctx, "")
	if err != ErrSessionNotFound {
		t.Errorf("err = %v, want ErrSessionNotFound", err)
	}
}

func TestValidateSession_Expired(t *testing.T) {
	tp := testDBPool(t)
	ctx := context.Background()

	hash, _ := HashPassword("testpass")
	sessionRepo := repository.NewSessionRepository(tp.Pool)
	// Create service with very short session duration
	svc := NewService("admin", hash, 1*time.Millisecond, sessionRepo)

	session, _ := svc.Login(ctx, "admin", "testpass")

	// Wait for session to expire
	time.Sleep(10 * time.Millisecond)

	_, err := svc.ValidateSession(ctx, session.ID)
	if err != ErrSessionExpired {
		t.Errorf("err = %v, want ErrSessionExpired", err)
	}

	// Expired session should have been cleaned up
	found, _ := sessionRepo.FindByID(ctx, session.ID)
	if found != nil {
		t.Error("expired session should have been deleted from DB")
	}
}

func TestLogout(t *testing.T) {
	tp := testDBPool(t)
	ctx := context.Background()

	hash, _ := HashPassword("testpass")
	sessionRepo := repository.NewSessionRepository(tp.Pool)
	svc := NewService("admin", hash, 24*time.Hour, sessionRepo)

	session, _ := svc.Login(ctx, "admin", "testpass")

	if err := svc.Logout(ctx, session.ID); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Session should no longer be valid
	_, err := svc.ValidateSession(ctx, session.ID)
	if err != ErrSessionNotFound {
		t.Errorf("err = %v, want ErrSessionNotFound after logout", err)
	}
}

func TestCleanupExpiredSessions(t *testing.T) {
	tp := testDBPool(t)
	ctx := context.Background()

	hash, _ := HashPassword("testpass")
	sessionRepo := repository.NewSessionRepository(tp.Pool)

	// Create one expired and one valid session directly
	_, err := sessionRepo.Create(ctx, "expired-token", time.Now().Add(-1*time.Hour))
	if err != nil {
		t.Fatalf("creating expired session: %v", err)
	}
	_, err = sessionRepo.Create(ctx, "valid-token", time.Now().Add(24*time.Hour))
	if err != nil {
		t.Fatalf("creating valid session: %v", err)
	}

	svc := NewService("admin", hash, 24*time.Hour, sessionRepo)
	deleted, err := svc.CleanupExpiredSessions(ctx)
	if err != nil {
		t.Fatalf("cleanup error: %v", err)
	}
	if deleted != 1 {
		t.Errorf("deleted = %d, want 1", deleted)
	}

	// Valid session should still exist
	valid, _ := sessionRepo.FindByID(ctx, "valid-token")
	if valid == nil {
		t.Error("valid session should still exist after cleanup")
	}
}
