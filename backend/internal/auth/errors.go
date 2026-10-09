package auth

import "errors"

var (
	// ErrInvalidCredentials is returned when the username or password is wrong.
	ErrInvalidCredentials = errors.New("invalid credentials")

	// ErrSessionNotFound is returned when the session token doesn't exist.
	ErrSessionNotFound = errors.New("session not found")

	// ErrSessionExpired is returned when the session has passed its expiry time.
	ErrSessionExpired = errors.New("session expired")
)
