package repository

import (
	"context"

	"time"

	"github.com/GoSec-Labs/mMAD/engines/internal/database/models"
	"github.com/GoSec-Labs/mMAD/engines/pkg/math"
)

// UserRepository defines user data operations
type UserRepository interface {
	Create(ctx context.Context, user *models.User) error
	GetByID(ctx context.Context, id string) (*models.User, error)
	GetByEmail(ctx context.Context, email string) (*models.User, error)
	GetByUsername(ctx context.Context, username string) (*models.User, error)
	Update(ctx context.Context, user *models.User) error
	Delete(ctx context.Context, id string) error
	List(ctx context.Context, filters UserFilters) ([]*models.User, error)
	Count(ctx context.Context, filters UserFilters) (int64, error)
	UpdateLastLogin(ctx context.Context, id string) error
	IncrementLoginAttempts(ctx context.Context, id string) error
	ResetLoginAttempts(ctx context.Context, id string) error
	LockAccount(ctx context.Context, id string, until time.Time) error
}

// AccountRepository defines account data operations
type AccountRepository interface {
	Create(ctx context.Context, account *models.Account) error
	GetByID(ctx context.Context, id string) (*models.Account, error)
	GetByAccountNumber(ctx context.Context, accountNumber string) (*models.Account, error)
	GetByUserID(ctx context.Context, userID string) ([]*models.Account, error)
	GetDefaultByUserID(ctx context.Context, userID string) (*models.Account, error)
	Update(ctx context.Context, account *models.Account) error
	UpdateBalance(ctx context.Context, id string, balance *math.Decimal) error
	Delete(ctx context.Context, id string) error
	List(ctx context.Context, filters AccountFilters) ([]*models.Account, error)
	Count(ctx context.Context, filters AccountFilters) (int64, error)
	FreezeAccount(ctx context.Context, id string, reason string) error
	UnfreezeAccount(ctx context.Context, id string) error
}

// TransactionRepository defines transaction data operations
type TransactionRepository interface {
	Create(ctx context.Context, transaction *models.Transaction) error
	GetByID(ctx context.Context, id string) (*models.Transaction, error)
	GetByReference(ctx context.Context, reference string) (*models.Transaction, error)
	GetByAccountID(ctx context.Context, accountID string, filters TransactionFilters) ([]*models.Transaction, error)
	Update(ctx context.Context, transaction *models.Transaction) error
	UpdateStatus(ctx context.Context, id string, status models.TransactionStatus) error
	List(ctx context.Context, filters TransactionFilters) ([]*models.Transaction, error)
	Count(ctx context.Context, filters TransactionFilters) (int64, error)
	GetDailyVolume(ctx context.Context, accountID string, date time.Time) (*math.Decimal, error)
	GetMonthlyVolume(ctx context.Context, accountID string, month time.Time) (*math.Decimal, error)
}

// ProofRepository defines ZK proof data operations
type ProofRepository interface {
	Create(ctx context.Context, proof *models.ZKProof) error
	GetByID(ctx context.Context, id string) (*models.ZKProof, error)
	GetByHash(ctx context.Context, hash string) (*models.ZKProof, error)
	GetByUserID(ctx context.Context, userID string, filters ProofFilters) ([]*models.ZKProof, error)
	GetByAccountID(ctx context.Context, accountID string, filters ProofFilters) ([]*models.ZKProof, error)
	Update(ctx context.Context, proof *models.ZKProof) error
	UpdateStatus(ctx context.Context, id string, status models.ProofStatus) error
	Delete(ctx context.Context, id string) error
	List(ctx context.Context, filters ProofFilters) ([]*models.ZKProof, error)
	Count(ctx context.Context, filters ProofFilters) (int64, error)
	GetLatestByType(ctx context.Context, proofType models.ProofType) (*models.ZKProof, error)
	CleanupExpired(ctx context.Context) (int64, error)
}

// ReserveRepository defines reserve monitoring data operations
type ReserveRepository interface {
	Create(ctx context.Context, reserve *models.Reserve) error
	GetByID(ctx context.Context, id string) (*models.Reserve, error)
	GetByAccountNumber(ctx context.Context, accountNumber string) (*models.Reserve, error)
	Update(ctx context.Context, reserve *models.Reserve) error
	UpdateBalance(ctx context.Context, id string, balance *math.Decimal) error
	Delete(ctx context.Context, id string) error
	List(ctx context.Context, filters ReserveFilters) ([]*models.Reserve, error)
	GetActiveReserves(ctx context.Context) ([]*models.Reserve, error)
	GetReservesNeedingCheck(ctx context.Context) ([]*models.Reserve, error)

	// Snapshot operations
	CreateSnapshot(ctx context.Context, snapshot *models.ReserveSnapshot) error
	GetSnapshotsByReserveID(ctx context.Context, reserveID string, filters SnapshotFilters) ([]*models.ReserveSnapshot, error)
	GetLatestSnapshot(ctx context.Context, reserveID string) (*models.ReserveSnapshot, error)
	GetSnapshotsForPeriod(ctx context.Context, start, end time.Time) ([]*models.ReserveSnapshot, error)
}

// ComplianceRepository defines compliance data operations
type ComplianceRepository interface {
	CreateCheck(ctx context.Context, check *models.ComplianceCheck) error
	GetCheckByID(ctx context.Context, id string) (*models.ComplianceCheck, error)
	GetChecksByUserID(ctx context.Context, userID string, filters ComplianceFilters) ([]*models.ComplianceCheck, error)
	UpdateCheck(ctx context.Context, check *models.ComplianceCheck) error
	UpdateCheckStatus(ctx context.Context, id string, status models.ComplianceStatus) error
	ListChecks(ctx context.Context, filters ComplianceFilters) ([]*models.ComplianceCheck, error)
	GetPendingChecks(ctx context.Context) ([]*models.ComplianceCheck, error)
	GetChecksNeedingRetry(ctx context.Context) ([]*models.ComplianceCheck, error)

	// Rule operations
	CreateRule(ctx context.Context, rule *models.ComplianceRule) error
	GetRuleByID(ctx context.Context, id string) (*models.ComplianceRule, error)
	GetActiveRules(ctx context.Context, checkType models.ComplianceCheckType) ([]*models.ComplianceRule, error)
	UpdateRule(ctx context.Context, rule *models.ComplianceRule) error
	DeleteRule(ctx context.Context, id string) error
}

// Filter types for repository queries
type UserFilters struct {
	Status    *models.UserStatus
	KYCLevel  *models.KYCLevel
	CreatedAt *TimeRange
	Limit     *int
	Offset    *int
	OrderBy   string
	OrderDir  string
}

type AccountFilters struct {
	UserID     *string
	Type       *models.AccountType
	Status     *models.AccountStatus
	Currency   *string
	MinBalance *math.Decimal
	MaxBalance *math.Decimal
	CreatedAt  *TimeRange
	Limit      *int
	Offset     *int
	OrderBy    string
	OrderDir   string
}

type TransactionFilters struct {
	AccountID *string
	Type      *models.TransactionType
	Status    *models.TransactionStatus
	Currency  *string
	MinAmount *math.Decimal
	MaxAmount *math.Decimal
	DateRange *TimeRange
	Reference *string
	Category  *string
	Limit     *int
	Offset    *int
	OrderBy   string
	OrderDir  string
}

type ProofFilters struct {
	Type      *models.ProofType
	Status    *models.ProofStatus
	UserID    *string
	AccountID *string
	CreatedAt *TimeRange
	Limit     *int
	Offset    *int
	OrderBy   string
	OrderDir  string
}

type ReserveFilters struct {
	Type       *models.ReserveType
	Status     *models.ReserveStatus
	Currency   *string
	IsIncluded *bool
	MinBalance *math.Decimal
	MaxBalance *math.Decimal
	Limit      *int
	Offset     *int
	OrderBy    string
	OrderDir   string
}

type SnapshotFilters struct {
	DateRange  *TimeRange
	IsVerified *bool
	Source     *string
	Limit      *int
	Offset     *int
	OrderBy    string
	OrderDir   string
}

type ComplianceFilters struct {
	Type      *models.ComplianceCheckType
	Status    *models.ComplianceStatus
	RiskLevel *models.RiskLevel
	UserID    *string
	Provider  *string
	CreatedAt *TimeRange
	Limit     *int
	Offset    *int
	OrderBy   string
	OrderDir  string
}

type TimeRange struct {
	Start time.Time
	End   time.Time
}

// Repository manager interface
type Manager interface {
	Users() UserRepository
	Accounts() AccountRepository
	Transactions() TransactionRepository
	Proofs() ProofRepository
	Reserves() ReserveRepository
	Compliance() ComplianceRepository

	// Transaction management
	BeginTx(ctx context.Context) (TxManager, error)
	WithTx(ctx context.Context, fn func(TxManager) error) error
}

// Transaction manager interface
type TxManager interface {
	Users() UserRepository
	Accounts() AccountRepository
	Transactions() TransactionRepository
	Proofs() ProofRepository
	Reserves() ReserveRepository
	Compliance() ComplianceRepository

	Commit() error
	Rollback() error
}
