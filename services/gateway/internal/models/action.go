package models

import "time"

// ActionEvaluationRequest represents an action an AI agent wants to perform.
type ActionEvaluationRequest struct {
	Action   string `json:"action"`
	Resource string `json:"resource"`
	Reason   string `json:"reason"`
	GrantID  string `json:"grant_id,omitempty"`
}

// ActionEvaluationResponse represents AgentShield's security decision.
type ActionEvaluationResponse struct {
	Decision   string    `json:"decision"`
	RiskScore  int       `json:"risk_score"`
	Reason     string    `json:"reason"`
	Action     string    `json:"action"`
	Resource   string    `json:"resource"`
	AgentID    string    `json:"agent_id"`
	SessionID  string    `json:"session_id"`
	GrantID    string    `json:"grant_id,omitempty"`
	Timestamp  time.Time `json:"timestamp"`
	ApprovalID string    `json:"approval_id,omitempty"`
}
