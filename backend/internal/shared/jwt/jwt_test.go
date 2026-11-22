package jwt_test

import (
	"errors"
	"strings"
	"testing"
	"time"

	"monorepo/backend/internal/shared/jwt"
)

func TestGenerateAndParseToken(t *testing.T) {
	secret := "test-secret"
	token, err := jwt.GenerateToken(secret, 30*time.Minute, "user-123", "alice@example.com", "Alice", "agent")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	claims, err := jwt.ParseToken(secret, token)
	if err != nil {
		t.Fatalf("unexpected parse error: %v", err)
	}
	if claims.Subject != "user-123" {
		t.Fatalf("expected subject user-123, got %q", claims.Subject)
	}
	if claims.Email != "alice@example.com" {
		t.Fatalf("expected email alice@example.com, got %q", claims.Email)
	}
	if claims.Name != "Alice" {
		t.Fatalf("expected name Alice, got %q", claims.Name)
	}
	if claims.Role != "agent" {
		t.Fatalf("expected role agent, got %q", claims.Role)
	}
	if claims.ExpiresAt-claims.IssuedAt != 1800 {
		t.Fatalf("expected 1800s TTL, got %d", claims.ExpiresAt-claims.IssuedAt)
	}
}

func TestParseTokenErrors(t *testing.T) {
	secret := "test-secret"
	token, err := jwt.GenerateToken(secret, 30*time.Minute, "user-123", "alice@example.com", "Alice", "requester")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if _, err := jwt.ParseToken("wrong-secret", token); !errors.Is(err, jwt.ErrInvalidToken) {
		t.Fatalf("expected ErrInvalidToken for wrong secret, got %v", err)
	}

	parts := strings.Split(token, ".")
	tampered := parts[0] + "." + parts[1] + "x." + parts[2]
	if _, err := jwt.ParseToken(secret, tampered); !errors.Is(err, jwt.ErrInvalidToken) {
		t.Fatalf("expected ErrInvalidToken for tampered token, got %v", err)
	}

	if _, err := jwt.ParseToken(secret, "not-a-jwt"); !errors.Is(err, jwt.ErrInvalidToken) {
		t.Fatalf("expected ErrInvalidToken for malformed token, got %v", err)
	}

	expiredToken, err := jwt.GenerateToken(secret, -1*time.Minute, "user-123", "alice@example.com", "Alice", "requester")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if _, err := jwt.ParseToken(secret, expiredToken); !errors.Is(err, jwt.ErrExpiredToken) {
		t.Fatalf("expected ErrExpiredToken for expired token, got %v", err)
	}
}
