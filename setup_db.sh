#!/bin/bash

# Database connection details
DB_URI="postgresql://postgres:npg_eiCzc53PmMRS@10.1.1.82:5432/launchstack"

echo "Connecting to TimescaleDB and applying migration..."
psql "$DB_URI" -f db_migration.sql

# Check if migration was successful
if [ $? -eq 0 ]; then
    echo "✅ Database migration completed successfully."
    echo "✅ TimescaleDB extension enabled."
    echo "✅ Resource usage table converted to hypertable."
    
    # Run the continuous aggregate script separately (outside transaction)
    echo "Setting up continuous aggregates..."
    psql "$DB_URI" -f db_migration_aggregate.sql
    
    if [ $? -eq 0 ]; then
        echo "✅ Continuous aggregates created for better performance."
        echo "✅ Retention policy set to 30 days."
        echo "✅ Compression policy applied to optimize storage."
    else
        echo "⚠️ Warning: Continuous aggregates setup had issues."
    fi
else
    echo "❌ Database migration failed."
    exit 1
fi

# Apply compression policy
echo "Setting up compression policy..."
psql "$DB_URI" -c "ALTER TABLE resource_usage SET (timescaledb.compress, timescaledb.compress_segmentby = 'instance_id');"
psql "$DB_URI" -c "SELECT add_compression_policy('resource_usage', INTERVAL '7 days', if_not_exists => TRUE);"

# Check TimescaleDB status
echo "Verifying TimescaleDB setup..."
psql "$DB_URI" -c "SELECT extname, extversion FROM pg_extension WHERE extname = 'timescaledb';"
psql "$DB_URI" -c "SELECT * FROM timescaledb_information.hypertables WHERE hypertable_name = 'resource_usage';"
psql "$DB_URI" -c "SELECT * FROM timescaledb_information.continuous_aggregates WHERE view_name = 'resource_usage_hourly';"

echo ""
echo "Testing connection to the database..."
psql "$DB_URI" -c "SELECT version();" 