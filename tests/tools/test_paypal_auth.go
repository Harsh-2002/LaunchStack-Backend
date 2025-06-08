package main

import (
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"strings"
)

func main() {
	// Get credentials from environment
	clientID := os.Getenv("PAYPAL_API_KEY")
	secret := os.Getenv("PAYPAL_SECRET")
	
	if clientID == "" || secret == "" {
		fmt.Println("Error: PayPal credentials not found in environment")
		os.Exit(1)
	}
	
	// Create basic auth header
	auth := base64.StdEncoding.EncodeToString([]byte(clientID + ":" + secret))
	
	// Prepare request
	data := url.Values{}
	data.Set("grant_type", "client_credentials")
	
	// Determine endpoint based on mode
	mode := os.Getenv("PAYPAL_MODE")
	var baseURL string
	if mode == "production" {
		baseURL = "https://api-m.paypal.com"
	} else {
		baseURL = "https://api-m.sandbox.paypal.com"
	}
	
	// Create request
	req, err := http.NewRequest("POST", baseURL+"/v1/oauth2/token", strings.NewReader(data.Encode()))
	if err != nil {
		fmt.Printf("Error creating request: %v\n", err)
		os.Exit(1)
	}
	
	// Set headers
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Add("Authorization", "Basic "+auth)
	
	// Execute request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("Error executing request: %v\n", err)
		os.Exit(1)
	}
	defer resp.Body.Close()
	
	// Read response
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("Error reading response: %v\n", err)
		os.Exit(1)
	}
	
	// Check status code
	if resp.StatusCode != http.StatusOK {
		fmt.Printf("Error: Status code %d\nResponse: %s\n", resp.StatusCode, string(body))
		os.Exit(1)
	}
	
	fmt.Printf("PayPal authentication successful!\nToken: %s\n", string(body))
} 