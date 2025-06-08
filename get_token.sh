#!/bin/bash

# Replace with your Clerk secret key
CLERK_SECRET_KEY="sk_test_P73mo2hGSZN8co43wbBVcVObQWB6mLfPT0toQ8nBnA"

# Replace with your user ID
USER_ID="user_2yDNOxLPr7zujKe3hz0Lqdu5TKH"

# Get a token for the user
curl -s -X POST "https://api.clerk.dev/v1/sign_in_tokens" \
  -H "Authorization: Bearer ${CLERK_SECRET_KEY}" \
  -H "Content-Type: application/json" \
  -d "{\"user_id\": \"${USER_ID}\"}" | grep -o '"token":"[^"]*"' | cut -d'"' -f4

echo ""
echo "Use this token in Postman with the Authorization header:"
echo "Authorization: Bearer <token>" 