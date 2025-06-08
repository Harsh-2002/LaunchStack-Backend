package main

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v4"
)

func main() {
	// Generate private key
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		fmt.Printf("Error generating RSA keys: %v\n", err)
		os.Exit(1)
	}

	// Create sample claims with hardcoded user ID
	claims := jwt.MapClaims{
		"sub":      "user_2yDNOxLPr7zujKe3hz0Lqdu5TKH",
		"user_id":  "user_2yDNOxLPr7zujKe3hz0Lqdu5TKH",
		"name":     "Test User",
		"email":    "test@example.com",
		"iat":      time.Now().Unix(),
		"exp":      time.Now().Add(time.Hour * 24).Unix(),
		"iss":     "LaunchStack",
		"aud":     "api.launchstack.io",
	}

	// Create token
	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	token.Header["kid"] = "test-key-1"

	// Sign the token
	tokenString, err := token.SignedString(privateKey)
	if err != nil {
		fmt.Printf("Error signing token: %v\n", err)
		os.Exit(1)
	}

	// Output the token
	fmt.Println(tokenString)

	// Output the public key in PEM format for verification
	publicKeyBytes, err := x509.MarshalPKIXPublicKey(&privateKey.PublicKey)
	if err != nil {
		fmt.Printf("Error encoding public key: %v\n", err)
		os.Exit(1)
	}

	publicKeyPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "PUBLIC KEY",
		Bytes: publicKeyBytes,
	})

	// Write public key to a file
	err = os.WriteFile("test_public_key.pem", publicKeyPEM, 0644)
	if err != nil {
		fmt.Printf("Error writing public key file: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("Public key written to test_public_key.pem")
} 