package mlclient

type PredictionRequest struct {
	EventCount             float64 `json:"event_count"`
	DenyRatio              float64 `json:"deny_ratio"`
	HighRiskRatio          float64 `json:"high_risk_ratio"`
	ActionDiversityRatio   float64 `json:"action_diversity_ratio"`
	ResourceDiversityRatio float64 `json:"resource_diversity_ratio"`
	AverageRiskScore       float64 `json:"average_risk_score"`
	ProductionAccessRatio  float64 `json:"production_access_ratio"`
	SensitiveActionRatio   float64 `json:"sensitive_action_ratio"`
}

type PredictionResponse struct {
	ReconstructionError float64 `json:"reconstruction_error"`
	Threshold           float64 `json:"threshold"`
	IsAnomaly           bool    `json:"is_anomaly"`
	ScoreRatio          float64 `json:"score_ratio"`
	Model               string  `json:"model"`
}
