-- Migration: 001_init.up.sql
-- Create initial schema for Vert Helper

-- ========== Services Table ==========
CREATE TABLE services (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) NOT NULL UNIQUE,
    description TEXT,
    enabled BOOLEAN DEFAULT true,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_services_name ON services(name);
CREATE INDEX idx_services_enabled ON services(enabled);

-- ========== Service Health Table ==========
CREATE TABLE service_health (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    service_id UUID NOT NULL REFERENCES services(id) ON DELETE CASCADE,
    status VARCHAR(50) NOT NULL,
    message TEXT,
    checked_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    expires_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_service_health_service_id ON service_health(service_id);
CREATE INDEX idx_service_health_status ON service_health(status);
CREATE INDEX idx_service_health_checked_at ON service_health(checked_at);
CREATE INDEX idx_service_health_expires_at ON service_health(expires_at);

-- ========== Actions Table ==========
CREATE TABLE actions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    service_id UUID NOT NULL REFERENCES services(id) ON DELETE CASCADE,
    slug VARCHAR(255) NOT NULL,
    title VARCHAR(255) NOT NULL,
    description TEXT,
    active BOOLEAN DEFAULT true,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE UNIQUE INDEX idx_actions_slug ON actions(slug);
CREATE INDEX idx_actions_service_id ON actions(service_id);
CREATE INDEX idx_actions_active ON actions(active);

-- ========== Questions Table ==========
CREATE TABLE questions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    action_id UUID NOT NULL REFERENCES actions(id) ON DELETE CASCADE,
    slug VARCHAR(255) NOT NULL,
    input_type VARCHAR(50) NOT NULL,
    label VARCHAR(255) NOT NULL,
    placeholder VARCHAR(255),
    required BOOLEAN DEFAULT false,
    options JSONB,
    parent_id UUID REFERENCES questions(id) ON DELETE SET NULL,
    parent_value VARCHAR(255),
    "order" INTEGER DEFAULT 0,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE UNIQUE INDEX idx_questions_action_slug ON questions(action_id, slug);
CREATE INDEX idx_questions_action_id ON questions(action_id);
CREATE INDEX idx_questions_parent_id ON questions(parent_id);

-- ========== Action Executions Table ==========
CREATE TABLE action_executions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    action_id UUID NOT NULL REFERENCES actions(id) ON DELETE CASCADE,
    status VARCHAR(50) NOT NULL,
    input JSONB,
    output JSONB,
    error TEXT,
    executed_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_action_executions_action_id ON action_executions(action_id);
CREATE INDEX idx_action_executions_status ON action_executions(status);
CREATE INDEX idx_action_executions_created_at ON action_executions(created_at);

-- ========== Workers Table ==========
CREATE TABLE workers (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    service_id UUID NOT NULL REFERENCES services(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    status VARCHAR(50) NOT NULL,
    last_check TIMESTAMP WITH TIME ZONE,
    last_error TEXT,
    processed_count BIGINT DEFAULT 0,
    failed_count BIGINT DEFAULT 0,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_workers_service_id ON workers(service_id);
CREATE INDEX idx_workers_status ON workers(status);
CREATE INDEX idx_workers_name ON workers(name);

-- ========== Worker Snapshots Table ==========
CREATE TABLE worker_snapshots (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    service_id UUID NOT NULL REFERENCES services(id) ON DELETE CASCADE,
    worker_id UUID NOT NULL REFERENCES workers(id) ON DELETE CASCADE,
    status VARCHAR(50) NOT NULL,
    error_message TEXT,
    processed_count BIGINT,
    failed_count BIGINT,
    uptime_seconds BIGINT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_worker_snapshots_service_id ON worker_snapshots(service_id);
CREATE INDEX idx_worker_snapshots_worker_id ON worker_snapshots(worker_id);
CREATE INDEX idx_worker_snapshots_created_at ON worker_snapshots(created_at);
CREATE INDEX idx_worker_snapshots_status ON worker_snapshots(status);
