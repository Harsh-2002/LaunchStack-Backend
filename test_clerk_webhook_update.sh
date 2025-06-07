#!/bin/bash

# Test script for Clerk webhook endpoint with user update data

# Get the ID of the last user we created
LAST_USER_ID=$(go run list_users.go | grep -A 5 "User #3" | grep "Clerk User ID" | awk '{print $4}')

if [ -z "$LAST_USER_ID" ]; then
  echo "Failed to get the last user ID. Make sure you have created a user first."
  exit 1
fi

echo "Updating user with ID: $LAST_USER_ID"

# Send a user.updated webhook request
echo "Sending user.updated webhook request..."
curl -X POST http://localhost:8080/api/v1/webhooks/clerk \
  -H "Content-Type: application/json" \
  -H "Svix-Id: msg_$(date +%s)" \
  -H "Svix-Timestamp: $(date +%s)" \
  -H "Svix-Signature: v1,test-signature" \
  -d '{
    "type": "user.updated",
    "data": {
      "id": "'$LAST_USER_ID'",
      "first_name": "Updated",
      "last_name": "UserName",
      "email_addresses": [
        {
          "id": "email_'$(date +%s)'",
          "email_address": "updated_'$(date +%s)'@example.com",
          "verification": {
            "status": "verified",
            "strategy": "email_code"
          }
        }
      ],
      "primary_email_address_id": "email_'$(date +%s)'",
      "created_at": '$(date +%s)',
      "updated_at": '$(date +%s)',
      "profile_image_url": "https://example.com/updated_avatar.jpg"
    },
    "object": "event"
  }'

echo ""
echo "Test completed! Check the users table to see if the user was updated." 