# LaunchStack API Documentation

## Base URL

All API requests should be prefixed with `/api/v1`.

## Authentication

All endpoints except public health checks require authentication. Use Bearer token authentication.

```
Authorization: Bearer <token>
```

## Error Responses

Error responses follow this format:

```json
{
  "error": "Error message description"
}
```

## Common HTTP Status Codes

- `200 OK`: Request successful
- `400 Bad Request`: Invalid parameters
- `401 Unauthorized`: Authentication failed
- `403 Forbidden`: Permission denied
- `404 Not Found`: Resource not found
- `500 Internal Server Error`: Server error

---

## Endpoints

### Health Check

#### GET /health

Returns the health status of the API with system metrics and response times.

**Response**:
```json
{
  "status": "ok",
  "version": "1.0.0",
  "environment": "production",
  "timestamp": "2023-06-08T12:34:56Z",
  "database": {
    "status": "ok",
    "response_time_ms": 5
  },
  "system": {
    "memory_usage_mb": 24.5,
    "cpu_cores": 4,
    "go_routines": 12,
    "uptime": "2h15m30s"
  },
  "response_time_ms": 8
}
```

**Status Values**:
- `ok`: All systems operational
- `degraded`: One or more systems have issues but the API is still functional
- `error`: Critical failure

**Response Fields**:
- `status`: Overall health status
- `version`: API version
- `environment`: Current environment (production, staging, development)
- `timestamp`: Current time when the health check was performed
- `database`: Database connection status
  - `status`: Connection status
  - `response_time_ms`: Time taken to complete a database ping
- `system`: System resource information
  - `memory_usage_mb`: Current memory usage in MB
  - `cpu_cores`: Number of CPU cores available
  - `go_routines`: Current number of Go routines
  - `uptime`: Service uptime
- `response_time_ms`: Total time taken to process the health check request

### Users

#### GET /users/me

Returns the current user's profile information.

**Response**:
```json
{
  "id": "user_2Pc5GFJ3kd89qJlPg95XilHK41N",
  "email": "user@example.com",
  "username": "username",
  "created_at": "2023-06-08T12:34:56Z",
  "updated_at": "2023-06-08T12:34:56Z"
}
```

#### PUT /users/me

Updates the current user's profile information.

**Request Body**:
```json
{
  "username": "new_username"
}
```

**Response**:
```json
{
  "id": "user_2Pc5GFJ3kd89qJlPg95XilHK41N",
  "email": "user@example.com",
  "username": "new_username",
  "created_at": "2023-06-08T12:34:56Z",
  "updated_at": "2023-06-08T12:34:56Z"
}
```

### Instances

#### GET /instances

Returns a list of all instances owned by the current user.

**Response**:
```json
[
  {
    "id": "123e4567-e89b-12d3-a456-426614174000",
    "name": "Production n8n",
    "description": "Production workflow automation",
    "status": "running",
    "created_at": "2023-06-08T12:34:56Z",
    "updated_at": "2023-06-08T12:34:56Z",
    "memory_limit": 536870912,
    "domain": "prod-n8n.launchstack.io"
  },
  {
    "id": "223e4567-e89b-12d3-a456-426614174001",
    "name": "Development n8n",
    "description": "Development workflow automation",
    "status": "stopped",
    "created_at": "2023-06-07T10:24:46Z",
    "updated_at": "2023-06-07T10:24:46Z",
    "memory_limit": 268435456,
    "domain": "dev-n8n.launchstack.io"
  }
]
```

#### GET /instances/:id

Returns details of a specific instance.

**URL Parameters**:
- `:id` - UUID of the instance

**Response**:
```json
{
  "id": "123e4567-e89b-12d3-a456-426614174000",
  "name": "Production n8n",
  "description": "Production workflow automation",
  "status": "running",
  "created_at": "2023-06-08T12:34:56Z",
  "updated_at": "2023-06-08T12:34:56Z",
  "memory_limit": 536870912,
  "domain": "prod-n8n.launchstack.io"
}
```

#### POST /instances

Creates a new instance.

**Request Body**:
```json
{
  "name": "New n8n Instance",
  "description": "My new n8n instance",
  "memory_limit": 536870912
}
```

**Response**:
```json
{
  "id": "323e4567-e89b-12d3-a456-426614174002",
  "name": "New n8n Instance",
  "description": "My new n8n instance",
  "status": "created",
  "created_at": "2023-06-08T12:34:56Z",
  "updated_at": "2023-06-08T12:34:56Z",
  "memory_limit": 536870912,
  "domain": "new-n8n.launchstack.io"
}
```

#### DELETE /instances/:id

Deletes an instance.

**URL Parameters**:
- `:id` - UUID of the instance

**Response**:
```json
{
  "message": "Instance deleted successfully"
}
```

#### POST /instances/:id/start

Starts an instance.

**URL Parameters**:
- `:id` - UUID of the instance

**Response**:
```json
{
  "message": "Instance started successfully"
}
```

#### POST /instances/:id/stop

Stops an instance.

**URL Parameters**:
- `:id` - UUID of the instance

**Response**:
```json
{
  "message": "Instance stopped successfully"
}
```

#### POST /instances/:id/restart

Restarts an instance.

**URL Parameters**:
- `:id` - UUID of the instance

**Response**:
```json
{
  "message": "Instance restarted successfully"
}
```

### Instance Metrics

#### GET /instances/:id/stats

Returns real-time resource usage for an instance.

**URL Parameters**:
- `:id` - UUID of the instance

**Response**:
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

#### GET /instances/:id/stats/history

Returns historical resource usage data for an instance.

**URL Parameters**:
- `:id` - UUID of the instance

**Query Parameters**:
- `period` - Time period to fetch data for (default: "1h")
  - Options: "10m", "1h", "6h", "24h"

**Response**:
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
  {
    "timestamp": "2023-06-08T12:33:56Z",
    "cpu_usage": 22.1,
    "memory_usage": 102400000,
    "memory_limit": 536870912,
    "memory_percentage": 19.1,
    "network_in": 1020000,
    "network_out": 510000
  }
]
```

## Implementation Notes

### Historical Metrics

- The `/instances/:id/stats/history` endpoint uses TimescaleDB to efficiently query time-series data
- Data points are evenly distributed using time bucketing
- Response is limited to 100 data points maximum to prevent excessive data transfer
- Timestamps are in ISO 8601 format (RFC3339)
- Metric fields:
  - `cpu_usage`: CPU usage percentage (0-100)
  - `memory_usage`: Memory usage in bytes
  - `memory_limit`: Memory limit in bytes
  - `memory_percentage`: Memory usage as percentage of limit (0-100)
  - `network_in`: Incoming network traffic in bytes
  - `network_out`: Outgoing network traffic in bytes

### Period Parameter

The `period` parameter determines the time window for historical metrics:

- `10m`: Last 10 minutes (high resolution)
- `1h`: Last hour (default)
- `6h`: Last 6 hours
- `24h`: Last 24 hours (lower resolution)

Each period automatically adjusts the data resolution to provide meaningful visualizations without excessive data points. 