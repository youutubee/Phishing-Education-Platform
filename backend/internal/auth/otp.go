package auth

import (
	"crypto/rand"
	"fmt"
	"math/big"
	"time"
)

func GenerateOTP() string {
	n, _ := rand.Int(rand.Reader, big.NewInt(900000))
	return fmt.Sprintf("%06d", n.Int64()+100000)
}

func GetOTPExpiry() time.Time {
	return time.Now().Add(5 * time.Minute) // 5 minutes default
}
