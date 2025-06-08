# TimescaleDB Migration Steps

This document outlines the steps taken to migrate the LaunchStack-Backend system from the original PostgreSQL database to TimescaleDB for improved time-series data handling.

## Migration Overview

The migration process was completed in several stages:

1. **Environment Setup**
   - Configured docker-compose for TimescaleDB
   - Updated environment variables for new database connection

2. **Database Schema Setup**
   - Created tables with TimescaleDB-compatible structure
   - Converted resource_usage to a hypertable
   - Added continuous aggregates and data lifecycle policies

3. **Data Migration**
   - Exported and imported users and instances
   - Migrated resource usage historical data
   - Verified data integrity after migration

4. **Application Updates**
   - Modified database connection code
   - Updated queries to leverage TimescaleDB features

## Migration Scripts

### 1. Database Schema Setup (`setup_db.sh`)

This script initializes the TimescaleDB schema with:
- TimescaleDB extension
- Tables for users, instances, and resource usage
- Hypertable conversion for resource_usage
- Retention and compression policies

### 2. Data Migration (`migrate_data.sh`)

This script transfers data from the old database to TimescaleDB:
- Connects to both databases
- Exports and imports table data
- Transforms resource usage data for the new structure
- Verifies record counts match between databases

## Verification Steps

After migration, the following checks were performed:

1. **Data Integrity**
   - Verified record counts match between old and new databases
   - Checked referential integrity between tables

2. **TimescaleDB Features**
   - Confirmed resource_usage is properly configured as a hypertable
   - Verified continuous aggregates are collecting data
   - Tested compression and retention policies

3. **Application Compatibility**
   - Validated database connection using new credentials
   - Confirmed queries return expected results
   - Tested time-series specific functionality

## Rollback Plan

If issues are encountered with the TimescaleDB migration, the following rollback steps are available:

1. Revert environment variables in `.env` to use the original database
2. Restart the application to reconnect to the original database
3. No schema changes are needed as the original database remains untouched

## Post-Migration Tasks

After successful migration, consider these follow-up tasks:

1. **Performance Tuning**
   - Adjust TimescaleDB chunk sizes for optimal performance
   - Fine-tune retention and compression policies based on usage

2. **Monitoring**
   - Set up monitoring for TimescaleDB specific metrics
   - Monitor disk space usage for compressed and uncompressed chunks

3. **Backup Strategy**
   - Implement regular backups of the TimescaleDB database
   - Test restoration procedures 