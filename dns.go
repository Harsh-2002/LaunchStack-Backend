package main

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

const (
	host     = "dns.srvr.site"
	username = "Pi"
	password = "9130458959"
	protocol = "https"
)

type DNSRewrite struct {
	Domain string `json:"domain"`
	Answer string `json:"answer"`
}

func createAuthHeader() string {
	auth := username + ":" + password
	return "Basic " + base64.StdEncoding.EncodeToString([]byte(auth))
}

func getDNSRewrites() ([]DNSRewrite, error) {
	client := &http.Client{
		Timeout: 10 * time.Second,
	}
	url := fmt.Sprintf("%s://%s/control/rewrite/list", protocol, host)
	
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	
	req.Header.Add("Authorization", createAuthHeader())
	
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API error: %s - %s", resp.Status, string(body))
	}
	
	var rewrites []DNSRewrite
	err = json.Unmarshal(body, &rewrites)
	if err != nil {
		return nil, err
	}
	
	return rewrites, nil
}

func addDNSRewrite(domain, answer string) error {
	client := &http.Client{
		Timeout: 10 * time.Second,
	}
	url := fmt.Sprintf("%s://%s/control/rewrite/add", protocol, host)
	
	rewrite := DNSRewrite{
		Domain: domain,
		Answer: answer,
	}
	
	jsonData, err := json.Marshal(rewrite)
	if err != nil {
		return err
	}
	
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return err
	}
	
	req.Header.Set("Content-Type", "application/json")
	req.Header.Add("Authorization", createAuthHeader())
	
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	
	body, _ := io.ReadAll(resp.Body)
	
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to add DNS rewrite: %s - %s", resp.Status, string(body))
	}
	
	fmt.Printf("Add response: %s\n", string(body))
	
	return nil
}

func findRewrite(domain string) (*DNSRewrite, error) {
	rewrites, err := getDNSRewrites()
	if err != nil {
		return nil, err
	}
	
	for _, rewrite := range rewrites {
		if rewrite.Domain == domain {
			return &rewrite, nil
		}
	}
	
	return nil, fmt.Errorf("DNS rewrite for domain %s not found", domain)
}

func deleteDNSRewrite(domain string) error {
	// First, get the current value to ensure we have the right answer field
	existingRewrite, err := findRewrite(domain)
	if err != nil {
		return fmt.Errorf("cannot delete rewrite: %v", err)
	}
	
	client := &http.Client{
		Timeout: 10 * time.Second,
	}
	url := fmt.Sprintf("%s://%s/control/rewrite/delete", protocol, host)
	
	// Use the existing rewrite's answer value
	rewrite := DNSRewrite{
		Domain: domain,
		Answer: existingRewrite.Answer,
	}
	
	jsonData, err := json.Marshal(rewrite)
	if err != nil {
		return err
	}
	
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return err
	}
	
	req.Header.Set("Content-Type", "application/json")
	req.Header.Add("Authorization", createAuthHeader())
	
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	
	body, _ := io.ReadAll(resp.Body)
	
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to delete DNS rewrite: %s - %s", resp.Status, string(body))
	}
	
	fmt.Printf("Delete response: %s\n", string(body))
	
	// Add a delay after deletion to allow the API to process
	time.Sleep(1 * time.Second)
	
	// Verify deletion
	verifyRewrites, err := getDNSRewrites()
	if err != nil {
		return fmt.Errorf("failed to verify deletion: %v", err)
	}
	
	for _, rewrite := range verifyRewrites {
		if rewrite.Domain == domain {
			// Record still exists, try one more time with alternative method
			fmt.Printf("Warning: Record still exists after deletion attempt. Trying alternative method.\n")
			
			// Try again with a different endpoint/method
			alternativeUrl := fmt.Sprintf("%s://%s/control/rewrite/delete", protocol, host)
			alternativeData := map[string]string{
				"domain": domain,
			}
			
			altJsonData, _ := json.Marshal(alternativeData)
			altReq, _ := http.NewRequest("POST", alternativeUrl, bytes.NewBuffer(altJsonData))
			altReq.Header.Set("Content-Type", "application/json")
			altReq.Header.Add("Authorization", createAuthHeader())
			
			altResp, altErr := client.Do(altReq)
			if altErr != nil {
				return fmt.Errorf("alternative deletion failed: %v", altErr)
			}
			defer altResp.Body.Close()
			
			// Wait a bit longer after the second attempt
			time.Sleep(2 * time.Second)
			
			// Check once more
			finalRewrites, _ := getDNSRewrites()
			for _, finalRewrite := range finalRewrites {
				if finalRewrite.Domain == domain {
					return fmt.Errorf("failed to delete DNS record after multiple attempts")
				}
			}
			
			return nil
		}
	}
	
	return nil
}

func main() {
	fmt.Println("AdGuard Home DNS Manager")
	fmt.Println("=======================")

	// Get current DNS rewrites
	fmt.Println("Fetching current DNS rewrites...")
	rewrites, err := getDNSRewrites()
	if err != nil {
		fmt.Printf("Error getting DNS rewrites: %v\n", err)
		return
	}
	
	fmt.Println("Current DNS rewrites:")
	if len(rewrites) == 0 {
		fmt.Println("No DNS rewrites found")
	} else {
		for _, rewrite := range rewrites {
			fmt.Printf("  %s -> %s\n", rewrite.Domain, rewrite.Answer)
		}
	}
	
	// Add a new DNS rewrite
	testDomain := "test.example.local"
	testAnswer := "192.168.1.100"
	
	fmt.Printf("\nAdding test DNS rewrite: %s -> %s\n", testDomain, testAnswer)
	err = addDNSRewrite(testDomain, testAnswer)
	if err != nil {
		fmt.Printf("Error adding DNS rewrite: %v\n", err)
		return
	}
	fmt.Println("DNS rewrite added successfully")
	
	// Verify it was added
	fmt.Println("\nVerifying DNS rewrite was added...")
	rewrites, err = getDNSRewrites()
	if err != nil {
		fmt.Printf("Error getting DNS rewrites: %v\n", err)
		return
	}
	
	found := false
	for _, rewrite := range rewrites {
		if rewrite.Domain == testDomain {
			found = true
			fmt.Printf("Found: %s -> %s\n", rewrite.Domain, rewrite.Answer)
			break
		}
	}
	
	if !found {
		fmt.Println("Warning: Added DNS rewrite was not found in the list")
	}
	
	// Delete the DNS rewrite
	fmt.Printf("\nCleaning up: Deleting test DNS rewrite: %s\n", testDomain)
	err = deleteDNSRewrite(testDomain)
	if err != nil {
		fmt.Printf("Error deleting DNS rewrite: %v\n", err)
		return
	}
	fmt.Println("DNS rewrite deleted successfully")
	
	// Verify it was deleted
	// Wait a moment before verifying deletion
	time.Sleep(2 * time.Second)
	
	fmt.Println("\nVerifying DNS rewrite was deleted...")
	rewrites, err = getDNSRewrites()
	if err != nil {
		fmt.Printf("Error getting updated DNS rewrites: %v\n", err)
		return
	}
	
	stillExists := false
	for _, rewrite := range rewrites {
		if rewrite.Domain == testDomain {
			stillExists = true
			fmt.Printf("Warning: DNS rewrite %s was not deleted\n", testDomain)
			break
		}
	}
	
	if !stillExists {
		fmt.Println("Test complete: DNS rewrite was successfully added and deleted")
	} else {
		fmt.Println("Warning: DNS record could not be deleted. Manual cleanup may be required.")
		fmt.Printf("Please visit %s://%s and delete the record manually.\n", protocol, host)
	}
}

