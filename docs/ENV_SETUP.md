# Environment Configuration

This document outlines the environment variables used in the LaunchStack-Backend system, with a focus on the TimescaleDB database configuration.

## TimescaleDB Configuration

The system now uses TimescaleDB for efficient time-series data storage. The following environment variables are used for database configuration:

```
# Main database connection string
DATABASE_URL=postgresql://postgres:password@localhost:5432/launchstack

# Individual connection parameters
DB_USER=postgres
DB_PASSWORD=password
DB_HOST=localhost
DB_PORT=5432
DB_NAME=launchstack
```

### Configuration Notes

- `DATABASE_URL` is the complete connection string used by many parts of the application
- The individual parameters (`DB_USER`, `DB_HOST`, etc.) are used by the updated database connection code
- When deploying to production, ensure you use strong passwords and consider using SSL

## Setting Up TimescaleDB

1. Deploy the TimescaleDB container using the provided docker-compose file
2. Run the setup script to initialize the database: `./setup_db.sh`
3. Verify the environment variables are correctly set in `.env`

## Complete Environment Variables

A complete list of environment variables used by the system:

### Database
```
DATABASE_URL=postgresql://postgres:password@localhost:5432/launchstack
DB_USER=postgres
DB_PASSWORD=password
DB_HOST=localhost
DB_PORT=5432
DB_NAME=launchstack
```

### Server
```
PORT=8080
APP_ENV=development
DOMAIN=launchstack.io
BACKEND_URL=http://localhost:8080
FRONTEND_URL=http://localhost:3000
JWT_SECRET=your_jwt_secret_here
```

### Docker
```
DOCKER_HOST=http://localhost:2375
DOCKER_NETWORK=n8n
DOCKER_NETWORK_SUBNET=10.1.2.0/24
N8N_CONTAINER_PORT=5678
N8N_BASE_IMAGE=n8nio/n8n:latest
```

### Monitoring
```
RESOURCE_MONITOR_INTERVAL=10s
LOG_LEVEL=debug
```

### Authentication and Payment
```
CLERK_SECRET_KEY=your_clerk_secret_key
CLERK_WEBHOOK_SECRET=your_clerk_webhook_secret
DISABLE_PAYMENTS=false
PAYPAL_API_KEY=your_paypal_api_key
PAYPAL_SECRET=your_paypal_secret
PAYPAL_MODE=sandbox
```

## Development vs. Production

Different environment settings are recommended based on the deployment environment:

### Development
```
APP_ENV=development
LOG_LEVEL=debug
```

### Production
```
APP_ENV=production
LOG_LEVEL=info
```

For production, consider setting up:
1. SSL/TLS for database connections
2. More restrictive network configurations
3. Proper monitoring and alerting
4. Database backup procedures

# Environment Variable Setup

## Required Environment Variables

Here's a list of all required environment variables for LaunchStack Backend:

### Core Configuration
- `PORT`: Server port (default: 8080)
- `APP_ENV`: Environment name (development/staging/production)
- `DOMAIN`: Domain for subdomains (e.g., srvr.site)
- `BACKEND_URL`: URL of the backend (e.g., http://localhost:8080)
- `FRONTEND_URL`: URL of the frontend (e.g., http://localhost:3000)
- `LOG_LEVEL`: Logging level (debug/info/warn/error)

### Database
- `DATABASE_URL`: PostgreSQL connection URL

### Authentication
- `JWT_SECRET`: Secret for JWT tokens
- `CLERK_SECRET_KEY`: Clerk API secret key
- `CLERK_WEBHOOK_SECRET`: Secret for Clerk webhooks
- `NEXT_PUBLIC_CLERK_PUBLISHABLE_KEY`: Clerk publishable key

### CORS
- `CORS_ORIGINS`: Comma-separated list of allowed origins

### AdGuard DNS Configuration
- `ADGUARD_HOST`: AdGuard Home DNS server hostname (e.g., dns.srvr.site)
- `ADGUARD_USERNAME`: AdGuard admin username
- `ADGUARD_PASSWORD`: AdGuard admin password
- `ADGUARD_PROTOCOL`: Protocol to use for AdGuard API (http/https)

### Docker Configuration
- `DOCKER_HOST`: Docker API endpoint (e.g., http://10.1.1.81:2375)
- `DOCKER_NETWORK`: Docker network name (e.g., n8n)
- `DOCKER_NETWORK_SUBNET`: Subnet for Docker network (e.g., 10.1.2.0/24)

### N8N Configuration
- `N8N_CONTAINER_PORT`: Port used inside N8N containers (default: 5678)
- `N8N_BASE_IMAGE`: N8N Docker image (e.g., n8nio/n8n:latest)
- `N8N_DATA_DIR`: Directory to store N8N data
- `N8N_WEBHOOK_SECRET`: Secret for N8N webhooks

### Payment Processing
- `DISABLE_PAYMENTS`: Set to "true" to bypass payment integration (development mode)
- `PAYPAL_API_KEY`: PayPal API key
- `PAYPAL_SECRET`: PayPal secret
- `PAYPAL_MODE`: PayPal mode (sandbox/live)

### Monitoring
- `RESOURCE_MONITOR_INTERVAL`: Interval for resource monitoring (e.g., 30s)

## Development Mode Setup

For local development, create a `.env` file with the following minimum configuration:

```
# Database connection
DATABASE_URL=postgresql://user:password@localhost:5432/launchstack?sslmode=disable

# Server configuration
PORT=8080
APP_ENV=development
DOMAIN=srvr.site
BACKEND_URL=http://localhost:8080
FRONTEND_URL=http://localhost:3000
JWT_SECRET=dev_secret_key
CORS_ORIGINS=http://localhost:3000

# AdGuard DNS Configuration
ADGUARD_HOST=dns.srvr.site
ADGUARD_USERNAME=your_adguard_username
ADGUARD_PASSWORD=your_adguard_password
ADGUARD_PROTOCOL=https

# Authentication (Clerk)
CLERK_SECRET_KEY=your_clerk_secret_key
CLERK_WEBHOOK_SECRET=your_clerk_webhook_secret
NEXT_PUBLIC_CLERK_PUBLISHABLE_KEY=your_clerk_publishable_key

# Development mode
DISABLE_PAYMENTS=true

# Docker configuration
DOCKER_HOST=http://localhost:2375
DOCKER_NETWORK=n8n
DOCKER_NETWORK_SUBNET=10.1.2.0/24

# N8N configuration
N8N_CONTAINER_PORT=5678
N8N_BASE_IMAGE=n8nio/n8n:latest
N8N_DATA_DIR=/path/to/n8n/data
N8N_WEBHOOK_SECRET=dev_webhook_secret

# Monitoring and logging
RESOURCE_MONITOR_INTERVAL=30s
LOG_LEVEL=debug
```

Replace placeholder values with your actual configuration.

## DNS Configuration

LaunchStack uses AdGuard DNS for routing traffic to containers. For each container:

1. Creates a subdomain mapping: `{subdomain}.{DOMAIN}` → `{subdomain}.docker`
2. Creates a Docker DNS mapping: `{subdomain}.docker` → Container IP

This enables easy access to instances via memorable subdomains without modifying Caddy or nginx configurations.

## IP Address Allocation

When creating new containers, the system allocates IP addresses from the subnet specified in `DOCKER_NETWORK_SUBNET`. The IP allocation starts from the 10th IP in the subnet (e.g., for subnet 10.1.2.0/24, the first allocated IP would be 10.1.2.10).

Each container gets a unique IP address but uses the same port specified in `N8N_CONTAINER_PORT` (default: 5678). 