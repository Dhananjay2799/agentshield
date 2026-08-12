package models

import "time"

type ApprovalRequest struct {
	ID          string     `json:"id"`
	AgentID     string     `json:"agent_id"`
	SessionID   string     `json:"session_id"`
	Action      string     `json:"action"`
	Resource    string     `json:"resource"`
	Reason      string     `json:"reason"`
	RiskScore   int        `json:"risk_score"`
	Status      string     `json:"status"`
	RequestedAt time.Time  `json:"requested_at"`
	ApprovedAt  *time.Time `json:"approved_at,omitempty"`
	DeniedAt    *time.Time `json:"denied_at,omitempty"`
	ExpiresAt   *time.Time `json:"expires_at,omitempty"`
}
