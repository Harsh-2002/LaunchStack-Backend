package container

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
)

// DNSRewrite represents a DNS rewrite rule in AdGuard
type DNSRewrite struct {
	Domain string `json:"domain"`
	Answer string `json:"answer"`
}

// DNSManager manages DNS entries in AdGuard
type DNSManager struct {
	logger   *logrus.Logger
	host     string
	username string
	password string
	protocol string
	cliPath  string
}

// NewDNSManager creates a new DNS manager with credentials from environment variables
func NewDNSManager(logger *logrus.Logger) *DNSManager {
	return &DNSManager{
		logger:   logger,
		host:     getEnv("ADGUARD_HOST", "dns.srvr.site"),
		username: getEnv("ADGUARD_USERNAME", "Pi"),
		password: getEnv("ADGUARD_PASSWORD", "9130458959"),
		protocol: getEnv("ADGUARD_PROTOCOL", "https"),
		cliPath:  getEnv("DNS_CLI_PATH", "./dns-cli"),
	}
}

// getEnv gets an environment variable or returns a default value
func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}

// createAuthHeader creates the authorization header for AdGuard requests
func (m *DNSManager) createAuthHeader() string {
	auth := m.username + ":" + m.password
	return "Basic " + base64.StdEncoding.EncodeToString([]byte(auth))
}

// runCLICommand runs the DNS CLI tool with the given arguments
func (m *DNSManager) runCLICommand(args ...string) (string, error) {
	// Check if CLI tool exists
	if _, err := os.Stat(m.cliPath); os.IsNotExist(err) {
		// Fall back to API if CLI tool doesn't exist
		m.logger.Warn("DNS CLI tool not found, falling back to API methods")
		return "", fmt.Errorf("DNS CLI tool not found at %s", m.cliPath)
	}
	
	// Prepare command
	cmd := exec.Command(m.cliPath, args...)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	
	// Run command
	err := cmd.Run()
	if err != nil {
		return "", fmt.Errorf("DNS CLI error: %v - %s", err, stderr.String())
	}
	
	return stdout.String(), nil
}

// GetDNSRewrites fetches all DNS rewrites from AdGuard
func (m *DNSManager) GetDNSRewrites() ([]DNSRewrite, error) {
	// Try CLI first
	output, err := m.runCLICommand("list")
	if err == nil {
		// Parse CLI output
		return m.parseListOutput(output), nil
	}
	
	// Fall back to API
	m.logger.WithError(err).Warn("Failed to get DNS rewrites via CLI, falling back to API")
	
	client := &http.Client{
		Timeout: 10 * time.Second,
	}
	url := fmt.Sprintf("%s://%s/control/rewrite/list", m.protocol, m.host)
	
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	
	req.Header.Add("Authorization", m.createAuthHeader())
	
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

// parseListOutput parses the output of the list command
func (m *DNSManager) parseListOutput(output string) []DNSRewrite {
	var rewrites []DNSRewrite
	lines := strings.Split(output, "\n")
	
	// Skip header lines
	startParsing := false
	for _, line := range lines {
		if strings.Contains(line, "-----") {
			startParsing = true
			continue
		}
		
		if !startParsing || strings.TrimSpace(line) == "" {
			continue
		}
		
		parts := strings.Split(line, "->")
		if len(parts) != 2 {
			continue
		}
		
		domain := strings.TrimSpace(parts[0])
		answer := strings.TrimSpace(parts[1])
		
		rewrites = append(rewrites, DNSRewrite{
			Domain: domain,
			Answer: answer,
		})
	}
	
	return rewrites
}

// FindDNSRewrite finds a specific DNS rewrite by domain
func (m *DNSManager) FindDNSRewrite(domain string) (*DNSRewrite, error) {
	// Try CLI first
	output, err := m.runCLICommand("get", "-domain", domain)
	if err == nil {
		// Parse CLI output
		parts := strings.Split(output, "->")
		if len(parts) == 2 {
			return &DNSRewrite{
				Domain: strings.TrimSpace(parts[0]),
				Answer: strings.TrimSpace(parts[1]),
			}, nil
		}
	}
	
	// Fall back to API
	m.logger.WithError(err).Warn("Failed to find DNS rewrite via CLI, falling back to API")
	
	rewrites, err := m.GetDNSRewrites()
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

// AddDNSRewrite adds a DNS rewrite to AdGuard
func (m *DNSManager) AddDNSRewrite(domain, answer string) error {
	m.logger.WithFields(logrus.Fields{
		"domain": domain,
		"answer": answer,
	}).Info("Adding DNS rewrite")

	// Try CLI first
	_, err := m.runCLICommand("add", "-domain", domain, "-answer", answer)
	if err == nil {
		return nil
	}
	
	// Fall back to API
	m.logger.WithError(err).Warn("Failed to add DNS rewrite via CLI, falling back to API")

	client := &http.Client{
		Timeout: 10 * time.Second,
	}
	url := fmt.Sprintf("%s://%s/control/rewrite/add", m.protocol, m.host)
	
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
	req.Header.Add("Authorization", m.createAuthHeader())
	
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("failed to add DNS rewrite: %s - %s", resp.Status, string(body))
	}
	
	return nil
}

// DeleteDNSRewrite removes a DNS rewrite from AdGuard
func (m *DNSManager) DeleteDNSRewrite(domain string) error {
	m.logger.WithField("domain", domain).Info("Deleting DNS rewrite")

	// Try CLI first - this approach is more reliable
	_, err := m.runCLICommand("delete", "-domain", domain)
	if err == nil {
		m.logger.WithField("domain", domain).Info("Successfully deleted DNS record via CLI")
		return nil
	}
	
	// Fall back to API with improved approach
	m.logger.WithError(err).Warn("Failed to delete DNS rewrite via CLI, falling back to API")

	// First, get the current value to ensure we have the right answer field
	existingRewrite, err := m.FindDNSRewrite(domain)
	if err != nil {
		m.logger.WithError(err).Warn("Cannot find DNS rewrite before deletion")
		// Continue with deletion attempt even if we can't find the record
	}

	client := &http.Client{
		Timeout: 10 * time.Second,
	}
	url := fmt.Sprintf("%s://%s/control/rewrite/delete", m.protocol, m.host)
	
	var jsonData []byte
	if existingRewrite != nil {
		// Use the existing rewrite's domain and answer values
		rewrite := DNSRewrite{
			Domain: domain,
			Answer: existingRewrite.Answer,
		}
		
		m.logger.WithFields(logrus.Fields{
			"domain": domain,
			"answer": existingRewrite.Answer,
		}).Info("Deleting DNS rewrite with full details")
		
		jsonData, err = json.Marshal(rewrite)
		if err != nil {
			return err
		}
	} else {
		// Fallback to just using the domain
		data := map[string]string{
			"domain": domain,
		}
		
		jsonData, err = json.Marshal(data)
		if err != nil {
			return err
		}
	}
	
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return err
	}
	
	req.Header.Set("Content-Type", "application/json")
	req.Header.Add("Authorization", m.createAuthHeader())
	
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	
	body, _ := io.ReadAll(resp.Body)
	
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to delete DNS rewrite: %s - %s", resp.Status, string(body))
	}
	
	// Add a delay after deletion to allow the API to process
	time.Sleep(1 * time.Second)
	
	// Verify deletion
	verifyRewrites, err := m.GetDNSRewrites()
	if err != nil {
		m.logger.WithError(err).Warn("Failed to verify DNS record deletion")
		return nil // Continue anyway, don't fail because of verification error
	}
	
	recordStillExists := false
	for _, rewrite := range verifyRewrites {
		if rewrite.Domain == domain {
			recordStillExists = true
			break
		}
	}
	
	if recordStillExists {
		// Record still exists, try one more time with alternative method
		m.logger.Warn("DNS record still exists after deletion attempt, trying alternative method")
		
		// Try again with a different payload structure
		alternativeData := map[string]string{
			"domain": domain,
		}
		
		altJsonData, _ := json.Marshal(alternativeData)
		altReq, _ := http.NewRequest("POST", url, bytes.NewBuffer(altJsonData))
		altReq.Header.Set("Content-Type", "application/json")
		altReq.Header.Add("Authorization", m.createAuthHeader())
		
		altResp, altErr := client.Do(altReq)
		if altErr != nil {
			m.logger.WithError(altErr).Warn("Alternative deletion method failed")
			return nil // Continue anyway, don't fail
		}
		defer altResp.Body.Close()
		
		// Wait a bit longer after the second attempt
		time.Sleep(2 * time.Second)
		
		// Check once more
		finalRewrites, finalErr := m.GetDNSRewrites()
		if finalErr != nil {
			m.logger.WithError(finalErr).Warn("Failed to verify final DNS record deletion")
			return nil // Continue anyway
		}
		
		for _, finalRewrite := range finalRewrites {
			if finalRewrite.Domain == domain {
				m.logger.WithField("domain", domain).Warn(
					"Failed to delete DNS record after multiple attempts. Manual cleanup may be required.",
				)
				return nil // Continue anyway, but log the warning
			}
		}
		
		m.logger.WithField("domain", domain).Info("DNS record deleted successfully on second attempt")
	} else {
		m.logger.WithField("domain", domain).Info("DNS record deleted successfully")
	}
	
	return nil
} 