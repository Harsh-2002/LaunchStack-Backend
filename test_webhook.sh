#!/bin/bash

# Test script for Clerk webhook endpoint

# Start the server with webhook secret disabled
export CLERK_WEBHOOK_SECRET=""

echo "Starting test webhook server..."
go run simple_webhook.go &
SERVER_PID=$!

# Give the server time to start
sleep 2

# Send a test webhook request
echo "Sending test webhook request..."
curl -X POST http://localhost:8080/api/v1/webhooks/clerk \
  -H "Content-Type: application/json" \
  -H "Svix-Id: msg_test123" \
  -H "Svix-Timestamp: $(date +%s)" \
  -H "Svix-Signature: v1,test-signature" \
  -d '{"type": "user.created", "data": {"id": "test-user-123", "email": "test@example.com"}}'

echo ""
echo "Test completed!"

# Cleanup
kill $SERVER_PID 