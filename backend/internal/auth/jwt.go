package auth

import (
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

func GenerateJWT(userID string, role string) (string, error) {
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		secret = "default-secret-key"
	}

	expiry := os.Getenv("JWT_EXPIRY")
	if expiry == "" {
		expiry = "24h"
	}

	expiryDuration, err := time.ParseDuration(expiry)
	if err != nil {
		expiryDuration = 24 * time.Hour
	}

	claims := jwt.MapClaims{
		"user_id": userID,
		"role":    role,
		"exp":     time.Now().Add(expiryDuration).Unix(),
		"iat":     time.Now().Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secret))
}
