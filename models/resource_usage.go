package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// ResourceUsage represents resource utilization for an n8n instance
type ResourceUsage struct {
	ID           uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	InstanceID   uuid.UUID `gorm:"type:uuid;index" json:"instance_id"`
	Timestamp    time.Time `json:"timestamp"`
	CPUUsage     float64   `json:"cpu_usage"` // Percentage (0-100)
	MemoryUsage  int64     `json:"memory_usage"` // In bytes
	DiskUsage    int64     `json:"disk_usage"` // In bytes
	NetworkIn    int64     `json:"network_in"` // In bytes
	NetworkOut   int64     `json:"network_out"` // In bytes
	CreatedAt    time.Time `json:"created_at"`
	
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

// formatBytes converts bytes to a human-readable format (KB, MB, GB)
func formatBytes(bytes int64) map[string]interface{} {
	const (
		KB = 1024
		MB = KB * 1024
		GB = MB * 1024
	)
	
	var value float64
	var unit string
	
	switch {
	case bytes >= GB:
		value = float64(bytes) / float64(GB)
		unit = "GB"
	case bytes >= MB:
		value = float64(bytes) / float64(MB)
		unit = "MB"
	case bytes >= KB:
		value = float64(bytes) / float64(KB)
		unit = "KB"
	default:
		value = float64(bytes)
		unit = "bytes"
	}
	
	return map[string]interface{}{
		"value": value,
		"unit":  unit,
		"bytes": bytes,
	}
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