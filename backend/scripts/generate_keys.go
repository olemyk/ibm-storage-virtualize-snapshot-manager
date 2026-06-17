package main

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
)

func main() {
	// Generate encryption key (32 bytes for AES-256)
	encKey := make([]byte, 32)
	rand.Read(encKey)

	// Generate JWT secret (32 bytes)
	jwtSecret := make([]byte, 32)
	rand.Read(jwtSecret)

	fmt.Println("# Generated secure keys")
	fmt.Printf("ENCRYPTION_KEY=%s\n", base64.StdEncoding.EncodeToString(encKey))
	fmt.Printf("JWT_SECRET=%s\n", base64.StdEncoding.EncodeToString(jwtSecret))
}

//
