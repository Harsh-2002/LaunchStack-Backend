# LaunchStack Backend

LaunchStack is a platform for deploying and managing n8n instances.

## Directory Structure

- `config/`: Configuration files and structures
- `container/`: Docker container management code
- `db/`: Database models and migrations
- `docs/`: Documentation files
- `middleware/`: Middleware for authentication, CORS, etc.
- `models/`: Data models
- `routes/`: API route handlers
- `tests/`: Test scripts and tools

## Setup

1. Clone the repository
2. Copy `.env.example` to `.env` and update with your configuration
3. Run `go build -o launchstack-backend main.go`
4. Run `./launchstack-backend`

## Testing

To run tests, use the `run_tests.sh` script:

```bash
./run_tests.sh [test_name]
```

For more information on testing, see the [tests/README.md](tests/README.md) file.

## Documentation

See the `docs/` directory for detailed documentation:

- [API Documentation](docs/API_DOCUMENTATION.md)
- [Architecture Diagram](docs/ARCHITECTURE_DIAGRAM.md)
- [Authentication Documentation](docs/AUTH_DOCUMENTATION.md)
- [Database Schema](docs/DATABASE_SCHEMA.md)
- [DNS Management](docs/DNS_MANAGEMENT.md)
- [Environment Setup](docs/ENV_SETUP.md)
- [Improvement Checklist](docs/IMPROVEMENT_CHECKLIST.md)
- [Resource Allocation](docs/RESOURCE_ALLOCATION.md)

## API Documentation

### Instance Endpoints

#### Get Current Instance Metrics
- **Endpoint**: `GET /api/v1/instances/:id/stats`
- **Description**: Returns real-time resource usage for an instance
- **Authentication**: Required
- **URL Parameters**: 
  - `:id` - UUID of the instance
- **Response**: 
  ```json
  {
    "cpu_usage": 23.5,
    "memory_usage": 104857600,
    "memory_limit": 536870912,
    "memory_percentage": 19.5,
    "network_in": 1024000,
    "network_out": 512000
  }
  ```

#### Get Historical Instance Metrics
- **Endpoint**: `GET /api/v1/instances/:id/stats/history`
- **Description**: Returns historical resource usage data for an instance
- **Authentication**: Required
- **URL Parameters**:
  - `:id` - UUID of the instance
- **Query Parameters**:
  - `period` - Time period to fetch data for (default: "1h")
    - Options: "10m", "1h", "6h", "24h"
- **Response**:
  - An array of data points, ordered from newest to oldest (max 100 points)
  ```json
  [
    {
      "timestamp": "2023-06-08T12:34:56Z",
      "cpu_usage": 23.5,
      "memory_usage": 104857600,
      "memory_limit": 536870912,
      "memory_percentage": 19.5,
      "network_in": 1024000,
      "network_out": 512000
    },
    ...
  ]
  ```

## Historical Metrics Visualization

The TimescaleDB integration provides optimized time-series data for instance resource metrics. 

### 1. API Endpoints for Frontend Integration

- `GET /api/v1/instances/:id/stats` - Get current resource usage
- `GET /api/v1/instances/:id/stats/history` - Get historical resource usage (specific for frontend)

#### Historical Stats Parameters:

- `period`: Time period to fetch data for (default: "1h")
  - Options: "10m", "1h", "6h", "24h"

Example request matching frontend expectations:
```
GET /api/v1/instances/123e4567-e89b-12d3-a456-426614174000/stats/history?period=1h
```

### 2. Response Format

The API returns an array of data points with this structure:
```json
[
  {
    "timestamp": "2023-06-08T12:34:56Z",
    "cpu_usage": 23.5,
    "memory_usage": 104857600,
    "memory_limit": 536870912,
    "memory_percentage": 19.5,
    "network_in": 1024000,
    "network_out": 512000
  },
  ...
]
```

Key points about the response format:
- Data is limited to 100 data points maximum
- Each data point contains all required metrics
- CPU usage is in percentage (0-100)
- Memory usage is in bytes
- Timestamps are in ISO 8601 format

### 3. Frontend Implementation

For optimal visualization in the frontend:

```javascript
// Fetch historical metrics
async function fetchInstanceMetrics(instanceId, period = '1h') {
  try {
    const response = await fetch(`/api/v1/instances/${instanceId}/stats/history?period=${period}`);
    if (!response.ok) {
      throw new Error(`Error fetching metrics: ${response.statusText}`);
    }
    return await response.json();
  } catch (error) {
    console.error(`Failed to fetch metrics for instance ${instanceId}:`, error);
    return []; // Return empty array as fallback
  }
}

// Create chart with the fetched data
function createMetricsChart(container, metricsData) {
  const ctx = document.getElementById(container).getContext('2d');
  
  // Extract data for chart
  const timestamps = metricsData.map(point => new Date(point.timestamp));
  const cpuData = metricsData.map(point => point.cpu_usage);
  const memoryData = metricsData.map(point => point.memory_percentage);
  
  const chart = new Chart(ctx, {
    type: 'line',
    data: {
      labels: timestamps,
      datasets: [
        {
          label: 'CPU Usage (%)',
          data: cpuData,
          borderColor: 'rgba(75, 192, 192, 1)',
          tension: 0.1,
          fill: false
        },
        {
          label: 'Memory Usage (%)',
          data: memoryData,
          borderColor: 'rgba(153, 102, 255, 1)',
          tension: 0.1,
          fill: false
        }
      ]
    },
    options: {
      responsive: true,
      scales: {
        x: {
          type: 'time',
          time: {
            unit: period === '10m' ? 'minute' : 
                  period === '1h' || period === '6h' ? 'hour' : 'day'
          }
        },
        y: {
          beginAtZero: true,
          max: 100 // Since both CPU and memory are percentages
        }
      }
    }
  });
  
  return chart;
}

// Usage example
document.addEventListener('DOMContentLoaded', async () => {
  const instanceId = '123e4567-e89b-12d3-a456-426614174000';
  const periodSelector = document.getElementById('period-selector');
  let chart = null;
  
  async function updateChart() {
    const period = periodSelector.value;
    const metrics = await fetchInstanceMetrics(instanceId, period);
    
    if (chart) {
      chart.destroy();
    }
    
    if (metrics.length > 0) {
      chart = createMetricsChart('metrics-chart', metrics);
    } else {
      // Handle empty data
      document.getElementById('metrics-chart').innerHTML = 
        '<div class="no-data">No metrics data available for this period</div>';
    }
  }
  
  periodSelector.addEventListener('change', updateChart);
  updateChart();
});
```

### 4. Multiple Instance Dashboard

For dashboards displaying metrics from multiple instances:

```javascript
// Fetch metrics for all instances
async function fetchAllInstancesMetrics(instanceIds, period = '1h') {
  const promises = instanceIds.map(id => fetchInstanceMetrics(id, period));
  const results = await Promise.allSettled(promises);
  
  // Create a map of instance ID to metrics
  const metricsMap = {};
  results.forEach((result, index) => {
    if (result.status === 'fulfilled') {
      metricsMap[instanceIds[index]] = result.value;
    } else {
      console.warn(`Failed to fetch metrics for instance ${instanceIds[index]}`);
      metricsMap[instanceIds[index]] = [];
    }
  });
  
  return metricsMap;
}

// Aggregate CPU/memory across all instances
function calculateAggregateMetrics(metricsMap) {
  // First, create a timeline of all unique timestamps
  const allTimestamps = new Set();
  Object.values(metricsMap).forEach(metrics => {
    metrics.forEach(point => allTimestamps.add(point.timestamp));
  });
  
  // Sort timestamps chronologically
  const sortedTimestamps = Array.from(allTimestamps).sort();
  
  // For each timestamp, calculate the average CPU and memory usage
  const aggregateMetrics = sortedTimestamps.map(timestamp => {
    let totalCpu = 0;
    let totalMemory = 0;
    let instanceCount = 0;
    
    Object.values(metricsMap).forEach(metrics => {
      const point = metrics.find(p => p.timestamp === timestamp);
      if (point) {
        totalCpu += point.cpu_usage;
        totalMemory += point.memory_percentage;
        instanceCount++;
      }
    });
    
    return {
      timestamp,
      cpu_usage: instanceCount > 0 ? totalCpu / instanceCount : 0,
      memory_percentage: instanceCount > 0 ? totalMemory / instanceCount : 0,
      instance_count: instanceCount
    };
  });
  
  return aggregateMetrics;
}
```

### 5. Best Practices

- Always check data points count; limit client-side processing for larger datasets
- Implement error handling for failed requests
- Use appropriate time units based on the selected period
- Consider implementing auto-refresh for real-time monitoring
- Calculate memory usage percentage client-side if needed: `(memory_usage / memory_limit) * 100`
- For multiple instances, fetch data in parallel with Promise.all
- Handle missing data points gracefully

## TimescaleDB Technical Details

### Metrics Storage and Aggregation

The system uses TimescaleDB's specialized features for time-series data:

1. **Hypertables**: Resource usage data is stored in hypertables, which automatically partition data by time for efficient queries
   ```sql
   SELECT create_hypertable('resource_usage', 'timestamp');
   ```

2. **Data Retention**: Old data is automatically removed after 30 days to manage storage
   ```sql
   SELECT add_retention_policy('resource_usage', INTERVAL '30 days');
   ```

3. **Data Compression**: Older data is automatically compressed to save storage space
   ```sql
   ALTER TABLE resource_usage SET (
     timescaledb.compress,
     timescaledb.compress_segmentby = 'instance_id'
   );
   SELECT add_compression_policy('resource_usage', INTERVAL '7 days');
   ```

4. **Continuous Aggregates**: Pre-computed aggregates for faster queries on historical data
   ```sql
   CREATE MATERIALIZED VIEW resource_usage_hourly
   WITH (timescaledb.continuous) AS
   SELECT
     time_bucket('1 hour', timestamp) AS bucket,
     instance_id,
     AVG(cpu_usage) AS avg_cpu,
     MAX(cpu_usage) AS max_cpu,
     AVG(memory_usage) AS avg_memory,
     MAX(memory_usage) AS max_memory,
     SUM(network_in) AS total_network_in,
     SUM(network_out) AS total_network_out
   FROM resource_usage
   GROUP BY bucket, instance_id;
   ```

### Query Optimization

For optimal performance, the API uses:

1. Time bucketing to aggregate data points evenly
2. Automatic resolution selection based on time period
3. Limit of 100 data points to prevent excessive data transfer
4. Continuous aggregates for longer time periods

This implementation provides significant performance improvements over standard PostgreSQL:
- 10-100x faster queries for time-series data
- Up to 90% reduction in storage requirements
- Automated data lifecycle management

## License

Proprietary. 