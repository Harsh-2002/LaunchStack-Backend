# LaunchStack Backend

Backend API service for LaunchStack, a PaaS solution for self-hosted n8n instances.

## Architecture

LaunchStack backend is built with Golang and the Gin framework, providing a robust API for managing n8n instances using Docker containers. The system handles user authentication via Clerk, subscription management with Stripe, and container orchestration through the Docker API.

## Prerequisites

- Go 1.21 or later
- Docker Engine
- PostgreSQL (or use the provided Neon cloud database)
- Clerk account for authentication
- Stripe account for payments (optional for development)

## Environment Setup

Three environment files are provided:

- `.env.example` - Template with all required variables
- `.env` - Development configuration with placeholder values
- `.env.production` - Production configuration with environment variable references

Configure your environment by copying and modifying the appropriate file:

```bash
# For development
cp .env.example .env
# Edit .env with your specific values
```

### Required Environment Variables

- `DATABASE_URL` - PostgreSQL connection string
- `CLERK_SECRET_KEY` - Secret key from Clerk dashboard
- `JWT_SECRET` - Secret for JWT token generation
- `DOCKER_HOST` - Docker socket path (default: unix:///var/run/docker.sock)
- `N8N_WEBHOOK_SECRET` - Secret for n8n webhook signatures

## Installation

### Local Development

1. Clone the repository:
   ```bash
   git clone https://github.com/yourusername/LaunchStack-Backend.git
   cd LaunchStack-Backend
   ```

2. Install dependencies:
   ```bash
   go mod download
   ```

3. Run the application:
   ```bash
   go run main.go
   ```

### Docker Deployment

1. Build and run with Docker Compose:
   ```bash
   docker-compose up -d
   ```

2. For production deployment:
   ```bash
   docker-compose -f docker-compose.yml --env-file .env.production up -d
   ```

## API Endpoints

Detailed API documentation is available in the [API_DOCUMENTATION.md](API_DOCUMENTATION.md) file.

Key endpoints include:

- **Health Check**: `/api/v1/health` - System health status with database and Docker connectivity
- **Authentication**: `/api/v1/webhooks/clerk` - Webhook for Clerk authentication events
- **User Profile**: `/api/v1/users/me` - Get current user profile
- **Instances**: `/api/v1/instances` - CRUD operations for n8n instances
- **Monitoring**: `/api/v1/usage/:instanceId` - Get instance resource usage metrics
- **Subscriptions**: `/api/v1/payments/checkout` - Create a subscription
- **N8N Webhooks**: `/api/v1/webhooks/n8n` - Receive events from n8n instances

## Database Schema

The system uses the following database models:

- **User**: User accounts linked to Clerk authentication
- **Instance**: n8n container instances
- **ResourceUsage**: Metrics and monitoring data
- **Payment**: Subscription and payment records

## Docker Container Management

LaunchStack uses the Docker API to:

1. Create containerized n8n instances
2. Manage container lifecycle (start, stop, delete)
3. Monitor resource usage
4. Handle subdomain routing

## Development

### Adding a New API Endpoint

1. Create a new handler in the appropriate package
2. Register the route in `routes/routes.go`
3. Add any required middleware (authentication, rate limiting)

### Testing

Run the test suite:

```bash
go test ./...
```

### Verifying System Health

You can check if the system is running properly by using the health endpoint:

```bash
curl http://localhost:8080/api/v1/health
```

This will return a JSON response with the status of various components.

## License

This project is licensed under the MIT License - see the LICENSE file for details.

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request. 