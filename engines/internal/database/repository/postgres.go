package repository

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/GoSec-Labs/mMAD/engines/pkg/config"
	"github.com/GoSec-Labs/mMAD/engines/pkg/logger"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

// PostgresManager implements the repository Manager interface
type PostgresManager struct {
	db *sqlx.DB
}

// NewPostgresManager creates a new PostgreSQL repository manager
func NewPostgresManager(cfg config.DatabaseConfig) (*PostgresManager, error) {
	dsn := fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		cfg.Host, cfg.Port, cfg.Username, cfg.Password, cfg.Database, cfg.SSLMode,
	)

	db, err := sqlx.Connect("postgres", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// Configure connection pool
	db.SetMaxOpenConns(cfg.MaxOpenConns)
	db.SetMaxIdleConns(cfg.MaxIdleConns)
	db.SetConnMaxLifetime(cfg.ConnMaxLifetime)

	// Test connection
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	logger.Info("Connected to PostgreSQL database",
		"host", cfg.Host,
		"port", cfg.Port,
		"database", cfg.Database)

	return &PostgresManager{db: db}, nil
}

// Repository implementations
func (m *PostgresManager) Users() UserRepository {
	return &PostgresUserRepository{db: m.db}
}

func (m *PostgresManager) Accounts() AccountRepository {
	return &PostgresAccountRepository{db: m.db}
}

func (m *PostgresManager) Transactions() TransactionRepository {
	return &PostgresTransactionRepository{db: m.db}
}

func (m *PostgresManager) Proofs() ProofRepository {
	return &PostgresProofRepository{db: m.db}
}

func (m *PostgresManager) Reserves() ReserveRepository {
	return &PostgresReserveRepository{db: m.db}
}

func (m *PostgresManager) Compliance() ComplianceRepository {
	return &PostgresComplianceRepository{db: m.db}
}

// Transaction management
func (m *PostgresManager) BeginTx(ctx context.Context) (TxManager, error) {
	tx, err := m.db.BeginTxx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}

	return &PostgresTxManager{tx: tx}, nil
}

func (m *PostgresManager) WithTx(ctx context.Context, fn func(TxManager) error) error {
	tx, err := m.BeginTx(ctx)
	if err != nil {
		return err
	}

	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			panic(r)
		}
	}()

	if err := fn(tx); err != nil {
		if rbErr := tx.Rollback(); rbErr != nil {
			logger.Error("Failed to rollback transaction", "error", rbErr)
		}
		return err
	}

	return tx.Commit()
}

// Close closes the database connection
func (m *PostgresManager) Close() error {
	return m.db.Close()
}

// PostgresTxManager implements transaction management
type PostgresTxManager struct {
	tx *sqlx.Tx
}

func (tm *PostgresTxManager) Users() UserRepository {
	return &PostgresUserRepository{db: tm.tx}
}

func (tm *PostgresTxManager) Accounts() AccountRepository {
	return &PostgresAccountRepository{db: tm.tx}
}

func (tm *PostgresTxManager) Transactions() TransactionRepository {
	return &PostgresTransactionRepository{db: tm.tx}
}

func (tm *PostgresTxManager) Proofs() ProofRepository {
	return &PostgresProofRepository{db: tm.tx}
}

func (tm *PostgresTxManager) Reserves() ReserveRepository {
	return &PostgresReserveRepository{db: tm.tx}
}

func (tm *PostgresTxManager) Compliance() ComplianceRepository {
	return &PostgresComplianceRepository{db: tm.tx}
}

func (tm *PostgresTxManager) Commit() error {
	return tm.tx.Commit()
}

func (tm *PostgresTxManager) Rollback() error {
	return tm.tx.Rollback()
}

// DBExecutor interface for both *sqlx.DB and *sqlx.Tx
type DBExecutor interface {
	GetContext(ctx context.Context, dest interface{}, query string, args ...interface{}) error
	SelectContext(ctx context.Context, dest interface{}, query string, args ...interface{}) error
	ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error)
	NamedExecContext(ctx context.Context, query string, arg interface{}) (sql.Result, error)
	NamedQuery(query string, arg interface{}) (*sqlx.Rows, error)
}
