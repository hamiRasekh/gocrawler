package auth

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
)

// GenerateRefreshToken returns a cryptographically secure random token string.
func GenerateRefreshToken() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", fmt.Errorf("failed to generate refresh token: %w", err)
	}

	return hex.EncodeToString(bytes), nil
}

