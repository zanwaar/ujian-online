package utils_test

import (
	"testing"
	"time"

	"golang-backend/utils"

	"github.com/golang-jwt/jwt/v5"
)

func TestValidateTokenRejectsNonHS256Token(t *testing.T) {
	t.Setenv("JWT_SECRET", "test-secret")
	token, err := jwt.NewWithClaims(jwt.SigningMethodHS512, jwt.MapClaims{"user_id": "user-1", "exp": time.Now().Add(time.Hour).Unix()}).SignedString([]byte("test-secret"))
	if err != nil {
		t.Fatalf("SignedString() error = %v", err)
	}

	if _, err := utils.ValidateToken(token); err == nil {
		t.Fatal("ValidateToken() should reject HS512 tokens")
	}
}

func TestValidateTokenRejectsExpiredToken(t *testing.T) {
	t.Setenv("JWT_SECRET", "test-secret")
	token, err := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{"user_id": "user-1", "exp": time.Now().Add(-time.Minute).Unix()}).SignedString([]byte("test-secret"))
	if err != nil {
		t.Fatalf("SignedString() error = %v", err)
	}

	parsed, err := utils.ValidateToken(token)
	if err == nil || parsed.Valid {
		t.Fatalf("ValidateToken() = token=%v, error=%v; want expired token rejection", parsed, err)
	}
}
