#!/bin/bash

# Database connection details
OLD_DB_URI="postgresql://launchstack_owner:npg_eiCzc53PmMRS@ep-noisy-mountain-a1y8gyu3-pooler.ap-southeast-1.aws.neon.tech/launchstack?sslmode=require"
NEW_DB_URI="postgresql://postgres:npg_eiCzc53PmMRS@10.1.1.82:5432/launchstack"

echo "🚀 Starting data migration from old database to TimescaleDB..."
echo "This script will migrate all data while preserving relationships."

# Check if we can connect to both databases
echo "Testing connections to databases..."
OLD_DB_VERSION=$(PGPASSWORD=npg_eiCzc53PmMRS psql -h ep-noisy-mountain-a1y8gyu3-pooler.ap-southeast-1.aws.neon.tech -U launchstack_owner -d launchstack -c "SELECT version();" -t 2>/dev/null)
NEW_DB_VERSION=$(PGPASSWORD=npg_eiCzc53PmMRS psql -h 10.1.1.82 -U postgres -d launchstack -c "SELECT version();" -t 2>/dev/null)

if [ -z "$OLD_DB_VERSION" ]; then
    echo "❌ Cannot connect to old database. Please check credentials and network."
    exit 1
fi

if [ -z "$NEW_DB_VERSION" ]; then
    echo "❌ Cannot connect to new TimescaleDB. Please check credentials and network."
    exit 1
fi

echo "✅ Successfully connected to both databases."
echo "Old DB: $OLD_DB_VERSION"
echo "New DB: $NEW_DB_VERSION"

# Migration Process
echo "Starting migration process..."

# Step 1: Export Users
echo "Migrating Users table..."
pg_dump -h ep-noisy-mountain-a1y8gyu3-pooler.ap-southeast-1.aws.neon.tech -U launchstack_owner -d launchstack -t users --data-only | \
  PGPASSWORD=npg_eiCzc53PmMRS psql -h 10.1.1.82 -U postgres -d launchstack
echo "✅ Users migrated"

# Step 2: Export Instances
echo "Migrating Instances table..."
pg_dump -h ep-noisy-mountain-a1y8gyu3-pooler.ap-southeast-1.aws.neon.tech -U launchstack_owner -d launchstack -t instances --data-only | \
  PGPASSWORD=npg_eiCzc53PmMRS psql -h 10.1.1.82 -U postgres -d launchstack
echo "✅ Instances migrated"

# Step 3: Export Resource Usage data with transformation
echo "Migrating Resource Usage data (this may take a while)..."

# Create temporary migration file
cat > migrate_resource_usage.sql << EOF
-- Temporary file to migrate resource usage data
COPY (
  SELECT 
    id, 
    instance_id, 
    timestamp, 
    cpu_usage, 
    memory_usage, 
    memory_limit, 
    network_in, 
    network_out
  FROM 
    resource_usages
) TO STDOUT;
EOF

# Execute the export and import
PGPASSWORD=npg_eiCzc53PmMRS psql -h ep-noisy-mountain-a1y8gyu3-pooler.ap-southeast-1.aws.neon.tech -U launchstack_owner -d launchstack -f migrate_resource_usage.sql | \
  PGPASSWORD=npg_eiCzc53PmMRS psql -h 10.1.1.82 -U postgres -d launchstack -c "COPY resource_usage (id, instance_id, timestamp, cpu_usage, memory_usage, memory_limit, network_in, network_out) FROM STDIN;"

rm migrate_resource_usage.sql
echo "✅ Resource Usage data migrated"

# Step 4: Export any other necessary tables (add more as needed)
echo "Migrating Payments table..."
pg_dump -h ep-noisy-mountain-a1y8gyu3-pooler.ap-southeast-1.aws.neon.tech -U launchstack_owner -d launchstack -t payments --data-only 2>/dev/null | \
  PGPASSWORD=npg_eiCzc53PmMRS psql -h 10.1.1.82 -U postgres -d launchstack 2>/dev/null || echo "⚠️ Payments table not found or empty"

# Verify migration
echo "Verifying migration..."

# Count users
OLD_USERS=$(PGPASSWORD=npg_eiCzc53PmMRS psql -h ep-noisy-mountain-a1y8gyu3-pooler.ap-southeast-1.aws.neon.tech -U launchstack_owner -d launchstack -c "SELECT COUNT(*) FROM users;" -t | tr -d ' ')
NEW_USERS=$(PGPASSWORD=npg_eiCzc53PmMRS psql -h 10.1.1.82 -U postgres -d launchstack -c "SELECT COUNT(*) FROM users;" -t | tr -d ' ')

# Count instances
OLD_INSTANCES=$(PGPASSWORD=npg_eiCzc53PmMRS psql -h ep-noisy-mountain-a1y8gyu3-pooler.ap-southeast-1.aws.neon.tech -U launchstack_owner -d launchstack -c "SELECT COUNT(*) FROM instances;" -t | tr -d ' ')
NEW_INSTANCES=$(PGPASSWORD=npg_eiCzc53PmMRS psql -h 10.1.1.82 -U postgres -d launchstack -c "SELECT COUNT(*) FROM instances;" -t | tr -d ' ')

# Count resource usage records
OLD_RESOURCE=$(PGPASSWORD=npg_eiCzc53PmMRS psql -h ep-noisy-mountain-a1y8gyu3-pooler.ap-southeast-1.aws.neon.tech -U launchstack_owner -d launchstack -c "SELECT COUNT(*) FROM resource_usages;" -t 2>/dev/null | tr -d ' ' || echo "0")
NEW_RESOURCE=$(PGPASSWORD=npg_eiCzc53PmMRS psql -h 10.1.1.82 -U postgres -d launchstack -c "SELECT COUNT(*) FROM resource_usage;" -t | tr -d ' ')

echo "Migration Summary:"
echo "Users: $OLD_USERS → $NEW_USERS"
echo "Instances: $OLD_INSTANCES → $NEW_INSTANCES"
echo "Resource Usage Records: $OLD_RESOURCE → $NEW_RESOURCE"

if [ "$OLD_USERS" == "$NEW_USERS" ] && [ "$OLD_INSTANCES" == "$NEW_INSTANCES" ]; then
    echo "✅ Migration completed successfully!"
    
    # Update the continuous aggregate view
    echo "Refreshing continuous aggregate views..."
    PGPASSWORD=npg_eiCzc53PmMRS psql -h 10.1.1.82 -U postgres -d launchstack -c "CALL refresh_continuous_aggregate('resource_usage_hourly', NULL, NULL);" 2>/dev/null || echo "⚠️ Continuous aggregate refresh not available yet"
else
    echo "⚠️ Migration completed with potential issues. Please verify the data manually."
fi 