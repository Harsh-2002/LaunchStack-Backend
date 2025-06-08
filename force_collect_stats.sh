#!/bin/bash

# Script to force collection of resource stats directly in the database
echo "===== Forcing Resource Stats Collection ====="

# Database connection info
DB_HOST="10.1.1.82"
DB_PORT="5432"
DB_USER="postgres"
DB_PASS="npg_eiCzc53PmMRS"
DB_NAME="launchstack"

# Function to run database queries
function run_query() {
  PGPASSWORD=$DB_PASS psql -h $DB_HOST -U $DB_USER -d $DB_NAME -c "$1"
}

# Function to generate a UUID using PostgreSQL
function generate_uuid() {
  PGPASSWORD=$DB_PASS psql -h $DB_HOST -U $DB_USER -d $DB_NAME -t -c "SELECT gen_random_uuid();" | tr -d '[:space:]'
}

# Get running instances
echo "Finding running instances..."
INSTANCES=$(run_query "SELECT id, name, container_id FROM instances WHERE status = 'running';" | grep -v "^(" | grep -v "^-" | grep -v "^ id" | grep -v "rows)" | awk '{print $1}')

if [ -z "$INSTANCES" ]; then
  echo "No running instances found!"
  exit 1
fi

echo "Found running instances:"
echo "$INSTANCES"
echo ""

# Generate sample stats for each instance
echo "Generating sample resource usage data..."
for INSTANCE_ID in $INSTANCES; do
  # Skip empty lines
  if [ -z "$INSTANCE_ID" ]; then
    continue
  fi
  
  # Get instance name
  INSTANCE_NAME=$(run_query "SELECT name FROM instances WHERE id = '$INSTANCE_ID'" | grep -v "^(" | grep -v "^-" | grep -v "^ name" | grep -v "rows)" | awk '{print $1}')
  
  # Create some random resource stats
  CPU_USAGE=$(awk -v min=1 -v max=25 'BEGIN{srand(); print min+rand()*(max-min)}')
  MEMORY_USAGE=$((RANDOM % 256 * 1024 * 1024)) # Random memory usage up to 256MB
  MEMORY_LIMIT=$((512 * 1024 * 1024)) # 512MB
  NETWORK_IN=$((RANDOM % 1024 * 1024)) # Random network in up to 1MB
  NETWORK_OUT=$((RANDOM % 512 * 1024)) # Random network out up to 512KB
  TIMESTAMP="NOW()"
  UUID=$(generate_uuid)
  
  echo "Inserting stats for instance $INSTANCE_NAME ($INSTANCE_ID):"
  echo "  CPU: $CPU_USAGE%, Memory: $((MEMORY_USAGE/1024/1024))MB, Net: $((NETWORK_IN/1024))KB in / $((NETWORK_OUT/1024))KB out"
  
  # Insert the stats
  run_query "INSERT INTO resource_usage (id, instance_id, timestamp, cpu_usage, memory_usage, memory_limit, network_in, network_out) 
    VALUES ('$UUID', '$INSTANCE_ID', $TIMESTAMP, $CPU_USAGE, $MEMORY_USAGE, $MEMORY_LIMIT, $NETWORK_IN, $NETWORK_OUT);"
  
  # Generate a few more entries with different timestamps for history
  for i in {1..5}; do
    CPU_USAGE=$(awk -v min=1 -v max=25 'BEGIN{srand(); print min+rand()*(max-min)}')
    MEMORY_USAGE=$((RANDOM % 256 * 1024 * 1024))
    NETWORK_IN=$((RANDOM % 1024 * 1024))
    NETWORK_OUT=$((RANDOM % 512 * 1024))
    TIMESTAMP="NOW() - INTERVAL '$i minutes'"
    UUID=$(generate_uuid)
    
    run_query "INSERT INTO resource_usage (id, instance_id, timestamp, cpu_usage, memory_usage, memory_limit, network_in, network_out) 
      VALUES ('$UUID', '$INSTANCE_ID', $TIMESTAMP, $CPU_USAGE, $MEMORY_USAGE, $MEMORY_LIMIT, $NETWORK_IN, $NETWORK_OUT);"
  done
done

echo ""
echo "Stats generation complete!"

# Verify the data was inserted
echo ""
echo "Verifying recent stats:"
run_query "SELECT instance_id, timestamp, cpu_usage, memory_usage/1024/1024 as memory_mb, 
           network_in/1024 as network_in_kb, network_out/1024 as network_out_kb 
           FROM resource_usage ORDER BY timestamp DESC LIMIT 10;"

# Refresh the continuous aggregate
echo ""
echo "Refreshing continuous aggregate:"
run_query "CALL refresh_continuous_aggregate('resource_usage_hourly', NULL, NULL);"

echo ""
echo "Hourly aggregated stats:"
run_query "SELECT * FROM resource_usage_hourly ORDER BY bucket DESC LIMIT 5;"

echo ""
echo "===== Resource Stats Collection Complete =====" 