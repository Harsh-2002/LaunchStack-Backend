package main

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"
)

// Configuration
var (
	host     string
	username string
	password string
	protocol string
)

// DNSRewrite represents the structure for AdGuard DNS rewrite rules
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

	return nil
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

	return nil
}

func listRewrites() {
	rewrites, err := getDNSRewrites()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error getting DNS rewrites: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("Current DNS rewrites:")
	fmt.Printf("%-40s -> %s\n", "DOMAIN", "ANSWER")
	fmt.Println(string(bytes.Repeat([]byte("-"), 70)))

	if len(rewrites) == 0 {
		fmt.Println("No DNS rewrites found")
	} else {
		for _, rewrite := range rewrites {
			fmt.Printf("%-40s -> %s\n", rewrite.Domain, rewrite.Answer)
		}
	}
}

func addRewrite(domain, answer string) {
	if domain == "" || answer == "" {
		fmt.Fprintf(os.Stderr, "Error: Both domain and answer are required for adding a DNS rewrite\n")
		os.Exit(1)
	}

	err := addDNSRewrite(domain, answer)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error adding DNS rewrite: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Successfully added DNS rewrite: %s -> %s\n", domain, answer)
}

func deleteRewrite(domain string) {
	if domain == "" {
		fmt.Fprintf(os.Stderr, "Error: Domain is required for deleting a DNS rewrite\n")
		os.Exit(1)
	}

	err := deleteDNSRewrite(domain)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error deleting DNS rewrite: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Successfully deleted DNS rewrite for domain: %s\n", domain)
}

func getRewrite(domain string) {
	if domain == "" {
		fmt.Fprintf(os.Stderr, "Error: Domain is required for getting a DNS rewrite\n")
		os.Exit(1)
	}

	rewrite, err := findRewrite(domain)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("DNS rewrite: %s -> %s\n", rewrite.Domain, rewrite.Answer)
}

func printUsage() {
	fmt.Println("AdGuard DNS CLI")
	fmt.Println("==============")
	fmt.Println("Usage:")
	fmt.Println("  adguard-dns-cli [command] [options]")
	fmt.Println("")
	fmt.Println("Commands:")
	fmt.Println("  list      List all DNS rewrites")
	fmt.Println("  add       Add a new DNS rewrite")
	fmt.Println("  delete    Delete a DNS rewrite")
	fmt.Println("  get       Get a specific DNS rewrite")
	fmt.Println("")
	fmt.Println("Options:")
	fmt.Println("  -host     AdGuard Home host (default: dns.srvr.site)")
	fmt.Println("  -username AdGuard Home username (default: Pi)")
	fmt.Println("  -password AdGuard Home password")
	fmt.Println("  -protocol Protocol (http or https, default: https)")
	fmt.Println("  -domain   Domain for add/delete/get commands")
	fmt.Println("  -answer   IP address or value for add command")
	fmt.Println("")
	fmt.Println("Examples:")
	fmt.Println("  adguard-dns-cli list")
	fmt.Println("  adguard-dns-cli add -domain example.com -answer 192.168.1.10")
	fmt.Println("  adguard-dns-cli delete -domain example.com")
	fmt.Println("  adguard-dns-cli get -domain example.com")
	fmt.Println("")
}

func main() {
	// Setup default values and flags
	var (
		domainFlag  string
		answerFlag  string
		showHelp    bool
	)

	// Define command-specific flag sets
	listCmd := flag.NewFlagSet("list", flag.ExitOnError)
	addCmd := flag.NewFlagSet("add", flag.ExitOnError)
	deleteCmd := flag.NewFlagSet("delete", flag.ExitOnError)
	getCmd := flag.NewFlagSet("get", flag.ExitOnError)
	
	// Common flags for all commands
	commonFlags := func(fs *flag.FlagSet) {
		fs.StringVar(&host, "host", "dns.srvr.site", "AdGuard Home host")
		fs.StringVar(&username, "username", "Pi", "AdGuard Home username")
		fs.StringVar(&password, "password", "9130458959", "AdGuard Home password")
		fs.StringVar(&protocol, "protocol", "https", "Protocol (http or https)")
		fs.BoolVar(&showHelp, "help", false, "Show help")
	}
	
	// Apply common flags to all command flagsets
	commonFlags(listCmd)
	commonFlags(addCmd)
	commonFlags(deleteCmd)
	commonFlags(getCmd)
	
	// Command-specific flags
	addCmd.StringVar(&domainFlag, "domain", "", "Domain for DNS rewrite")
	addCmd.StringVar(&answerFlag, "answer", "", "Answer (IP or value) for DNS rewrite")
	
	deleteCmd.StringVar(&domainFlag, "domain", "", "Domain for DNS rewrite")
	
	getCmd.StringVar(&domainFlag, "domain", "", "Domain for DNS rewrite")
	
	// Global flags for backwards compatibility
	flag.StringVar(&host, "host", "dns.srvr.site", "AdGuard Home host")
	flag.StringVar(&username, "username", "Pi", "AdGuard Home username")
	flag.StringVar(&password, "password", "9130458959", "AdGuard Home password")
	flag.StringVar(&protocol, "protocol", "https", "Protocol (http or https)")
	flag.StringVar(&domainFlag, "domain", "", "Domain for DNS rewrite")
	flag.StringVar(&answerFlag, "answer", "", "Answer (IP or value) for DNS rewrite")
	flag.BoolVar(&showHelp, "help", false, "Show help")

	// Check if we have enough arguments to determine the command
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(0)
	}

	// Get the command (first argument)
	command := os.Args[1]

	// Handle global help flag
	if command == "-help" || command == "--help" || command == "help" {
		printUsage()
		os.Exit(0)
	}

	// Process the command
	switch command {
	case "list":
		listCmd.Parse(os.Args[2:])
		if showHelp {
			fmt.Println("Usage: adguard-dns-cli list [options]")
			listCmd.PrintDefaults()
			os.Exit(0)
		}
		listRewrites()
		
	case "add":
		addCmd.Parse(os.Args[2:])
		if showHelp {
			fmt.Println("Usage: adguard-dns-cli add -domain <domain> -answer <answer> [options]")
			addCmd.PrintDefaults()
			os.Exit(0)
		}
		if domainFlag == "" || answerFlag == "" {
			fmt.Fprintf(os.Stderr, "Error: Both -domain and -answer flags are required\n")
			os.Exit(1)
		}
		addRewrite(domainFlag, answerFlag)
		
	case "delete":
		deleteCmd.Parse(os.Args[2:])
		if showHelp {
			fmt.Println("Usage: adguard-dns-cli delete -domain <domain> [options]")
			deleteCmd.PrintDefaults()
			os.Exit(0)
		}
		if domainFlag == "" {
			fmt.Fprintf(os.Stderr, "Error: -domain flag is required\n")
			os.Exit(1)
		}
		deleteRewrite(domainFlag)
		
	case "get":
		getCmd.Parse(os.Args[2:])
		if showHelp {
			fmt.Println("Usage: adguard-dns-cli get -domain <domain> [options]")
			getCmd.PrintDefaults()
			os.Exit(0)
		}
		if domainFlag == "" {
			fmt.Fprintf(os.Stderr, "Error: -domain flag is required\n")
			os.Exit(1)
		}
		getRewrite(domainFlag)
		
	default:
		fmt.Fprintf(os.Stderr, "Error: Unknown command '%s'\n\n", command)
		printUsage()
		os.Exit(1)
	}
}