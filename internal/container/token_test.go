package container

import (
	"encoding/hex"
	"testing"
)

func TestGenerateToken_Length(t *testing.T) {
	token, err := GenerateToken()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// 32 bytes = 64 hex characters.
	if len(token) != 64 {
		t.Fatalf("expected 64-char token, got %d chars: %s", len(token), token)
	}
}

func TestGenerateToken_ValidHex(t *testing.T) {
	token, err := GenerateToken()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if _, err := hex.DecodeString(token); err != nil {
		t.Fatalf("token is not valid hex: %v", err)
	}
}

func TestGenerateToken_Unique(t *testing.T) {
	token1, err := GenerateToken()
	if err != nil {
		t.Fatalf("unexpected error generating token 1: %v", err)
	}

	token2, err := GenerateToken()
	if err != nil {
		t.Fatalf("unexpected error generating token 2: %v", err)
	}

	if token1 == token2 {
		t.Fatal("expected two different tokens, got identical values")
	}
}
