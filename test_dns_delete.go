package main

import (
	"fmt"
	"os"

	"github.com/launchstack/backend/container"
	"github.com/sirupsen/logrus"
)

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
		fmt.Println("Usage: go run test_dns_delete.go <command> [domain] [ip]")
		fmt.Println("Commands: list, add, delete")
		fmt.Println("Example: go run test_dns_delete.go add test.docker 192.168.1.10")
		fmt.Println("Example: go run test_dns_delete.go delete test.docker")
		fmt.Println("Example: go run test_dns_delete.go list")
		os.Exit(1)
	}
	
	command := os.Args[1]
	
	switch command {
	case "list":
		// List all DNS records
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
			fmt.Println("Usage: go run test_dns_delete.go add <domain> <ip>")
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
		
	case "delete":
		// Delete a DNS record
		domain := os.Args[2]
		
		logger.WithField("domain", domain).Info("Deleting DNS record")
		
		err := dnsManager.DeleteDNSRewrite(domain)
		if err != nil {
			logger.WithError(err).Fatal("Failed to delete DNS record")
		}
		
		logger.Info("DNS record deleted successfully")
		
	default:
		fmt.Printf("Unknown command: %s\n", command)
		os.Exit(1)
	}
} 