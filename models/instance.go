package models

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// ServiceType defines the type of service an instance is running
type ServiceType string

const (
	ServiceN8N ServiceType = "n8n"
)

// InstanceStatus defines the status of an n8n instance
type InstanceStatus string

const (
	StatusRunning InstanceStatus = "running"
	StatusStopped InstanceStatus = "stopped"
	StatusError   InstanceStatus = "error"
	StatusPending InstanceStatus = "pending"
	StatusDeleted InstanceStatus = "deleted"
)

// Instance represents a user's n8n instance
type Instance struct {
	ID            uuid.UUID       `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	UserID        uuid.UUID       `gorm:"type:uuid" json:"user_id"`
	Name          string          `gorm:"size:255;not null" json:"name"`
	Description   string          `gorm:"size:1000" json:"description"`
	Status        InstanceStatus  `gorm:"size:50;not null" json:"status"`
	Host          string          `gorm:"size:255" json:"host"`
	Port          int             `json:"port"`
	URL           string          `gorm:"size:255" json:"url"`
	CPULimit      float64         `json:"cpu_limit"`
	MemoryLimit   int             `json:"memory_limit"` // in MB
	StorageLimit  int             `json:"storage_limit"` // in GB
	ContainerID   string          `gorm:"size:255" json:"container_id"`
	IPAddress     string          `gorm:"size:50" json:"ip_address"`
	CreatedAt     time.Time       `json:"created_at"`
	UpdatedAt     time.Time       `json:"updated_at"`
	DeletedAt     gorm.DeletedAt  `gorm:"index" json:"-"`
	
	// Relationships
	User          User            `gorm:"foreignKey:UserID" json:"-"`
	ResourceUsage []ResourceUsage `gorm:"foreignKey:InstanceID" json:"resource_usage,omitempty"`
}

// TableName sets the table name for the Instance model
func (Instance) TableName() string {
	return "instances"
}

// BeforeCreate hook is called before creating a new instance
func (i *Instance) BeforeCreate(tx *gorm.DB) error {
	if i.ID == uuid.Nil {
		i.ID = uuid.New()
	}
	return nil
}

// ToPublicResponse returns a public representation of the instance for API responses
func (i *Instance) ToPublicResponse() map[string]interface{} {
	return map[string]interface{}{
		"id":           i.ID,
		"name":         i.Name,
		"description":  i.Description,
		"status":       i.Status,
		"url":          i.URL,
		"cpu_limit":    i.CPULimit,
		"memory_limit": i.MemoryLimit,
		"storage_limit": i.StorageLimit,
		"created_at":   i.CreatedAt,
		"updated_at":   i.UpdatedAt,
	}
}

// GetURL returns the full URL to access the instance
func (i *Instance) GetURL(domain string) string {
	if i.URL == "" {
		return ""
	}
	return fmt.Sprintf("https://%s.%s", i.URL, domain)
}

// GetDockerName returns the container name for Docker
func (i *Instance) GetDockerName() string {
	return fmt.Sprintf("n8n-%s", i.ID.String())
}

// GetDataVolumeName returns the volume name for Docker
func (i *Instance) GetDataVolumeName() string {
	return fmt.Sprintf("n8n-data-%s", i.ID.String())
}

// CanStart checks if the instance can be started
func (i *Instance) CanStart() bool {
	return i.Status == StatusStopped
}

// CanStop checks if the instance can be stopped
func (i *Instance) CanStop() bool {
	return i.Status == StatusRunning
}

// CanDelete checks if the instance can be deleted
func (i *Instance) CanDelete() bool {
	return i.Status != StatusDeleted
}

// GetResourceStatus returns the latest resource usage data
func (i *Instance) GetResourceStatus(db *gorm.DB) (*ResourceUsage, error) {
	var usage ResourceUsage
	if err := db.Where("instance_id = ?", i.ID).Order("timestamp desc").First(&usage).Error; err != nil {
		return nil, err
	}
	return &usage, nil
}

// GetUptime returns the instance uptime based on the last status change
func (i *Instance) GetUptime() string {
	if i.Status != StatusRunning {
		return "0m"
	}

	// For demo, we'll use the time since creation
	// In a real implementation, we'd track the last start time
	uptime := time.Since(i.CreatedAt)
	
	days := int(uptime.Hours() / 24)
	hours := int(uptime.Hours()) % 24
	minutes := int(uptime.Minutes()) % 60
	
	if days > 0 {
		return fmt.Sprintf("%dd %dh %dm", days, hours, minutes)
	} else if hours > 0 {
		return fmt.Sprintf("%dh %dm", hours, minutes)
	}
	return fmt.Sprintf("%dm", minutes)
}

// ToDetailedResponse returns a detailed representation of the instance for API responses
func (i *Instance) ToDetailedResponse(domain string, resourceUsage *ResourceUsage) map[string]interface{} {
	response := map[string]interface{}{
		"id":           i.ID,
		"name":         i.Name,
		"description":  i.Description,
		"status":       i.Status,
		"url":          i.GetURL(domain),
		"cpu_limit":    i.CPULimit,
		"memory_limit": i.MemoryLimit,
		"storage_limit": i.StorageLimit,
		"created_at":   i.CreatedAt,
		"updated_at":   i.UpdatedAt,
	}

	if resourceUsage != nil {
		response["current_usage"] = map[string]interface{}{
			"cpu_usage":     resourceUsage.CPUUsage,
			"memory_usage":  resourceUsage.MemoryUsage,
			"disk_usage":    resourceUsage.DiskUsage,
			"uptime":        i.GetUptime(),
		}
	}

	return response
} 