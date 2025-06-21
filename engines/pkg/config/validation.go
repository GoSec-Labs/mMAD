package config

import (
	"fmt"
	"net/url"
	"strconv"
	"strings"
)

// Validator handles configuration validation
type Validator struct{}

// NewValidator creates a new configuration validator
func NewValidator() *Validator {
	return &Validator{}
}

// Validate validates the entire configuration
func (v *Validator) Validate(config *Config) error {
	validators := []func(*Config) error{
		v.validateServer,
		v.validateDatabase,
		v.validateLogging,
		v.validateCrypto,
		v.validateZKProof,
		v.validateReserve,
		v.validateCompliance,
		v.validateBlockchain,
		v.validateMessaging,
		v.validateMonitoring,
	}

	for _, validator := range validators {
		if err := validator(config); err != nil {
			return err
		}
	}

	return nil
}

// validateServer validates server configuration
func (v *Validator) validateServer(config *Config) error {
	server := config.Server

	if server.Host == "" {
		return fmt.Errorf("server host cannot be empty")
	}

	if server.Port <= 0 || server.Port > 65535 {
		return fmt.Errorf("invalid server port: %d", server.Port)
	}

	if server.GRPCPort <= 0 || server.GRPCPort > 65535 {
		return fmt.Errorf("invalid gRPC port: %d", server.GRPCPort)
	}

	if server.Port == server.GRPCPort {
		return fmt.Errorf("server port and gRPC port cannot be the same")
	}

	if server.EnableTLS {
		if server.TLSCertPath == "" || server.TLSKeyPath == "" {
			return fmt.Errorf("TLS cert and key paths required when TLS is enabled")
		}
	}

	return nil
}

// validateDatabase validates database configuration
func (v *Validator) validateDatabase(config *Config) error {
	db := config.Database

	validTypes := map[string]bool{
		"postgres": true,
		"mysql":    true,
		"sqlite":   true,
	}

	if !validTypes[db.Type] {
		return fmt.Errorf("unsupported database type: %s", db.Type)
	}

	if db.Type != "sqlite" {
		if db.Host == "" {
			return fmt.Errorf("database host cannot be empty")
		}

		if db.Port <= 0 || db.Port > 65535 {
			return fmt.Errorf("invalid database port: %d", db.Port)
		}

		if db.Username == "" {
			return fmt.Errorf("database username cannot be empty")
		}

		if db.Database == "" {
			return fmt.Errorf("database name cannot be empty")
		}
	}

	if db.MaxOpenConns < 0 {
		return fmt.Errorf("max open connections cannot be negative")
	}

	if db.MaxIdleConns < 0 {
		return fmt.Errorf("max idle connections cannot be negative")
	}

	return nil
}

// validateLogging validates logging configuration
func (v *Validator) validateLogging(config *Config) error {
	logging := config.Logging

	validLevels := map[string]bool{
		"debug": true,
		"info":  true,
		"warn":  true,
		"error": true,
		"fatal": true,
		"panic": true,
	}

	if !validLevels[strings.ToLower(logging.Level)] {
		return fmt.Errorf("invalid log level: %s", logging.Level)
	}

	validFormats := map[string]bool{
		"json": true,
		"text": true,
	}

	if !validFormats[strings.ToLower(logging.Format)] {
		return fmt.Errorf("invalid log format: %s", logging.Format)
	}

	return nil
}

// validateCrypto validates crypto configuration
func (v *Validator) validateCrypto(config *Config) error {
	crypto := config.Crypto

	validHashFunctions := map[string]bool{
		"sha256":    true,
		"keccak256": true,
		"poseidon":  true,
	}

	if !validHashFunctions[crypto.DefaultHashFunction] {
		return fmt.Errorf("invalid hash function: %s", crypto.DefaultHashFunction)
	}

	validKeySizes := map[int]bool{
		16: true, // AES-128
		24: true, // AES-192
		32: true, // AES-256
	}

	if !validKeySizes[crypto.KeySize] {
		return fmt.Errorf("invalid key size: %d", crypto.KeySize)
	}

	return nil
}

// validateZKProof validates ZK proof configuration
func (v *Validator) validateZKProof(config *Config) error {
	zkproof := config.ZKProof

	if zkproof.CircuitPath == "" {
		return fmt.Errorf("circuit path cannot be empty")
	}

	if zkproof.TrustedSetupPath == "" {
		return fmt.Errorf("trusted setup path cannot be empty")
	}

	if zkproof.ProofCacheSize < 0 {
		return fmt.Errorf("proof cache size cannot be negative")
	}

	if zkproof.ProofTimeout <= 0 {
		return fmt.Errorf("proof timeout must be positive")
	}

	if zkproof.BatchSize <= 0 {
		return fmt.Errorf("batch size must be positive")
	}

	if zkproof.MaxConcurrentProofs <= 0 {
		return fmt.Errorf("max concurrent proofs must be positive")
	}

	return nil
}

// validateReserve validates reserve configuration
func (v *Validator) validateReserve(config *Config) error {
	reserve := config.Reserve

	if reserve.BankAPIEndpoint != "" {
		if _, err := url.Parse(reserve.BankAPIEndpoint); err != nil {
			return fmt.Errorf("invalid bank API endpoint: %w", err)
		}
	}

	if reserve.CheckInterval <= 0 {
		return fmt.Errorf("check interval must be positive")
	}

	if reserve.MinReserveThreshold != "" {
		if _, err := strconv.ParseFloat(reserve.MinReserveThreshold, 64); err != nil {
			return fmt.Errorf("invalid min reserve threshold: %w", err)
		}
	}

	return nil
}

// validateCompliance validates compliance configuration
func (v *Validator) validateCompliance(config *Config) error {
	compliance := config.Compliance

	if compliance.CheckTimeout <= 0 {
		return fmt.Errorf("check timeout must be positive")
	}

	if compliance.RetryAttempts < 0 {
		return fmt.Errorf("retry attempts cannot be negative")
	}

	if compliance.CacheExpiration < 0 {
		return fmt.Errorf("cache expiration cannot be negative")
	}

	return nil
}

// validateBlockchain validates blockchain configuration
func (v *Validator) validateBlockchain(config *Config) error {
	blockchain := config.Blockchain

	if len(blockchain.Networks) == 0 {
		return fmt.Errorf("at least one blockchain network must be configured")
	}

	if blockchain.DefaultNetwork == "" {
		return fmt.Errorf("default network must be specified")
	}

	if _, exists := blockchain.Networks[blockchain.DefaultNetwork]; !exists {
		return fmt.Errorf("default network not found in networks: %s", blockchain.DefaultNetwork)
	}

	for name, network := range blockchain.Networks {
		if err := v.validateNetwork(name, network); err != nil {
			return err
		}
	}

	return nil
}

// validateNetwork validates individual network configuration
func (v *Validator) validateNetwork(name string, network NetworkConfig) error {
	if network.Name == "" {
		return fmt.Errorf("network %s: name cannot be empty", name)
	}

	if network.ChainID <= 0 {
		return fmt.Errorf("network %s: invalid chain ID", name)
	}

	if network.RPCURL == "" {
		return fmt.Errorf("network %s: RPC URL cannot be empty", name)
	}

	if _, err := url.Parse(network.RPCURL); err != nil {
		return fmt.Errorf("network %s: invalid RPC URL: %w", name, err)
	}

	if network.ContractAddress == "" {
		return fmt.Errorf("network %s: contract address cannot be empty", name)
	}

	return nil
}

// validateMessaging validates messaging configuration
func (v *Validator) validateMessaging(config *Config) error {
	messaging := config.Messaging

	validProviders := map[string]bool{
		"rabbitmq": true,
		"kafka":    true,
		"nats":     true,
	}

	if messaging.Provider != "" && !validProviders[messaging.Provider] {
		return fmt.Errorf("unsupported messaging provider: %s", messaging.Provider)
	}

	if messaging.BrokerURL != "" {
		if _, err := url.Parse(messaging.BrokerURL); err != nil {
			return fmt.Errorf("invalid broker URL: %w", err)
		}
	}

	return nil
}

// validateMonitoring validates monitoring configuration
func (v *Validator) validateMonitoring(config *Config) error {
	monitoring := config.Monitoring

	if monitoring.Enabled {
		if monitoring.MetricsPort <= 0 || monitoring.MetricsPort > 65535 {
			return fmt.Errorf("invalid metrics port: %d", monitoring.MetricsPort)
		}

		if monitoring.HealthCheckPort <= 0 || monitoring.HealthCheckPort > 65535 {
			return fmt.Errorf("invalid health check port: %d", monitoring.HealthCheckPort)
		}

		if monitoring.JaegerEnabled && monitoring.JaegerEndpoint == "" {
			return fmt.Errorf("Jaeger endpoint required when Jaeger is enabled")
		}
	}

	return nil
}
