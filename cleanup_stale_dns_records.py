#!/usr/bin/env python3

import os
import sys
import json
import base64
import logging
import time
import subprocess
import argparse
import requests
from urllib.parse import urljoin

# Configure logging
logging.basicConfig(
    level=logging.INFO,
    format="%(asctime)s [%(levelname)s] %(message)s",
    handlers=[logging.StreamHandler()],
)
logger = logging.getLogger(__name__)

# Default AdGuard DNS configuration
DEFAULT_CONFIG = {
    "protocol": os.environ.get("ADGUARD_PROTOCOL", "https"),
    "host": os.environ.get("ADGUARD_HOST", "dns.srvr.site"),
    "username": os.environ.get("ADGUARD_USERNAME", "Pi"),
    "password": os.environ.get("ADGUARD_PASSWORD", "9130458959"),
}

class AdGuardDNSManager:
    def __init__(self, protocol=None, host=None, username=None, password=None):
        self.protocol = protocol or DEFAULT_CONFIG["protocol"]
        self.host = host or DEFAULT_CONFIG["host"]
        self.username = username or DEFAULT_CONFIG["username"]
        self.password = password or DEFAULT_CONFIG["password"]
        self.base_url = f"{self.protocol}://{self.host}"
        self.auth_header = self._generate_auth_header()
        
    def _generate_auth_header(self):
        """Generate Basic Auth header from username and password"""
        credentials = f"{self.username}:{self.password}"
        encoded = base64.b64encode(credentials.encode("utf-8")).decode("utf-8")
        return {"Authorization": f"Basic {encoded}"}
        
    def _make_request(self, endpoint, method="GET", data=None, params=None):
        """Make a request to the AdGuard DNS API"""
        url = urljoin(self.base_url, endpoint)
        headers = {**self.auth_header}
        
        if data:
            headers["Content-Type"] = "application/json"
            data = json.dumps(data)
            
        try:
            if method == "GET":
                response = requests.get(url, headers=headers, params=params)
            elif method == "POST":
                response = requests.post(url, headers=headers, data=data, params=params)
            elif method == "PUT":
                response = requests.put(url, headers=headers, data=data, params=params)
            elif method == "DELETE":
                response = requests.delete(url, headers=headers, data=data, params=params)
            else:
                raise ValueError(f"Unsupported HTTP method: {method}")
                
            response.raise_for_status()
            
            # Some endpoints return empty responses
            if response.text:
                return response.json() if response.headers.get("content-type", "").startswith("application/json") else response.text
            return {}
            
        except requests.exceptions.RequestException as e:
            logger.error(f"Request failed: {e}")
            if hasattr(e, "response") and e.response is not None:
                logger.error(f"Response: {e.response.text}")
            return None
            
    def get_dns_records(self):
        """Get all DNS rewrites"""
        return self._make_request("control/rewrite/list")
        
    def add_dns_record(self, domain, answer):
        """Add a DNS record"""
        logger.info(f"Adding DNS record: {domain} -> {answer}")
        return self._make_request("control/rewrite/add", method="POST", data={"domain": domain, "answer": answer})
        
    def delete_dns_record(self, domain):
        """Delete a DNS record"""
        logger.info(f"Deleting DNS record: {domain}")
        return self._make_request("control/rewrite/delete", method="POST", data={"domain": domain})
        
    def deduplicate_records(self, records):
        """Remove duplicate DNS records based on domain name"""
        unique_records = {}
        for record in records:
            domain = record.get("domain")
            if domain:
                unique_records[domain] = record
        return list(unique_records.values())

def get_active_containers():
    """Get a list of active Docker containers"""
    try:
        # Get active container names
        result = subprocess.run(
            ["docker", "ps", "--format", "{{.Names}}"],
            capture_output=True,
            text=True,
            check=True
        )
        container_names = result.stdout.strip().split("\n")
        return [name for name in container_names if name]  # Filter out empty names
    except (subprocess.SubprocessError, FileNotFoundError) as e:
        logger.error(f"Error getting active containers: {e}")
        return []

def get_test_active_containers():
    """Get a test list of active containers for demo purposes"""
    # Return a fixed list of container names for testing
    return ["blue-wolf", "pearl-puppy", "jolly-willow"]

def is_container_dns_record(domain):
    """Check if a DNS record domain name looks like a container name"""
    # Only consider records ending with .docker
    if not domain.endswith(".docker"):
        return False
    
    # Special excluded records
    excluded = ["test.docker", "test-record.docker"]
    if domain in excluded:
        return False
        
    return True

def cleanup_stale_dns_records(dry_run=False, force_cleanup=False, test_mode=False):
    """
    Clean up stale DNS records by comparing active containers with DNS records.
    
    Args:
        dry_run: If True, only print what would be done without making changes
        force_cleanup: If True, perform a full cleanup by deleting all container DNS records
                      and recreating only those for active containers
        test_mode: If True, use test data instead of real Docker containers
    """
    # Initialize DNS manager
    dns_manager = AdGuardDNSManager()
    
    # Get active container names
    if test_mode:
        active_containers = get_test_active_containers()
        logger.info("Using test container data")
    else:
        active_containers = get_active_containers()
        
    logger.info(f"Found {len(active_containers)} active containers: {', '.join(active_containers)}")
    
    # Get all DNS records
    dns_records = dns_manager.get_dns_records()
    if dns_records is None:
        logger.error("Failed to get DNS records")
        return False
        
    # Deduplicate DNS records
    dns_records = dns_manager.deduplicate_records(dns_records)
    logger.info(f"Found {len(dns_records)} unique DNS records")
    
    # Identify container DNS records
    container_dns_records = []
    other_dns_records = []
    for record in dns_records:
        domain = record.get("domain", "")
        if is_container_dns_record(domain):
            container_dns_records.append(record)
        else:
            other_dns_records.append(record)
            
    logger.info(f"Found {len(container_dns_records)} container DNS records")
    logger.info(f"Found {len(other_dns_records)} other DNS records")
    
    # Identify stale records (container records without active containers)
    stale_records = []
    active_records = []
    for record in container_dns_records:
        domain = record.get("domain", "")
        container_name = domain.replace(".docker", "")
        if container_name in active_containers:
            active_records.append(record)
        else:
            stale_records.append(record)
            
    logger.info(f"Found {len(stale_records)} stale container DNS records")
    logger.info(f"Found {len(active_records)} active container DNS records")
    
    if dry_run:
        logger.info("DRY RUN: No changes will be made")
        
        logger.info("Active containers:")
        for container in active_containers:
            logger.info(f"  {container}")
            
        logger.info("Stale DNS records that would be deleted:")
        for record in stale_records:
            domain = record.get("domain", "")
            answer = record.get("answer", "")
            logger.info(f"  {domain} -> {answer}")
            
        return True
        
    # Delete stale records
    if stale_records:
        if force_cleanup:
            logger.info("Force cleanup mode: Deleting all container DNS records and recreating active ones")
            
            # Delete all container DNS records
            for record in container_dns_records:
                domain = record.get("domain", "")
                dns_manager.delete_dns_record(domain)
                
            # Recreate active container DNS records
            for record in active_records:
                domain = record.get("domain", "")
                answer = record.get("answer", "")
                dns_manager.add_dns_record(domain, answer)
                
            logger.info("Force cleanup completed")
        else:
            logger.info("Deleting stale DNS records")
            deleted_count = 0
            for record in stale_records:
                domain = record.get("domain", "")
                dns_manager.delete_dns_record(domain)
                deleted_count += 1
                
            logger.info(f"Attempted to delete {deleted_count} stale DNS records")
            
        # Verify cleanup
        time.sleep(2)  # Wait for changes to take effect
        updated_records = dns_manager.get_dns_records()
        if updated_records is None:
            logger.error("Failed to get updated DNS records")
            return False
            
        # Check for remaining stale records
        remaining_stale = 0
        for record in updated_records:
            domain = record.get("domain", "")
            if is_container_dns_record(domain):
                container_name = domain.replace(".docker", "")
                if container_name not in active_containers:
                    remaining_stale += 1
                    
        if remaining_stale > 0:
            logger.warning(f"There are still {remaining_stale} stale DNS records remaining")
            if not force_cleanup:
                logger.info("Consider running with --force to perform a complete cleanup")
        else:
            logger.info("All stale DNS records have been removed")
            
    else:
        logger.info("No stale DNS records found")
        
    return True

def main():
    parser = argparse.ArgumentParser(description="Clean up stale DNS records for Docker containers")
    parser.add_argument("--dry-run", action="store_true", help="Print what would be done without making changes")
    parser.add_argument("--force", action="store_true", help="Force a complete cleanup by recreating all DNS records")
    parser.add_argument("--test", action="store_true", help="Run in test mode with sample data")
    args = parser.parse_args()
    
    if cleanup_stale_dns_records(dry_run=args.dry_run, force_cleanup=args.force, test_mode=args.test):
        return 0
    else:
        return 1

if __name__ == "__main__":
    sys.exit(main()) 