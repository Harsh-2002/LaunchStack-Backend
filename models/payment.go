package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// PaymentStatus defines the status of a payment
type PaymentStatus string

const (
	PaymentStatusPending   PaymentStatus = "pending"
	PaymentStatusSucceeded PaymentStatus = "succeeded"
	PaymentStatusFailed    PaymentStatus = "failed"
	PaymentStatusRefunded  PaymentStatus = "refunded"
)

// BillingPeriod defines the billing period of a subscription
type BillingPeriod string

const (
	BillingMonthly BillingPeriod = "monthly"
	BillingYearly  BillingPeriod = "yearly"
)

// Payment represents a payment record
type Payment struct {
	ID              uuid.UUID     `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	UserID          uuid.UUID     `gorm:"type:uuid;index" json:"user_id"`
	Amount          int           `json:"amount"` // In cents
	Currency        string        `gorm:"type:varchar(3);default:'usd'" json:"currency"`
	Status          PaymentStatus `gorm:"type:varchar(20)" json:"status"`
	PayPalPaymentID string        `json:"paypal_payment_id,omitempty"`
	PayPalOrderID   string        `json:"paypal_order_id,omitempty"`
	InvoiceURL      string        `json:"invoice_url,omitempty"`
	Description     string        `json:"description"`
	Metadata        string        `gorm:"type:jsonb" json:"metadata,omitempty"`
	CreatedAt       time.Time     `json:"created_at"`
	UpdatedAt       time.Time     `json:"updated_at"`
	
	// Relationships
	User            User         `gorm:"foreignKey:UserID" json:"-"`
}

// TableName sets the table name for the Payment model
func (Payment) TableName() string {
	return "payments"
}

// BeforeCreate hook is called before creating a new payment
func (p *Payment) BeforeCreate(tx *gorm.DB) error {
	if p.ID == uuid.Nil {
		p.ID = uuid.New()
	}
	return nil
}

// ToPublicResponse returns a public representation of the payment for API responses
func (p *Payment) ToPublicResponse() map[string]interface{} {
	return map[string]interface{}{
		"id":           p.ID,
		"amount":       float64(p.Amount) / 100, // Convert cents to dollars
		"currency":     p.Currency,
		"status":       p.Status,
		"description":  p.Description,
		"invoice_url":  p.InvoiceURL,
		"created_at":   p.CreatedAt,
	}
}

// GetPlanPrice returns the price for a plan and billing period
func GetPlanPrice(plan SubscriptionPlan, billingPeriod BillingPeriod) float64 {
	// Monthly prices
	prices := map[SubscriptionPlan]float64{
		PlanStarter: 2.0,
		PlanPro:     29.0,
	}

	// Apply yearly discount (2 months free)
	if billingPeriod == BillingYearly {
		return prices[plan] * 10 // 12 months - 2 months free
	}

	return prices[plan]
}

// NewPayment creates a new payment record
func NewPayment(userID uuid.UUID, paypalPaymentID string, plan SubscriptionPlan, billingPeriod BillingPeriod) *Payment {
	amount := GetPlanPrice(plan, billingPeriod)
	
	return &Payment{
		UserID:         userID,
		PayPalPaymentID: paypalPaymentID,
		Amount:         int(amount * 100), // Convert dollars to cents
		Currency:       "usd",
		Status:         PaymentStatusPending,
		Description:    "Subscription payment",
	}
}

// CompletePayment marks a payment as completed
func (p *Payment) CompletePayment() {
	p.Status = PaymentStatusSucceeded
}

// FailPayment marks a payment as failed
func (p *Payment) FailPayment() {
	p.Status = PaymentStatusFailed
}

// RefundPayment marks a payment as refunded
func (p *Payment) RefundPayment() {
	p.Status = PaymentStatusRefunded
} 