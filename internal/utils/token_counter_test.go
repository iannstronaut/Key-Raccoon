package utils

import (
	"testing"
)

func TestCountTokens(t *testing.T) {
	tests := []struct {
		input    string
		expected int64
	}{
		{"", 0},
		{"a", 0},
		{"abcd", 1},
		{"abcdefghijklmnop", 4},
		{"Hello world! This is a test.", 7},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := CountTokens(tt.input)
			if result != tt.expected {
				t.Fatalf("CountTokens(%q) = %d, want %d", tt.input, result, tt.expected)
			}
		})
	}
}

func TestCountMessageTokens(t *testing.T) {
	messages := []map[string]string{
		{"role": "user", "content": "Hello!"},
		{"role": "assistant", "content": "Hi there!"},
	}

	result := CountMessageTokens(messages)
	if result <= 0 {
		t.Fatalf("CountMessageTokens() = %d, want > 0", result)
	}

	// Rough estimate: "Hello!" + "Hi there!" + overhead
	expectedMin := int64(2) // at least some tokens
	if result < expectedMin {
		t.Fatalf("CountMessageTokens() = %d, want >= %d", result, expectedMin)
	}
}

func TestCountResponseTokens(t *testing.T) {
	result := CountResponseTokens("Hello world")
	expected := int64(2) // 11 chars / 4 = 2
	if result != expected {
		t.Fatalf("CountResponseTokens() = %d, want %d", result, expected)
	}
}
