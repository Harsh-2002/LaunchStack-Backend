package main

import (
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

func main() {
	// Create a default router
	router := gin.Default()

	// Simple CORS middleware
	router.Use(func(c *gin.Context) {
		origin := c.Request.Header.Get("Origin")
		if origin != "" {
			c.Writer.Header().Set("Access-Control-Allow-Origin", origin)
		} else {
			c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		}
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	})

	// Health check endpoint
	router.GET("/api/v1/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status": "ok",
			"version": "0.1.0",
		})
	})

	// Clerk webhook endpoint
	router.POST("/api/v1/webhooks/clerk", func(c *gin.Context) {
		// Log headers
		fmt.Println("Request headers:")
		for key, values := range c.Request.Header {
			fmt.Printf("  %s: %s\n", key, strings.Join(values, ", "))
		}

		// Read body
		var body map[string]interface{}
		if err := c.BindJSON(&body); err != nil {
			fmt.Printf("Error parsing JSON: %v\n", err)
			c.JSON(400, gin.H{"error": "Invalid JSON"})
			return
		}

		// Log the event
		fmt.Printf("Received event: %v\n", body)
		
		// Success response
		c.JSON(200, gin.H{"message": "Webhook processed successfully"})
	})

	// Log all registered routes
	for _, routeInfo := range router.Routes() {
		fmt.Printf("Registered route: %s %s\n", routeInfo.Method, routeInfo.Path)
	}

	// Start the server
	port := ":8080"
	fmt.Printf("Starting server on port %s...\n", port)
	if err := router.Run("0.0.0.0" + port); err != nil {
		log.Fatal(err)
	}
} 