package main

import (
	"fmt"
	"os"
	"time"

	"github.com/sirupsen/logrus"
)

// Import our container package that contains the DNS manager
// Update this import path if needed to match your project structure
import "github.com/launchstack/backend/container"

func main() {
	// Setup logger
	logger := logrus.New()
	logger.SetLevel(logrus.InfoLevel)
	logger.SetFormatter(&logrus.TextFormatter{
		FullTimestamp: true,
	})
	
	// Create DNS manager
	dnsManager := container.NewDNSManager(logger)
	
	// Check command-line arguments
	if len(os.Args) < 2 {
		fmt.Println("Usage: go run test_improved_dns_delete.go <command> [domain] [ip]")
		fmt.Println("Commands: list, add, delete, test")
		fmt.Println("Examples:")
		fmt.Println("  go run test_improved_dns_delete.go list")
		fmt.Println("  go run test_improved_dns_delete.go add test-dns.docker 10.1.2.3")
		fmt.Println("  go run test_improved_dns_delete.go delete test-dns.docker")
		fmt.Println("  go run test_improved_dns_delete.go test")
		os.Exit(1)
	}
	
	command := os.Args[1]
	
	switch command {
	case "list":
		// List all DNS records
		fmt.Println("Listing all DNS records...")
		records, err := dnsManager.GetDNSRewrites()
		if err != nil {
			logger.WithError(err).Fatal("Failed to list DNS records")
		}
		
		fmt.Println("DNS Records:")
		for _, record := range records {
			fmt.Printf("Domain: %-30s Answer: %s\n", record.Domain, record.Answer)
		}
		
	case "add":
		// Add a DNS record
		if len(os.Args) < 4 {
			fmt.Println("Usage: go run test_improved_dns_delete.go add <domain> <ip>")
			os.Exit(1)
		}
		
		domain := os.Args[2]
		ip := os.Args[3]
		
		logger.WithFields(logrus.Fields{
			"domain": domain,
			"ip":     ip,
		}).Info("Adding DNS record")
		
		err := dnsManager.AddDNSRewrite(domain, ip)
		if err != nil {
			logger.WithError(err).Fatal("Failed to add DNS record")
		}
		
		logger.Info("DNS record added successfully")
		
		// Verify the record was added
		records, err := dnsManager.GetDNSRewrites()
		if err != nil {
			logger.WithError(err).Fatal("Failed to verify DNS record was added")
		}
		
		found := false
		for _, record := range records {
			if record.Domain == domain {
				found = true
				logger.WithFields(logrus.Fields{
					"domain": record.Domain,
					"ip":     record.Answer,
				}).Info("Verified DNS record was added")
				break
			}
		}
		
		if !found {
			logger.Warn("Could not verify DNS record was added")
		}
		
	case "delete":
		// Delete a DNS record
		if len(os.Args) < 3 {
			fmt.Println("Usage: go run test_improved_dns_delete.go delete <domain>")
			os.Exit(1)
		}
		
		domain := os.Args[2]
		
		// Check if record exists first
		existingRecord, err := dnsManager.FindDNSRewrite(domain)
		if err != nil {
			logger.WithError(err).Warn("Record not found before deletion attempt")
		} else {
			logger.WithFields(logrus.Fields{
				"domain": existingRecord.Domain,
				"ip":     existingRecord.Answer,
			}).Info("Found DNS record to delete")
		}
		
		// Delete the record with our improved approach
		logger.WithField("domain", domain).Info("Deleting DNS record")
		
		err = dnsManager.DeleteDNSRewrite(domain)
		if err != nil {
			logger.WithError(err).Fatal("Failed to delete DNS record")
		}
		
		logger.Info("DNS deletion operation completed")
		
		// Wait a moment to let changes propagate
		time.Sleep(2 * time.Second)
		
		// Verify the record was deleted
		records, err := dnsManager.GetDNSRewrites()
		if err != nil {
			logger.WithError(err).Fatal("Failed to verify DNS record deletion")
		}
		
		stillExists := false
		for _, record := range records {
			if record.Domain == domain {
				stillExists = true
				logger.WithFields(logrus.Fields{
					"domain": record.Domain,
					"ip":     record.Answer,
				}).Warn("DNS record still exists after deletion")
				break
			}
		}
		
		if !stillExists {
			logger.Info("Verified DNS record was successfully deleted")
		}
		
	case "test":
		// Run a full test cycle: add, verify, delete, verify
		testDomain := "test-improved-dns.docker"
		testIP := "10.1.2.99"
		
		logger.Info("=== Starting DNS API Test ===")
		
		// Add test record
		logger.WithFields(logrus.Fields{
			"domain": testDomain,
			"ip":     testIP,
		}).Info("Adding test DNS record")
		
		err := dnsManager.AddDNSRewrite(testDomain, testIP)
		if err != nil {
			logger.WithError(err).Fatal("Failed to add test DNS record")
		}
		
		logger.Info("Test DNS record added successfully")
		
		// Wait a moment
		time.Sleep(2 * time.Second)
		
		// Verify the record was added
		records, err := dnsManager.GetDNSRewrites()
		if err != nil {
			logger.WithError(err).Fatal("Failed to verify test DNS record was added")
		}
		
		found := false
		for _, record := range records {
			if record.Domain == testDomain {
				found = true
				logger.WithFields(logrus.Fields{
					"domain": record.Domain,
					"ip":     record.Answer,
				}).Info("Verified test DNS record was added")
				break
			}
		}
		
		if !found {
			logger.Warn("Could not verify test DNS record was added")
		}
		
		// Delete the test record
		logger.WithField("domain", testDomain).Info("Deleting test DNS record")
		
		err = dnsManager.DeleteDNSRewrite(testDomain)
		if err != nil {
			logger.WithError(err).Fatal("Failed to delete test DNS record")
		}
		
		logger.Info("DNS deletion operation completed")
		
		// Wait a moment to let changes propagate
		time.Sleep(2 * time.Second)
		
		// Verify the record was deleted
		records, err = dnsManager.GetDNSRewrites()
		if err != nil {
			logger.WithError(err).Fatal("Failed to verify test DNS record deletion")
		}
		
		stillExists := false
		for _, record := range records {
			if record.Domain == testDomain {
				stillExists = true
				logger.WithFields(logrus.Fields{
					"domain": record.Domain,
					"ip":     record.Answer,
				}).Warn("Test DNS record still exists after deletion")
				break
			}
		}
		
		if !stillExists {
			logger.Info("Verified test DNS record was successfully deleted")
			logger.Info("=== DNS API Test Completed Successfully ===")
		} else {
			logger.Warn("=== DNS API Test Completed with Issues ===")
		}
		
	default:
		fmt.Printf("Unknown command: %s\n", command)
		os.Exit(1)
	}
} 