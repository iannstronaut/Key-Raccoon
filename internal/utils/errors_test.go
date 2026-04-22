package utils_test

import (
	"testing"

	"keyraccoon/internal/utils"
)

func TestNewAppError(t *testing.T) {
	err := utils.NewAppError("bad_request", "invalid payload", 400)
	if err.Code != "bad_request" || err.Message != "invalid payload" || err.StatusCode != 400 {
		t.Fatalf("unexpected app error: %+v", err)
	}
	if err.Error() != "invalid payload" {
		t.Fatalf("Error() = %q, want %q", err.Error(), "invalid payload")
	}
}
