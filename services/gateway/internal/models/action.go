package models

import "time"

type ActionEvaluationRequest struct {
	Action   string `json:"action"`
	Resource string `json:"resource"`
	Reason   string `json:"reason"`
}

type DecisionTrace struct {
	Request       RequestTrace       `json:"request"`
	Risk          RiskTrace          `json:"risk"`
	Policy        PolicyTrace        `json:"policy"`
	Authorization AuthorizationTrace `json:"authorization"`
	Final         FinalTrace         `json:"final"`
}

type RequestTrace struct {
	AgentID     string `json:"agent_id"`
	AgentType   string `json:"agent_type"`
	SessionID   string `json:"session_id"`
	Action      string `json:"action"`
	Resource    string `json:"resource"`
	Environment string `json:"environment"`
}

type RiskTrace struct {
	Score  int    `json:"score"`
	Reason string `json:"reason"`
}

type PolicyTrace struct {
	Engine   string `json:"engine"`
	Matched  bool   `json:"matched"`
	ID       string `json:"id,omitempty"`
	Name     string `json:"name,omitempty"`
	Priority int    `json:"priority,omitempty"`
	Effect   string `json:"effect,omitempty"`
	Version  int    `json:"version,omitempty"`
	Source   string `json:"source,omitempty"`
	Decision string `json:"decision"`
	Reason   string `json:"reason"`
}

type AuthorizationTrace struct {
	Required   bool   `json:"required"`
	ApprovalID string `json:"approval_id,omitempty"`
	GrantUsed  bool   `json:"grant_used"`
	GrantID    string `json:"grant_id,omitempty"`
}

type FinalTrace struct {
	Decision string `json:"decision"`
	Reason   string `json:"reason"`
}

type ActionEvaluationResponse struct {
	Decision   string        `json:"decision"`
	RiskScore  int           `json:"risk_score"`
	Reason     string        `json:"reason"`
	Action     string        `json:"action"`
	Resource   string        `json:"resource"`
	AgentID    string        `json:"agent_id"`
	SessionID  string        `json:"session_id"`
	GrantID    string        `json:"grant_id,omitempty"`
	Timestamp  time.Time     `json:"timestamp"`
	ApprovalID string        `json:"approval_id,omitempty"`
	Trace      DecisionTrace `json:"trace"`
}
