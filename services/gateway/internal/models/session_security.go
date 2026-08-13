package models

import "time"

type SessionSecuritySummary struct {
	ID               string     `json:"id"`
	AgentID          string     `json:"agent_id"`
	TaskID           string     `json:"task_id"`
	Status           string     `json:"status"`
	StartedAt        time.Time  `json:"started_at"`
	EndedAt          *time.Time `json:"ended_at,omitempty"`
	ExpiresAt        *time.Time `json:"expires_at,omitempty"`
	ActionCount      int        `json:"action_count"`
	AllowedCount     int        `json:"allowed_count"`
	DeniedCount      int        `json:"denied_count"`
	ApprovalCount    int        `json:"approval_count"`
	HighestRiskScore int        `json:"highest_risk_score"`
	LastActionAt     *time.Time `json:"last_action_at,omitempty"`
}
