package registry

import (
	"crypto/rand"
	"fmt"
)

// GenerateConnectionId generates a unique connection identifier
func GenerateConnectionId() (string, error) {
	bytes := make([]byte, 16)
	_, err := rand.Read(bytes)
	if err != nil {
		return "", err
	}

	// Format as hex string (32 characters)
	return fmt.Sprintf("%x", bytes), nil
}
