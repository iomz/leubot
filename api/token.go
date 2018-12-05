package api

import (
	"crypto/rand"
	"fmt"
)

type Token struct {
	Token string `json:"token"`
}

func GenerateToken() string {
	b := make([]byte, 16) // 16-byte token
	rand.Read(b)
	return fmt.Sprintf("%x", b)
}
