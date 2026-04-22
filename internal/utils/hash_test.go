package utils_test

import (
	"testing"

	"keyraccoon/internal/utils"
)

func TestHashPasswordAndCheckPasswordHash(t *testing.T) {
	hashed, err := utils.HashPassword("super-secret")
	if err != nil {
		t.Fatalf("HashPassword() error = %v", err)
	}
	if hashed == "super-secret" {
		t.Fatal("hashed password should not equal plaintext")
	}
	if !utils.CheckPasswordHash("super-secret", hashed) {
		t.Fatal("CheckPasswordHash() = false, want true")
	}
	if utils.CheckPasswordHash("wrong", hashed) {
		t.Fatal("CheckPasswordHash() = true for wrong password, want false")
	}
}
