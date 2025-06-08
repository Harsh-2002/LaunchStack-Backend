-- Enable TimescaleDB extension
CREATE EXTENSION IF NOT EXISTS timescaledb CASCADE;

-- Create base tables if they don't exist (simplified version)
CREATE TABLE IF NOT EXISTS users (
    id UUID PRIMARY KEY,
    username VARCHAR(50) UNIQUE NOT NULL,
    email VARCHAR(100) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS instances (
    id UUID PRIMARY KEY,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    name VARCHAR(100) NOT NULL,
    container_id VARCHAR(255),
    status VARCHAR(20) NOT NULL DEFAULT 'stopped',
    port_external INT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Create resource_usage table without a primary key constraint on id
-- We'll use (id, timestamp) as the primary key instead
DROP TABLE IF EXISTS resource_usage CASCADE;
CREATE TABLE IF NOT EXISTS resource_usage (
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

-- Create index on instance_id
CREATE INDEX IF NOT EXISTS idx_resource_usage_instance_id ON resource_usage(instance_id);

-- Convert resource_usage table to a TimescaleDB hypertable
-- The if_not_exists parameter prevents errors if already a hypertable
SELECT create_hypertable('resource_usage', 'timestamp', if_not_exists => TRUE, migrate_data => TRUE);

-- Now add the TimescaleDB features after confirming the hypertable exists
DO $$
BEGIN
    -- Check if resource_usage is a hypertable
    IF EXISTS (
        SELECT 1 
        FROM timescaledb_information.hypertables 
        WHERE hypertable_name = 'resource_usage'
    ) THEN
        -- Set retention policy - keep data for 30 days
        PERFORM add_retention_policy('resource_usage', INTERVAL '30 days', if_not_exists => TRUE);
        
        -- Create continuous aggregates for better query performance
        BEGIN
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
            PERFORM add_continuous_aggregate_policy('resource_usage_hourly',
                start_offset => INTERVAL '2 days',
                end_offset => INTERVAL '1 hour',
                schedule_interval => INTERVAL '1 hour',
                if_not_exists => TRUE);
                
            -- Create a compression policy
            ALTER TABLE resource_usage SET (
                timescaledb.compress,
                timescaledb.compress_segmentby = 'instance_id'
            );
            
            PERFORM add_compression_policy('resource_usage', INTERVAL '7 days', if_not_exists => TRUE);
        EXCEPTION WHEN OTHERS THEN
            RAISE NOTICE 'Error setting up TimescaleDB policies: %', SQLERRM;
        END;
    ELSE
        RAISE NOTICE 'resource_usage is not a hypertable, skipping TimescaleDB policies';
    END IF;
END $$; 