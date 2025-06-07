package models

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// ResourceUsage represents the resource usage of an instance
type ResourceUsage struct {
	ID              uuid.UUID      `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	InstanceID      uuid.UUID      `json:"instance_id" gorm:"type:uuid;index"`
	Timestamp       time.Time      `json:"timestamp"`
	CPUUsage        float64        `json:"cpu_usage"`        // CPU usage percentage
	MemoryUsage     int64          `json:"memory_usage"`     // Memory usage in bytes
	MemoryLimit     int64          `json:"memory_limit"`     // Memory limit in bytes
	MemoryPercentage float64       `json:"memory_percentage"` // Memory usage percentage
	DiskUsage       int64          `json:"disk_usage"`       // Disk usage in bytes
	NetworkIn       int64          `json:"network_in"`       // Network traffic in (bytes)
	NetworkOut      int64          `json:"network_out"`      // Network traffic out (bytes)
	CreatedAt       time.Time      `json:"created_at"`
	UpdatedAt       time.Time      `json:"updated_at"`
	DeletedAt       gorm.DeletedAt `json:"-" gorm:"index"`
	
	// Relationships
	Instance     Instance  `gorm:"foreignKey:InstanceID" json:"-"`
}

// TableName sets the table name for the ResourceUsage model
func (ResourceUsage) TableName() string {
	return "resource_usages"
}

// BeforeCreate hook is called before creating a new resource usage record
func (r *ResourceUsage) BeforeCreate(tx *gorm.DB) error {
	if r.ID == uuid.Nil {
		r.ID = uuid.New()
	}
	if r.Timestamp.IsZero() {
		r.Timestamp = time.Now()
	}
	return nil
}

// ToPublicResponse returns a public representation of the resource usage for API responses
func (r *ResourceUsage) ToPublicResponse() map[string]interface{} {
	return map[string]interface{}{
		"timestamp":     r.Timestamp,
		"cpu_usage":     r.CPUUsage,
		"memory_usage":  formatBytes(r.MemoryUsage),
		"disk_usage":    formatBytes(r.DiskUsage),
		"network_in":    formatBytes(r.NetworkIn),
		"network_out":   formatBytes(r.NetworkOut),
	}
}

// FormatStats returns formatted resource usage data
func (r *ResourceUsage) FormatStats() map[string]interface{} {
	return map[string]interface{}{
		"id":                r.ID,
		"instance_id":       r.InstanceID,
		"timestamp":         r.Timestamp,
		"cpu_usage":         r.CPUUsage,
		"cpu_formatted":     formatPercentage(r.CPUUsage),
		"memory_usage":      r.MemoryUsage,
		"memory_limit":      r.MemoryLimit,
		"memory_percentage": r.MemoryPercentage,
		"memory_formatted":  formatBytes(r.MemoryUsage) + " / " + formatBytes(r.MemoryLimit) + " (" + formatPercentage(r.MemoryPercentage) + ")",
		"disk_usage":        r.DiskUsage,
		"disk_formatted":    formatBytes(r.DiskUsage),
		"network_in":        r.NetworkIn,
		"network_out":       r.NetworkOut,
		"network_formatted": formatBytes(r.NetworkIn) + " in / " + formatBytes(r.NetworkOut) + " out",
	}
}

// Helper functions for formatting
func formatPercentage(value float64) string {
	if value < 0.01 {
		return "0.00%"
	}
	if value < 10 {
		return fmt.Sprintf("%.2f%%", value)
	}
	return fmt.Sprintf("%.1f%%", value)
}

// formatBytes formats bytes to a human-readable format
func formatBytes(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}

// NewResourceUsage creates a new ResourceUsage record
func NewResourceUsage(instanceID uuid.UUID, cpuUsage float64, memoryUsage, diskUsage, networkIn, networkOut int64) *ResourceUsage {
	return &ResourceUsage{
		InstanceID:  instanceID,
		CPUUsage:    cpuUsage,
		MemoryUsage: memoryUsage,
		DiskUsage:   diskUsage,
		NetworkIn:   networkIn,
		NetworkOut:  networkOut,
		Timestamp:   time.Now(),
	}
}

// GetResourceSummary calculates summary statistics for resource usage
func GetResourceSummary(db *gorm.DB, instanceID uuid.UUID, period time.Duration) (map[string]interface{}, error) {
	var usages []ResourceUsage
	since := time.Now().Add(-period)

	if err := db.Where("instance_id = ? AND timestamp > ?", instanceID, since).
		Order("timestamp asc").
		Find(&usages).Error; err != nil {
		return nil, err
	}

	if len(usages) == 0 {
		return map[string]interface{}{
			"avg_cpu":      0.0,
			"max_cpu":      0.0,
			"avg_memory":   0,
			"max_memory":   0,
			"avg_disk":     0,
			"total_network": 0,
		}, nil
	}

	// Calculate summary statistics
	var totalCPU float64
	var maxCPU float64
	var totalMemory int64
	var maxMemory int64
	var totalDisk int64
	var maxDisk int64
	var totalNetworkIn int64
	var totalNetworkOut int64

	for _, usage := range usages {
		totalCPU += usage.CPUUsage
		if usage.CPUUsage > maxCPU {
			maxCPU = usage.CPUUsage
		}

		totalMemory += usage.MemoryUsage
		if usage.MemoryUsage > int64(maxMemory) {
			maxMemory = usage.MemoryUsage
		}

		totalDisk += usage.DiskUsage
		if usage.DiskUsage > maxDisk {
			maxDisk = usage.DiskUsage
		}

		totalNetworkIn += usage.NetworkIn
		totalNetworkOut += usage.NetworkOut
	}

	count := int64(len(usages))
	avgCPU := totalCPU / float64(count)
	avgMemory := totalMemory / count
	avgDisk := totalDisk / count
	totalNetwork := totalNetworkIn + totalNetworkOut

	return map[string]interface{}{
		"avg_cpu":       avgCPU,
		"max_cpu":       maxCPU,
		"avg_memory":    avgMemory,
		"max_memory":    maxMemory,
		"avg_disk":      avgDisk,
		"max_disk":      maxDisk,
		"total_network": totalNetwork,
	}, nil
} 