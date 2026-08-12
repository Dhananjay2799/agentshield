package models

import "time"

// Agent represents an autonomous AI agent registered with AgentShield.
type Agent struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	AgentType   string    `json:"agent_type"`
	Owner       string    `json:"owner"`
	Framework   string    `json:"framework"`
	Model       string    `json:"model"`
	Environment string    `json:"environment"`
	Status      string    `json:"status"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// CreateAgentRequest represents the payload used to register a new agent.
type CreateAgentRequest struct {
	Name        string `json:"name"`
	AgentType   string `json:"agent_type"`
	Owner       string `json:"owner"`
	Framework   string `json:"framework"`
	Model       string `json:"model"`
	Environment string `json:"environment"`
}
