package auth

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"time"

	"golang.org/x/crypto/bcrypt"

	"github.com/stevin/referral-tracker/backend/internal/repository"
)

// Service handles authentication and session management.
type Service struct {
	username        string
	passwordHash    string
	sessionDuration time.Duration
	sessions        *repository.SessionRepository
}

// NewService creates a new auth Service.
func NewService(username, passwordHash string, sessionDuration time.Duration, sessions *repository.SessionRepository) *Service {
	return &Service{
		username:        username,
		passwordHash:    passwordHash,
		sessionDuration: sessionDuration,
		sessions:        sessions,
	}
}

// VerifyPassword checks if the given plaintext password matches the stored hash.
func VerifyPassword(plaintext, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(plaintext))
	return err == nil
}

// HashPassword generates a bcrypt hash from a plaintext password.
// Used by the hash-password CLI tool and tests.
func HashPassword(plaintext string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(plaintext), 12)
	if err != nil {
		return "", fmt.Errorf("hashing password: %w", err)
	}
	return string(hash), nil
}

// dummyHash is a pre-computed bcrypt hash used when the username doesn't match.
// This ensures bcrypt always runs, preventing timing-based user enumeration.
var dummyHash = func() []byte {
	h, err := bcrypt.GenerateFromPassword([]byte("dummy"), 12)
	if err != nil {
		panic(fmt.Sprintf("auth: failed to generate dummy hash: %v", err))
	}
	return h
}()

// Login verifies credentials and creates a new session on success.
// Returns the session or an error if credentials are invalid.
func (s *Service) Login(ctx context.Context, username, password string) (*repository.Session, error) {
	if username != s.username {
		// Always run bcrypt to prevent timing oracle on username enumeration
		bcrypt.CompareHashAndPassword(dummyHash, []byte(password))
		return nil, ErrInvalidCredentials
	}
	if !VerifyPassword(password, s.passwordHash) {
		return nil, ErrInvalidCredentials
	}

	token, err := generateToken()
	if err != nil {
		return nil, fmt.Errorf("generating session token: %w", err)
	}

	expiresAt := time.Now().Add(s.sessionDuration)
	session, err := s.sessions.Create(ctx, token, expiresAt)
	if err != nil {
		return nil, fmt.Errorf("creating session: %w", err)
	}

	return session, nil
}

// ValidateSession checks if a session token is valid and not expired.
// Returns the session or an error.
func (s *Service) ValidateSession(ctx context.Context, token string) (*repository.Session, error) {
	if token == "" {
		return nil, ErrSessionNotFound
	}

	session, err := s.sessions.FindByID(ctx, token)
	if err != nil {
		return nil, fmt.Errorf("finding session: %w", err)
	}
	if session == nil {
		return nil, ErrSessionNotFound
	}

	if time.Now().After(session.ExpiresAt) {
		// Clean up the expired session
		_ = s.sessions.Delete(ctx, token)
		return nil, ErrSessionExpired
	}

	return session, nil
}

// Logout deletes the session associated with the given token.
func (s *Service) Logout(ctx context.Context, token string) error {
	if err := s.sessions.Delete(ctx, token); err != nil {
		return fmt.Errorf("deleting session: %w", err)
	}
	return nil
}

// CleanupExpiredSessions removes all expired sessions from the database.
func (s *Service) CleanupExpiredSessions(ctx context.Context) (int64, error) {
	return s.sessions.DeleteExpired(ctx)
}

// generateToken creates a cryptographically random 32-byte hex-encoded token.
func generateToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}
