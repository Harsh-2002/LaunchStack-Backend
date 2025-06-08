# DNS Management Solution

This document outlines the solution for managing DNS records in the LaunchStack-Backend system.

## Problem Statement

The system uses AdGuard DNS to manage DNS records for container subdomains. However, the AdGuard DNS API has proven to be unreliable in two specific ways:

1. The API sometimes returns a success response for delete operations without actually deleting the DNS record.
2. Stale DNS records accumulate over time, which can lead to confusion and potential security issues.

## Solution

The solution implements a multi-layered approach to ensure DNS records are properly managed:

### 1. DNS CLI Tool

A custom DNS CLI tool (`dns-cli.go`) has been implemented to provide more reliable DNS record management:

- Provides a simple command-line interface for managing DNS records
- Uses both domain and answer fields for more reliable record management
- Includes verification steps to ensure operations are successful
- Can be used both programmatically and manually

### 2. Enhanced DNS Record Deletion in Go Code

The DNS deletion logic in `container/dns.go` has been improved to:

- Use the new DNS CLI tool as the primary method for DNS operations
- Fall back to API methods if the CLI tool is not available
- Retrieve the existing DNS record before attempting deletion to include both domain and answer fields
- Add verification steps to confirm that records are actually deleted
- Implement retry logic with alternative deletion methods
- Add proper delays between operations

### 3. Automated Cleanup Cron Job

A Python-based cleanup script (`cleanup_stale_dns_records.py`) has been implemented to:

- Run regularly via cron (every 6 hours)
- Compare active containers with DNS records
- Remove any stale DNS records that don't correspond to running containers
- Log all cleanup operations for auditing

### 4. Manual Cleanup Options

For immediate DNS cleanup, multiple convenience scripts have been added:

- `run_dns_cleanup_now.sh`: Runs the cleanup script with the `--force` option
- `dns-cli`: Direct interaction with AdGuard DNS records
- `setup_dns_cli.sh`: Builds and installs the DNS CLI tool

### 5. Database Schema Updates

Database schema issues have been addressed:

- Added proper migration for the `resource_usages` table to ensure all required columns exist
- Fixed the missing `memory_limit` column that was causing errors when saving resource usage data
- Ensured migrations run on application startup regardless of migration timing

## Usage

### Using the DNS CLI Tool

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

### Setting up the Cron Job

```bash
sudo ./setup_dns_cleanup_cron.sh
```

This will:
- Configure the cleanup script to run every 6 hours
- Create the necessary log file
- Run an initial dry run to verify the script works

### Running Manual Cleanup

```bash
sudo ./run_dns_cleanup_now.sh
```

This will:
- Run the cleanup script immediately with the force option
- Remove any stale DNS records

## Technical Implementation

The solution uses multiple approaches to maximize reliability:

1. **DNS CLI Tool (Primary Method):**
   - Direct interaction with AdGuard DNS API
   - Includes comprehensive error handling
   - Provides verification of operations
   - Can be used for manual and automated management

2. **During Container Deletion:**
   - First try using the DNS CLI tool
   - Fall back to API methods if needed
   - Always include both domain and answer fields
   - Verify successful deletion
   - Log warnings when deletion appears to fail

3. **During Scheduled Cleanup:**
   - Get list of all active containers from Docker
   - Get list of all DNS records from AdGuard
   - Identify container DNS records (.docker domains)
   - Remove any container DNS records that don't have a corresponding active container

4. **Database Schema Management:**
   - Explicit column checks and creation during migration
   - Verification of all required columns for resource usage tracking

## Future Improvements

1. Consider replacing AdGuard DNS with a more reliable DNS management solution
2. Implement monitoring to alert on DNS record deletion failures
3. Add additional verification steps and cleanup operations on app restart 