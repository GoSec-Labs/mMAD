package main

import (
	"os"

	"mMMAD/engines/cmd/mmad/command"
	"mMMAD/engines/internal/config"
	"mMMAD/engines/pkg/logger"
)

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		logger.Error("Failed to load configuration: " + err.Error())
		os.Exit(1)
	}

	// Initialize logger
	logger.Init(cfg.Logging) // ✅ Pass the LoggingConfig directly

	// Execute root command
	if err := command.Execute(cfg); err != nil {
		logger.Error("Command execution failed: " + err.Error())
		os.Exit(1)
	}
}
