package config

import "time"

// GetDefaults returns the default configuration
func GetDefaults() *Config {
	return &Config{
		Server: ServerConfig{
			Host:           "0.0.0.0",
			Port:           8080,
			GRPCPort:       9090,
			ReadTimeout:    30 * time.Second,
			WriteTimeout:   30 * time.Second,
			MaxHeaderBytes: 1 << 20, // 1MB
			EnableTLS:      false,
			CORSEnabled:    true,
			CORSOrigins:    []string{"*"},
		},
		Database: DatabaseConfig{
			Type:            "postgres",
			Host:            "localhost",
			Port:            5432,
			Username:        "mmad",
			Password:        "",
			Database:        "mmad_engines",
			SSLMode:         "disable",
			MaxOpenConns:    25,
			MaxIdleConns:    5,
			ConnMaxLifetime: 300 * time.Second,
			MigrationPath:   "./migrations",
		},
		Logging: LoggingConfig{
			Level:  "info",
			Format: "json",
			Output: "stdout",
		},
		Crypto: CryptoConfig{
			DefaultHashFunction: "keccak256",
			KeySize:             32, // AES-256
			EncryptionEnabled:   true,
			KeyDerivationIter:   100000,
		},
		ZKProof: ZKProofConfig{
			CircuitPath:         "./circuits",
			TrustedSetupPath:    "./setup",
			ProofCacheSize:      1000,
			ProofTimeout:        5 * time.Minute,
			BatchSize:           10,
			MaxConcurrentProofs: 5,
			EnableBatching:      true,
		},
		Reserve: ReserveConfig{
			CheckInterval:       15 * time.Minute,
			MinReserveThreshold: "1000000", // $1M
			AlertThreshold:      "500000",  // $500K
			CurrencyCode:        "USD",
		},
		Compliance: ComplianceConfig{
			CheckTimeout:    30 * time.Second,
			RetryAttempts:   3,
			CacheExpiration: 24 * time.Hour,
			EnableSanctions: true,
		},
		Blockchain: BlockchainConfig{
			DefaultNetwork: "ethereum",
			Networks: map[string]NetworkConfig{
				"ethereum": {
					Name:               "Ethereum Mainnet",
					ChainID:            1,
					RPCURL:             "https://mainnet.infura.io/v3/YOUR_PROJECT_ID",
					ContractAddress:    "",
					GasLimit:           3000000,
					ConfirmationBlocks: 12,
				},
				"polygon": {
					Name:               "Polygon Mainnet",
					ChainID:            137,
					RPCURL:             "https://polygon-rpc.com",
					ContractAddress:    "",
					GasLimit:           3000000,
					ConfirmationBlocks: 20,
				},
			},
		},
		Messaging: MessagingConfig{
			Provider:    "rabbitmq",
			BrokerURL:   "amqp://localhost:5672",
			Exchange:    "mmad",
			QueuePrefix: "mmad_",
			Topics: map[string]string{
				"proof_generated":     "proofs.generated",
				"reserve_checked":     "reserves.checked",
				"compliance_verified": "compliance.verified",
			},
		},
		Monitoring: MonitoringConfig{
			Enabled:           true,
			MetricsPort:       2112,
			HealthCheckPort:   8081,
			PrometheusEnabled: true,
			JaegerEnabled:     false,
			MetricsInterval:   30 * time.Second,
		},
	}
}
