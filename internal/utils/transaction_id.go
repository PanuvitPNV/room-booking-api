package utils

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"time"
)

// GenerateTransactionID generates a unique transaction ID
func GenerateTransactionID() string {
	// Use timestamp for ordering
	timestamp := time.Now().UnixNano()

	// Add random bytes for uniqueness
	randomBytes := make([]byte, 8)
	_, err := rand.Read(randomBytes)
	if err != nil {
		// Fallback if random generation fails
		return fmt.Sprintf("tx-%d", timestamp)
	}

	// Combine timestamp and random bytes
	return fmt.Sprintf("tx-%d-%s", timestamp, hex.EncodeToString(randomBytes))
}
