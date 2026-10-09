package auth

import (
	"testing"
)

func TestVerifyPassword(t *testing.T) {
	hash, err := HashPassword("correctpassword")
	if err != nil {
		t.Fatalf("hashing password: %v", err)
	}

	tests := []struct {
		name      string
		plaintext string
		hash      string
		want      bool
	}{
		{
			name:      "correct password",
			plaintext: "correctpassword",
			hash:      hash,
			want:      true,
		},
		{
			name:      "wrong password",
			plaintext: "wrongpassword",
			hash:      hash,
			want:      false,
		},
		{
			name:      "empty password",
			plaintext: "",
			hash:      hash,
			want:      false,
		},
		{
			name:      "invalid hash",
			plaintext: "correctpassword",
			hash:      "not-a-valid-hash",
			want:      false,
		},
		{
			name:      "empty hash",
			plaintext: "correctpassword",
			hash:      "",
			want:      false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := VerifyPassword(tt.plaintext, tt.hash)
			if got != tt.want {
				t.Errorf("VerifyPassword() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestHashPassword(t *testing.T) {
	hash, err := HashPassword("mypassword")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if hash == "" {
		t.Fatal("hash should not be empty")
	}

	// Hash should start with bcrypt prefix
	if hash[0:4] != "$2a$" {
		t.Errorf("hash should start with $2a$, got %q", hash[0:4])
	}

	// Same password should produce different hashes (salt)
	hash2, err := HashPassword("mypassword")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if hash == hash2 {
		t.Error("two hashes of the same password should differ (different salts)")
	}

	// Both hashes should verify against the original password
	if !VerifyPassword("mypassword", hash) {
		t.Error("first hash should verify")
	}
	if !VerifyPassword("mypassword", hash2) {
		t.Error("second hash should verify")
	}
}

func TestGenerateToken(t *testing.T) {
	token1, err := generateToken()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(token1) != 64 { // 32 bytes = 64 hex chars
		t.Errorf("token length = %d, want 64", len(token1))
	}

	token2, err := generateToken()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if token1 == token2 {
		t.Error("two generated tokens should differ")
	}
}
