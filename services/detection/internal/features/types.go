package features

import "time"

type Event struct {
	AgentID    string
	SessionID  string
	Action     string
	Resource   string
	Decision   string
	RiskScore  int
	OccurredAt time.Time
}

type Vector struct {
	AgentID string `json:"agent_id"`

	EventCount int `json:"event_count"`

	DeniedCount int `json:"denied_count"`

	HighRiskCount int `json:"high_risk_count"`

	DistinctActionCount int `json:"distinct_action_count"`

	DistinctResourceCount int `json:"distinct_resource_count"`

	AverageRiskScore float64 `json:"average_risk_score"`

	MaximumRiskScore int `json:"maximum_risk_score"`

	DenyRatio float64 `json:"deny_ratio"`

	HighRiskRatio float64 `json:"high_risk_ratio"`

	ActionDiversityRatio float64 `json:"action_diversity_ratio"`

	ResourceDiversityRatio float64 `json:"resource_diversity_ratio"`

	ProductionAccessCount int `json:"production_access_count"`

	ProductionAccessRatio float64 `json:"production_access_ratio"`

	SensitiveActionCount int `json:"sensitive_action_count"`

	SensitiveActionRatio float64 `json:"sensitive_action_ratio"`

	WindowStartedAt time.Time `json:"window_started_at"`

	WindowEndedAt time.Time `json:"window_ended_at"`
}
