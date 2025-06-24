package main

import (
	"context"
	"log"

	"github.com/GoSec-Labs/mMAD/engines/internal/database/models"
	"github.com/GoSec-Labs/mMAD/engines/internal/database/repository"
	"github.com/GoSec-Labs/mMAD/engines/pkg/config"
	"github.com/GoSec-Labs/mMAD/engines/pkg/logger"
	"github.com/GoSec-Labs/mMAD/engines/pkg/math"
)

func main() {
	// Load configuration
	configManager := config.NewManager()
	cfg, err := configManager.Load("")
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Initialize logger
	logger.Init(cfg.Logging)

	// Initialize database
	repoManager, err := repository.NewPostgresManager(cfg.Database)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer repoManager.Close()

	ctx := context.Background()

	// Example: Create a user
	user := &models.User{
		ID:        "user123",
		Email:     "alice@example.com",
		Username:  "alice",
		FirstName: "Alice",
		LastName:  "Smith",
		Status:    models.UserStatusActive,
		KYCLevel:  models.KYCLevelBasic,
	}

	if err := repoManager.Users().Create(ctx, user); err != nil {
		logger.Error("Failed to create user", "error", err)
		return
	}

	// Example: Create an account
	balance, _ := math.NewDecimal("1000.50")
	account := &models.Account{
		ID:               "acc123",
		UserID:           user.ID,
		AccountNumber:    "ACC-001",
		AccountType:      models.AccountTypeChecking,
		Currency:         "USD",
		Balance:          balance,
		AvailableBalance: balance,
		ReservedBalance:  math.NewDecimalFromInt(0),
		Status:           models.AccountStatusActive,
		IsDefault:        true,
	}

	if err := repoManager.Accounts().Create(ctx, account); err != nil {
		logger.Error("Failed to create account", "error", err)
		return
	}

	// Example: Create a transaction
	amount, _ := math.NewDecimal("100.00")
	transaction := &models.Transaction{
		ID:          "tx123",
		ToAccountID: &account.ID,
		Type:        models.TransactionTypeDeposit,
		Amount:      amount,
		Currency:    "USD",
		Status:      models.TransactionStatusCompleted,
		Reference:   "DEP-001",
		Description: "Initial deposit",
	}

	if err := repoManager.Transactions().Create(ctx, transaction); err != nil {
		logger.Error("Failed to create transaction", "error", err)
		return
	}

	logger.Info("✅ Database operations completed successfully!")
}
