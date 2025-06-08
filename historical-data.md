graph TD
    A[Frontend React App] --> B[API Endpoints]
    B --> C[Controller Layer]
    C --> D[Data Access Layer]
    D --> E[TimescaleDB]
    
    subgraph "Frontend Components"
        A1[Instance Dashboard] --> A2[Time Period Selector]
        A1 --> A3[Chart Component]
        A1 --> A4[Metrics Card Display]
    end
    
    subgraph "API Endpoints"
        B1[GET /instances/:id/stats] --> C
        B2[GET /instances/:id/historical-stats] --> C
    end
    
    subgraph "Controller Layer"
        C1[GetInstanceStats] --> D
        C2[GetInstanceHistoricalStats] --> D
    end
    
    subgraph "Data Access Layer"
        D1[GetLatestResourceUsage] --> E
        D2[GetResourceUsageHistorical] --> E
        D3[GetResourceUsageAggregates] --> E
    end
    
    subgraph "TimescaleDB"
        E1[resource_usage table] --> E3[Hypertable]
        E2[resource_usage_hourly] --> E4[Continuous Aggregate]
    end
    
    D2 -- "Short periods (< 3h)" --> E1
    D3 -- "Longer periods (≥ 3h)" --> E2

    -----------------------------------

    sequenceDiagram
    participant F as Frontend
    participant B as Backend API
    participant DB as TimescaleDB
    
    F->>B: GET /api/v1/instances/:id/stats/history?period=1h
    Note over B: Authenticate and authorize request
    B->>DB: Query time-bucketed data<br/>with proper resolution
    Note over DB: TimescaleDB processes<br/>time-series query efficiently
    DB->>B: Return aggregated metrics
    Note over B: Format response with<br/>timestamps in ISO 8601 format
    B->>F: Return array of data points (max 100)
    Note over F: Process and display<br/>metrics in charts
    
    F->>F: User selects different time period (e.g., 24h)
    F->>B: GET /api/v1/instances/:id/stats/history?period=24h
    Note over B: Calculate appropriate<br/>time bucket size
    B->>DB: Query with larger time bucket
    DB->>B: Return data with lower resolution
    B->>F: Return array of data points
    Note over F: Update chart with<br/>new time period data