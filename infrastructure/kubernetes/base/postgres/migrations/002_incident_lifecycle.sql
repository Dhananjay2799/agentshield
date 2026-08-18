-- ============================================================
-- AgentShield
-- 002_incident_lifecycle.sql
--
-- Adds SOC investigation and incident lifecycle metadata.
-- ============================================================

ALTER TABLE security_incidents
    ADD COLUMN IF NOT EXISTS assigned_to TEXT,

    ADD COLUMN IF NOT EXISTS investigation_note TEXT,

    ADD COLUMN IF NOT EXISTS resolution TEXT,

    ADD COLUMN IF NOT EXISTS investigating_at TIMESTAMPTZ,

    ADD COLUMN IF NOT EXISTS updated_at TIMESTAMPTZ
        NOT NULL
        DEFAULT NOW();


CREATE INDEX IF NOT EXISTS idx_security_incidents_assigned_to
    ON security_incidents (assigned_to)
    WHERE assigned_to IS NOT NULL;


CREATE INDEX IF NOT EXISTS idx_security_incidents_status_severity
    ON security_incidents (
        status,
        severity,
        last_seen_at DESC
    );