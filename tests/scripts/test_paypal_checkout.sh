#!/bin/bash

# First, let's get a JWT token from Clerk (use the token from get_token.sh)
JWT_TOKEN=$(./get_token.sh | head -1)

if [ -z "$JWT_TOKEN" ]; then
    echo "Error: Failed to get JWT token"
    exit 1
fi

echo "Using JWT token: $JWT_TOKEN"

# Create a checkout session for the Pro plan
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

# Extract the checkout URL
CHECKOUT_URL=$(echo $RESPONSE | grep -o '"checkout_url":"[^"]*"' | cut -d'"' -f4)

if [ -n "$CHECKOUT_URL" ]; then
    echo ""
    echo "Open this URL in your browser to complete the PayPal payment:"
    echo $CHECKOUT_URL
    echo ""
    echo "After payment, PayPal will redirect you to the success URL"
fi 