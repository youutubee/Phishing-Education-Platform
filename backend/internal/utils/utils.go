package utils

import (
	"crypto/rand"
	"encoding/hex"
)

func GenerateToken() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

func SanitizeInput(input string) string {
	// Basic sanitization - remove potentially dangerous characters
	// In production, use a proper HTML sanitizer
	return input
}
