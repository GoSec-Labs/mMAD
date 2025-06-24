-- ZK Proofs table
CREATE TABLE IF NOT EXISTS zk_proofs (
    id VARCHAR(36) PRIMARY KEY DEFAULT gen_random_uuid()::text,
    type VARCHAR(20) NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'pending',
    user_id VARCHAR(36) REFERENCES users(id),
    account_id VARCHAR(36) REFERENCES accounts(id),
    proof_data JSONB,
    public_inputs JSONB,
    private_inputs JSONB, -- Encrypted
    circuit_hash VARCHAR(64) NOT NULL,
    verification_key TEXT,
    proof_hash VARCHAR(64),
    merkle_root VARCHAR(64),
    block_number BIGINT,
    timestamp TIMESTAMP WITH TIME ZONE NOT NULL,
    expires_at TIMESTAMP WITH TIME ZONE,
    generated_at TIMESTAMP WITH TIME ZONE,
    verified_at TIMESTAMP WITH TIME ZONE,
    failed_at TIMESTAMP WITH TIME ZONE,
    failure_reason TEXT,
    generation_time INTERVAL,
    verification_time INTERVAL,
    metadata JSONB DEFAULT '{}',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Indexes
CREATE INDEX idx_proofs_type ON zk_proofs(type);
CREATE INDEX idx_proofs_status ON zk_proofs(status);
CREATE INDEX idx_proofs_user_id ON zk_proofs(user_id);
CREATE INDEX idx_proofs_account_id ON zk_proofs(account_id);
CREATE INDEX idx_proofs_circuit_hash ON zk_proofs(circuit_hash);
CREATE INDEX idx_proofs_proof_hash ON zk_proofs(proof_hash);
CREATE INDEX idx_proofs_merkle_root ON zk_proofs(merkle_root);
CREATE INDEX idx_proofs_timestamp ON zk_proofs(timestamp);
CREATE INDEX idx_proofs_expires_at ON zk_proofs(expires_at);
CREATE INDEX idx_proofs_created_at ON zk_proofs(created_at);

-- Constraints
ALTER TABLE zk_proofs ADD CONSTRAINT chk_proofs_type 
    CHECK (type IN ('reserve', 'solvency', 'balance', 'compliance'));

ALTER TABLE zk_proofs ADD CONSTRAINT chk_proofs_status 
    CHECK (status IN ('pending', 'generated', 'verified', 'failed', 'expired'));

-- Updated timestamp trigger
CREATE TRIGGER update_proofs_updated_at 
    BEFORE UPDATE ON zk_proofs 
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();