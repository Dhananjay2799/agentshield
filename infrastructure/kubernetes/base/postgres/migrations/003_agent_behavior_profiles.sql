CREATE TABLE IF NOT EXISTS agent_behavior_profiles (
    agent_id UUID PRIMARY KEY
        REFERENCES agents(id)
        ON DELETE CASCADE,

    sample_count BIGINT NOT NULL DEFAULT 0,

    mean_event_count DOUBLE PRECISION NOT NULL DEFAULT 0,

    mean_deny_ratio DOUBLE PRECISION NOT NULL DEFAULT 0,

    mean_high_risk_ratio DOUBLE PRECISION NOT NULL DEFAULT 0,

    mean_action_diversity_ratio DOUBLE PRECISION NOT NULL DEFAULT 0,

    mean_resource_diversity_ratio DOUBLE PRECISION NOT NULL DEFAULT 0,

    mean_average_risk_score DOUBLE PRECISION NOT NULL DEFAULT 0,

    mean_production_access_ratio DOUBLE PRECISION NOT NULL DEFAULT 0,

    mean_sensitive_action_ratio DOUBLE PRECISION NOT NULL DEFAULT 0,

    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_agent_behavior_profiles_updated_at
    ON agent_behavior_profiles(updated_at DESC);