package baseline

type Observation struct {
	EventCount             float64
	DenyRatio              float64
	HighRiskRatio          float64
	ActionDiversityRatio   float64
	ResourceDiversityRatio float64
	AverageRiskScore       float64
	ProductionAccessRatio  float64
	SensitiveActionRatio   float64
}

type Profile struct {
	AgentID string

	SampleCount int64

	Mean Observation
}

type Deviation struct {
	EventCount             float64
	DenyRatio              float64
	HighRiskRatio          float64
	ActionDiversityRatio   float64
	ResourceDiversityRatio float64
	AverageRiskScore       float64
	ProductionAccessRatio  float64
	SensitiveActionRatio   float64
}

type Score struct {
	Value       float64
	IsAnomalous bool
	WarmedUp    bool
	SampleCount int64
	Explanation map[string]float64
}
