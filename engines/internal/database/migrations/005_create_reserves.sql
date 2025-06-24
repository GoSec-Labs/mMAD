-- Reserves table
CREATE TABLE IF NOT EXISTS reserves (
    id VARCHAR(36) PRIMARY KEY DEFAULT gen_random_uuid()::text,
    name VARCHAR(255) NOT NULL,
    type VARCHAR(20) NOT NULL,
    currency VARCHAR(3) NOT NULL,
    account_number VARCHAR(100) NOT NULL,
    bank_name VARCHAR(255),
    bank_code VARCHAR(50),
    api_endpoint TEXT,
    api_credentials JSONB, -- Encrypted
    status VARCHAR(20) NOT NULL DEFAULT 'active',
    current_balance DECIMAL(28,18),
    last_balance DECIMAL(28,18),
    min_threshold DECIMAL(28,18),
    alert_threshold DECIMAL(28,18),
    max_threshold DECIMAL(28,18),
    last_checked_at TIMESTAMP WITH TIME ZONE,
    last_update_at TIMESTAMP WITH TIME ZONE,
    error_count INTEGER DEFAULT 0,
    last_error TEXT,
    check_interval INTERVAL NOT NULL DEFAULT '15 minutes',
    next_check_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    is_included BOOLEAN DEFAULT TRUE,
    weight DECIMAL(8,6) DEFAULT 1.0,
    metadata JSONB DEFAULT '{}',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    deleted_at TIMESTAMP WITH TIME ZONE
);

-- Reserve snapshots table
CREATE TABLE IF NOT EXISTS reserve_snapshots (
    id VARCHAR(36) PRIMARY KEY DEFAULT gen_random_uuid()::text,
    reserve_id VARCHAR(36) NOT NULL REFERENCES reserves(id) ON DELETE CASCADE,
    balance DECIMAL(28,18) NOT NULL,
    currency VARCHAR(3) NOT NULL,
    snapshot_time TIMESTAMP WITH TIME ZONE NOT NULL,
    source VARCHAR(50) NOT NULL,
    reference VARCHAR(255),
    proof_id VARCHAR(36) REFERENCES zk_proofs(id),
    is_verified BOOLEAN DEFAULT FALSE,
    verified_at TIMESTAMP WITH TIME ZONE,
    metadata JSONB DEFAULT '{}',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Indexes for reserves
CREATE INDEX idx_reserves_type ON reserves(type) WHERE deleted_at IS NULL;
CREATE INDEX idx_reserves_status ON reserves(status) WHERE deleted_at IS NULL;
CREATE INDEX idx_reserves_currency ON reserves(currency) WHERE deleted_at IS NULL;
CREATE INDEX idx_reserves_account_number ON reserves(account_number) WHERE deleted_at IS NULL;
CREATE INDEX idx_reserves_next_check_at ON reserves(next_check_at) WHERE status = 'active' AND deleted_at IS NULL;
CREATE INDEX idx_reserves_is_included ON reserves(is_included) WHERE deleted_at IS NULL;
CREATE INDEX idx_reserves_created_at ON reserves(created_at);

-- Indexes for snapshots
CREATE INDEX idx_snapshots_reserve_id ON reserve_snapshots(reserve_id);
CREATE INDEX idx_snapshots_snapshot_time ON reserve_snapshots(snapshot_time);
CREATE INDEX idx_snapshots_source ON reserve_snapshots(source);
CREATE INDEX idx_snapshots_is_verified ON reserve_snapshots(is_verified);
CREATE INDEX idx_snapshots_proof_id ON reserve_snapshots(proof_id);
CREATE INDEX idx_snapshots_created_at ON reserve_snapshots(created_at);

-- Constraints for reserves
ALTER TABLE reserves ADD CONSTRAINT chk_reserves_type 
   CHECK (type IN ('bank_account', 'crypto', 'commodity', 'securities'));

ALTER TABLE reserves ADD CONSTRAINT chk_reserves_status 
   CHECK (status IN ('active', 'inactive', 'unavailable', 'error'));

ALTER TABLE reserves ADD CONSTRAINT chk_reserves_balances 
   CHECK (current_balance IS NULL OR current_balance >= 0);

ALTER TABLE reserves ADD CONSTRAINT chk_reserves_thresholds 
   CHECK (min_threshold IS NULL OR min_threshold >= 0);

-- Constraints for snapshots
ALTER TABLE reserve_snapshots ADD CONSTRAINT chk_snapshots_balance_positive 
   CHECK (balance >= 0);

-- Updated timestamp trigger for reserves
CREATE TRIGGER update_reserves_updated_at 
   BEFORE UPDATE ON reserves 
   FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();