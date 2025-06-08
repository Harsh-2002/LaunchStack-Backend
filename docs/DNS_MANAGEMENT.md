# DNS Management

The LaunchStack system uses AdGuard DNS for dynamic routing to Docker containers. Each container gets a DNS record mapping a subdomain (e.g., `container-name.docker`) to the container's IP address.

## Improved DNS Management

We've implemented an improved approach for DNS record deletion that addresses the reliability issues with the AdGuard DNS API. The key improvements include:

1. Finding the existing DNS record with its complete details before deletion
2. Using both domain and answer values in the deletion request
3. Adding delays to allow the API time to process the deletion
4. Implementing verification and retry mechanisms
5. Gracefully handling deletion failures without blocking container management

## DNS Management Tools

The system now uses both Go implementations and supplementary Python tools for DNS management:

### Go Implementations

1. `container/dns.go` - The primary DNS management implementation with the improved approach
2. `dns.go` - A standalone test utility for DNS operations
3. `test_dns_delete.go` - A testing utility for DNS deletion operations

### Cleanup Scripts

1. `cleanup_stale_dns_records.py` - A Python script for identifying and removing stale DNS records
   ```
   ./cleanup_stale_dns_records.py --dry-run       # Show what would be cleaned up without making changes
   ./cleanup_stale_dns_records.py --force         # Forcefully clean up stale DNS records
   ./cleanup_stale_dns_records.py --test          # Run in test mode with sample data
   ```

2. `setup_dns_cleanup_cron.sh` - A script to set up daily DNS cleanup
   ```
   sudo ./setup_dns_cleanup_cron.sh               # Set up daily cron job for DNS cleanup
   ```

## Known Issues

The AdGuard DNS API has a fundamental issue with DNS record deletion:

1. The API endpoint for deleting DNS records (`control/rewrite/delete`) returns success responses but doesn't always delete the records.
2. This can lead to accumulation of stale DNS records over time.
3. Our improved implementation mitigates this issue with multiple deletion attempts and verification steps.

## Current Solution

Our solution addresses the AdGuard DNS API issues through several approaches:

1. **Improved DNS Management Code**:
   - Gets existing record details before deletion
   - Includes IP address in deletion requests
   - Verifies deletions and implements retry logic
   - Handles failures gracefully

2. **Robust Container Management**:
   - Continues with container deletion even if DNS record deletion fails
   - Logs comprehensive information about DNS operations
   - Properly handles edge cases like missing DNS records

3. **Periodic Cleanup**:
   - Automated script to identify and clean up stale DNS records
   - Compares DNS records with active containers
   - Scheduled to run daily via cron
   - Includes dry-run and force options

For more detailed information about this issue and solution, see `DNS_MANAGEMENT_SOLUTION.md`.

## Environment Variables

The DNS tools use the following environment variables:

- `ADGUARD_PROTOCOL`: Protocol to use for AdGuard API (default: `https`)
- `ADGUARD_HOST`: AdGuard DNS server hostname (default: `dns.srvr.site`)
- `ADGUARD_USERNAME`: AdGuard username (default: `Pi`)
- `ADGUARD_PASSWORD`: AdGuard password (default: `9130458959`)

These can be set in your `.env` file or directly in the environment. 