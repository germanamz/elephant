package container

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
)

const tokenBytes = 32

// GenerateToken returns a cryptographically random 256-bit hex-encoded token.
func GenerateToken() (string, error) {
	b := make([]byte, tokenBytes)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("generate token: %w", err)
	}

	return hex.EncodeToString(b), nil
}
