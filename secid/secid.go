// Package secid (i.e. secure identifier) creates identifiers using cryptographically secure random number generation
package secid

import (
	"crypto/rand"
	"errors"
	"fmt"

	"github.com/lytics/base62"
)

// String creates a secure identifier string, uses length for how many bytes of randomness it will use
func String(length int) (string, error) {
	if length < 1 {
		return "", errors.New("length must be positive")
	}
	// Create the array for the random bytes
	randomData := make([]byte, length)
	// Fill it with random bytes from crypto/rand
	written, err := rand.Read(randomData)
	if err != nil {
		return "", fmt.Errorf("could not read random bytes: %w", err)
	}
	// Encode to base62 and return the resulting string
	return base62.StdEncoding.EncodeToString(randomData[:written]), nil
}
