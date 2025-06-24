-- Accounts table
CREATE TABLE IF NOT EXISTS accounts (
    id VARCHAR(36) PRIMARY KEY DEFAULT gen_random_uuid()::text,
    user_id VARCHAR(36) NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    account_number VARCHAR(50) UNIQUE NOT NULL,
    account_type VARCHAR(20) NOT NULL,
    currency VARCHAR(3) NOT NULL DEFAULT 'USD',
    balance DECIMAL(28,18) NOT NULL DEFAULT 0,
    available_balance DECIMAL(28,18) NOT NULL DEFAULT 0,
    reserved_balance DECIMAL(28,18) NOT NULL DEFAULT 0,
    status VARCHAR(20) NOT NULL DEFAULT 'active',
    is_default BOOLEAN DEFAULT FALSE,
    interest_rate DECIMAL(8,6),
    overdraft_limit DECIMAL(28,18),
    daily_limit DECIMAL(28,18),
    monthly_limit DECIMAL(28,18),
    last_transaction_at TIMESTAMP WITH TIME ZONE,
    last_interest_at TIMESTAMP WITH TIME ZONE,
    frozen_at TIMESTAMP WITH TIME ZONE,
    frozen_reason TEXT,
    blockchain_address VARCHAR(42),
    metadata JSONB DEFAULT '{}',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    deleted_at TIMESTAMP WITH TIME ZONE
);

-- Indexes
CREATE INDEX idx_accounts_user_id ON accounts(user_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_accounts_account_number ON accounts(account_number) WHERE deleted_at IS NULL;
CREATE INDEX idx_accounts_type ON accounts(account_type) WHERE deleted_at IS NULL;
CREATE INDEX idx_accounts_status ON accounts(status) WHERE deleted_at IS NULL;
CREATE INDEX idx_accounts_currency ON accounts(currency) WHERE deleted_at IS NULL;
CREATE INDEX idx_accounts_balance ON accounts(balance);
CREATE INDEX idx_accounts_blockchain_address ON accounts(blockchain_address) WHERE blockchain_address IS NOT NULL;
CREATE INDEX idx_accounts_created_at ON accounts(created_at);

-- Constraints
ALTER TABLE accounts ADD CONSTRAINT chk_accounts_type 
    CHECK (account_type IN ('checking', 'savings', 'reserve', 'custodial'));

ALTER TABLE accounts ADD CONSTRAINT chk_accounts_status 
    CHECK (status IN ('active', 'frozen', 'closed', 'suspended'));

ALTER TABLE accounts ADD CONSTRAINT chk_accounts_balance_positive 
    CHECK (balance >= 0);

ALTER TABLE accounts ADD CONSTRAINT chk_accounts_available_balance 
    CHECK (available_balance >= 0);

ALTER TABLE accounts ADD CONSTRAINT chk_accounts_reserved_balance 
    CHECK (reserved_balance >= 0);

-- Ensure only one default account per user per currency
CREATE UNIQUE INDEX idx_accounts_user_default 
    ON accounts(user_id, currency) 
    WHERE is_default = TRUE AND deleted_at IS NULL;

-- Updated timestamp trigger
CREATE TRIGGER update_accounts_updated_at 
    BEFORE UPDATE ON accounts 
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();