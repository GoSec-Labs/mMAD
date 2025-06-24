package models

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"time"
)

// UserStatus represents user account status
type UserStatus string

const (
	UserStatusPending   UserStatus = "pending"
	UserStatusActive    UserStatus = "active"
	UserStatusSuspended UserStatus = "suspended"
	UserStatusClosed    UserStatus = "closed"
)

// KYCLevel represents KYC verification level
type KYCLevel string

const (
	KYCLevelNone     KYCLevel = "none"
	KYCLevelBasic    KYCLevel = "basic"
	KYCLevelStandard KYCLevel = "standard"
	KYCLevelEnhanced KYCLevel = "enhanced"
)

// User represents a user in the system
type User struct {
	ID               string                 `json:"id" db:"id"`
	Email            string                 `json:"email" db:"email"`
	Username         string                 `json:"username" db:"username"`
	PasswordHash     string                 `json:"-" db:"password_hash"`
	FirstName        string                 `json:"first_name" db:"first_name"`
	LastName         string                 `json:"last_name" db:"last_name"`
	DateOfBirth      *time.Time             `json:"date_of_birth" db:"date_of_birth"`
	PhoneNumber      string                 `json:"phone_number" db:"phone_number"`
	Address          *Address               `json:"address" db:"address"`
	Status           UserStatus             `json:"status" db:"status"`
	KYCLevel         KYCLevel               `json:"kyc_level" db:"kyc_level"`
	KYCVerifiedAt    *time.Time             `json:"kyc_verified_at" db:"kyc_verified_at"`
	TwoFactorEnabled bool                   `json:"two_factor_enabled" db:"two_factor_enabled"`
	TwoFactorSecret  string                 `json:"-" db:"two_factor_secret"`
	LastLoginAt      *time.Time             `json:"last_login_at" db:"last_login_at"`
	LoginAttempts    int                    `json:"login_attempts" db:"login_attempts"`
	LockedUntil      *time.Time             `json:"locked_until" db:"locked_until"`
	Metadata         map[string]interface{} `json:"metadata" db:"metadata"`
	CreatedAt        time.Time              `json:"created_at" db:"created_at"`
	UpdatedAt        time.Time              `json:"updated_at" db:"updated_at"`
	DeletedAt        *time.Time             `json:"deleted_at" db:"deleted_at"`
}

// Address represents a user's address
type Address struct {
	Street     string `json:"street"`
	City       string `json:"city"`
	State      string `json:"state"`
	PostalCode string `json:"postal_code"`
	Country    string `json:"country"`
}

// Value implements driver.Valuer for Address
func (a Address) Value() (driver.Value, error) {
	return json.Marshal(a)
}

// Scan implements sql.Scanner for Address
func (a *Address) Scan(value interface{}) error {
	if value == nil {
		return nil
	}

	bytes, ok := value.([]byte)
	if !ok {
		return fmt.Errorf("cannot scan %T into Address", value)
	}

	return json.Unmarshal(bytes, a)
}

// IsActive checks if user is active
func (u *User) IsActive() bool {
	return u.Status == UserStatusActive && u.DeletedAt == nil
}

// IsKYCVerified checks if user has completed KYC
func (u *User) IsKYCVerified() bool {
	return u.KYCLevel != KYCLevelNone && u.KYCVerifiedAt != nil
}

// IsLocked checks if user account is locked
func (u *User) IsLocked() bool {
	return u.LockedUntil != nil && u.LockedUntil.After(time.Now())
}

// GetFullName returns user's full name
func (u *User) GetFullName() string {
	return fmt.Sprintf("%s %s", u.FirstName, u.LastName)
}
