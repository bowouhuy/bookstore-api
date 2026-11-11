package auth

import (
	"errors"
	"strings"
	"testing"
	"time"
)

const testSecret = "test-secret-key-that-is-at-least-32-characters"

func TestTokenManager_GenerateAndVerify(t *testing.T) {
	m := NewTokenManager(testSecret, time.Hour)

	token, expiresAt, err := m.Generate(42)
	if err != nil {
		t.Fatalf("generate: unexpected error: %v", err)
	}

	if time.Until(expiresAt) <= 0 {
		t.Errorf("expiresAt should be in the future, got %v", expiresAt)
	}

	userID, err := m.Verify(token)
	if err != nil {
		t.Fatalf("verify: unexpected error: %v", err)
	}
	if userID != 42 {
		t.Errorf("userID = %d, want 42", userID)
	}
}

func TestTokenManager_Verify_RejectsInvalidTokens(t *testing.T) {
	m := NewTokenManager(testSecret, time.Hour)

	// Signed with a different secret.
	otherManager := NewTokenManager("a-completely-different-secret-key-for-testing", time.Hour)
	tokenFromOtherSecret, _, _ := otherManager.Generate(1)

	// Already expired (negative TTL).
	expiredManager := NewTokenManager(testSecret, -time.Minute)
	expiredToken, _, _ := expiredManager.Generate(1)

	// Payload swapped to impersonate user 2.
	tokenUser1, _, _ := m.Generate(1)
	tokenUser2, _, _ := m.Generate(2)
	parts1 := strings.Split(tokenUser1, ".")
	parts2 := strings.Split(tokenUser2, ".")
	tamperedToken := parts1[0] + "." + parts2[1] + "." + parts1[2]

	tests := []struct {
		name  string
		token string
	}{
		{name: "empty string", token: ""},
		{name: "garbage", token: "this-is-not-a-jwt"},
		{name: "signed with different secret", token: tokenFromOtherSecret},
		{name: "expired", token: expiredToken},
		{name: "tampered payload", token: tamperedToken},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := m.Verify(tt.token)
			if !errors.Is(err, ErrInvalidToken) {
				t.Errorf("expected ErrInvalidToken, got %v", err)
			}
		})
	}
}
