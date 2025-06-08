# LaunchStack Authentication System

This document explains how authentication works in the LaunchStack platform, including the JWT token flow, Clerk integration, and authentication middleware.

## Overview

LaunchStack uses [Clerk](https://clerk.dev) as the authentication provider. Clerk handles user registration, login, session management, and JWT token generation. The frontend includes the JWT token in API requests, and the backend validates the token and associates the request with the authenticated user.

## Authentication Flow

1. **User Authentication**: 
   - User signs up or logs in through Clerk on the frontend
   - Clerk issues a JWT token after successful authentication
   - The frontend stores this token (managed by Clerk's SDK)

2. **API Requests**:
   - Frontend includes the JWT token in the `Authorization` header
   - Format: `Authorization: Bearer <clerk_jwt_token>`
   - Backend middleware validates the token and identifies the user
   - If valid, the request proceeds; if invalid, a 401 Unauthorized response is returned

3. **User Synchronization**:
   - Clerk webhooks keep user data in sync between Clerk and LaunchStack
   - Events like user creation, updates, and deletion are handled
   - The webhook endpoint is `/api/v1/auth/webhook`

## JWT Token Configuration

The JWT token contains claims that identify the user and provide additional information:

```json
{
  "sub": "<clerk_user_id>",
  "user_id": "<clerk_user_id>",
  "email": "<user_email>",
  "name": "<user_name>",
  "plan": "<subscription_plan>"
}
```

The token is validated using Clerk's JWKS (JSON Web Key Set) endpoint, which provides the public keys needed to verify the token's signature.

## Backend Middleware

The authentication middleware in `middleware/auth.go` performs the following steps:

1. Extracts the JWT token from the `Authorization` header
2. Validates the token using Clerk's JWKS endpoint
3. Extracts the user ID from the token claims
4. Retrieves the user from the database
5. Adds the user to the request context for route handlers to use

Public endpoints like `/api/v1/health` and webhook endpoints are excluded from authentication.

## Development Mode

For development, the system has a bypass mode activated when `DISABLE_PAYMENTS=true` and `APP_ENV=development`:

- Uses a fixed development user ID
- No token validation required
- Useful for local testing without Clerk integration

## Frontend Integration

The frontend uses Clerk's React SDK to handle authentication:

```jsx
import { useAuth } from '@clerk/clerk-react';

function Component() {
  const { getToken } = useAuth();
  
  async function fetchData() {
    const token = await getToken();
    
    // Include token in API request
    const response = await fetch('/api/v1/resource', {
      headers: {
        'Authorization': `Bearer ${token}`
      }
    });
    // Handle response
  }
}
```

## Environment Configuration

The following environment variables are used for authentication:

```
# Clerk Authentication
CLERK_SECRET_KEY=sk_test_your_clerk_secret_key
CLERK_WEBHOOK_SECRET=whsec_your_clerk_webhook_secret
NEXT_PUBLIC_CLERK_PUBLISHABLE_KEY=pk_test_your_clerk_publishable_key
CLERK_ISSUER=your-clerk-instance.clerk.accounts.dev
```

## Security Considerations

1. Always use HTTPS in production
2. Keep Clerk secrets secure and never expose them to the client
3. Set reasonable JWT token expiration times
4. Implement proper error handling for authentication failures
5. Regularly rotate secrets 