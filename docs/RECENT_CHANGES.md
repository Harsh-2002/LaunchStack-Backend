# Recent Changes

## TimescaleDB Migration
- Migrated from PostgreSQL to TimescaleDB for optimized time-series data storage
- Implemented hypertables for efficient storage of resource usage metrics
- Added continuous aggregates for fast querying of historical data
- Set up data retention and compression policies to manage storage growth
- Optimized database queries for time-series data access patterns

## Resource Usage Monitoring Improvements
- Updated CPU usage calculation to report percentage values (0-100%)
- Implemented minimum reporting threshold of 0.01% for non-zero CPU usage
- Removed disk usage tracking as requested
- Increased monitoring frequency to 10 seconds
- Implemented parallel stats collection for better performance

## Docker Integration Enhancements
- Switched from host filesystem bind mounts to Docker volumes for better data persistence
- Ensured all Docker operations use the API rather than CLI commands
- Improved container resource usage tracking
- Enhanced error handling for Docker operations

## Server Port Configuration
- Updated server to use environment variable for port configuration
- Default port: 8080
- Can be changed by setting the `PORT` environment variable (e.g., `PORT=9090 go run main.go`)

## Environment Variables
- Organized and restructured .env file
- Added sections for different types of variables
- Removed unused port range variables (`N8N_PORT_RANGE_START` and `N8N_PORT_RANGE_END`)
- Added documentation in ENV_SETUP.md for all required variables

## Container Management
- Updated IP allocation algorithm to properly increment IPs
- Added support for configurable subnets via `DOCKER_NETWORK_SUBNET`
- All containers now use the same port (`N8N_CONTAINER_PORT=5678`) but different IPs
- Updated Caddy configuration to point to container's actual IP address

## Error Handling
- Added robust error handling for missing environment variables
- Added default values for important configuration
- Added warnings when using default values

## Mock Payment Routes
- Added mock payment routes for development mode
- These routes simulate payment flows without actual payment processing
- Routes use the development user ID for authentication
- Available routes:
  - `GET /api/v1/payments` - Returns mock payment history
  - `GET /api/v1/payments/subscriptions` - Returns mock active subscription
  - `POST /api/v1/payments/checkout` - Simulates payment checkout
  - `POST /api/v1/payments/subscriptions/:id/cancel` - Simulates subscription cancellation

## Trailing Slash Issue
- Fixed the issue with trailing slashes in routes
- All routes now support both trailing and non-trailing slashes
- This prevents unnecessary redirects (307 responses)

## Documentation
- Updated DEVELOPMENT_NOTES.md with details about development mode
- Created ENV_SETUP.md with environment variable documentation
- Updated FRONTEND.md with information about mock payment routes
- Created this RECENT_CHANGES.md file

## Frontend Integration
- Updated frontend documentation to include backend API URL
- Added instructions for integrating with mock payment routes

## Performance
- Improved container IP allocation for better scalability
- Using consistent port (5678) for all containers improves resource utilization 