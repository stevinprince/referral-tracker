package auth

import "errors"

var (
	// ErrInvalidCredentials is returned when the password is wrong.
	ErrInvalidCredentials = errors.New("invalid credentials")

	// ErrUserNotFound is returned when the username doesn't exist.
	ErrUserNotFound = errors.New("user not found")

	// ErrSessionNotFound is returned when the session token doesn't exist.
	ErrSessionNotFound = errors.New("session not found")

	// ErrSessionExpired is returned when the session has passed its expiry time.
	ErrSessionExpired = errors.New("session expired")
)
