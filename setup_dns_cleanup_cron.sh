#!/bin/bash

# This script sets up a cron job to clean up stale DNS records
# It should be run with root privileges

set -e

# Default script location - adjust if needed
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
CLEANUP_SCRIPT="${SCRIPT_DIR}/cleanup_stale_dns_records.py"
LOG_FILE="/var/log/dns_cleanup.log"

# Check if the cleanup script exists and is executable
if [ ! -f "$CLEANUP_SCRIPT" ]; then
  echo "Error: Cleanup script not found at $CLEANUP_SCRIPT"
  exit 1
fi

if [ ! -x "$CLEANUP_SCRIPT" ]; then
  echo "Making cleanup script executable"
  chmod +x "$CLEANUP_SCRIPT"
fi

# Create the cron job entry - runs every 6 hours
CRON_ENTRY="0 */6 * * * root $CLEANUP_SCRIPT --force >> $LOG_FILE 2>&1"

# Check if the entry already exists in crontab
if grep -q "$CLEANUP_SCRIPT" /etc/crontab; then
  echo "Cron job already exists, updating it"
  # Remove existing entry
  sed -i "\|$CLEANUP_SCRIPT|d" /etc/crontab
fi

# Add new entry
echo "$CRON_ENTRY" >> /etc/crontab
echo "Cron job added to /etc/crontab"

# Create the log file if it doesn't exist
if [ ! -f "$LOG_FILE" ]; then
  touch "$LOG_FILE"
  chmod 644 "$LOG_FILE"
  echo "Log file created at $LOG_FILE"
fi

# Run the script once immediately to verify it works
echo "Running initial cleanup (dry run) to verify the script works"
$CLEANUP_SCRIPT --dry-run

echo "Setup complete. DNS records will be cleaned up every 6 hours."
echo "Logs will be written to $LOG_FILE"
echo ""
echo "To manually clean up DNS records, run:"
echo "  $CLEANUP_SCRIPT --force"
echo ""
echo "To check what would be cleaned up without making changes, run:"
echo "  $CLEANUP_SCRIPT --dry-run" 