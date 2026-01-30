-- Talos PostgreSQL Schema Migration
-- Version: 001_initial_schema.sql
-- Description: Initial database schema for production PostgreSQL deployment

-- Enable UUID extension
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- Actions table (replaces SQLite ledger)
CREATE TABLE actions (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    resource_id VARCHAR(255) NOT NULL,
    action_type VARCHAR(100) NOT NULL,
    status VARCHAR(50) NOT NULL DEFAULT 'PENDING',
    checksum VARCHAR(64) NOT NULL,
    payload JSONB,
    risk_score DECIMAL(3,1),
    estimated_savings DECIMAL(10,2),
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    started_at TIMESTAMP,
    completed_at TIMESTAMP,
    error_message TEXT,
    
    -- Indexes for performance
    CONSTRAINT actions_status_check CHECK (status IN ('PENDING', 'IN_PROGRESS', 'COMPLETED', 'FAILED', 'ROLLED_BACK'))
);

CREATE INDEX idx_actions_status ON actions(status);
CREATE INDEX idx_actions_resource ON actions(resource_id);
CREATE INDEX idx_actions_created ON actions(created_at DESC);
CREATE INDEX idx_actions_checksum ON actions(checksum);

-- AI decisions table (for memory/context)
CREATE TABLE ai_decisions (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    resource_id VARCHAR(255) NOT NULL,
    model VARCHAR(100) NOT NULL,
    decision TEXT NOT NULL,
    reasoning TEXT,
    confidence DECIMAL(3,2),
    tokens_used INTEGER,
    latency_ms INTEGER,
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_ai_decisions_resource ON ai_decisions(resource_id);
CREATE INDEX idx_ai_decisions_created ON ai_decisions(created_at DESC);

-- Token tracking table
CREATE TABLE token_usage (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    model VARCHAR(100) NOT NULL,
    tokens INTEGER NOT NULL,
    cost_usd DECIMAL(10,4) NOT NULL,
    request_type VARCHAR(50),
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_token_usage_model ON token_usage(model);
CREATE INDEX idx_token_usage_created ON token_usage(created_at DESC);

-- Savings tracking table
CREATE TABLE savings_events (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    action_id UUID REFERENCES actions(id),
    resource_id VARCHAR(255) NOT NULL,
    optimization_type VARCHAR(100),
    estimated_savings DECIMAL(10,2),
    actual_savings DECIMAL(10,2),
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_savings_action ON savings_events(action_id);
CREATE INDEX idx_savings_created ON savings_events(created_at DESC);

-- Organizations table (for multi-tenancy)
CREATE TABLE organizations (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR(255) NOT NULL UNIQUE,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    settings JSONB
);

-- Users table (for RBAC)
CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    email VARCHAR(255) NOT NULL UNIQUE,
    org_id UUID REFERENCES organizations(id),
    role VARCHAR(50) NOT NULL DEFAULT 'viewer',
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    last_login TIMESTAMP,
    
    CONSTRAINT users_role_check CHECK (role IN ('admin', 'operator', 'viewer'))
);

CREATE INDEX idx_users_org ON users(org_id);
CREATE INDEX idx_users_email ON users(email);

-- Resources table (cloud resource inventory)
CREATE TABLE resources (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    org_id UUID REFERENCES organizations(id),
    cloud_resource_id VARCHAR(255) NOT NULL,
    cloud_provider VARCHAR(50) NOT NULL,
    resource_type VARCHAR(100),
    region VARCHAR(50),
    tags JSONB,
    metadata JSONB,
    monthly_cost DECIMAL(10,2),
    last_scanned TIMESTAMP,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    
    UNIQUE(org_id, cloud_resource_id, cloud_provider)
);

CREATE INDEX idx_resources_org ON resources(org_id);
CREATE INDEX idx_resources_provider ON resources(cloud_provider);
CREATE INDEX idx_resources_scanned ON resources(last_scanned DESC);

-- Audit log table
CREATE TABLE audit_log (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID REFERENCES users(id),
    action VARCHAR(100) NOT NULL,
    resource_type VARCHAR(100),
    resource_id VARCHAR(255),
    details JSONB,
    ip_address INET,
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_audit_user ON audit_log(user_id);
CREATE INDEX idx_audit_created ON audit_log(created_at DESC);

-- Create default organization for migration
INSERT INTO organizations (name, settings) 
VALUES ('Default', '{"mode": "personal", "risk_threshold": 5.0}'::jsonb);

-- Comments for documentation
COMMENT ON TABLE actions IS 'Idempotent action ledger for crash-safe execution';
COMMENT ON TABLE ai_decisions IS 'AI model decisions and reasoning for audit trail';
COMMENT ON TABLE token_usage IS 'Token consumption tracking for ROI calculation';
COMMENT ON TABLE savings_events IS 'Actual vs estimated savings tracking';
