package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// SubscriptionPlan defines the available subscription plans
type SubscriptionPlan string

const (
	PlanFree     SubscriptionPlan = "free"
	PlanStarter  SubscriptionPlan = "starter" // Same as free but with trial
	PlanPro      SubscriptionPlan = "pro"
)

// SubscriptionStatus defines the user's subscription status
type SubscriptionStatus string

const (
	StatusTrial     SubscriptionStatus = "trial"
	StatusActive    SubscriptionStatus = "active"
	StatusCanceled  SubscriptionStatus = "canceled"
	StatusExpired   SubscriptionStatus = "expired"
)

// User represents a user in the system
type User struct {
	ID            uuid.UUID       `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	ClerkUserID   string          `gorm:"uniqueIndex" json:"clerk_user_id"`
	Email         string          `gorm:"uniqueIndex" json:"email"`
	Username      string          `gorm:"uniqueIndex" json:"username"`
	PasswordHash  string          `json:"-"` // Store hashed password, never expose in JSON
	FirstName     string          `json:"first_name"`
	LastName      string          `json:"last_name"`
	Plan          SubscriptionPlan `gorm:"type:varchar(20);default:'free'" json:"plan"`
	PayPalCustomerID string       `json:"paypal_customer_id,omitempty"`
	SubscriptionID   string       `json:"subscription_id,omitempty"`
	SubscriptionStatus SubscriptionStatus `gorm:"type:varchar(50)" json:"subscription_status,omitempty"`
	CurrentPeriodEnd time.Time    `json:"current_period_end,omitempty"`
	CreatedAt     time.Time       `json:"created_at"`
	UpdatedAt     time.Time       `json:"updated_at"`
	DeletedAt     gorm.DeletedAt  `gorm:"index" json:"-"`
	
	// Relationships
	Instances        []Instance         `gorm:"foreignKey:UserID" json:"instances,omitempty"`
	Payments         []Payment          `gorm:"foreignKey:UserID" json:"payments,omitempty"`
}

// TableName sets the table name for the User model
func (User) TableName() string {
	return "users"
}

// BeforeCreate hook is called before creating a new user
func (u *User) BeforeCreate(tx *gorm.DB) error {
	if u.ID == uuid.Nil {
		u.ID = uuid.New()
	}
	return nil
}

// GetPlanResourceLimits returns the resource limits for the user's plan
func (u *User) GetPlanResourceLimits() map[string]interface{} {
	limits := make(map[string]interface{})
	
	switch u.Plan {
	case PlanFree, PlanStarter:
		limits["max_instances"] = 1
		limits["cpu_limit"] = 0.5
		limits["memory_limit"] = 512 // MB
		limits["storage_limit"] = 1  // GB
	case PlanPro:
		limits["max_instances"] = 10
		limits["cpu_limit"] = 1.0
		limits["memory_limit"] = 1024 // MB
		limits["storage_limit"] = 20  // GB
	default:
		// Default to free plan limits
		limits["max_instances"] = 1
		limits["cpu_limit"] = 0.5
		limits["memory_limit"] = 512 // MB
		limits["storage_limit"] = 1  // GB
	}
	
	return limits
}

// GetInstancesLimit returns the maximum number of instances a user can create based on their plan
func (u *User) GetInstancesLimit() int {
	// For testing purposes
	if u.Plan == "unlimited" {
		return 100 // Allow many instances for testing
	}

	switch u.Plan {
	case PlanFree, PlanStarter:
		return 1
	case PlanPro:
		return 10
	default:
		return 1 // Default to free plan
	}
}

// GetCPULimit returns the CPU limit per instance based on subscription plan
func (u *User) GetCPULimit() float64 {
	switch u.Plan {
	case PlanFree, PlanStarter:
		return 0.5
	case PlanPro:
		return 1.0
	default:
		return 0.5 // Default to free plan
	}
}

// GetMemoryLimit returns the memory limit per instance in MB based on subscription plan
func (u *User) GetMemoryLimit() int {
	switch u.Plan {
	case PlanFree, PlanStarter:
		return 512
	case PlanPro:
		return 1024
	default:
		return 512 // Default to free plan
	}
}

// GetStorageLimit returns the storage limit per instance in GB based on subscription plan
func (u *User) GetStorageLimit() int {
	switch u.Plan {
	case PlanFree, PlanStarter:
		return 1
	case PlanPro:
		return 20
	default:
		return 1 // Default to free plan
	}
}

// IsTrialActive checks if the user's trial is active
func (u *User) IsTrialActive() bool {
	if u.CurrentPeriodEnd.IsZero() {
		return false
	}
	return u.SubscriptionStatus == StatusTrial && time.Now().Before(u.CurrentPeriodEnd)
}

// TrialDaysLeft returns the number of days left in the trial
func (u *User) TrialDaysLeft() int {
	if u.CurrentPeriodEnd.IsZero() || u.SubscriptionStatus != StatusTrial {
		return 0
	}
	
	daysLeft := int(time.Until(u.CurrentPeriodEnd).Hours() / 24)
	if daysLeft < 0 {
		return 0
	}
	return daysLeft
}

// StartTrial starts the user's trial period
func (u *User) StartTrial() {
	now := time.Now()
	trialDays := 7 // Both plans have 7-day trial period
	endDate := now.AddDate(0, 0, trialDays)
	
	u.CurrentPeriodEnd = endDate
	u.SubscriptionStatus = StatusTrial
}

// ToPublicResponse returns a public representation of the user for API responses
func (u *User) ToPublicResponse() map[string]interface{} {
	return map[string]interface{}{
		"id":                 u.ID,
		"email":              u.Email,
		"first_name":         u.FirstName,
		"last_name":          u.LastName,
		"plan":               u.Plan,
		"subscription_status": u.SubscriptionStatus,
		"current_period_end": u.CurrentPeriodEnd,
		"instances_limit":    u.GetInstancesLimit(),
	}
} 