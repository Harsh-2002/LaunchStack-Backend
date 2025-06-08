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