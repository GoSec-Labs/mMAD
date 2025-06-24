-- Transactions table
CREATE TABLE IF NOT EXISTS transactions (
    id VARCHAR(36) PRIMARY KEY DEFAULT gen_random_uuid()::text,
    from_account_id VARCHAR(36) REFERENCES accounts(id),
    to_account_id VARCHAR(36) REFERENCES accounts(id),
    type VARCHAR(20) NOT NULL,
    amount DECIMAL(28,18) NOT NULL,
    currency VARCHAR(3) NOT NULL,
    fee DECIMAL(28,18),
    status VARCHAR(20) NOT NULL DEFAULT 'pending',
    reference VARCHAR(100) UNIQUE NOT NULL,
    external_reference VARCHAR(100),
    description TEXT,
    category VARCHAR(50),
    tags TEXT[],
    from_balance DECIMAL(28,18),
    to_balance DECIMAL(28,18),
    exchange_rate DECIMAL(18,8),
    processed_at TIMESTAMP WITH TIME ZONE,
    completed_at TIMESTAMP WITH TIME ZONE,
    failed_at TIMESTAMP WITH TIME ZONE,
    failure_reason TEXT,
    blockchain_tx_hash VARCHAR(66),
    block_number BIGINT,
    gas_used BIGINT,
    metadata JSONB DEFAULT '{}',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Indexes
CREATE INDEX idx_transactions_from_account ON transactions(from_account_id);
CREATE INDEX idx_transactions_to_account ON transactions(to_account_id);
CREATE INDEX idx_transactions_type ON transactions(type);
CREATE INDEX idx_transactions_status ON transactions(status);
CREATE INDEX idx_transactions_currency ON transactions(currency);
CREATE INDEX idx_transactions_reference ON transactions(reference);
CREATE INDEX idx_transactions_external_reference ON transactions(external_reference);
CREATE INDEX idx_transactions_amount ON transactions(amount);
CREATE INDEX idx_transactions_created_at ON transactions(created_at);
CREATE INDEX idx_transactions_completed_at ON transactions(completed_at);
CREATE INDEX idx_transactions_blockchain_tx_hash ON transactions(blockchain_tx_hash);
CREATE INDEX idx_transactions_category ON transactions(category);

-- Constraints
ALTER TABLE transactions ADD CONSTRAINT chk_transactions_type 
    CHECK (type IN ('deposit', 'withdrawal', 'transfer', 'interest', 'fee', 'refund', 'adjustment'));

ALTER TABLE transactions ADD CONSTRAINT chk_transactions_status 
    CHECK (status IN ('pending', 'processed', 'completed', 'failed', 'cancelled', 'reversed'));

ALTER TABLE transactions ADD CONSTRAINT chk_transactions_amount_positive 
    CHECK (amount > 0);

ALTER TABLE transactions ADD CONSTRAINT chk_transactions_fee_positive 
    CHECK (fee IS NULL OR fee >= 0);

-- Ensure at least one account is specified
ALTER TABLE transactions ADD CONSTRAINT chk_transactions_accounts 
    CHECK (from_account_id IS NOT NULL OR to_account_id IS NOT NULL);

-- Updated timestamp trigger
CREATE TRIGGER update_transactions_updated_at 
    BEFORE UPDATE ON transactions 
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();