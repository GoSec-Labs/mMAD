package models

import (
	"fmt"
	"time"

	"github.com/GoSec-Labs/mMAD/engines/pkg/math"
)

// AccountType represents the type of account
type AccountType string

const (
	AccountTypeChecking  AccountType = "checking"
	AccountTypeSavings   AccountType = "savings"
	AccountTypeReserve   AccountType = "reserve"
	AccountTypeCustodial AccountType = "custodial"
)

// AccountStatus represents account status
type AccountStatus string

const (
	AccountStatusActive    AccountStatus = "active"
	AccountStatusFrozen    AccountStatus = "frozen"
	AccountStatusClosed    AccountStatus = "closed"
	AccountStatusSuspended AccountStatus = "suspended"
)

// Account represents a financial account
type Account struct {
	ID                string                 `json:"id" db:"id"`
	UserID            string                 `json:"user_id" db:"user_id"`
	AccountNumber     string                 `json:"account_number" db:"account_number"`
	AccountType       AccountType            `json:"account_type" db:"account_type"`
	Currency          string                 `json:"currency" db:"currency"`
	Balance           *math.Decimal          `json:"balance" db:"balance"`
	AvailableBalance  *math.Decimal          `json:"available_balance" db:"available_balance"`
	ReservedBalance   *math.Decimal          `json:"reserved_balance" db:"reserved_balance"`
	Status            AccountStatus          `json:"status" db:"status"`
	IsDefault         bool                   `json:"is_default" db:"is_default"`
	InterestRate      *math.Decimal          `json:"interest_rate" db:"interest_rate"`
	OverdraftLimit    *math.Decimal          `json:"overdraft_limit" db:"overdraft_limit"`
	DailyLimit        *math.Decimal          `json:"daily_limit" db:"daily_limit"`
	MonthlyLimit      *math.Decimal          `json:"monthly_limit" db:"monthly_limit"`
	LastTransactionAt *time.Time             `json:"last_transaction_at" db:"last_transaction_at"`
	LastInterestAt    *time.Time             `json:"last_interest_at" db:"last_interest_at"`
	FrozenAt          *time.Time             `json:"frozen_at" db:"frozen_at"`
	FrozenReason      string                 `json:"frozen_reason" db:"frozen_reason"`
	BlockchainAddress string                 `json:"blockchain_address" db:"blockchain_address"`
	Metadata          map[string]interface{} `json:"metadata" db:"metadata"`
	CreatedAt         time.Time              `json:"created_at" db:"created_at"`
	UpdatedAt         time.Time              `json:"updated_at" db:"updated_at"`
	DeletedAt         *time.Time             `json:"deleted_at" db:"deleted_at"`

	// Relations
	User         *User          `json:"user,omitempty"`
	Transactions []*Transaction `json:"transactions,omitempty"`
}

// IsActive checks if account is active
func (a *Account) IsActive() bool {
	return a.Status == AccountStatusActive && a.DeletedAt == nil
}

// IsFrozen checks if account is frozen
func (a *Account) IsFrozen() bool {
	return a.Status == AccountStatusFrozen
}

// CanDebit checks if account can be debited for amount
func (a *Account) CanDebit(amount *math.Decimal) bool {
	if !a.IsActive() {
		return false
	}

	availableWithOverdraft := a.AvailableBalance
	if a.OverdraftLimit != nil && a.OverdraftLimit.IsPositive() {
		availableWithOverdraft = a.AvailableBalance.Add(a.OverdraftLimit)
	}

	return availableWithOverdraft.Cmp(amount) >= 0
}

// UpdateBalance updates account balance
func (a *Account) UpdateBalance(amount *math.Decimal) {
	a.Balance = a.Balance.Add(amount)
	a.AvailableBalance = a.Balance.Sub(a.ReservedBalance)
	a.LastTransactionAt = &time.Time{}
	*a.LastTransactionAt = time.Now()
	a.UpdatedAt = time.Now()
}

// ReserveBalance reserves amount from available balance
func (a *Account) ReserveBalance(amount *math.Decimal) error {
	if a.AvailableBalance.Cmp(amount) < 0 {
		return fmt.Errorf("insufficient available balance")
	}

	a.ReservedBalance = a.ReservedBalance.Add(amount)
	a.AvailableBalance = a.Balance.Sub(a.ReservedBalance)
	a.UpdatedAt = time.Now()

	return nil
}

// ReleaseReserve releases reserved amount
func (a *Account) ReleaseReserve(amount *math.Decimal) {
	if a.ReservedBalance.Cmp(amount) >= 0 {
		a.ReservedBalance = a.ReservedBalance.Sub(amount)
		a.AvailableBalance = a.Balance.Sub(a.ReservedBalance)
		a.UpdatedAt = time.Now()
	}
}
