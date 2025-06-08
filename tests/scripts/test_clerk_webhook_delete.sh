#!/bin/bash

# Test script for Clerk webhook endpoint with user deletion data

# Get the ID of the last user we created
LAST_USER_ID=$(go run list_users.go | grep -A 5 "User #3" | grep "Clerk User ID" | awk '{print $4}')

if [ -z "$LAST_USER_ID" ]; then
  echo "Failed to get the last user ID. Make sure you have created a user first."
  exit 1
fi

echo "Deleting user with ID: $LAST_USER_ID"

# Send a user.deleted webhook request
echo "Sending user.deleted webhook request..."
curl -X POST http://localhost:8080/api/v1/webhooks/clerk \
  -H "Content-Type: application/json" \
  -H "Svix-Id: msg_$(date +%s)" \
  -H "Svix-Timestamp: $(date +%s)" \
  -H "Svix-Signature: v1,test-signature" \
  -d '{
    "type": "user.deleted",
    "data": {
      "id": "'$LAST_USER_ID'",
      "deleted": true
    },
    "object": "event"
  }'

echo ""
echo "Test completed! Check the users table to see if the user was deleted (soft deleted)." 