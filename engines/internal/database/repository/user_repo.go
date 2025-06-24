package repository

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/GoSec-Labs/mMAD/engines/internal/database/models"
)

// PostgresUserRepository implements UserRepository for PostgreSQL
type PostgresUserRepository struct {
	db DBExecutor
}

// Create creates a new user
func (r *PostgresUserRepository) Create(ctx context.Context, user *models.User) error {
	query := `
        INSERT INTO users (
            id, email, username, password_hash, first_name, last_name,
            date_of_birth, phone_number, address, status, kyc_level,
            two_factor_enabled, metadata
        ) VALUES (
            :id, :email, :username, :password_hash, :first_name, :last_name,
            :date_of_birth, :phone_number, :address, :status, :kyc_level,
            :two_factor_enabled, :metadata
        )`

	_, err := r.db.NamedExecContext(ctx, query, user)
	if err != nil {
		return fmt.Errorf("failed to create user: %w", err)
	}

	return nil
}

// GetByID retrieves a user by ID
func (r *PostgresUserRepository) GetByID(ctx context.Context, id string) (*models.User, error) {
	query := `
        SELECT * FROM users 
        WHERE id = $1 AND deleted_at IS NULL`

	user := &models.User{}
	err := r.db.GetContext(ctx, user, query, id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("user not found: %s", id)
		}
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	return user, nil
}

// GetByEmail retrieves a user by email
func (r *PostgresUserRepository) GetByEmail(ctx context.Context, email string) (*models.User, error) {
	query := `
        SELECT * FROM users 
        WHERE email = $1 AND deleted_at IS NULL`

	user := &models.User{}
	err := r.db.GetContext(ctx, user, query, email)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("user not found: %s", email)
		}
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	return user, nil
}

// GetByUsername retrieves a user by username
func (r *PostgresUserRepository) GetByUsername(ctx context.Context, username string) (*models.User, error) {
	query := `
        SELECT * FROM users 
        WHERE username = $1 AND deleted_at IS NULL`

	user := &models.User{}
	err := r.db.GetContext(ctx, user, query, username)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("user not found: %s", username)
		}
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	return user, nil
}

// Update updates a user
func (r *PostgresUserRepository) Update(ctx context.Context, user *models.User) error {
	query := `
        UPDATE users SET
            email = :email,
            username = :username,
            first_name = :first_name,
            last_name = :last_name,
            date_of_birth = :date_of_birth,
            phone_number = :phone_number,
            address = :address,
            status = :status,
            kyc_level = :kyc_level,
            kyc_verified_at = :kyc_verified_at,
            two_factor_enabled = :two_factor_enabled,
            metadata = :metadata,
            updated_at = NOW()
        WHERE id = :id AND deleted_at IS NULL`

	result, err := r.db.NamedExecContext(ctx, query, user)
	if err != nil {
		return fmt.Errorf("failed to update user: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return fmt.Errorf("user not found: %s", user.ID)
	}

	return nil
}

// Delete soft deletes a user
func (r *PostgresUserRepository) Delete(ctx context.Context, id string) error {
	query := `
        UPDATE users SET 
            deleted_at = NOW(),
            updated_at = NOW()
        WHERE id = $1 AND deleted_at IS NULL`

	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete user: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return fmt.Errorf("user not found: %s", id)
	}

	return nil
}

// List retrieves users with filters
func (r *PostgresUserRepository) List(ctx context.Context, filters UserFilters) ([]*models.User, error) {
	query := "SELECT * FROM users WHERE deleted_at IS NULL"
	args := []interface{}{}
	argIndex := 1

	// Apply filters
	if filters.Status != nil {
		query += fmt.Sprintf(" AND status = $%d", argIndex)
		args = append(args, *filters.Status)
		argIndex++
	}

	if filters.KYCLevel != nil {
		query += fmt.Sprintf(" AND kyc_level = $%d", argIndex)
		args = append(args, *filters.KYCLevel)
		argIndex++
	}

	if filters.CreatedAt != nil {
		query += fmt.Sprintf(" AND created_at BETWEEN $%d AND $%d", argIndex, argIndex+1)
		args = append(args, filters.CreatedAt.Start, filters.CreatedAt.End)
		argIndex += 2
	}

	// Apply ordering
	orderBy := "created_at"
	if filters.OrderBy != "" {
		orderBy = filters.OrderBy
	}

	orderDir := "DESC"
	if filters.OrderDir != "" {
		orderDir = strings.ToUpper(filters.OrderDir)
	}

	query += fmt.Sprintf(" ORDER BY %s %s", orderBy, orderDir)

	// Apply pagination
	if filters.Limit != nil {
		query += fmt.Sprintf(" LIMIT $%d", argIndex)
		args = append(args, *filters.Limit)
		argIndex++
	}

	if filters.Offset != nil {
		query += fmt.Sprintf(" OFFSET $%d", argIndex)
		args = append(args, *filters.Offset)
		argIndex++
	}

	var users []*models.User
	err := r.db.SelectContext(ctx, &users, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to list users: %w", err)
	}

	return users, nil
}

// Count counts users with filters
func (r *PostgresUserRepository) Count(ctx context.Context, filters UserFilters) (int64, error) {
	query := "SELECT COUNT(*) FROM users WHERE deleted_at IS NULL"
	args := []interface{}{}
	argIndex := 1

	// Apply same filters as List
	if filters.Status != nil {
		query += fmt.Sprintf(" AND status = $%d", argIndex)
		args = append(args, *filters.Status)
		argIndex++
	}

	if filters.KYCLevel != nil {
		query += fmt.Sprintf(" AND kyc_level = $%d", argIndex)
		args = append(args, *filters.KYCLevel)
		argIndex++
	}

	if filters.CreatedAt != nil {
		query += fmt.Sprintf(" AND created_at BETWEEN $%d AND $%d", argIndex, argIndex+1)
		args = append(args, filters.CreatedAt.Start, filters.CreatedAt.End)
		argIndex += 2
	}

	var count int64
	err := r.db.GetContext(ctx, &count, query, args...)
	if err != nil {
		return 0, fmt.Errorf("failed to count users: %w", err)
	}

	return count, nil
}

// UpdateLastLogin updates the last login timestamp
func (r *PostgresUserRepository) UpdateLastLogin(ctx context.Context, id string) error {
	query := `
        UPDATE users SET 
            last_login_at = NOW(),
            login_attempts = 0,
            updated_at = NOW()
        WHERE id = $1 AND deleted_at IS NULL`

	_, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to update last login: %w", err)
	}

	return nil
}

// IncrementLoginAttempts increments login attempts
func (r *PostgresUserRepository) IncrementLoginAttempts(ctx context.Context, id string) error {
	query := `
        UPDATE users SET 
            login_attempts = login_attempts + 1,
            updated_at = NOW()
        WHERE id = $1 AND deleted_at IS NULL`

	_, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to increment login attempts: %w", err)
	}

	return nil
}

// ResetLoginAttempts resets login attempts to 0
func (r *PostgresUserRepository) ResetLoginAttempts(ctx context.Context, id string) error {
	query := `
        UPDATE users SET 
            login_attempts = 0,
            locked_until = NULL,
            updated_at = NOW()
        WHERE id = $1 AND deleted_at IS NULL`

	_, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to reset login attempts: %w", err)
	}

	return nil
}

// LockAccount locks user account until specified time
func (r *PostgresUserRepository) LockAccount(ctx context.Context, id string, until time.Time) error {
	query := `
        UPDATE users SET 
            locked_until = $2,
            updated_at = NOW()
        WHERE id = $1 AND deleted_at IS NULL`

	_, err := r.db.ExecContext(ctx, query, id, until)
	if err != nil {
		return fmt.Errorf("failed to lock account: %w", err)
	}

	return nil
}
