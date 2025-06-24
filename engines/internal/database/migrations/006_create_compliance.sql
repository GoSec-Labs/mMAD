-- Compliance checks table
CREATE TABLE IF NOT EXISTS compliance_checks (
    id VARCHAR(36) PRIMARY KEY DEFAULT gen_random_uuid()::text,
    user_id VARCHAR(36) NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    type VARCHAR(20) NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'pending',
    risk_level VARCHAR(20),
    provider VARCHAR(100),
    provider_ref VARCHAR(255),
    request_data JSONB,
    response_data JSONB,
    score DECIMAL(5,4),
    reason TEXT,
    notes TEXT,
    documents TEXT[],
    expires_at TIMESTAMP WITH TIME ZONE,
    checked_at TIMESTAMP WITH TIME ZONE,
    reviewed_at TIMESTAMP WITH TIME ZONE,
    reviewed_by VARCHAR(36) REFERENCES users(id),
    retry_count INTEGER DEFAULT 0,
    next_retry_at TIMESTAMP WITH TIME ZONE,
    metadata JSONB DEFAULT '{}',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Compliance rules table
CREATE TABLE IF NOT EXISTS compliance_rules (
    id VARCHAR(36) PRIMARY KEY DEFAULT gen_random_uuid()::text,
    name VARCHAR(255) NOT NULL,
    type VARCHAR(20) NOT NULL,
    description TEXT,
    is_enabled BOOLEAN DEFAULT TRUE,
    priority INTEGER DEFAULT 0,
    conditions JSONB NOT NULL DEFAULT '{}',
    actions JSONB NOT NULL DEFAULT '{}',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Indexes for compliance checks
CREATE INDEX idx_compliance_checks_user_id ON compliance_checks(user_id);
CREATE INDEX idx_compliance_checks_type ON compliance_checks(type);
CREATE INDEX idx_compliance_checks_status ON compliance_checks(status);
CREATE INDEX idx_compliance_checks_risk_level ON compliance_checks(risk_level);
CREATE INDEX idx_compliance_checks_provider ON compliance_checks(provider);
CREATE INDEX idx_compliance_checks_expires_at ON compliance_checks(expires_at);
CREATE INDEX idx_compliance_checks_checked_at ON compliance_checks(checked_at);
CREATE INDEX idx_compliance_checks_next_retry_at ON compliance_checks(next_retry_at) 
    WHERE status = 'failed' AND next_retry_at IS NOT NULL;
CREATE INDEX idx_compliance_checks_created_at ON compliance_checks(created_at);

-- Indexes for compliance rules
CREATE INDEX idx_compliance_rules_type ON compliance_rules(type);
CREATE INDEX idx_compliance_rules_is_enabled ON compliance_rules(is_enabled);
CREATE INDEX idx_compliance_rules_priority ON compliance_rules(priority);

-- Constraints for compliance checks
ALTER TABLE compliance_checks ADD CONSTRAINT chk_compliance_checks_type 
    CHECK (type IN ('kyc', 'aml', 'sanctions', 'pep', 'watchlist'));

ALTER TABLE compliance_checks ADD CONSTRAINT chk_compliance_checks_status 
    CHECK (status IN ('pending', 'passed', 'failed', 'review', 'expired'));

ALTER TABLE compliance_checks ADD CONSTRAINT chk_compliance_checks_risk_level 
    CHECK (risk_level IN ('low', 'medium', 'high', 'critical'));

ALTER TABLE compliance_checks ADD CONSTRAINT chk_compliance_checks_score 
    CHECK (score IS NULL OR (score >= 0 AND score <= 1));

-- Constraints for compliance rules
ALTER TABLE compliance_rules ADD CONSTRAINT chk_compliance_rules_type 
    CHECK (type IN ('kyc', 'aml', 'sanctions', 'pep', 'watchlist'));

-- Updated timestamp triggers
CREATE TRIGGER update_compliance_checks_updated_at 
    BEFORE UPDATE ON compliance_checks 
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_compliance_rules_updated_at 
    BEFORE UPDATE ON compliance_rules 
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();