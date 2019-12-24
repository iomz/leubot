package api

import (
	"crypto/rand"
	"fmt"
)

// Token for the user token
type Token struct {
	Token string `json:"token"`
}

// GenerateToken creates a new token
func GenerateToken() string {
	b := make([]byte, 16) // 16-byte token
	rand.Read(b)
	return fmt.Sprintf("%x", b)
}
