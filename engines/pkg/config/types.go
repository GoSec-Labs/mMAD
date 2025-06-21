package config

import "time"

// Config represents the complete application configuration
type Config struct {
	Server     ServerConfig     `json:"server" yaml:"server"`
	Database   DatabaseConfig   `json:"database" yaml:"database"`
	Logging    LoggingConfig    `json:"logging" yaml:"logging"`
	Crypto     CryptoConfig     `json:"crypto" yaml:"crypto"`
	ZKProof    ZKProofConfig    `json:"zkproof" yaml:"zkproof"`
	Reserve    ReserveConfig    `json:"reserve" yaml:"reserve"`
	Compliance ComplianceConfig `json:"compliance" yaml:"compliance"`
	Blockchain BlockchainConfig `json:"blockchain" yaml:"blockchain"`
	Messaging  MessagingConfig  `json:"messaging" yaml:"messaging"`
	Monitoring MonitoringConfig `json:"monitoring" yaml:"monitoring"`
}

// ServerConfig defines HTTP/gRPC server settings
type ServerConfig struct {
	Host           string        `json:"host" yaml:"host" env:"SERVER_HOST"`
	Port           int           `json:"port" yaml:"port" env:"SERVER_PORT"`
	GRPCPort       int           `json:"grpc_port" yaml:"grpc_port" env:"GRPC_PORT"`
	ReadTimeout    time.Duration `json:"read_timeout" yaml:"read_timeout"`
	WriteTimeout   time.Duration `json:"write_timeout" yaml:"write_timeout"`
	MaxHeaderBytes int           `json:"max_header_bytes" yaml:"max_header_bytes"`
	EnableTLS      bool          `json:"enable_tls" yaml:"enable_tls" env:"ENABLE_TLS"`
	TLSCertPath    string        `json:"tls_cert_path" yaml:"tls_cert_path" env:"TLS_CERT_PATH"`
	TLSKeyPath     string        `json:"tls_key_path" yaml:"tls_key_path" env:"TLS_KEY_PATH"`
	CORSEnabled    bool          `json:"cors_enabled" yaml:"cors_enabled"`
	CORSOrigins    []string      `json:"cors_origins" yaml:"cors_origins"`
}

// DatabaseConfig defines database connection settings
type DatabaseConfig struct {
	Type            string        `json:"type" yaml:"type" env:"DB_TYPE"`
	Host            string        `json:"host" yaml:"host" env:"DB_HOST"`
	Port            int           `json:"port" yaml:"port" env:"DB_PORT"`
	Username        string        `json:"username" yaml:"username" env:"DB_USERNAME"`
	Password        string        `json:"password" yaml:"password" env:"DB_PASSWORD"`
	Database        string        `json:"database" yaml:"database" env:"DB_NAME"`
	SSLMode         string        `json:"ssl_mode" yaml:"ssl_mode" env:"DB_SSL_MODE"`
	MaxOpenConns    int           `json:"max_open_conns" yaml:"max_open_conns"`
	MaxIdleConns    int           `json:"max_idle_conns" yaml:"max_idle_conns"`
	ConnMaxLifetime time.Duration `json:"conn_max_lifetime" yaml:"conn_max_lifetime"`
	MigrationPath   string        `json:"migration_path" yaml:"migration_path"`
}

// LoggingConfig defines logging settings
type LoggingConfig struct {
	Level  string `json:"level" yaml:"level" env:"LOG_LEVEL"`
	Format string `json:"format" yaml:"format" env:"LOG_FORMAT"`
	Output string `json:"output" yaml:"output" env:"LOG_OUTPUT"`
}

// CryptoConfig defines cryptographic settings
type CryptoConfig struct {
	DefaultHashFunction string `json:"default_hash_function" yaml:"default_hash_function"`
	KeySize             int    `json:"key_size" yaml:"key_size"`
	EncryptionEnabled   bool   `json:"encryption_enabled" yaml:"encryption_enabled"`
	KeyDerivationIter   int    `json:"key_derivation_iterations" yaml:"key_derivation_iterations"`
}

// ZKProofConfig defines zero-knowledge proof settings
type ZKProofConfig struct {
	CircuitPath         string        `json:"circuit_path" yaml:"circuit_path" env:"CIRCUIT_PATH"`
	TrustedSetupPath    string        `json:"trusted_setup_path" yaml:"trusted_setup_path" env:"TRUSTED_SETUP_PATH"`
	ProofCacheSize      int           `json:"proof_cache_size" yaml:"proof_cache_size"`
	ProofTimeout        time.Duration `json:"proof_timeout" yaml:"proof_timeout"`
	BatchSize           int           `json:"batch_size" yaml:"batch_size"`
	MaxConcurrentProofs int           `json:"max_concurrent_proofs" yaml:"max_concurrent_proofs"`
	EnableBatching      bool          `json:"enable_batching" yaml:"enable_batching"`
}

// ReserveConfig defines reserve monitoring settings
type ReserveConfig struct {
	BankAPIEndpoint     string        `json:"bank_api_endpoint" yaml:"bank_api_endpoint" env:"BANK_API_ENDPOINT"`
	BankAPIKey          string        `json:"bank_api_key" yaml:"bank_api_key" env:"BANK_API_KEY"`
	CheckInterval       time.Duration `json:"check_interval" yaml:"check_interval"`
	MinReserveThreshold string        `json:"min_reserve_threshold" yaml:"min_reserve_threshold"`
	AlertThreshold      string        `json:"alert_threshold" yaml:"alert_threshold"`
	AccountNumbers      []string      `json:"account_numbers" yaml:"account_numbers"`
	CurrencyCode        string        `json:"currency_code" yaml:"currency_code"`
}

// ComplianceConfig defines compliance checking settings
type ComplianceConfig struct {
	KYCProvider       string        `json:"kyc_provider" yaml:"kyc_provider" env:"KYC_PROVIDER"`
	KYCAPIKey         string        `json:"kyc_api_key" yaml:"kyc_api_key" env:"KYC_API_KEY"`
	AMLProvider       string        `json:"aml_provider" yaml:"aml_provider" env:"AML_PROVIDER"`
	AMLAPIKey         string        `json:"aml_api_key" yaml:"aml_api_key" env:"AML_API_KEY"`
	CheckTimeout      time.Duration `json:"check_timeout" yaml:"check_timeout"`
	RetryAttempts     int           `json:"retry_attempts" yaml:"retry_attempts"`
	CacheExpiration   time.Duration `json:"cache_expiration" yaml:"cache_expiration"`
	EnableSanctions   bool          `json:"enable_sanctions" yaml:"enable_sanctions"`
	SanctionsProvider string        `json:"sanctions_provider" yaml:"sanctions_provider"`
}

// BlockchainConfig defines blockchain integration settings
type BlockchainConfig struct {
	Networks       map[string]NetworkConfig `json:"networks" yaml:"networks"`
	DefaultNetwork string                   `json:"default_network" yaml:"default_network"`
}

// NetworkConfig defines settings for a specific blockchain network
type NetworkConfig struct {
	Name               string `json:"name" yaml:"name"`
	ChainID            int64  `json:"chain_id" yaml:"chain_id"`
	RPCURL             string `json:"rpc_url" yaml:"rpc_url" env:"RPC_URL"`
	WSUrl              string `json:"ws_url" yaml:"ws_url" env:"WS_URL"`
	ContractAddress    string `json:"contract_address" yaml:"contract_address"`
	PrivateKey         string `json:"private_key" yaml:"private_key" env:"PRIVATE_KEY"`
	GasLimit           uint64 `json:"gas_limit" yaml:"gas_limit"`
	GasPrice           string `json:"gas_price" yaml:"gas_price"`
	ConfirmationBlocks int    `json:"confirmation_blocks" yaml:"confirmation_blocks"`
}

// MessagingConfig defines event messaging settings
type MessagingConfig struct {
	Provider    string            `json:"provider" yaml:"provider" env:"MESSAGE_PROVIDER"`
	BrokerURL   string            `json:"broker_url" yaml:"broker_url" env:"BROKER_URL"`
	Username    string            `json:"username" yaml:"username" env:"MESSAGE_USERNAME"`
	Password    string            `json:"password" yaml:"password" env:"MESSAGE_PASSWORD"`
	Exchange    string            `json:"exchange" yaml:"exchange"`
	QueuePrefix string            `json:"queue_prefix" yaml:"queue_prefix"`
	Topics      map[string]string `json:"topics" yaml:"topics"`
}

// MonitoringConfig defines monitoring and metrics settings
type MonitoringConfig struct {
	Enabled           bool          `json:"enabled" yaml:"enabled"`
	MetricsPort       int           `json:"metrics_port" yaml:"metrics_port"`
	HealthCheckPort   int           `json:"health_check_port" yaml:"health_check_port"`
	PrometheusEnabled bool          `json:"prometheus_enabled" yaml:"prometheus_enabled"`
	JaegerEnabled     bool          `json:"jaeger_enabled" yaml:"jaeger_enabled"`
	JaegerEndpoint    string        `json:"jaeger_endpoint" yaml:"jaeger_endpoint"`
	MetricsInterval   time.Duration `json:"metrics_interval" yaml:"metrics_interval"`
}
