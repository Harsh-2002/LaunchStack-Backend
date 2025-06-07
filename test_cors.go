package main

import (
	"fmt"
	"os"
	"strings"
	
	"github.com/joho/godotenv"
)

func main() {
	// Load environment variables
	err := godotenv.Load()
	if err != nil {
		fmt.Println("Error loading .env file, using environment variables")
	}
	
	// Get CORS origins from environment
	corsOrigins := os.Getenv("CORS_ORIGINS")
	if corsOrigins == "" {
		fmt.Println("❌ CORS_ORIGINS environment variable is not set")
		os.Exit(1)
	}
	
	// Parse CORS origins
	origins := strings.Split(corsOrigins, ",")
	
	// Print CORS origins
	fmt.Println("Configured CORS origins:")
	for i, origin := range origins {
		fmt.Printf("%d. %s\n", i+1, origin)
	}
	
	// Check if dev.srvr.site is in the allowed origins
	devSite := "https://dev.srvr.site"
	found := false
	for _, origin := range origins {
		if origin == devSite {
			found = true
			break
		}
	}
	
	if found {
		fmt.Printf("✅ %s is in the allowed origins\n", devSite)
	} else {
		fmt.Printf("❌ %s is NOT in the allowed origins\n", devSite)
	}
	
	// Print testing instructions
	fmt.Println("\nTo test CORS manually, you can use the following curl command:")
	fmt.Println("curl -X OPTIONS -H \"Origin: https://dev.srvr.site\" -H \"Access-Control-Request-Method: GET\" -v http://10.1.1.79:8080/api/v1/health")
} 