package lib

import (
	"crypto/rand"
	"fmt"
	"time"
)

// GenerateUniqueString creates a unique string by combining a username, the current date, and random characters.
func GenerateUniqueString(username string) string {
	// Get the current date in YYYYMMDD format
	currentDate := time.Now().Format("20060102")

	// Generate a random string of 8 characters
	randString := generateRandomString(8)

	// Combine the username, date, and random string
	uniqueString := fmt.Sprintf("%s-%s-%s", username, currentDate, randString)

	return uniqueString
}

// generateRandomString creates a random alphanumeric string of the specified length.
func generateRandomString(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	randBytes := make([]byte, length)
	_, err := rand.Read(randBytes)
	if err != nil {
		panic("Failed to generate random bytes")
	}

	for i, b := range randBytes {
		randBytes[i] = charset[b%byte(len(charset))]
	}

	return string(randBytes)
}
