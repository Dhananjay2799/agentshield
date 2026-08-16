package models

import (
	"encoding/json"
	"time"
)

type AuditEvent struct {
	ID        string          `json:"id"`
	AgentID   *string         `json:"agent_id,omitempty"`
	SessionID *string         `json:"session_id,omitempty"`
	EventType string          `json:"event_type"`
	Action    string          `json:"action"`
	Resource  string          `json:"resource"`
	Decision  string          `json:"decision"`
	RiskScore int             `json:"risk_score"`
	Metadata  json.RawMessage `json:"metadata"`
	CreatedAt time.Time       `json:"created_at"`
}
