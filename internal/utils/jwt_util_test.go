package utils_test

import (
	"testing"

	"keyraccoon/internal/utils"
)

func TestGenerateAndValidateAccessToken(t *testing.T) {
	token, err := utils.GenerateAccessToken(7, "user@example.com", "admin", "test-secret", 30)
	if err != nil {
		t.Fatalf("GenerateAccessToken() error = %v", err)
	}

	claims, err := utils.ValidateToken(token, "test-secret")
	if err != nil {
		t.Fatalf("ValidateToken() error = %v", err)
	}

	if claims.UserID != 7 || claims.Email != "user@example.com" || claims.Role != "admin" {
		t.Fatalf("unexpected claims: %+v", claims)
	}
	if claims.TokenType != "access" {
		t.Fatalf("TokenType = %q, want access", claims.TokenType)
	}
}

func TestGenerateRefreshToken(t *testing.T) {
	token, err := utils.GenerateRefreshToken(9, "refresh@example.com", "user", "test-secret")
	if err != nil {
		t.Fatalf("GenerateRefreshToken() error = %v", err)
	}

	claims, err := utils.ValidateToken(token, "test-secret")
	if err != nil {
		t.Fatalf("ValidateToken() error = %v", err)
	}
	if claims.TokenType != "refresh" {
		t.Fatalf("TokenType = %q, want refresh", claims.TokenType)
	}
}

func TestValidateTokenErrorCases(t *testing.T) {
	if _, err := utils.ValidateToken("whatever", ""); err == nil {
		t.Fatal("ValidateToken() error = nil for empty secret")
	}

	token, err := utils.GenerateAccessToken(1, "user@example.com", "user", "right-secret", 30)
	if err != nil {
		t.Fatalf("GenerateAccessToken() error = %v", err)
	}

	if _, err := utils.ValidateToken(token, "wrong-secret"); err == nil {
		t.Fatal("ValidateToken() error = nil for wrong secret")
	}
}
