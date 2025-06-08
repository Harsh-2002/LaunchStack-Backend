#!/bin/bash

# Test the PayPal webhook with a mock payment capture event
echo "Sending mock PAYMENT.CAPTURE.COMPLETED webhook event..."

curl -s -X POST "http://localhost:8080/api/v1/webhooks/paypal" \
  -H "Content-Type: application/json" \
  -d '{
    "id": "WH-123456789",
    "event_type": "PAYMENT.CAPTURE.COMPLETED",
    "resource": {
      "id": "5O190127TN364715T",
      "status": "COMPLETED",
      "amount": {
        "total": "5.00",
        "currency": "USD"
      },
      "custom_id": "user_2yDNOxLPr7zujKe3hz0Lqdu5TKH",
      "invoice_id": "INV-123456"
    },
    "create_time": "2025-06-08T00:00:00Z"
  }'

echo ""
echo "Sending mock BILLING.SUBSCRIPTION.CREATED webhook event..."

curl -s -X POST "http://localhost:8080/api/v1/webhooks/paypal" \
  -H "Content-Type: application/json" \
  -d '{
    "id": "WH-987654321",
    "event_type": "BILLING.SUBSCRIPTION.CREATED",
    "resource": {
      "id": "I-12345ABCDEF",
      "status": "ACTIVE",
      "plan_id": "P-12345ABCDEF",
      "custom_id": "user_2yDNOxLPr7zujKe3hz0Lqdu5TKH",
      "subscriber": {
        "email_address": "test@example.com",
        "name": {
          "given_name": "Test",
          "surname": "User"
        }
      },
      "billing_info": {
        "cycle_executions": [{
          "tenure_type": "REGULAR",
          "sequence": 1,
          "cycles_completed": 0,
          "cycles_remaining": 0,
          "total_cycles": 0
        }],
        "next_billing_time": "2025-07-08T00:00:00Z"
      }
    },
    "create_time": "2025-06-08T00:00:00Z"
  }' 