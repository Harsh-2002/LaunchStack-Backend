#!/bin/bash

# TimescaleDB verification script
echo "===== TimescaleDB Resource Usage Verification ====="
echo ""

# Database connection info
DB_HOST="10.1.1.82"
DB_PORT="5432"
DB_USER="postgres"
DB_PASS="npg_eiCzc53PmMRS"
DB_NAME="launchstack"

# Connect function
function run_query() {
  PGPASSWORD=$DB_PASS psql -h $DB_HOST -U $DB_USER -d $DB_NAME -c "$1"
}

# Verify TimescaleDB is installed
echo "Checking TimescaleDB installation:"
run_query "SELECT extname, extversion FROM pg_extension WHERE extname = 'timescaledb';"
echo ""

# Verify hypertable configuration
echo "Checking hypertable configuration:"
run_query "SELECT * FROM timescaledb_information.hypertables WHERE hypertable_name = 'resource_usage';"
echo ""

# Check compression policy
echo "Checking compression policy:"
run_query "SELECT * FROM timescaledb_information.compression_settings WHERE hypertable_name = 'resource_usage';"
echo ""

# Check retention policy
echo "Checking retention policy:"
run_query "SELECT * FROM timescaledb_information.jobs WHERE proc_name = 'policy_retention';"
echo ""

# Check continuous aggregates
echo "Checking continuous aggregates:"
run_query "SELECT * FROM timescaledb_information.continuous_aggregates WHERE hypertable_name = 'resource_usage';"
echo ""

# Check most recent resource usage records
echo "Recent resource usage records:"
run_query "SELECT * FROM resource_usage ORDER BY timestamp DESC LIMIT 5;"
echo ""

# Check for instances with no resource usage data
echo "Instances with no resource usage data:"
run_query "
SELECT i.id, i.name, i.container_id, i.status 
FROM instances i 
LEFT JOIN resource_usage ru ON i.id = ru.instance_id 
WHERE ru.id IS NULL AND i.status = 'running' 
LIMIT 5;
"
echo ""

# Check aggregated stats
echo "Hourly aggregated stats from the last 24 hours:"
run_query "
SELECT 
    date_trunc('hour', timestamp) AS hour,
    instance_id,
    count(*) as sample_count,
    avg(cpu_usage) as avg_cpu,
    max(cpu_usage) as max_cpu,
    avg(memory_usage)/1024/1024 as avg_memory_mb
FROM resource_usage 
WHERE timestamp > NOW() - INTERVAL '24 hours'
GROUP BY hour, instance_id 
ORDER BY hour DESC
LIMIT 10;
"
echo ""

# Force collection of stats if none exist
if [ $(run_query "SELECT COUNT(*) FROM resource_usage;" | grep -oP '\d+' | head -1) -eq 0 ]; then
    echo "No resource usage data found. Checking available instances:"
    
    # List running instances
    run_query "SELECT id, name, container_id, status FROM instances WHERE status = 'running';"
    
    echo ""
    echo "To force resource collection for a specific instance, you can use the API endpoint:"
    echo "curl -X GET 'http://localhost:8080/api/v1/instances/[instance-id]/stats'"
    echo ""
fi

echo "===== Verification Complete =====" 