#!/bin/bash

# Get a JWT token from Clerk
JWT_TOKEN=$(./get_token.sh | head -1)

if [ -z "$JWT_TOKEN" ]; then
    echo "Error: Failed to get JWT token"
    exit 1
fi

echo "Using JWT token: $JWT_TOKEN"

# Test the mock checkout endpoint
echo "Creating mock checkout session for Pro plan..."

RESPONSE=$(curl -s -X POST "http://localhost:8080/api/v1/payments/checkout" \
  -H "Authorization: Bearer $JWT_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "plan": "pro",
    "success_url": "http://localhost:3000/dashboard?success=true",
    "cancel_url": "http://localhost:3000/pricing?canceled=true"
  }')

echo "Checkout Response:"
echo $RESPONSE

# Extract and display the checkout URL
CHECKOUT_URL=$(echo $RESPONSE | grep -o '"checkout_url":"[^"]*"' | cut -d'"' -f4)

if [ -n "$CHECKOUT_URL" ]; then
    echo ""
    echo "Mock checkout URL:"
    echo $CHECKOUT_URL
fi

# Now let's test getting the subscription
echo ""
echo "Getting subscription details..."

RESPONSE=$(curl -s -X GET "http://localhost:8080/api/v1/payments/subscriptions" \
  -H "Authorization: Bearer $JWT_TOKEN")

echo "Subscription Response:"
echo $RESPONSE 