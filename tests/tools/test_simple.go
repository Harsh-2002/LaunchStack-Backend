package main

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"strings"
)

func main() {
	// Get PayPal credentials from environment
	clientID := os.Getenv("PAYPAL_API_KEY")
	secret := os.Getenv("PAYPAL_SECRET")
	
	if clientID == "" || secret == "" {
		fmt.Println("Error: PayPal credentials not found in environment")
		os.Exit(1)
	}
	
	// Create basic auth header
	auth := base64.StdEncoding.EncodeToString([]byte(clientID + ":" + secret))
	
	// Get access token
	tokenResponse, err := getAccessToken(auth)
	if err != nil {
		fmt.Printf("Error getting access token: %v\n", err)
		os.Exit(1)
	}
	
	// Extract access token
	accessToken := tokenResponse["access_token"].(string)
	fmt.Printf("Access Token: %s\n\n", accessToken)
	
	// Create an order
	orderResponse, err := createOrder(accessToken)
	if err != nil {
		fmt.Printf("Error creating order: %v\n", err)
		os.Exit(1)
	}
	
	fmt.Printf("Order Response:\n%s\n\n", prettyPrint(orderResponse))
	
	// Get the approval URL
	var approvalURL string
	for _, link := range orderResponse["links"].([]interface{}) {
		linkObj := link.(map[string]interface{})
		if linkObj["rel"].(string) == "approve" {
			approvalURL = linkObj["href"].(string)
			break
		}
	}
	
	if approvalURL != "" {
		fmt.Printf("Approval URL: %s\n", approvalURL)
		fmt.Println("\nOpen this URL in your browser to complete the payment")
	}
}

func getAccessToken(auth string) (map[string]interface{}, error) {
	// Determine endpoint based on mode
	mode := os.Getenv("PAYPAL_MODE")
	var baseURL string
	if mode == "production" {
		baseURL = "https://api-m.paypal.com"
	} else {
		baseURL = "https://api-m.sandbox.paypal.com"
	}
	
	// Prepare request
	data := url.Values{}
	data.Set("grant_type", "client_credentials")
	
	// Create request
	req, err := http.NewRequest("POST", baseURL+"/v1/oauth2/token", strings.NewReader(data.Encode()))
	if err != nil {
		return nil, err
	}
	
	// Set headers
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Add("Authorization", "Basic "+auth)
	
	// Execute request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	
	// Read response
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	
	// Check status code
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("error: status code %d, response: %s", resp.StatusCode, string(body))
	}
	
	// Parse response
	var result map[string]interface{}
	err = json.Unmarshal(body, &result)
	if err != nil {
		return nil, err
	}
	
	return result, nil
}

func createOrder(accessToken string) (map[string]interface{}, error) {
	// Determine endpoint based on mode
	mode := os.Getenv("PAYPAL_MODE")
	var baseURL string
	if mode == "production" {
		baseURL = "https://api-m.paypal.com"
	} else {
		baseURL = "https://api-m.sandbox.paypal.com"
	}
	
	// Create order payload
	orderData := map[string]interface{}{
		"intent": "CAPTURE",
		"purchase_units": []map[string]interface{}{
			{
				"amount": map[string]interface{}{
					"currency_code": "USD",
					"value": "5.00",
				},
				"description": "LaunchStack Pro Plan",
			},
		},
		"application_context": map[string]interface{}{
			"return_url": "http://localhost:3000/dashboard?success=true",
			"cancel_url": "http://localhost:3000/pricing?canceled=true",
		},
	}
	
	// Convert to JSON
	orderJSON, err := json.Marshal(orderData)
	if err != nil {
		return nil, err
	}
	
	// Create request
	req, err := http.NewRequest("POST", baseURL+"/v2/checkout/orders", bytes.NewBuffer(orderJSON))
	if err != nil {
		return nil, err
	}
	
	// Set headers
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Authorization", "Bearer "+accessToken)
	
	// Execute request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	
	// Read response
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	
	// Check status code
	if resp.StatusCode != http.StatusCreated {
		return nil, fmt.Errorf("error: status code %d, response: %s", resp.StatusCode, string(body))
	}
	
	// Parse response
	var result map[string]interface{}
	err = json.Unmarshal(body, &result)
	if err != nil {
		return nil, err
	}
	
	return result, nil
}

func prettyPrint(v interface{}) string {
	jsonData, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return fmt.Sprintf("%v", v)
	}
	return string(jsonData)
} 