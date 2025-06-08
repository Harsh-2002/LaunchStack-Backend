#!/bin/bash

# Use the manually generated token
TOKEN=$(go run create_test_token.go | head -1)

echo "Using token: $TOKEN"

# Create a mock checkout session
echo "Creating checkout session for Pro plan..."

RESPONSE=$(curl -s -X POST "http://localhost:8080/api/v1/payments/checkout" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "plan": "pro",
    "success_url": "http://localhost:3000/dashboard?success=true",
    "cancel_url": "http://localhost:3000/pricing?canceled=true"
  }')

echo "Checkout Response:"
echo $RESPONSE

# Extract the checkout URL
CHECKOUT_URL=$(echo $RESPONSE | grep -o '"checkout_url":"[^"]*"' | cut -d'"' -f4)

if [ -n "$CHECKOUT_URL" ]; then
    echo ""
    echo "Mock checkout URL:"
    echo $CHECKOUT_URL
    echo ""
    echo "You can open this URL in a browser to complete the test checkout process."
fi 