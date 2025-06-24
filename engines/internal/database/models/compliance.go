package models

import (
	"time"
)

// ComplianceCheckType represents the type of compliance check
type ComplianceCheckType string

const (
	ComplianceCheckKYC       ComplianceCheckType = "kyc"
	ComplianceCheckAML       ComplianceCheckType = "aml"
	ComplianceCheckSanctions ComplianceCheckType = "sanctions"
	ComplianceCheckPEP       ComplianceCheckType = "pep"
	ComplianceCheckWatchlist ComplianceCheckType = "watchlist"
)

// ComplianceStatus represents the status of a compliance check
type ComplianceStatus string

const (
	ComplianceStatusPending ComplianceStatus = "pending"
	ComplianceStatusPassed  ComplianceStatus = "passed"
	ComplianceStatusFailed  ComplianceStatus = "failed"
	ComplianceStatusReview  ComplianceStatus = "review"
	ComplianceStatusExpired ComplianceStatus = "expired"
)

// RiskLevel represents the risk level assessment
type RiskLevel string

const (
	RiskLevelLow      RiskLevel = "low"
	RiskLevelMedium   RiskLevel = "medium"
	RiskLevelHigh     RiskLevel = "high"
	RiskLevelCritical RiskLevel = "critical"
)

// ComplianceCheck represents a compliance verification check
type ComplianceCheck struct {
	ID           string                 `json:"id" db:"id"`
	UserID       string                 `json:"user_id" db:"user_id"`
	Type         ComplianceCheckType    `json:"type" db:"type"`
	Status       ComplianceStatus       `json:"status" db:"status"`
	RiskLevel    RiskLevel              `json:"risk_level" db:"risk_level"`
	Provider     string                 `json:"provider" db:"provider"`
	ProviderRef  string                 `json:"provider_ref" db:"provider_ref"`
	RequestData  map[string]interface{} `json:"request_data" db:"request_data"`
	ResponseData map[string]interface{} `json:"response_data" db:"response_data"`
	Score        *float64               `json:"score" db:"score"`
	Reason       string                 `json:"reason" db:"reason"`
	Notes        string                 `json:"notes" db:"notes"`
	Documents    []string               `json:"documents" db:"documents"`
	ExpiresAt    *time.Time             `json:"expires_at" db:"expires_at"`
	CheckedAt    *time.Time             `json:"checked_at" db:"checked_at"`
	ReviewedAt   *time.Time             `json:"reviewed_at" db:"reviewed_at"`
	ReviewedBy   *string                `json:"reviewed_by" db:"reviewed_by"`
	RetryCount   int                    `json:"retry_count" db:"retry_count"`
	NextRetryAt  *time.Time             `json:"next_retry_at" db:"next_retry_at"`
	Metadata     map[string]interface{} `json:"metadata" db:"metadata"`
	CreatedAt    time.Time              `json:"created_at" db:"created_at"`
	UpdatedAt    time.Time              `json:"updated_at" db:"updated_at"`

	// Relations
	User *User `json:"user,omitempty"`
}

// ComplianceRule represents a compliance rule configuration
type ComplianceRule struct {
	ID          string                 `json:"id" db:"id"`
	Name        string                 `json:"name" db:"name"`
	Type        ComplianceCheckType    `json:"type" db:"type"`
	Description string                 `json:"description" db:"description"`
	IsEnabled   bool                   `json:"is_enabled" db:"is_enabled"`
	Priority    int                    `json:"priority" db:"priority"`
	Conditions  map[string]interface{} `json:"conditions" db:"conditions"`
	Actions     map[string]interface{} `json:"actions" db:"actions"`
	CreatedAt   time.Time              `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at" db:"updated_at"`
}

// IsValid checks if compliance check is valid and not expired
func (c *ComplianceCheck) IsValid() bool {
	if c.Status != ComplianceStatusPassed {
		return false
	}

	if c.ExpiresAt != nil && c.ExpiresAt.Before(time.Now()) {
		return false
	}

	return true
}

// IsExpired checks if compliance check has expired
func (c *ComplianceCheck) IsExpired() bool {
	return c.ExpiresAt != nil && c.ExpiresAt.Before(time.Now())
}

// NeedsReview checks if compliance check needs manual review
func (c *ComplianceCheck) NeedsReview() bool {
	return c.Status == ComplianceStatusReview ||
		c.RiskLevel == RiskLevelHigh ||
		c.RiskLevel == RiskLevelCritical
}

// CanRetry checks if compliance check can be retried
func (c *ComplianceCheck) CanRetry() bool {
	return c.Status == ComplianceStatusFailed &&
		c.RetryCount < 3 &&
		(c.NextRetryAt == nil || c.NextRetryAt.Before(time.Now()))
}

// MarkPassed marks compliance check as passed
func (c *ComplianceCheck) MarkPassed(score *float64) {
	c.Status = ComplianceStatusPassed
	c.Score = score
	now := time.Now()
	c.CheckedAt = &now
	c.UpdatedAt = now
}

// MarkFailed marks compliance check as failed
func (c *ComplianceCheck) MarkFailed(reason string) {
	c.Status = ComplianceStatusFailed
	c.Reason = reason
	c.RetryCount++
	now := time.Now()
	c.CheckedAt = &now
	c.UpdatedAt = now

	// Schedule retry in 1 hour
	retryTime := now.Add(time.Hour)
	c.NextRetryAt = &retryTime
}

// MarkForReview marks compliance check for manual review
func (c *ComplianceCheck) MarkForReview(reason string) {
	c.Status = ComplianceStatusReview
	c.Reason = reason
	now := time.Now()
	c.CheckedAt = &now
	c.UpdatedAt = now
}
