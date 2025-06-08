# Resource Allocation Strategy

This document outlines the resource allocation strategy for LaunchStack instances based on subscription plans.

## Overview

Each instance created on LaunchStack is allocated specific computing resources based on the user's subscription plan. Resources are enforced at the container level using Docker's resource constraints.

## Resource Allocation by Plan

| Plan    | Pricing            | Instances | CPU per Instance | Memory per Instance | Storage per Instance |
|---------|-------------------|-----------|------------------|---------------------|---------------------|
| Starter | $2/mo or $20/yr   | 1         | 0.5 CPU          | 512 MB              | 1 GB                |
| Pro     | $5/mo or $50/yr   | 10        | 1.0 CPU          | 1 GB                | 20 GB               |

The Starter plan includes a 7-day free trial period. After the trial ends, users are billed according to their selected billing cycle (monthly or yearly). If payment fails, instances will be marked as expired and scheduled for deletion after a grace period.

## Implementation Details

### CPU Limits

CPU limits are soft limits that restrict the percentage of host CPU time that a container can use. For example:
- 0.5 CPU means the container can use up to 50% of a single CPU core
- 1.0 CPU means the container can use up to 100% of a single CPU core

In practice, Docker allows CPU-hungry applications to burst above their limit temporarily if host resources are available, but will throttle them back to their limit over time.

### Memory Limits

Memory limits are hard limits that restrict the amount of RAM a container can use. If a container exceeds its memory limit, it may be killed by the host operating system's OOM (Out of Memory) killer.

Memory limits are specified in MB (megabytes):
- 512 MB for Starter tier
- 1024 MB (1 GB) for Pro tier

### Storage Limits

Storage limits represent the amount of persistent disk storage allocated to each instance. This storage is used for:
- n8n workflow data
- Execution history
- Custom files and data

Storage is implemented using Docker volumes and host-mounted directories.

## Resource Usage Monitoring

Users can monitor their instance resource usage through the dashboard or API:
- Current CPU usage (percentage)
- Memory usage (MB and percentage of limit)
- Disk space usage (GB and percentage of limit)

Resource monitoring data is available through the `/api/v1/instances/:id/stats` endpoint.

## Scaling Considerations

- Instances cannot exceed their plan's resource limits
- Users can upgrade their plan to get access to more resources
- During peak load, instances will be throttled to stay within their resource limits
- If an instance consistently hits resource limits, users should consider upgrading to a higher tier plan

## Technical Implementation

Resource limits are implemented in the Docker container creation process in `container/docker.go`. The limits are determined by the user's subscription plan and are retrieved from the `User` model's methods such as:

- `GetCPULimit()`
- `GetMemoryLimit()`
- `GetStorageLimit()`
- `GetInstancesLimit()`

These limits are applied when creating containers and are enforced by Docker's cgroups system. 