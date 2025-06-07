# Development Mode Notes

## Fixed Development User ID

For development mode, we use a fixed user ID to ensure consistent behavior:

```
f2814e7b-75a0-44d4-b345-e5ef5a84aab3
```

This user is automatically created in the database when the server starts in development mode with `DISABLE_PAYMENTS=true`. 

## Database Schema Updates

The following database schema changes were made:
- Added `ip_address` column to the `instances` table (VARCHAR(50))

## Environment Variables

For development mode, make sure to set:
```
DISABLE_PAYMENTS=true
```

This enables several development features:
1. Authentication bypass with fixed user ID
2. Skipping payment-related routes and checks
3. Mock payment routes with dummy data

### Mock Payment Routes

In development mode, the following mock payment routes are available:
- `GET /api/v1/payments` - Returns mock payment history
- `GET /api/v1/payments/subscriptions` - Returns a mock active subscription
- `POST /api/v1/payments/checkout` - Simulates payment checkout
- `POST /api/v1/payments/subscriptions/:id/cancel` - Simulates subscription cancellation

These mock routes help you develop and test payment-related features without actual payment processing.

### IP and Port Configuration

The container manager now uses these environment variables:
- `DOCKER_NETWORK_SUBNET`: The subnet for container IP addresses (default: `10.1.2.0/24`)
- `N8N_CONTAINER_PORT`: The port that n8n runs on inside containers (default: `5678`)
- `DOMAIN`: The domain used for subdomains (default: `srvr.site`)

Each container gets a unique IP address from the subnet, but uses the same port (5678).
The Caddy reverse proxy configuration points to each container's unique IP address.

## Common Issues

### Foreign Key Constraint Errors

If you see an error like:
```
ERROR: insert or update on table "instances" violates foreign key constraint "fk_users_instances"
```

This means the user ID being used doesn't exist in the database. Make sure:
1. `DISABLE_PAYMENTS=true` is set in your .env file
2. The server is running in development mode
3. The development user was created successfully (check logs)

### Address Already in Use

If you see:
```
listen tcp :8080: bind: address already in use
```

Kill all running Go processes and try again:
```
pkill -f "go run"
```

Or use a different port:
```
PORT=9090 go run main.go
```

## DNS Management

### Improved DNS Record Management

We use AdGuard DNS for dynamic routing to Docker containers. Each container gets a DNS record mapping a subdomain (e.g., `{subdomain}.docker`) to the container's IP address.

We've implemented an improved approach for DNS record deletion that addresses reliability issues with the AdGuard DNS API. The key improvements include:

1. Finding the existing DNS record with its complete details before deletion
2. Using both domain and answer values in the deletion request
3. Adding delays to allow the API time to process the deletion
4. Implementing verification and retry mechanisms
5. Gracefully handling deletion failures without blocking container management

**Note**: The AdGuard DNS API may still occasionally fail to delete records despite these improvements. Our solution ensures that even if some records fail to delete, they will be cleaned up later during the periodic maintenance.

### DNS Cleanup

To address potential stale DNS records, we maintain a two-part solution:

1. The Docker manager attempts to delete DNS records when containers are deleted, with comprehensive error handling
2. A periodic cleanup script runs daily to identify and remove any stale DNS records

#### Testing DNS Operations

You can use our test utility to test DNS operations:

```
# Build the test utility
go build -o test_improved_dns test_improved_dns_delete.go

# List all DNS records
./test_improved_dns list

# Add a test DNS record
./test_improved_dns add test-domain.docker 10.1.2.3

# Delete a DNS record
./test_improved_dns delete test-domain.docker

# Run a full test cycle (add, verify, delete, verify)
./test_improved_dns test
```

#### Running the DNS Cleanup Script Manually

To check for stale DNS records without deleting them:
```
./cleanup_stale_dns_records.py --dry-run
```

To delete stale DNS records:
```
./cleanup_stale_dns_records.py --force
```

#### Setting Up Automatic DNS Cleanup

To set up a cron job for automatic DNS cleanup, run:
```
sudo ./setup_dns_cleanup_cron.sh
```

This will set up a cron job to run the cleanup script daily at 2:30 AM.

#### Manual DNS Management

If needed, you can still manually manage DNS records through the AdGuard web interface at `https://dns.srvr.site`. 

## Troubleshooting Common Issues

### Database Migrations

If you see errors related to missing columns in the database, you may need to run migrations manually:

```bash
go run main.go --force-migrations
```

This will force the system to run all migrations regardless of when they were last executed.

### Resource Usage Tracking

If you encounter database errors when collecting container stats like:

```
ERROR: column "memory_limit" of relation "resource_usages" does not exist
```

This indicates a schema mismatch between the model and the database. The system should automatically fix this issue on startup with the latest updates, but if it persists:

1. Connect to the database and check the schema:
   ```sql
   \d resource_usages
   ```

2. Manually add missing columns if needed:
   ```sql
   ALTER TABLE resource_usages ADD COLUMN memory_limit BIGINT;
   ALTER TABLE resource_usages ADD COLUMN memory_percentage FLOAT;
   ```

3. Restart the application to verify the fix.

### DNS Record Management

The system uses AdGuard DNS for managing container subdomains. There are known issues with the AdGuard DNS API occasionally failing to delete records despite returning success responses.

#### DNS CLI Tool

We now have a dedicated DNS CLI tool for more reliable DNS management:

```bash
# Build and set up the DNS CLI tool
./setup_dns_cli.sh

# List all DNS records
./dns-cli list

# Add a DNS record
./dns-cli add -domain example.docker -answer 10.1.2.3

# Delete a DNS record
./dns-cli delete -domain example.docker

# Get details of a specific record
./dns-cli get -domain example.docker
```

The container management code now uses this CLI tool as the primary method for DNS operations, with API fallback if the CLI tool is not available.

#### Handling Stale DNS Records

If you notice stale DNS records that weren't properly deleted:

1. Use the DNS CLI tool to delete the record directly:
   ```bash
   ./dns-cli delete -domain stale-record.docker
   ```

2. Run the manual cleanup script:
   ```bash
   sudo ./run_dns_cleanup_now.sh
   ```

3. Check the cleanup log for details:
   ```bash
   tail -f /var/log/dns_cleanup.log
   ```

4. In the development environment, you can run the Python cleanup script with additional options:
   ```bash
   ./cleanup_stale_dns_records.py --dry-run  # Shows what would be deleted without making changes
   ./cleanup_stale_dns_records.py --force    # Forcefully recreates all DNS records
   ```

The system now has a scheduled task that runs every 6 hours to clean up stale DNS records. This helps prevent the accumulation of orphaned records over time.

If needed, you can still manually manage DNS records through the AdGuard web interface at `https://dns.srvr.site`. 

## Development Environment Setup

### Docker and Container Management

The system requires Docker to be installed and running. In development mode, you need to:

1. Ensure Docker is running with the correct permissions
2. Configure the appropriate DNS and network settings in `docker-compose.yml`
3. Make sure the data directories for container volumes exist with the right permissions

### Environment Variables

Key environment variables for development:

```bash
export DISABLE_PAYMENTS=true  # Enable development mode with mock payments
export CLERK_API_KEY=test123  # Mock Clerk API key for development
export DB_URL="postgresql://postgres:postgres@localhost:5432/launchstack?sslmode=disable"
```

See `ENV_SETUP.md` for a complete list of environment variables.

### Testing Routes and Webhooks

Use the provided test scripts for verifying different components:

- `test_clerk_webhook.sh`: Test Clerk auth webhook handling
- `test_webhook.sh`: General webhook testing
- `debug_webhooks.sh`: Advanced webhook debugging

## API Documentation

See `API_DOCUMENTATION.md` for the complete API reference. 