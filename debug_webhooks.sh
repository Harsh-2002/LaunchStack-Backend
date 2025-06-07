#!/bin/bash

# Debug script for Clerk webhook functionality
# This script helps diagnose common issues with webhook processing

echo "=== LaunchStack Webhook Debug Script ==="
echo ""

# Check if server is running
echo "1. Checking if server is running..."
if ! curl -s http://localhost:8080/api/v1/health > /dev/null; then
  echo "ERROR: Server is not running at http://localhost:8080"
  echo "Please start the server first with 'go run main.go'"
  exit 1
fi
echo "✓ Server is running"
echo ""

# Check database connection
echo "2. Checking database connection..."
# We'll use a separate Go script to test the DB connection directly
cat > check_db.go << 'EOF'
package main

import (
	"fmt"
	"os"

	"github.com/joho/godotenv"
	"github.com/launchstack/backend/config"
	"github.com/launchstack/backend/db"
	"github.com/launchstack/backend/models"
)

func main() {
	// Load environment variables
	godotenv.Load()
	
	// Load config
	cfg, err := config.NewConfig()
	if err != nil {
		fmt.Printf("Failed to load config: %v\n", err)
		os.Exit(1)
	}
	
	// Connect to database
	fmt.Println("Connecting to database...")
	err = db.Initialize(cfg.Database.URL)
	if err != nil {
		fmt.Printf("Failed to connect to database: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("Successfully connected to database")
	
	// Check if users table exists
	var users []models.User
	result := db.DB.Limit(1).Find(&users)
	if result.Error != nil {
		fmt.Printf("Failed to query users table: %v\n", result.Error)
		os.Exit(1)
	}
	
	fmt.Printf("Database connection successful, found %d users\n", result.RowsAffected)
	
	// Check if any webhooks created users
	var clerkUsers []models.User
	result = db.DB.Where("clerk_user_id LIKE ?", "user_test_webhook_%").Find(&clerkUsers)
	if result.Error != nil {
		fmt.Printf("Failed to query for webhook-created users: %v\n", result.Error)
		os.Exit(1)
	}
	
	fmt.Printf("Found %d users created by webhook tests\n", result.RowsAffected)
	for i, user := range clerkUsers {
		fmt.Printf("  %d. User ID: %s, Clerk ID: %s, Email: %s\n", 
			i+1, user.ID, user.ClerkUserID, user.Email)
	}
}
EOF

echo "Compiling database check tool..."
if ! go build -o check_db check_db.go; then
  echo "ERROR: Failed to compile database check tool"
  echo "Make sure your Go environment is properly set up"
  exit 1
fi

echo "Running database check..."
./check_db
DB_STATUS=$?

if [ $DB_STATUS -ne 0 ]; then
  echo "✗ Database connection failed"
  echo "Check your database configuration in .env file"
else
  echo "✓ Database connection successful"
fi
echo ""

# Check webhook routes
echo "3. Checking webhook routes..."
ROUTES=$(curl -s http://localhost:8080/api/v1/health | grep -i webhook || echo "No webhook routes found")
echo "Registered routes: $ROUTES"
echo ""

# Generate test webhook event
TIMESTAMP=$(date +%s)
USER_ID="user_debug_webhook_${TIMESTAMP}"
EMAIL="debug${TIMESTAMP}@example.com"

echo "4. Sending test webhook event..."
echo "   User ID: $USER_ID"
echo "   Email: $EMAIL"
RESPONSE=$(curl -s -X POST http://localhost:8080/api/v1/webhooks/clerk \
  -H "Content-Type: application/json" \
  -H "Svix-Id: msg_${TIMESTAMP}" \
  -H "Svix-Timestamp: ${TIMESTAMP}" \
  -H "Svix-Signature: v1,debug-signature" \
  -d '{
    "type": "user.created",
    "data": {
      "id": "'$USER_ID'",
      "first_name": "Debug",
      "last_name": "User",
      "email_addresses": [
        {
          "id": "email_'$TIMESTAMP'",
          "email_address": "'$EMAIL'",
          "verification": {
            "status": "verified",
            "strategy": "email_code"
          }
        }
      ],
      "primary_email_address_id": "email_'$TIMESTAMP'",
      "created_at": '$TIMESTAMP',
      "updated_at": '$TIMESTAMP'
    },
    "object": "event"
  }')

echo "Server response: $RESPONSE"
echo ""

# Check if user was created
echo "5. Checking if test user was created..."
./check_db
FINAL_STATUS=$?

if [ $FINAL_STATUS -ne 0 ]; then
  echo "✗ Failed to verify user creation"
else
  echo "✓ Verification complete"
fi

echo ""
echo "=== Debug Summary ==="
echo "If you're still having issues with webhooks, check these common problems:"
echo "1. Make sure your .env file has the correct database configuration"
echo "2. Check server logs for detailed error messages"
echo "3. Verify that Clerk webhook routes are properly registered"
echo "4. Make sure webhook signature verification is disabled in development mode"
echo "5. Check if the database tables were properly created by the migrations"
echo ""
echo "You can run this script again after making changes to verify the fix."

# Cleanup
rm -f check_db check_db.go 