package models

import (
	"time"

	"github.com/GoSec-Labs/mMAD/engines/pkg/math"
)

// ReserveType represents the type of reserve
type ReserveType string

const (
	ReserveTypeBankAccount ReserveType = "bank_account"
	ReserveTypeCrypto      ReserveType = "crypto"
	ReserveTypeCommodity   ReserveType = "commodity"
	ReserveTypeSecurities  ReserveType = "securities"
)

// ReserveStatus represents reserve monitoring status
type ReserveStatus string

const (
	ReserveStatusActive      ReserveStatus = "active"
	ReserveStatusInactive    ReserveStatus = "inactive"
	ReserveStatusUnavailable ReserveStatus = "unavailable"
	ReserveStatusError       ReserveStatus = "error"
)

// Reserve represents a reserve account being monitored
type Reserve struct {
	ID             string                 `json:"id" db:"id"`
	Name           string                 `json:"name" db:"name"`
	Type           ReserveType            `json:"type" db:"type"`
	Currency       string                 `json:"currency" db:"currency"`
	AccountNumber  string                 `json:"account_number" db:"account_number"`
	BankName       string                 `json:"bank_name" db:"bank_name"`
	BankCode       string                 `json:"bank_code" db:"bank_code"`
	APIEndpoint    string                 `json:"api_endpoint" db:"api_endpoint"`
	APICredentials map[string]string      `json:"-" db:"api_credentials"` // Encrypted
	Status         ReserveStatus          `json:"status" db:"status"`
	CurrentBalance *math.Decimal          `json:"current_balance" db:"current_balance"`
	LastBalance    *math.Decimal          `json:"last_balance" db:"last_balance"`
	MinThreshold   *math.Decimal          `json:"min_threshold" db:"min_threshold"`
	AlertThreshold *math.Decimal          `json:"alert_threshold" db:"alert_threshold"`
	MaxThreshold   *math.Decimal          `json:"max_threshold" db:"max_threshold"`
	LastCheckedAt  *time.Time             `json:"last_checked_at" db:"last_checked_at"`
	LastUpdateAt   *time.Time             `json:"last_update_at" db:"last_update_at"`
	ErrorCount     int                    `json:"error_count" db:"error_count"`
	LastError      string                 `json:"last_error" db:"last_error"`
	CheckInterval  time.Duration          `json:"check_interval" db:"check_interval"`
	NextCheckAt    time.Time              `json:"next_check_at" db:"next_check_at"`
	IsIncluded     bool                   `json:"is_included" db:"is_included"`
	Weight         *math.Decimal          `json:"weight" db:"weight"`
	Metadata       map[string]interface{} `json:"metadata" db:"metadata"`
	CreatedAt      time.Time              `json:"created_at" db:"created_at"`
	UpdatedAt      time.Time              `json:"updated_at" db:"updated_at"`
	DeletedAt      *time.Time             `json:"deleted_at" db:"deleted_at"`

	// Relations
	Snapshots []*ReserveSnapshot `json:"snapshots,omitempty"`
}

// ReserveSnapshot represents a point-in-time balance snapshot
type ReserveSnapshot struct {
	ID           string                 `json:"id" db:"id"`
	ReserveID    string                 `json:"reserve_id" db:"reserve_id"`
	Balance      *math.Decimal          `json:"balance" db:"balance"`
	Currency     string                 `json:"currency" db:"currency"`
	SnapshotTime time.Time              `json:"snapshot_time" db:"snapshot_time"`
	Source       string                 `json:"source" db:"source"`
	Reference    string                 `json:"reference" db:"reference"`
	ProofID      *string                `json:"proof_id" db:"proof_id"`
	IsVerified   bool                   `json:"is_verified" db:"is_verified"`
	VerifiedAt   *time.Time             `json:"verified_at" db:"verified_at"`
	Metadata     map[string]interface{} `json:"metadata" db:"metadata"`
	CreatedAt    time.Time              `json:"created_at" db:"created_at"`

	// Relations
	Reserve *Reserve `json:"reserve,omitempty"`
	Proof   *ZKProof `json:"proof,omitempty"`
}

// IsActive checks if reserve is actively monitored
func (r *Reserve) IsActive() bool {
	return r.Status == ReserveStatusActive && r.DeletedAt == nil
}

// IsHealthy checks if reserve is healthy (no recent errors)
func (r *Reserve) IsHealthy() bool {
	return r.ErrorCount == 0 || r.LastError == ""
}

// IsBelowThreshold checks if current balance is below minimum threshold
func (r *Reserve) IsBelowThreshold() bool {
	if r.MinThreshold == nil || r.CurrentBalance == nil {
		return false
	}
	return r.CurrentBalance.Cmp(r.MinThreshold) < 0
}

// IsNearThreshold checks if balance is near alert threshold
func (r *Reserve) IsNearThreshold() bool {
	if r.AlertThreshold == nil || r.CurrentBalance == nil {
		return false
	}
	return r.CurrentBalance.Cmp(r.AlertThreshold) < 0
}

// NeedsCheck checks if reserve needs to be checked
func (r *Reserve) NeedsCheck() bool {
	return r.IsActive() && time.Now().After(r.NextCheckAt)
}

// UpdateBalance updates the current balance
func (r *Reserve) UpdateBalance(balance *math.Decimal) {
	r.LastBalance = r.CurrentBalance
	r.CurrentBalance = balance
	now := time.Now()
	r.LastUpdateAt = &now
	r.LastCheckedAt = &now
	r.NextCheckAt = now.Add(r.CheckInterval)
	r.ErrorCount = 0
	r.LastError = ""
	r.UpdatedAt = now
}

// RecordError records an error during balance check
func (r *Reserve) RecordError(errorMsg string) {
	r.ErrorCount++
	r.LastError = errorMsg
	now := time.Now()
	r.LastCheckedAt = &now
	r.NextCheckAt = now.Add(r.CheckInterval * 2) // Backoff on error
	r.UpdatedAt = now

	if r.ErrorCount >= 5 {
		r.Status = ReserveStatusError
	}
}
