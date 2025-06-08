# TimescaleDB Migration Guide

This document outlines the migration from regular PostgreSQL to TimescaleDB for better time-series data management in LaunchStack-Backend.

## Overview

TimescaleDB is a PostgreSQL extension that optimizes storage and querying of time-series data. It maintains full compatibility with standard PostgreSQL while adding specialized capabilities that significantly improve performance for time-series workloads.

## Migration Benefits

1. **Improved Query Performance**: Up to 10-100x faster queries on time-series data
2. **Better Data Compression**: Reduces storage requirements by up to 90%
3. **Automated Data Lifecycle Management**: Automatic data retention and downsampling
4. **Full PostgreSQL Compatibility**: Works with existing PostgreSQL tools and libraries

## Docker Setup

The TimescaleDB database is configured using Docker Compose:

```yaml
version: "3.8"

services:
  postgres:
    image: timescale/timescaledb:latest-pg17
    container_name: launchstack-postgres
    environment:
      POSTGRES_DB: launchstack
      POSTGRES_USER: postgres
      POSTGRES_PASSWORD: postgres
      TIMESCALEDB_TUNE: "true"
      TIMESCALEDB_TUNE_MAX_CONNS: "100"
    networks:
      macvlan:
        ipv4_address: 10.1.1.82
    restart: unless-stopped
    volumes:
      - /SSD/Postgres/LaunchStack:/var/lib/postgresql/data
    command: postgres -c shared_preload_libraries=timescaledb
    ports:
      - "5432:5432"

networks:
  macvlan:
    external: true
```

## Database Schema Changes

### 1. Resource Usage Table

The `resource_usage` table has been restructured to work with TimescaleDB:

```sql
CREATE TABLE resource_usage (
    id UUID NOT NULL,
    instance_id UUID NOT NULL REFERENCES instances(id) ON DELETE CASCADE,
    timestamp TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    cpu_usage FLOAT NOT NULL DEFAULT 0,
    memory_usage BIGINT NOT NULL DEFAULT 0,
    memory_limit BIGINT NOT NULL DEFAULT 0,
    network_in BIGINT NOT NULL DEFAULT 0,
    network_out BIGINT NOT NULL DEFAULT 0,
    PRIMARY KEY (id, timestamp)
);
```

Key changes:
- Primary key includes `timestamp` for partitioning
- Table is converted to a hypertable with `create_hypertable('resource_usage', 'timestamp')`

### 2. Continuous Aggregates

For efficient querying of historical data, we've created continuous aggregates:

```sql
CREATE MATERIALIZED VIEW resource_usage_hourly
WITH (timescaledb.continuous) AS
SELECT
    instance_id,
    time_bucket('1 hour', timestamp) AS bucket,
    AVG(cpu_usage) AS avg_cpu,
    MAX(cpu_usage) AS max_cpu,
    AVG(memory_usage) AS avg_memory,
    MAX(memory_usage) AS max_memory,
    SUM(network_in) AS total_network_in,
    SUM(network_out) AS total_network_out
FROM resource_usage
GROUP BY instance_id, bucket;
```

### 3. Data Lifecycle Policies

- **Retention Policy**: Automatically drops data older than 30 days
- **Compression Policy**: Compresses data older than 7 days
- **Refresh Policy**: Updates aggregates hourly

## Application Code Changes

The application code has been updated to work with TimescaleDB:

1. **Database Connection**:
   - Updated to connect to TimescaleDB
   - Uses environment variables for configuration

2. **Hypertable Management**:
   - Added automatic creation of hypertable during migration
   - Added data compression setup

3. **Resource Usage Queries**:
   - Updated to use TimescaleDB-specific functions for time-based queries
   - Added support for querying both raw data and aggregates

## Querying Time-Series Data

### Raw Data Queries

```sql
-- Get recent stats for an instance
SELECT * FROM resource_usage 
WHERE instance_id = 'instance-uuid' 
ORDER BY timestamp DESC 
LIMIT 60;
```

### Aggregate Queries

```sql
-- Get hourly average CPU usage for the past day
SELECT bucket, avg_cpu
FROM resource_usage_hourly
WHERE instance_id = 'instance-uuid'
  AND bucket > NOW() - INTERVAL '1 day'
ORDER BY bucket DESC;
```

## Migration Process

1. Deploy TimescaleDB using docker-compose
2. Run the migration script (`setup_db.sh`)
3. Verify the migration with test queries
4. Update application code to use the new database structure

## Troubleshooting

If you encounter issues during migration:

1. Check TimescaleDB extension status:
   ```sql
   SELECT extname, extversion FROM pg_extension WHERE extname = 'timescaledb';
   ```

2. Verify hypertable creation:
   ```sql
   SELECT * FROM timescaledb_information.hypertables 
   WHERE hypertable_name = 'resource_usage';
   ```

3. Check continuous aggregates:
   ```sql
   SELECT * FROM timescaledb_information.continuous_aggregates 
   WHERE view_name = 'resource_usage_hourly';
   ``` 