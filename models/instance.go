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

// Instance represents an n8n instance
type Instance struct {
	ID              uuid.UUID      `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	UserID          uuid.UUID      `gorm:"type:uuid;index" json:"user_id"`
	Name            string         `json:"name"`
	Description     string         `json:"description"`
	ContainerID     string         `json:"container_id,omitempty"`
	Status          InstanceStatus `gorm:"type:varchar(20);default:'pending'" json:"status"`
	Host            string         `json:"host,omitempty"`
	Port            int            `json:"port"`
	URL             string         `json:"url"`
	CPULimit        float64        `json:"cpu_limit"`
	MemoryLimit     int            `json:"memory_limit"` // In MB
	StorageLimit    int            `json:"storage_limit"` // In GB
	CreatedAt       time.Time      `json:"created_at"`
	UpdatedAt       time.Time      `json:"updated_at"`
	DeletedAt       gorm.DeletedAt `gorm:"index" json:"-"`
	
	// Relationships
	User            User           `gorm:"foreignKey:UserID" json:"-"`
	ResourceUsages  []ResourceUsage `gorm:"foreignKey:InstanceID" json:"resource_usages,omitempty"`
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