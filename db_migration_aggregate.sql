-- This script must be run outside a transaction block 
-- and after the main migration has completed

-- Create continuous aggregates for better query performance
CREATE MATERIALIZED VIEW IF NOT EXISTS resource_usage_hourly
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

-- Set refresh policy for the continuous aggregate
SELECT add_continuous_aggregate_policy('resource_usage_hourly',
    start_offset => INTERVAL '2 days',
    end_offset => INTERVAL '1 hour',
    schedule_interval => INTERVAL '1 hour',
    if_not_exists => TRUE); 