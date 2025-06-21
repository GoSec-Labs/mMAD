package config

import (
	"fmt"
	"os"
	"path/filepath"
)

// Manager handles configuration loading and management
type Manager struct {
	config    *Config
	loader    *Loader
	validator *Validator
}

// NewManager creates a new configuration manager
func NewManager() *Manager {
	return &Manager{
		loader:    NewLoader(),
		validator: NewValidator(),
	}
}

// Load loads configuration from various sources
func (m *Manager) Load(configPath string) (*Config, error) {
	// Start with defaults
	config := GetDefaults()

	// Load from file if provided
	if configPath != "" {
		if err := m.loader.LoadFromFile(configPath, config); err != nil {
			return nil, fmt.Errorf("failed to load config file: %w", err)
		}
	}

	// Override with environment variables
	if err := m.loader.LoadFromEnv(config); err != nil {
		return nil, fmt.Errorf("failed to load environment variables: %w", err)
	}

	// Validate configuration
	if err := m.validator.Validate(config); err != nil {
		return nil, fmt.Errorf("config validation failed: %w", err)
	}

	m.config = config
	return config, nil
}

// Get returns the current configuration
func (m *Manager) Get() *Config {
	return m.config
}

// GetServer returns server configuration
func (m *Manager) GetServer() ServerConfig {
	if m.config == nil {
		return GetDefaults().Server
	}
	return m.config.Server
}

// GetDatabase returns database configuration
func (m *Manager) GetDatabase() DatabaseConfig {
	if m.config == nil {
		return GetDefaults().Database
	}
	return m.config.Database
}

// GetZKProof returns ZK proof configuration
func (m *Manager) GetZKProof() ZKProofConfig {
	if m.config == nil {
		return GetDefaults().ZKProof
	}
	return m.config.ZKProof
}

// GetBlockchain returns blockchain configuration
func (m *Manager) GetBlockchain() BlockchainConfig {
	if m.config == nil {
		return GetDefaults().Blockchain
	}
	return m.config.Blockchain
}

// GetNetworkConfig returns configuration for a specific network
func (m *Manager) GetNetworkConfig(network string) (NetworkConfig, error) {
	blockchain := m.GetBlockchain()

	if networkConfig, exists := blockchain.Networks[network]; exists {
		return networkConfig, nil
	}

	return NetworkConfig{}, fmt.Errorf("network configuration not found: %s", network)
}

// IsProduction checks if running in production mode
func (m *Manager) IsProduction() bool {
	return os.Getenv("ENVIRONMENT") == "production" ||
		os.Getenv("ENV") == "production"
}

// IsDevelopment checks if running in development mode
func (m *Manager) IsDevelopment() bool {
	env := os.Getenv("ENVIRONMENT")
	if env == "" {
		env = os.Getenv("ENV")
	}
	return env == "development" || env == "dev" || env == ""
}

// GetConfigPath returns the configuration file path
func GetConfigPath() string {
	if path := os.Getenv("CONFIG_PATH"); path != "" {
		return path
	}

	// Try common locations
	locations := []string{
		"./config.yaml",
		"./config.yml",
		"./config.json",
		"./configs/config.yaml",
		"/etc/mmad/config.yaml",
	}

	for _, location := range locations {
		if _, err := os.Stat(location); err == nil {
			return location
		}
	}

	return ""
}

// WriteExample writes an example configuration file
func WriteExample(path string) error {
	config := GetDefaults()

	// Determine format from file extension
	ext := filepath.Ext(path)

	loader := NewLoader()
	switch ext {
	case ".yaml", ".yml":
		return loader.WriteYAML(path, config)
	case ".json":
		return loader.WriteJSON(path, config)
	default:
		return fmt.Errorf("unsupported file format: %s", ext)
	}
}
