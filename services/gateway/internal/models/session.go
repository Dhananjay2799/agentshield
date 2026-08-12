package models

import "time"

// AgentSession represents a short-lived execution identity for an agent.
type AgentSession struct {
	ID        string     `json:"id"`
	AgentID   string     `json:"agent_id"`
	TaskID    string     `json:"task_id"`
	Status    string     `json:"status"`
	StartedAt time.Time  `json:"started_at"`
	EndedAt   *time.Time `json:"ended_at,omitempty"`
	ExpiresAt *time.Time `json:"expires_at,omitempty"`
}

// CreateSessionRequest represents the payload for creating a new agent session.
type CreateSessionRequest struct {
	TaskID     string `json:"task_id"`
	TTLMinutes int    `json:"ttl_minutes"`
}
