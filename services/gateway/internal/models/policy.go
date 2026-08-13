package models

import "time"

type Policy struct {
	ID            string    `json:"id"`
	Name          string    `json:"name"`
	Description   string    `json:"description"`
	Effect        string    `json:"effect"`
	Status        string    `json:"status"`
	Priority      int       `json:"priority"`
	AgentType     *string   `json:"agent_type,omitempty"`
	Action        string    `json:"action"`
	ActionMatch   string    `json:"action_match"`
	Resource      string    `json:"resource"`
	ResourceMatch string    `json:"resource_match"`
	Environment   *string   `json:"environment,omitempty"`
	Version       int       `json:"version"`
	Source        string    `json:"source"`
	CreatedBy     string    `json:"created_by"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

type CreatePolicyRequest struct {
	Name          string  `json:"name"`
	Description   string  `json:"description"`
	Effect        string  `json:"effect"`
	Priority      int     `json:"priority"`
	AgentType     *string `json:"agent_type"`
	Action        string  `json:"action"`
	ActionMatch   string  `json:"action_match"`
	Resource      string  `json:"resource"`
	ResourceMatch string  `json:"resource_match"`
	Environment   *string `json:"environment"`
	CreatedBy     string  `json:"created_by"`
}

type PolicyValidationCheck struct {
	Status  string `json:"status"`
	Message string `json:"message"`
}

type PolicyValidationChecks struct {
	Schema           PolicyValidationCheck `json:"schema"`
	Matchers         PolicyValidationCheck `json:"matchers"`
	OPACompatibility PolicyValidationCheck `json:"opa_compatibility"`
}

type PolicyValidationResponse struct {
	Valid    bool                   `json:"valid"`
	PolicyID string                 `json:"policy_id"`
	Status   string                 `json:"status"`
	Checks   PolicyValidationChecks `json:"checks"`
}
