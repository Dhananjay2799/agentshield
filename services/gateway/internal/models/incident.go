package models

import (
	"encoding/json"
	"time"
)

type SecurityIncident struct {
	ID           string          `json:"id"`
	AgentID      string          `json:"agent_id"`
	SessionID    *string         `json:"session_id,omitempty"`
	IncidentType string          `json:"incident_type"`
	Severity     string          `json:"severity"`
	Title        string          `json:"title"`
	Description  *string         `json:"description,omitempty"`
	Status       string          `json:"status"`
	EventCount   int             `json:"event_count"`
	FirstSeenAt  time.Time       `json:"first_seen_at"`
	LastSeenAt   time.Time       `json:"last_seen_at"`
	CreatedAt    time.Time       `json:"created_at"`
	ResolvedAt   *time.Time      `json:"resolved_at,omitempty"`
	Metadata     json.RawMessage `json:"metadata"`
}

type IncidentStatusResponse struct {
	ID         string     `json:"id"`
	Status     string     `json:"status"`
	UpdatedAt  time.Time  `json:"updated_at"`
	ResolvedAt *time.Time `json:"resolved_at,omitempty"`
}
