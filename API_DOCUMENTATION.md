# LaunchStack Backend API Documentation

## Overview
LaunchStack is a PaaS solution providing self-hosted n8n workflow automation instances with Docker-based architecture. The platform supports multiple subscription tiers: Free, Basic, Pro, and Enterprise, each with different resource allocations.

## Base URL
All API endpoints use the following base URL:
```
https://gw.srvr.site/api/v1
```

## Trailing Slashes
All endpoints support both trailing and non-trailing slashes. For example, both of these are valid:
```
GET /api/v1/instances
GET /api/v1/instances/
```

## Authentication
LaunchStack uses Clerk for authentication. All protected API endpoints require a valid JWT token.

### Authentication Headers
For all protected endpoints, include the following header:
```
Authorization: Bearer <clerk_jwt_token>
```

## API Endpoints

### Health Check

#### Check API Health
```
GET /api/v1/health
```

**Response (200 OK)**:
```json
{
  "status": "ok",
  "version": "0.1.0",
  "environment": "production",
  "go_version": "go1.21.0",
  "timestamp": "2025-06-07T19:26:11+05:30",
  "database": {
    "status": "ok"
  },
  "docker": {
    "status": "ok"
  },
  "api": {
    "endpoints": [
      "/api/v1/instances",
      "/api/v1/users/me",
      "/api/v1/auth/webhook",
      "/api/v1/health"
    ]
  }
}
```

### User Management

#### Get Current User
```
GET /api/v1/users/me
```

**Response (200 OK)**:
```json
{
  "id": "123e4567-e89b-12d3-a456-426614174000",
  "email": "user@example.com",
  "first_name": "John",
  "last_name": "Doe",
  "plan": "pro",
  "subscription_status": "active",
  "current_period_end": "2024-06-01T00:00:00Z",
  "instances": {
    "current": 2,
    "limit": 10
  },
  "resource_limits": {
    "max_instances": 10,
    "cpu_limit": 2.0,
    "memory_limit": 2048,
    "storage_limit": 20
  }
}
```

#### Update Current User
```
PUT /api/v1/users/me
```

**Request Body**:
```json
{
  "first_name": "John",
  "last_name": "Smith"
}
```

**Response (200 OK)**:
```json
{
  "id": "123e4567-e89b-12d3-a456-426614174000",
  "email": "user@example.com",
  "first_name": "John",
  "last_name": "Smith",
  "plan": "pro",
  "subscription_status": "active",
  "current_period_end": "2024-06-01T00:00:00Z",
  "instances_limit": 10
}
```

### Instance Management

#### List All Instances
```
GET /api/v1/instances
```

**Response (200 OK)**:
```json
[
  {
    "id": "123e4567-e89b-12d3-a456-426614174000",
    "name": "Production Workflows",
    "description": "Production automation workflows",
    "status": "running",
    "url": "prod-workflows-abc123.launchstack.io",
    "cpu_limit": 2.0,
    "memory_limit": 2048,
    "storage_limit": 20,
    "created_at": "2024-01-01T00:00:00Z",
    "updated_at": "2024-01-01T12:00:00Z"
  },
  {
    "id": "223e4567-e89b-12d3-a456-426614174000",
    "name": "Development Workflows",
    "description": "Development and testing workflows",
    "status": "stopped",
    "url": "dev-workflows-def456.launchstack.io",
    "cpu_limit": 2.0,
    "memory_limit": 2048,
    "storage_limit": 20,
    "created_at": "2024-01-02T00:00:00Z",
    "updated_at": "2024-01-02T12:00:00Z"
  }
]
```

#### Create Instance
```
POST /api/v1/instances
```

**Request Body**:
```json
{
  "name": "Marketing Workflows",
  "description": "Automation workflows for marketing team"
}
```

**Response (201 Created)**:
```json
{
  "id": "323e4567-e89b-12d3-a456-426614174000",
  "name": "Marketing Workflows",
  "description": "Automation workflows for marketing team",
  "status": "running",
  "url": "marketing-workflows-ghi789.launchstack.io",
  "cpu_limit": 2.0,
  "memory_limit": 2048,
  "storage_limit": 20,
  "created_at": "2024-04-20T00:00:00Z",
  "updated_at": "2024-04-20T00:00:00Z"
}
```

#### Get Instance Details
```
GET /api/v1/instances/:id
```

**Response (200 OK)**:
```json
{
  "id": "123e4567-e89b-12d3-a456-426614174000",
  "name": "Production Workflows",
  "description": "Production automation workflows",
  "status": "running",
  "url": "prod-workflows-abc123.launchstack.io",
  "cpu_limit": 2.0,
  "memory_limit": 2048,
  "storage_limit": 20,
  "created_at": "2024-01-01T00:00:00Z",
  "updated_at": "2024-01-01T12:00:00Z"
}
```

#### Delete Instance
```
DELETE /api/v1/instances/:id
```

**Response (200 OK)**:
```json
{
  "message": "Instance deleted successfully"
}
```

#### Start Instance
```
POST /api/v1/instances/:id/start
```

**Response (200 OK)**:
```json
{
  "message": "Instance started successfully"
}
```

#### Stop Instance
```
POST /api/v1/instances/:id/stop
```

**Response (200 OK)**:
```json
{
  "message": "Instance stopped successfully"
}
```

#### Restart Instance
```
POST /api/v1/instances/:id/restart
```

**Response (200 OK)**:
```json
{
  "message": "Instance restarted successfully"
}
```

### Resource Usage

#### Get Instance Resource Stats
```
GET /api/v1/instances/:id/stats
```

**Response (200 OK)**:
```json
{
  "id": "123e4567-e89b-12d3-a456-426614174000",
  "instance_id": "123e4567-e89b-12d3-a456-426614174000",
  "timestamp": "2025-06-07T19:26:11+05:30",
  "cpu_usage": 1.25,
  "cpu_formatted": "1.3%",
  "memory_usage": 268435456,
  "memory_limit": 2147483648,
  "memory_percentage": 12.5,
  "memory_formatted": "256.0 MB / 2.0 GB (12.5%)",
  "disk_usage": 52428800,
  "disk_formatted": "50.0 MB",
  "network_in": 1048576,
  "network_out": 524288,
  "network_formatted": "1.0 MB in / 512.0 KB out"
}
```

### Payment Management (When Enabled)

#### Get Payments History
```
GET /api/v1/payments
```

**Response (200 OK)**:
```json
[
  {
    "id": "123e4567-e89b-12d3-a456-426614174000",
    "amount": 29.00,
    "currency": "usd",
    "status": "succeeded",
    "description": "Subscription payment",
    "invoice_url": "https://invoice.paypal.com/i/xxx",
    "created_at": "2024-01-01T00:00:00Z"
  },
  {
    "id": "223e4567-e89b-12d3-a456-426614174000",
    "amount": 29.00,
    "currency": "usd",
    "status": "succeeded",
    "description": "Subscription payment",
    "invoice_url": "https://invoice.paypal.com/i/yyy",
    "created_at": "2024-02-01T00:00:00Z"
  }
]
```

#### Create Checkout Session
```
POST /api/v1/payments/checkout
```

**Request Body**:
```json
{
  "plan": "pro",
  "success_url": "https://launchstack.io/dashboard?success=true",
  "cancel_url": "https://launchstack.io/pricing?canceled=true"
}
```

**Response (200 OK)**:
```json
{
  "checkout_url": "https://www.paypal.com/checkoutnow?token=xxx",
  "order_id": "order_12345"
}
```

#### Get Subscriptions
```
GET /api/v1/payments/subscriptions
```

**Response (200 OK)**:
```json
{
  "id": "sub_xxxxx",
  "plan": "pro",
  "status": "active",
  "current_period_end": "2024-05-01T00:00:00Z",
  "cancel_at_period_end": false
}
```

#### Cancel Subscription
```
POST /api/v1/payments/subscriptions/:id/cancel
```

**Response (200 OK)**:
```json
{
  "status": "success",
  "message": "Subscription will be canceled at the end of the current billing period"
}
```

### Webhooks

#### Clerk Webhook (User Management)
```
POST /api/v1/auth/webhook
```

Handles Clerk webhook events for user lifecycle management.

**Events**:
- `user.created`
- `user.updated`
- `user.deleted`

#### PayPal Webhook (Payment Processing)
```
POST /api/v1/webhooks/paypal
```

Handles PayPal webhook events for payment processing.

**Events**:
- `PAYMENT.CAPTURE.COMPLETED`
- `BILLING.SUBSCRIPTION.CREATED`
- `BILLING.SUBSCRIPTION.UPDATED`
- `BILLING.SUBSCRIPTION.CANCELLED`

## CORS Support

The API implements a permissive CORS policy that:
- Allows requests from any origin
- Supports credentials
- Permits common HTTP methods (GET, POST, PUT, DELETE, OPTIONS)
- Allows standard headers

## Error Responses

All API endpoints return appropriate HTTP status codes:

- `200 OK`: Request successful
- `201 Created`: Resource created successfully
- `400 Bad Request`: Invalid request parameters
- `401 Unauthorized`: Missing or invalid authentication
- `403 Forbidden`: Valid authentication but insufficient permissions
- `404 Not Found`: Resource not found
- `409 Conflict`: Resource conflict (e.g., name already exists)
- `429 Too Many Requests`: Rate limit exceeded
- `500 Internal Server Error`: Server error

Error response body format:

```json
{
  "error": "Error message describing the issue"
}
```

## Development Mode

When running in development mode with `DISABLE_PAYMENTS=true`, authentication and payment integration are bypassed. The API will:

1. Not verify Clerk tokens
2. Provide a development user with Pro plan features
3. Skip PayPal integration for payments and subscriptions
4. Allow full access to all endpoints

This mode is intended for development and testing only. 