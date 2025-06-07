#!/bin/bash

# This script sets up the DNS CLI tool

set -e

# Build the DNS CLI tool
echo "Building DNS CLI tool..."
go build -o dns-cli dns-cli.go

# Make it executable
chmod +x dns-cli

# Move it to a location in PATH if requested
if [ "$1" == "--install" ]; then
  echo "Installing DNS CLI tool to /usr/local/bin/"
  sudo cp dns-cli /usr/local/bin/
  echo "DNS CLI tool installed to /usr/local/bin/dns-cli"
else
  echo "DNS CLI tool built at ./dns-cli"
  echo "To install globally, run: $0 --install"
fi

echo "Testing DNS CLI tool..."
./dns-cli list

echo ""
echo "DNS CLI tool is ready to use. You can now use it to manage DNS records:"
echo "  ./dns-cli list                            # List all DNS records"
echo "  ./dns-cli add -domain example.docker -answer 10.1.2.3   # Add a record"
echo "  ./dns-cli delete -domain example.docker             # Delete a record"
echo ""
echo "The DNS management in container/dns.go has been updated to use this tool,"
echo "which should make DNS record management more reliable." 