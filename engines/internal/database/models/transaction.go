package models

import (
	"time"

	"github.com/GoSec-Labs/mMAD/engines/pkg/math"
)

// TransactionType represents the type of transaction
type TransactionType string

const (
	TransactionTypeDeposit    TransactionType = "deposit"
	TransactionTypeWithdrawal TransactionType = "withdrawal"
	TransactionTypeTransfer   TransactionType = "transfer"
	TransactionTypeInterest   TransactionType = "interest"
	TransactionTypeFee        TransactionType = "fee"
	TransactionTypeRefund     TransactionType = "refund"
	TransactionTypeAdjustment TransactionType = "adjustment"
)

// TransactionStatus represents transaction status
type TransactionStatus string

const (
	TransactionStatusPending   TransactionStatus = "pending"
	TransactionStatusProcessed TransactionStatus = "processed"
	TransactionStatusCompleted TransactionStatus = "completed"
	TransactionStatusFailed    TransactionStatus = "failed"
	TransactionStatusCancelled TransactionStatus = "cancelled"
	TransactionStatusReversed  TransactionStatus = "reversed"
)

// Transaction represents a financial transaction
type Transaction struct {
	ID                string                 `json:"id" db:"id"`
	FromAccountID     *string                `json:"from_account_id" db:"from_account_id"`
	ToAccountID       *string                `json:"to_account_id" db:"to_account_id"`
	Type              TransactionType        `json:"type" db:"type"`
	Amount            *math.Decimal          `json:"amount" db:"amount"`
	Currency          string                 `json:"currency" db:"currency"`
	Fee               *math.Decimal          `json:"fee" db:"fee"`
	Status            TransactionStatus      `json:"status" db:"status"`
	Reference         string                 `json:"reference" db:"reference"`
	ExternalReference string                 `json:"external_reference" db:"external_reference"`
	Description       string                 `json:"description" db:"description"`
	Category          string                 `json:"category" db:"category"`
	Tags              []string               `json:"tags" db:"tags"`
	FromBalance       *math.Decimal          `json:"from_balance" db:"from_balance"`
	ToBalance         *math.Decimal          `json:"to_balance" db:"to_balance"`
	ExchangeRate      *math.Decimal          `json:"exchange_rate" db:"exchange_rate"`
	ProcessedAt       *time.Time             `json:"processed_at" db:"processed_at"`
	CompletedAt       *time.Time             `json:"completed_at" db:"completed_at"`
	FailedAt          *time.Time             `json:"failed_at" db:"failed_at"`
	FailureReason     string                 `json:"failure_reason" db:"failure_reason"`
	BlockchainTxHash  string                 `json:"blockchain_tx_hash" db:"blockchain_tx_hash"`
	BlockNumber       *int64                 `json:"block_number" db:"block_number"`
	GasUsed           *int64                 `json:"gas_used" db:"gas_used"`
	Metadata          map[string]interface{} `json:"metadata" db:"metadata"`
	CreatedAt         time.Time              `json:"created_at" db:"created_at"`
	UpdatedAt         time.Time              `json:"updated_at" db:"updated_at"`

	// Relations
	FromAccount *Account `json:"from_account,omitempty"`
	ToAccount   *Account `json:"to_account,omitempty"`
}

// IsCompleted checks if transaction is completed
func (t *Transaction) IsCompleted() bool {
	return t.Status == TransactionStatusCompleted
}

// IsPending checks if transaction is pending
func (t *Transaction) IsPending() bool {
	return t.Status == TransactionStatusPending
}

// CanReverse checks if transaction can be reversed
func (t *Transaction) CanReverse() bool {
	return t.IsCompleted() &&
		t.Type != TransactionTypeInterest &&
		t.Type != TransactionTypeFee
}

// GetNetAmount returns amount minus fee
func (t *Transaction) GetNetAmount() *math.Decimal {
	if t.Fee == nil || t.Fee.IsZero() {
		return t.Amount
	}
	return t.Amount.Sub(t.Fee)
}

// MarkCompleted marks transaction as completed
func (t *Transaction) MarkCompleted() {
	t.Status = TransactionStatusCompleted
	now := time.Now()
	t.CompletedAt = &now
	t.UpdatedAt = now
}

// MarkFailed marks transaction as failed
func (t *Transaction) MarkFailed(reason string) {
	t.Status = TransactionStatusFailed
	t.FailureReason = reason
	now := time.Now()
	t.FailedAt = &now
	t.UpdatedAt = now
}
