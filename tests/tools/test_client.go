package main

import (
	"fmt"
	"io"
	"net/http"
	"time"
)

func main() {
	// Simple client to test the instances endpoint
	fmt.Println("Sending request to /api/v1/instances/...")
	
	// Create a client with a timeout
	client := &http.Client{
		Timeout: 10 * time.Second,
	}
	
	// Create the request
	req, err := http.NewRequest("GET", "http://localhost:8080/api/v1/instances/", nil)
	if err != nil {
		fmt.Printf("Error creating request: %v\n", err)
		return
	}
	
	// Set headers
	req.Header.Set("Origin", "https://dev.srvr.site")
	// Add a dummy authorization header - this works because in development mode, authentication is bypassed
	// See middleware/auth.go:AuthMiddleware
	req.Header.Set("Authorization", "Bearer test-token")
	
	// Send the request
	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("Error sending request: %v\n", err)
		return
	}
	defer resp.Body.Close()
	
	// Read the response
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("Error reading response: %v\n", err)
		return
	}
	
	// Print the response
	fmt.Printf("Status Code: %d\n", resp.StatusCode)
	fmt.Printf("Response Body: %s\n", string(body))
} 