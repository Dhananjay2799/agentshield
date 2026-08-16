package models

import "time"

type AuthorizationGrant struct {
	ID         string     `json:"id"`
	ApprovalID string     `json:"approval_id"`
	AgentID    string     `json:"agent_id"`
	SessionID  string     `json:"session_id"`
	Action     string     `json:"action"`
	Resource   string     `json:"resource"`
	Status     string     `json:"status"`
	IssuedAt   time.Time  `json:"issued_at"`
	ExpiresAt  time.Time  `json:"expires_at"`
	UsedAt     *time.Time `json:"used_at,omitempty"`
}

type ClaimGrantRequest struct {
	AgentID   string `json:"agent_id"`
	SessionID string `json:"session_id"`
	Action    string `json:"action"`
	Resource  string `json:"resource"`
}
