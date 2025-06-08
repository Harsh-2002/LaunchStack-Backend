#!/bin/bash

# Test script for Clerk webhook endpoint with real user data

# Make sure the server is running with webhook signature verification disabled in development mode

# Generate unique test IDs
TIMESTAMP=$(date +%s)
USER_ID="user_test_webhook_${TIMESTAMP}"
EMAIL_ID="email_${TIMESTAMP}"
EMAIL="testuser${TIMESTAMP}@example.com"

echo "=== Clerk Webhook Test Script ==="
echo "Testing with:"
echo "  User ID: $USER_ID"
echo "  Email: $EMAIL"
echo "  Email ID: $EMAIL_ID"
echo ""

# First check if the server is running
echo "Checking if server is running..."
if ! curl -s http://localhost:8080/api/v1/health > /dev/null; then
  echo "ERROR: Server is not running at http://localhost:8080"
  echo "Please start the server first with 'go run main.go'"
  exit 1
fi
echo "Server is running."
echo ""

# Send a user.created webhook request
echo "Sending user.created webhook request..."
RESPONSE=$(curl -s -X POST http://localhost:8080/api/v1/webhooks/clerk \
  -H "Content-Type: application/json" \
  -H "Svix-Id: msg_${TIMESTAMP}" \
  -H "Svix-Timestamp: ${TIMESTAMP}" \
  -H "Svix-Signature: v1,test-signature" \
  -d '{
    "type": "user.created",
    "data": {
      "id": "'$USER_ID'",
      "first_name": "Test",
      "last_name": "User",
      "email_addresses": [
        {
          "id": "'$EMAIL_ID'",
          "email_address": "'$EMAIL'",
          "verification": {
            "status": "verified",
            "strategy": "email_code"
          }
        }
      ],
      "primary_email_address_id": "'$EMAIL_ID'",
      "created_at": '$TIMESTAMP',
      "updated_at": '$TIMESTAMP',
      "profile_image_url": "https://example.com/avatar.jpg"
    },
    "object": "event"
  }')

echo "Server response: $RESPONSE"
echo ""

# Verify if the user was created (if you have access to psql)
if command -v psql &> /dev/null; then
  echo "Checking database for created user..."
  echo "If this fails, make sure your database credentials are correct."
  echo "The following command should show the user if it was created:"
  echo "  psql <YOUR_CONNECTION_STRING> -c \"SELECT id, clerk_user_id, email, first_name, last_name FROM users WHERE clerk_user_id = '$USER_ID';\""
  echo ""
fi

echo "Test completed! To verify database changes, check if a user with Clerk ID '$USER_ID' was created."
echo "Common issues if the test fails:"
echo "1. Database connection not initialized in the application"
echo "2. Tables not created - make sure migrations were run"
echo "3. Webhook route not properly registered"
echo "4. Error in webhook handler logic"
echo ""
echo "Check server logs for more details on any errors." 