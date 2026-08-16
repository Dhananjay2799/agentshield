package models

import "time"

type SecurityEvent struct {
	EventType  string         `json:"event_type"`
	AgentID    string         `json:"agent_id"`
	SessionID  string         `json:"session_id"`
	Action     string         `json:"action"`
	Resource   string         `json:"resource"`
	Decision   string         `json:"decision"`
	RiskScore  int            `json:"risk_score"`
	Metadata   map[string]any `json:"metadata,omitempty"`
	OccurredAt time.Time      `json:"occurred_at"`
}

func (e SecurityEvent) Valid() bool {
	return e.EventType != "" &&
		e.Action != "" &&
		e.Resource != "" &&
		e.Decision != "" &&
		!e.OccurredAt.IsZero()
}
