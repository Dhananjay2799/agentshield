package behavior

type Detection struct {
	Type        string
	Severity    string
	Title       string
	Description string
	Metadata    map[string]any
}

type Engine struct {
	DeniedThreshold            int
	HighRiskThreshold          int
	BurstThreshold             int
	ActionDiversityThreshold   int
	ResourceDiversityThreshold int
}

func NewEngine() *Engine {
	return &Engine{
		DeniedThreshold:            5,
		HighRiskThreshold:          3,
		BurstThreshold:             10,
		ActionDiversityThreshold:   6,
		ResourceDiversityThreshold: 8,
	}
}

func (e *Engine) Evaluate(
	event Event,
	snapshot Snapshot,
) []Detection {
	detections := make(
		[]Detection,
		0,
	)

	if snapshot.DeniedCount >= e.DeniedThreshold {
		detections = append(
			detections,
			Detection{
				Type:     "repeated_denied_actions",
				Severity: "critical",

				Title: "Repeated denied actions detected",

				Description: "Agent generated repeated denied actions within the behavioral observation window.",

				Metadata: map[string]any{
					"denied_count_window": snapshot.DeniedCount,

					"threshold": e.DeniedThreshold,

					"action": event.Action,

					"resource": event.Resource,

					"risk_score": event.RiskScore,
				},
			},
		)
	}

	if snapshot.HighRiskCount >=
		e.HighRiskThreshold {

		detections = append(
			detections,
			Detection{
				Type: "repeated_high_risk_behavior",

				Severity: "critical",

				Title: "Repeated high-risk agent behavior",

				Description: "Agent generated repeated high-risk actions within the behavioral observation window.",

				Metadata: map[string]any{
					"high_risk_count_window": snapshot.HighRiskCount,

					"threshold": e.HighRiskThreshold,

					"action": event.Action,

					"resource": event.Resource,

					"risk_score": event.RiskScore,
				},
			},
		)
	}

	if snapshot.EventCount >=
		e.BurstThreshold {

		detections = append(
			detections,
			Detection{
				Type: "agent_action_burst",

				Severity: "high",

				Title: "Abnormal agent action burst detected",

				Description: "Agent generated an unusually high number of actions within the behavioral observation window.",

				Metadata: map[string]any{
					"event_count_window": snapshot.EventCount,

					"threshold": e.BurstThreshold,

					"action": event.Action,

					"resource": event.Resource,

					"decision": event.Decision,

					"risk_score": event.RiskScore,
				},
			},
		)
	}

	if snapshot.DistinctActionCount >=
		e.ActionDiversityThreshold {

		detections = append(
			detections,
			Detection{
				Type: "high_action_diversity",

				Severity: "high",

				Title: "Unusual action diversity detected",

				Description: "Agent attempted an unusually broad set of actions within the behavioral observation window.",

				Metadata: map[string]any{
					"distinct_action_count": snapshot.DistinctActionCount,

					"threshold": e.ActionDiversityThreshold,

					"action": event.Action,

					"resource": event.Resource,
				},
			},
		)
	}

	if snapshot.DistinctResourceCount >=
		e.ResourceDiversityThreshold {

		detections = append(
			detections,
			Detection{
				Type: "high_resource_diversity",

				Severity: "high",

				Title: "Unusual resource exploration detected",

				Description: "Agent accessed an unusually broad set of resources within the behavioral observation window.",

				Metadata: map[string]any{
					"distinct_resource_count": snapshot.DistinctResourceCount,

					"threshold": e.ResourceDiversityThreshold,

					"action": event.Action,

					"resource": event.Resource,
				},
			},
		)
	}

	return detections
}
