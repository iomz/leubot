package api

import (
	"crypto/rand"
	"fmt"
)

// Token represents the token object
type Token struct {
	Token string `json:"token"`
}

// GenerateToken generates a new token and return as a string
func GenerateToken() string {
	b := make([]byte, 16) // 16-byte token
	rand.Read(b)
	return fmt.Sprintf("%x", b)
}
