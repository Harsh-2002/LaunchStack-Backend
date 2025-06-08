package db

import (
	"time"

	"github.com/google/uuid"
	"github.com/launchstack/backend/models"
)

// CreateResourceUsage saves a resource usage record to the database
func CreateResourceUsage(usage *models.ResourceUsage) error {
	result := DB.Create(usage)
	return result.Error
}

// GetResourceUsageByInstanceID retrieves resource usage records for an instance
func GetResourceUsageByInstanceID(instanceID uuid.UUID, limit int) ([]models.ResourceUsage, error) {
	var usages []models.ResourceUsage
	
	// Set a default limit if not specified
	if limit <= 0 {
		limit = 10
	}
	
	result := DB.Where("instance_id = ?", instanceID).
		Order("timestamp DESC").
		Limit(limit).
		Find(&usages)
	
	return usages, result.Error
}

// GetResourceUsageHistorical retrieves historical resource usage with TimescaleDB
func GetResourceUsageHistorical(instanceID uuid.UUID, period time.Duration, resolution string) ([]map[string]interface{}, error) {
	var results []map[string]interface{}
	
	// Calculate time bounds
	endTime := time.Now()
	startTime := endTime.Add(-period)
	
	// Choose time bucket size based on requested resolution and period
	var timeBucket string
	switch resolution {
	case "high":
		timeBucket = "10 seconds"
	case "medium":
		timeBucket = "1 minute"
	case "low":
		if period > 24*time.Hour {
			timeBucket = "1 hour"
		} else if period > 6*time.Hour {
			timeBucket = "10 minutes"
		} else {
			timeBucket = "1 minute"
		}
	default:
		// Automatically select resolution based on period
		if period > 7*24*time.Hour {
			timeBucket = "1 hour"
		} else if period > 24*time.Hour {
			timeBucket = "15 minutes"
		} else if period > 6*time.Hour {
			timeBucket = "5 minutes"
		} else if period > time.Hour {
			timeBucket = "1 minute"
		} else {
			timeBucket = "10 seconds"
		}
	}
	
	// Use time_bucket for proper time-series visualization with even intervals
	query := `
		SELECT 
			time_bucket($1, timestamp) AS time,
			AVG(cpu_usage) AS cpu_avg,
			MAX(cpu_usage) AS cpu_max,
			AVG(memory_usage) AS memory_avg,
			MAX(memory_usage) AS memory_max,
			AVG(memory_usage * 100.0 / NULLIF(memory_limit, 0)) AS memory_percentage_avg,
			SUM(network_in) AS network_in_total,
			SUM(network_out) AS network_out_total,
			COUNT(*) AS sample_count
		FROM resource_usage
		WHERE instance_id = $2 AND timestamp BETWEEN $3 AND $4
		GROUP BY time
		ORDER BY time DESC
		LIMIT 100
	`
	
	// Execute the query
	rows, err := DB.Raw(query, timeBucket, instanceID, startTime, endTime).Rows()
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	
	// Process the results
	for rows.Next() {
		var (
			timeVal            time.Time
			cpuAvg             float64
			cpuMax             float64
			memoryAvg          int64
			memoryMax          int64
			memoryPercentageAvg float64
			networkInTotal     int64
			networkOutTotal    int64
			sampleCount        int
		)
		
		if err := rows.Scan(&timeVal, &cpuAvg, &cpuMax, &memoryAvg, &memoryMax, 
			&memoryPercentageAvg, &networkInTotal, &networkOutTotal, &sampleCount); err != nil {
			return nil, err
		}
		
		// Format values for frontend display
		point := map[string]interface{}{
			"timestamp":         timeVal.Format(time.RFC3339),
			"cpu_avg":           cpuAvg,
			"cpu_max":           cpuMax,
			"memory_avg":        memoryAvg,
			"memory_max":        memoryMax,
			"memory_percentage": memoryPercentageAvg,
			"network_in":        networkInTotal,
			"network_out":       networkOutTotal,
			"sample_count":      sampleCount,
		}
		
		results = append(results, point)
	}
	
	return results, nil
}

// GetLatestResourceUsage retrieves the most recent resource usage record for an instance
func GetLatestResourceUsage(instanceID uuid.UUID) (*models.ResourceUsage, error) {
	var usage models.ResourceUsage
	
	result := DB.Where("instance_id = ?", instanceID).
		Order("timestamp DESC").
		First(&usage)
	
	if result.Error != nil {
		return nil, result.Error
	}
	
	return &usage, nil
}

// GetResourceUsageAggregates returns hourly aggregated stats for a specified time period
func GetResourceUsageAggregates(instanceID uuid.UUID, period time.Duration) ([]map[string]interface{}, error) {
	var results []map[string]interface{}
	
	// Calculate time bounds
	endTime := time.Now()
	startTime := endTime.Add(-period)
	
	// For longer time periods, use the continuous aggregate view for better performance
	if period > 2*time.Hour {
		query := `
			SELECT 
				bucket AS time,
				avg_cpu,
				max_cpu,
				avg_memory,
				max_memory,
				total_network_in,
				total_network_out
			FROM resource_usage_hourly
			WHERE instance_id = $1 AND bucket BETWEEN $2 AND $3
			ORDER BY bucket ASC
		`
		
		rows, err := DB.Raw(query, instanceID, startTime, endTime).Rows()
		if err != nil {
			return nil, err
		}
		defer rows.Close()
		
		for rows.Next() {
			var (
				timeVal        time.Time
				avgCpu         float64
				maxCpu         float64
				avgMemory      int64
				maxMemory      int64
				networkInTotal int64
				networkOutTotal int64
			)
			
			if err := rows.Scan(&timeVal, &avgCpu, &maxCpu, &avgMemory, &maxMemory, 
				&networkInTotal, &networkOutTotal); err != nil {
				return nil, err
			}
			
			point := map[string]interface{}{
				"timestamp":      timeVal,
				"cpu_avg":        avgCpu,
				"cpu_max":        maxCpu,
				"memory_avg":     avgMemory,
				"memory_max":     maxMemory,
				"network_in":     networkInTotal,
				"network_out":    networkOutTotal,
			}
			
			results = append(results, point)
		}
		
		return results, nil
	}
	
	// For shorter periods, compute aggregates on-the-fly with appropriate time buckets
	return GetResourceUsageHistorical(instanceID, period, "low")
}

// PruneResourceUsage deletes old resource usage records to prevent excessive database growth
// This is now handled by TimescaleDB's retention policy
func PruneResourceUsage(maxRecordsPerInstance int) error {
	// No longer needed with TimescaleDB retention policy
	return nil
} 