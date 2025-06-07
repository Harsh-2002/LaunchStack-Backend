#!/bin/bash

# This script runs the DNS cleanup immediately with the force option
# It should be run with root privileges

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
CLEANUP_SCRIPT="${SCRIPT_DIR}/cleanup_stale_dns_records.py"

# Check if the cleanup script exists and is executable
if [ ! -f "$CLEANUP_SCRIPT" ]; then
  echo "Error: Cleanup script not found at $CLEANUP_SCRIPT"
  exit 1
fi

if [ ! -x "$CLEANUP_SCRIPT" ]; then
  echo "Making cleanup script executable"
  chmod +x "$CLEANUP_SCRIPT"
fi

echo "Running DNS cleanup with force option..."
$CLEANUP_SCRIPT --force

echo "DNS cleanup completed. Any stale DNS records should now be removed."
echo "If you still see stale DNS records, you can try running the script again after waiting a few minutes." 