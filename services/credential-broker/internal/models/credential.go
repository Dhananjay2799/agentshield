package models

import "time"

type CredentialScope struct {
	Action   string `json:"action"`
	Resource string `json:"resource"`
}

type IssueCredentialRequest struct {
	GrantID   string `json:"grant_id"`
	AgentID   string `json:"agent_id"`
	SessionID string `json:"session_id"`
	Action    string `json:"action"`
	Resource  string `json:"resource"`
}

type IssuedCredential struct {
	ID        string          `json:"id"`
	GrantID   string          `json:"grant_id"`
	AgentID   string          `json:"agent_id"`
	SessionID string          `json:"session_id"`
	Scope     CredentialScope `json:"scope"`
	Token     string          `json:"token"`
	IssuedAt  time.Time       `json:"issued_at"`
	ExpiresAt time.Time       `json:"expires_at"`
	Status    string          `json:"status"`
}
