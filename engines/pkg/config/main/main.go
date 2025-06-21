// main.go
package main

import (
	"fmt"
	"log"

	"github.com/GoSec-Labs/mMAD/engines/pkg/config"
	"github.com/GoSec-Labs/mMAD/engines/pkg/logger"
)

func main() {
	// Load configuration
	configManager := config.NewManager()
	cfg, err := configManager.Load(config.GetConfigPath())
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Initialize logger with config
	if err := logger.Init(cfg.Logging); err != nil {
		log.Fatalf("Failed to initialize logger: %v", err)
	}

	logger.Info("🚀 mMAD Engines starting...")

	// Use configuration
	fmt.Printf("Server running on %s:%d\n",
		cfg.Server.Host, cfg.Server.Port)
	fmt.Printf("Database: %s://%s:%d/%s\n",
		cfg.Database.Type, cfg.Database.Host,
		cfg.Database.Port, cfg.Database.Database)
}
