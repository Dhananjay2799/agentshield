-- ============================================================
-- AgentShield
-- 001_initial_schema.sql
--
-- Core PostgreSQL schema for the AgentShield control plane.
-- ============================================================

CREATE EXTENSION IF NOT EXISTS pgcrypto;


-- ============================================================
-- AGENTS
-- ============================================================

CREATE TABLE IF NOT EXISTS agents (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),

    name TEXT NOT NULL,
    agent_type TEXT NOT NULL,
    owner TEXT NOT NULL,
    framework TEXT NOT NULL,
    model TEXT NOT NULL,
    environment TEXT NOT NULL DEFAULT 'development',

    status TEXT NOT NULL DEFAULT 'active',

    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_agents_created_at
    ON agents (created_at DESC);

CREATE INDEX IF NOT EXISTS idx_agents_status
    ON agents (status);


-- ============================================================
-- AGENT SESSIONS
-- ============================================================

CREATE TABLE IF NOT EXISTS agent_sessions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),

    agent_id UUID NOT NULL
        REFERENCES agents(id)
        ON DELETE CASCADE,

    task_id TEXT NOT NULL,

    status TEXT NOT NULL DEFAULT 'active',

    started_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    ended_at TIMESTAMPTZ,
    expires_at TIMESTAMPTZ
);

CREATE INDEX IF NOT EXISTS idx_agent_sessions_agent_id
    ON agent_sessions (agent_id);

CREATE INDEX IF NOT EXISTS idx_agent_sessions_agent_started
    ON agent_sessions (agent_id, started_at DESC);

CREATE INDEX IF NOT EXISTS idx_agent_sessions_status
    ON agent_sessions (status);

CREATE INDEX IF NOT EXISTS idx_agent_sessions_expires_at
    ON agent_sessions (expires_at);


-- ============================================================
-- POLICIES
-- ============================================================

CREATE TABLE IF NOT EXISTS policies (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),

    name TEXT NOT NULL,
    description TEXT NOT NULL DEFAULT '',

    effect TEXT NOT NULL,

    status TEXT NOT NULL DEFAULT 'draft',

    priority INTEGER NOT NULL DEFAULT 100,

    agent_type TEXT,

    action TEXT NOT NULL,
    action_match TEXT NOT NULL DEFAULT 'exact',

    resource TEXT NOT NULL,
    resource_match TEXT NOT NULL DEFAULT 'prefix',

    environment TEXT,

    version INTEGER NOT NULL DEFAULT 1,

    source TEXT NOT NULL DEFAULT 'control_plane',

    created_by TEXT NOT NULL DEFAULT 'soc-analyst',

    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_policies_status_priority
    ON policies (status, priority);

CREATE INDEX IF NOT EXISTS idx_policies_created_at
    ON policies (created_at DESC);


-- ============================================================
-- APPROVAL REQUESTS
-- ============================================================

CREATE TABLE IF NOT EXISTS approval_requests (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),

    agent_id UUID NOT NULL
        REFERENCES agents(id)
        ON DELETE CASCADE,

    session_id UUID NOT NULL
        REFERENCES agent_sessions(id)
        ON DELETE CASCADE,

    action TEXT NOT NULL,
    resource TEXT NOT NULL,
    reason TEXT NOT NULL DEFAULT '',

    risk_score INTEGER NOT NULL DEFAULT 0,

    status TEXT NOT NULL DEFAULT 'pending',

    requested_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    approved_at TIMESTAMPTZ,
    denied_at TIMESTAMPTZ,
    expires_at TIMESTAMPTZ
);

CREATE INDEX IF NOT EXISTS idx_approval_requests_status_expiry
    ON approval_requests (status, expires_at);

CREATE INDEX IF NOT EXISTS idx_approval_requests_agent
    ON approval_requests (agent_id, requested_at DESC);

CREATE INDEX IF NOT EXISTS idx_approval_requests_session
    ON approval_requests (session_id, requested_at);

CREATE INDEX IF NOT EXISTS idx_approval_requests_requested_at
    ON approval_requests (requested_at DESC);


-- ============================================================
-- AUTHORIZATION GRANTS
-- ============================================================

CREATE TABLE IF NOT EXISTS authorization_grants (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),

    approval_id UUID NOT NULL
        REFERENCES approval_requests(id)
        ON DELETE CASCADE,

    agent_id UUID NOT NULL
        REFERENCES agents(id)
        ON DELETE CASCADE,

    session_id UUID NOT NULL
        REFERENCES agent_sessions(id)
        ON DELETE CASCADE,

    action TEXT NOT NULL,
    resource TEXT NOT NULL,

    status TEXT NOT NULL DEFAULT 'active',

    issued_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    expires_at TIMESTAMPTZ NOT NULL,
    used_at TIMESTAMPTZ
);

CREATE INDEX IF NOT EXISTS idx_authorization_grants_active_lookup
    ON authorization_grants (
        agent_id,
        session_id,
        action,
        resource,
        status,
        expires_at
    );

CREATE INDEX IF NOT EXISTS idx_authorization_grants_approval
    ON authorization_grants (approval_id);

CREATE INDEX IF NOT EXISTS idx_authorization_grants_expiration
    ON authorization_grants (status, expires_at);

CREATE INDEX IF NOT EXISTS idx_authorization_grants_issued
    ON authorization_grants (issued_at DESC);


-- ============================================================
-- AUDIT EVENTS
-- ============================================================

CREATE TABLE IF NOT EXISTS audit_events (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),

    agent_id UUID
        REFERENCES agents(id)
        ON DELETE SET NULL,

    session_id UUID
        REFERENCES agent_sessions(id)
        ON DELETE SET NULL,

    event_type TEXT NOT NULL,

    action TEXT NOT NULL,
    resource TEXT NOT NULL,

    decision TEXT NOT NULL,

    risk_score INTEGER NOT NULL DEFAULT 0,

    metadata JSONB NOT NULL DEFAULT '{}'::jsonb,

    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_audit_events_agent_created
    ON audit_events (agent_id, created_at DESC);

CREATE INDEX IF NOT EXISTS idx_audit_events_session_created
    ON audit_events (session_id, created_at);

CREATE INDEX IF NOT EXISTS idx_audit_events_created_at
    ON audit_events (created_at DESC);

CREATE INDEX IF NOT EXISTS idx_audit_events_decision
    ON audit_events (decision);

CREATE INDEX IF NOT EXISTS idx_audit_events_metadata
    ON audit_events
    USING GIN (metadata);


-- ============================================================
-- SECURITY INCIDENTS
-- ============================================================

CREATE TABLE IF NOT EXISTS security_incidents (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),

    agent_id UUID NOT NULL
        REFERENCES agents(id)
        ON DELETE CASCADE,

    session_id UUID
        REFERENCES agent_sessions(id)
        ON DELETE SET NULL,

    incident_type TEXT NOT NULL,

    severity TEXT NOT NULL,

    title TEXT NOT NULL,
    description TEXT,

    status TEXT NOT NULL DEFAULT 'open',

    event_count INTEGER NOT NULL DEFAULT 1,

    first_seen_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    last_seen_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    resolved_at TIMESTAMPTZ,

    metadata JSONB NOT NULL DEFAULT '{}'::jsonb
);

CREATE INDEX IF NOT EXISTS idx_security_incidents_agent
    ON security_incidents (agent_id);

CREATE INDEX IF NOT EXISTS idx_security_incidents_status
    ON security_incidents (status);

CREATE INDEX IF NOT EXISTS idx_security_incidents_last_seen
    ON security_incidents (last_seen_at DESC);

CREATE INDEX IF NOT EXISTS idx_security_incidents_detection_lookup
    ON security_incidents (
        agent_id,
        incident_type,
        status,
        last_seen_at DESC
    );

CREATE INDEX IF NOT EXISTS idx_security_incidents_metadata
    ON security_incidents
    USING GIN (metadata);