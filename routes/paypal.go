package routes

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/launchstack/backend/config"
	"github.com/launchstack/backend/db"
	"github.com/launchstack/backend/models"
	"github.com/sirupsen/logrus"
)

// PayPalTokenResponse represents the response from PayPal OAuth token endpoint
type PayPalTokenResponse struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	ExpiresIn   int    `json:"expires_in"`
}

// PayPalOrderResponse represents the response from PayPal create order API
type PayPalOrderResponse struct {
	ID     string `json:"id"`
	Status string `json:"status"`
	Links  []struct {
		Href   string `json:"href"`
		Rel    string `json:"rel"`
		Method string `json:"method"`
	} `json:"links"`
}

// PayPalSubscriptionResponse represents the response from PayPal create subscription API
type PayPalSubscriptionResponse struct {
	ID     string `json:"id"`
	Status string `json:"status"`
	Links  []struct {
		Href   string `json:"href"`
		Rel    string `json:"rel"`
		Method string `json:"method"`
	} `json:"links"`
}

// PayPalHandler handles PayPal related operations
type PayPalHandler struct {
	Config *config.Config
	Logger *logrus.Logger
}

// NewPayPalHandler creates a new PayPal handler
func NewPayPalHandler(cfg *config.Config, logger *logrus.Logger) *PayPalHandler {
	return &PayPalHandler{
		Config: cfg,
		Logger: logger,
	}
}

// GetAccessToken gets an access token from PayPal API
func (h *PayPalHandler) GetAccessToken() (string, error) {
	baseURL := "https://api-m.sandbox.paypal.com"
	if h.Config.PayPal.Mode == "production" {
		baseURL = "https://api-m.paypal.com"
	}

	url := fmt.Sprintf("%s/v1/oauth2/token", baseURL)
	auth := base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf("%s:%s", h.Config.PayPal.APIKey, h.Config.PayPal.Secret)))

	req, err := http.NewRequest("POST", url, strings.NewReader("grant_type=client_credentials"))
	if err != nil {
		return "", err
	}

	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Add("Authorization", fmt.Sprintf("Basic %s", auth))

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("failed to get access token: %s", string(body))
	}

	var tokenResp PayPalTokenResponse
	if err := json.NewDecoder(resp.Body).Decode(&tokenResp); err != nil {
		return "", err
	}

	return tokenResp.AccessToken, nil
}

// CreateCheckoutSession creates a PayPal checkout session for subscription
func CreateCheckoutSession(c *gin.Context) {
	// Get user ID from context (set by auth middleware)
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	// Parse request body
	var req struct {
		Plan       string `json:"plan"`
		SuccessURL string `json:"success_url"`
		CancelURL  string `json:"cancel_url"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format"})
		return
	}

	// Validate plan
	if req.Plan != string(models.PlanPro) && req.Plan != string(models.PlanStarter) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid plan selected"})
		return
	}

	// Create PayPal handler
	cfg, _ := config.NewConfig()
	logger := logrus.New()
	handler := NewPayPalHandler(cfg, logger)

	// Get access token
	token, err := handler.GetAccessToken()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to authenticate with PayPal"})
		return
	}

	// Create order
	baseURL := "https://api-m.sandbox.paypal.com"
	if cfg.PayPal.Mode == "production" {
		baseURL = "https://api-m.paypal.com"
	}

	// Determine amount based on plan
	var amount float64
	if req.Plan == string(models.PlanPro) {
		amount = 5.00 // $5 per month for Pro plan
	} else {
		// Starter plan
		amount = 2.00 // $2 per month for Starter plan
	}

	// Create order payload
	orderData := map[string]interface{}{
		"intent": "CAPTURE",
		"purchase_units": []map[string]interface{}{
			{
				"amount": map[string]interface{}{
					"currency_code": "USD",
					"value":         fmt.Sprintf("%.2f", amount),
				},
				"description": fmt.Sprintf("LaunchStack %s Plan Subscription", req.Plan),
			},
		},
		"application_context": map[string]interface{}{
			"return_url": req.SuccessURL,
			"cancel_url": req.CancelURL,
		},
	}

	orderJSON, _ := json.Marshal(orderData)
	orderReq, err := http.NewRequest("POST", fmt.Sprintf("%s/v2/checkout/orders", baseURL), bytes.NewBuffer(orderJSON))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create PayPal order request"})
		return
	}

	orderReq.Header.Add("Content-Type", "application/json")
	orderReq.Header.Add("Authorization", fmt.Sprintf("Bearer %s", token))

	// Add user ID to PayPal request for webhook correlation
	orderReq.Header.Add("PayPal-Request-Id", userID.(uuid.UUID).String())

	// Execute order request
	client := &http.Client{}
	resp, err := client.Do(orderReq)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to communicate with PayPal"})
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("PayPal error: %s", string(body))})
		return
	}

	// Parse response
	var orderResp PayPalOrderResponse
	if err := json.NewDecoder(resp.Body).Decode(&orderResp); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to parse PayPal response"})
		return
	}

	// Find approval URL
	var checkoutURL string
	for _, link := range orderResp.Links {
		if link.Rel == "approve" {
			checkoutURL = link.Href
			break
		}
	}

	if checkoutURL == "" {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "No checkout URL found in PayPal response"})
		return
	}

	// Create payment record in pending state
	payment := models.Payment{
		UserID:        userID.(uuid.UUID),
		PayPalOrderID: orderResp.ID,
		Amount:        int(amount * 100), // Convert to cents
		Currency:      "usd",
		Status:        models.PaymentStatusPending,
		Description:   fmt.Sprintf("Subscription to %s plan", req.Plan),
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}

	if err := db.DB.Create(&payment).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to record payment"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"checkout_url": checkoutURL,
		"order_id":     orderResp.ID,
	})
}

// GetPayments gets payment history for the current user
func GetPayments(c *gin.Context) {
	// Get user ID from context (set by auth middleware)
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	// Get payment history from database
	var payments []models.Payment
	if err := db.DB.Where("user_id = ?", userID).Order("created_at DESC").Find(&payments).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch payment history"})
		return
	}

	// Convert to public response format
	response := make([]map[string]interface{}, len(payments))
	for i, payment := range payments {
		response[i] = payment.ToPublicResponse()
	}

	c.JSON(http.StatusOK, response)
}

// GetSubscriptions gets subscription details for the current user
func GetSubscriptions(c *gin.Context) {
	// Get user ID from context (set by auth middleware)
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	// Find user in database
	var user models.User
	if err := db.DB.Where("id = ?", userID).First(&user).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	// Check if user has an active subscription
	if user.SubscriptionID == "" {
		c.JSON(http.StatusOK, gin.H{
			"status": "no_subscription",
			"plan":   user.Plan,
		})
		return
	}

	// Return subscription details
	c.JSON(http.StatusOK, gin.H{
		"id":                  user.SubscriptionID,
		"plan":                user.Plan,
		"status":              user.SubscriptionStatus,
		"current_period_end":  user.CurrentPeriodEnd,
		"cancel_at_period_end": user.SubscriptionStatus == models.StatusCanceled,
	})
}

// CancelSubscription cancels the user's subscription
func CancelSubscription(c *gin.Context) {
	// Get user ID from context (set by auth middleware)
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	// Get subscription ID from URL
	subscriptionID := c.Param("id")
	if subscriptionID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Subscription ID is required"})
		return
	}

	// Find user in database
	var user models.User
	if err := db.DB.Where("id = ?", userID).First(&user).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	// Verify that subscription belongs to user
	if user.SubscriptionID != subscriptionID {
		c.JSON(http.StatusForbidden, gin.H{"error": "Not authorized to cancel this subscription"})
		return
	}

	// Create PayPal handler
	cfg, _ := config.NewConfig()
	logger := logrus.New()
	handler := NewPayPalHandler(cfg, logger)

	// Get access token
	token, err := handler.GetAccessToken()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to authenticate with PayPal"})
		return
	}

	// Cancel subscription with PayPal
	baseURL := "https://api-m.sandbox.paypal.com"
	if cfg.PayPal.Mode == "production" {
		baseURL = "https://api-m.paypal.com"
	}

	cancelReq, err := http.NewRequest("POST", fmt.Sprintf("%s/v1/billing/subscriptions/%s/cancel", baseURL, subscriptionID), strings.NewReader(`{"reason": "Customer requested cancellation"}`))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create cancellation request"})
		return
	}

	cancelReq.Header.Add("Content-Type", "application/json")
	cancelReq.Header.Add("Authorization", fmt.Sprintf("Bearer %s", token))

	// Execute cancellation request
	client := &http.Client{}
	resp, err := client.Do(cancelReq)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to communicate with PayPal"})
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent {
		body, _ := io.ReadAll(resp.Body)
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("PayPal error: %s", string(body))})
		return
	}

	// Update user subscription status
	user.SubscriptionStatus = models.StatusCanceled
	user.UpdatedAt = time.Now()

	if err := db.DB.Save(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update subscription status"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "Subscription will be canceled at the end of the current billing period",
	})
}

// PayPalWebhook handles webhook events from PayPal
func PayPalWebhook(c *gin.Context) {
	// Read request body
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to read request body"})
		return
	}

	// Parse event data
	var event map[string]interface{}
	if err := json.Unmarshal(body, &event); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON payload"})
		return
	}

	// Get event type
	eventType, ok := event["event_type"].(string)
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Missing event type"})
		return
	}

	// Get logger from context
	logger, exists := c.Get("logger")
	if !exists {
		logger = logrus.New()
	}

	// Handle different event types
	switch eventType {
	case "PAYMENT.CAPTURE.COMPLETED":
		handlePaymentCaptureCompleted(c, event, logger.(*logrus.Logger))
	case "BILLING.SUBSCRIPTION.CREATED":
		handleSubscriptionCreated(c, event, logger.(*logrus.Logger))
	case "BILLING.SUBSCRIPTION.UPDATED":
		handleSubscriptionUpdated(c, event, logger.(*logrus.Logger))
	case "BILLING.SUBSCRIPTION.CANCELLED":
		handleSubscriptionCancelled(c, event, logger.(*logrus.Logger))
	default:
		// Acknowledge receipt of the webhook but take no action
		logger.(*logrus.Logger).WithField("type", eventType).Info("Received unhandled PayPal event type")
		c.JSON(http.StatusOK, gin.H{"status": "acknowledged"})
	}
}

// handlePaymentCaptureCompleted handles the PAYMENT.CAPTURE.COMPLETED event from PayPal
func handlePaymentCaptureCompleted(c *gin.Context, event map[string]interface{}, logger *logrus.Logger) {
	// Extract data from the event
	resource, ok := event["resource"].(map[string]interface{})
	if !ok {
		logger.Error("Missing resource in PayPal event")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid event format"})
		return
	}

	// Get payment details
	paymentID, _ := resource["id"].(string)
	orderID, _ := resource["parent_payment"].(string)
	status, _ := resource["status"].(string)

	if paymentID == "" || status != "COMPLETED" {
		logger.Error("Missing payment ID or status not completed")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid payment data"})
		return
	}

	// Find the payment record
	var payment models.Payment
	if err := db.DB.Where("paypal_order_id = ?", orderID).First(&payment).Error; err != nil {
		logger.WithError(err).Error("Failed to find payment record")
		c.JSON(http.StatusNotFound, gin.H{"error": "Payment record not found"})
		return
	}

	// Update payment status and ID
	payment.Status = models.PaymentStatusSucceeded
	payment.PayPalPaymentID = paymentID
	payment.UpdatedAt = time.Now()

	if err := db.DB.Save(&payment).Error; err != nil {
		logger.WithError(err).Error("Failed to update payment record")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update payment"})
		return
	}

	logger.WithFields(logrus.Fields{
		"payment_id": payment.ID,
		"paypal_payment_id": paymentID,
	}).Info("Payment completed successfully")
	c.JSON(http.StatusOK, gin.H{"status": "success"})
}

// handleSubscriptionCreated handles the BILLING.SUBSCRIPTION.CREATED event from PayPal
func handleSubscriptionCreated(c *gin.Context, event map[string]interface{}, logger *logrus.Logger) {
	// Extract data from the event
	resource, ok := event["resource"].(map[string]interface{})
	if !ok {
		logger.Error("Missing resource in PayPal event")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid event format"})
		return
	}

	// Get subscription details
	subscriptionID, _ := resource["id"].(string)
	status, _ := resource["status"].(string)
	customID, _ := resource["custom_id"].(string)

	if subscriptionID == "" || status == "" {
		logger.Error("Missing subscription details")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid subscription data"})
		return
	}

	// Parse user ID from custom ID
	userID, err := uuid.Parse(customID)
	if err != nil {
		logger.WithError(err).Error("Invalid user ID format")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	// Find user in database
	var user models.User
	if err := db.DB.Where("id = ?", userID).First(&user).Error; err != nil {
		logger.WithError(err).Error("Failed to find user")
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	// Update user subscription details
	user.SubscriptionID = subscriptionID
	user.SubscriptionStatus = models.SubscriptionStatus(status)
	user.CurrentPeriodEnd = time.Now().AddDate(0, 1, 0) // Assuming monthly subscription
	user.UpdatedAt = time.Now()

	if err := db.DB.Save(&user).Error; err != nil {
		logger.WithError(err).Error("Failed to update user subscription")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update subscription"})
		return
	}

	logger.WithFields(logrus.Fields{
		"user_id": user.ID,
		"subscription_id": subscriptionID,
	}).Info("Subscription created successfully")
	c.JSON(http.StatusOK, gin.H{"status": "success"})
}

// handleSubscriptionUpdated handles the BILLING.SUBSCRIPTION.UPDATED event from PayPal
func handleSubscriptionUpdated(c *gin.Context, event map[string]interface{}, logger *logrus.Logger) {
	// Extract data from the event
	resource, ok := event["resource"].(map[string]interface{})
	if !ok {
		logger.Error("Missing resource in PayPal event")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid event format"})
		return
	}

	// Get subscription details
	subscriptionID, _ := resource["id"].(string)
	status, _ := resource["status"].(string)

	if subscriptionID == "" {
		logger.Error("Missing subscription ID")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Missing subscription ID"})
		return
	}

	// Find user by subscription ID
	var user models.User
	if err := db.DB.Where("subscription_id = ?", subscriptionID).First(&user).Error; err != nil {
		logger.WithError(err).Error("Failed to find user by subscription ID")
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	// Update user subscription status
	user.SubscriptionStatus = models.SubscriptionStatus(status)
	user.UpdatedAt = time.Now()

	if err := db.DB.Save(&user).Error; err != nil {
		logger.WithError(err).Error("Failed to update user subscription")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update subscription"})
		return
	}

	logger.WithFields(logrus.Fields{
		"user_id": user.ID,
		"subscription_id": subscriptionID,
		"status": status,
	}).Info("Subscription updated successfully")
	c.JSON(http.StatusOK, gin.H{"status": "success"})
}

// handleSubscriptionCancelled handles the BILLING.SUBSCRIPTION.CANCELLED event from PayPal
func handleSubscriptionCancelled(c *gin.Context, event map[string]interface{}, logger *logrus.Logger) {
	// Extract data from the event
	resource, ok := event["resource"].(map[string]interface{})
	if !ok {
		logger.Error("Missing resource in PayPal event")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid event format"})
		return
	}

	// Get subscription details
	subscriptionID, _ := resource["id"].(string)

	if subscriptionID == "" {
		logger.Error("Missing subscription ID")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Missing subscription ID"})
		return
	}

	// Find user by subscription ID
	var user models.User
	if err := db.DB.Where("subscription_id = ?", subscriptionID).First(&user).Error; err != nil {
		logger.WithError(err).Error("Failed to find user by subscription ID")
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	// Update user subscription status
	user.SubscriptionStatus = models.StatusCanceled
	user.UpdatedAt = time.Now()

	if err := db.DB.Save(&user).Error; err != nil {
		logger.WithError(err).Error("Failed to update user subscription status")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update subscription status"})
		return
	}

	logger.WithFields(logrus.Fields{
		"user_id": user.ID,
		"subscription_id": subscriptionID,
	}).Info("Subscription cancelled successfully")
	c.JSON(http.StatusOK, gin.H{"status": "success"})
} 