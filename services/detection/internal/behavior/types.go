package behavior

import "time"

type Event struct {
	EventType  string
	AgentID    string
	SessionID  string
	Action     string
	Resource   string
	Decision   string
	RiskScore  int
	OccurredAt time.Time
}

type AgentActivity struct {
	DeniedEvents []time.Time
	HighRisk     []time.Time
	AllEvents    []time.Time
	Actions      []TimedValue
	Resources    []TimedValue
}

type TimedValue struct {
	Value string
	At    time.Time
}

type Snapshot struct {
	DeniedCount           int
	HighRiskCount         int
	EventCount            int
	DistinctActionCount   int
	DistinctResourceCount int
}
