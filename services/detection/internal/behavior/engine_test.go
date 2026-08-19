package behavior

import (
	"testing"
	"time"
)

func TestEngineRepeatedDeniedActions(t *testing.T) {
	engine := NewEngine()

	event := Event{
		AgentID:    "agent-1",
		Action:     "database.read",
		Resource:   "production/orders",
		Decision:   "DENY",
		RiskScore:  60,
		OccurredAt: time.Now().UTC(),
	}

	snapshot := Snapshot{
		DeniedCount:   5,
		HighRiskCount: 0,
		EventCount:    5,
	}

	detections := engine.Evaluate(
		event,
		snapshot,
	)

	if len(detections) != 1 {
		t.Fatalf(
			"expected 1 detection, got %d",
			len(detections),
		)
	}

	if detections[0].Type !=
		"repeated_denied_actions" {

		t.Fatalf(
			"expected repeated_denied_actions, got %s",
			detections[0].Type,
		)
	}

	if detections[0].Severity != "critical" {
		t.Fatalf(
			"expected critical severity, got %s",
			detections[0].Severity,
		)
	}
}

func TestEngineRepeatedHighRiskBehavior(t *testing.T) {
	engine := NewEngine()

	event := Event{
		AgentID:    "agent-2",
		Action:     "secrets.read",
		Resource:   "production/database",
		Decision:   "REQUIRE_APPROVAL",
		RiskScore:  90,
		OccurredAt: time.Now().UTC(),
	}

	snapshot := Snapshot{
		DeniedCount:   0,
		HighRiskCount: 3,
		EventCount:    3,
	}

	detections := engine.Evaluate(
		event,
		snapshot,
	)

	if len(detections) != 1 {
		t.Fatalf(
			"expected 1 detection, got %d",
			len(detections),
		)
	}

	if detections[0].Type !=
		"repeated_high_risk_behavior" {

		t.Fatalf(
			"expected repeated_high_risk_behavior, got %s",
			detections[0].Type,
		)
	}
}

func TestEngineActionBurst(t *testing.T) {
	engine := NewEngine()

	event := Event{
		AgentID:    "agent-3",
		Action:     "logs.read",
		Resource:   "production/api",
		Decision:   "ALLOW",
		RiskScore:  10,
		OccurredAt: time.Now().UTC(),
	}

	snapshot := Snapshot{
		DeniedCount:   0,
		HighRiskCount: 0,
		EventCount:    10,
	}

	detections := engine.Evaluate(
		event,
		snapshot,
	)

	if len(detections) != 1 {
		t.Fatalf(
			"expected 1 detection, got %d",
			len(detections),
		)
	}

	if detections[0].Type !=
		"agent_action_burst" {

		t.Fatalf(
			"expected agent_action_burst, got %s",
			detections[0].Type,
		)
	}
}

func TestEngineNoDetectionBelowThresholds(
	t *testing.T,
) {
	engine := NewEngine()

	event := Event{
		AgentID:    "agent-4",
		Action:     "logs.read",
		Resource:   "development/api",
		Decision:   "ALLOW",
		RiskScore:  20,
		OccurredAt: time.Now().UTC(),
	}

	snapshot := Snapshot{
		DeniedCount:   1,
		HighRiskCount: 1,
		EventCount:    4,
	}

	detections := engine.Evaluate(
		event,
		snapshot,
	)

	if len(detections) != 0 {
		t.Fatalf(
			"expected no detections, got %d",
			len(detections),
		)
	}
}

func TestEngineCanReturnMultipleDetections(
	t *testing.T,
) {
	engine := NewEngine()

	event := Event{
		AgentID:    "agent-5",
		Action:     "database.delete",
		Resource:   "production/orders",
		Decision:   "DENY",
		RiskScore:  95,
		OccurredAt: time.Now().UTC(),
	}

	snapshot := Snapshot{
		DeniedCount:   5,
		HighRiskCount: 3,
		EventCount:    10,
	}

	detections := engine.Evaluate(
		event,
		snapshot,
	)

	if len(detections) != 3 {
		t.Fatalf(
			"expected 3 detections, got %d",
			len(detections),
		)
	}
}

func TestEngineHighActionDiversity(t *testing.T) {
	engine := NewEngine()

	event := Event{
		AgentID:    "agent-diversity-actions",
		Action:     "iam.read",
		Resource:   "production/iam",
		Decision:   "ALLOW",
		RiskScore:  40,
		OccurredAt: time.Now().UTC(),
	}

	snapshot := Snapshot{
		DeniedCount:           0,
		HighRiskCount:         0,
		EventCount:            6,
		DistinctActionCount:   6,
		DistinctResourceCount: 2,
	}

	detections := engine.Evaluate(
		event,
		snapshot,
	)

	found := false

	for _, detection := range detections {
		if detection.Type == "high_action_diversity" {
			found = true

			if detection.Severity != "high" {
				t.Fatalf(
					"expected high severity, got %s",
					detection.Severity,
				)
			}
		}
	}

	if !found {
		t.Fatal(
			"expected high_action_diversity detection",
		)
	}
}

func TestEngineHighResourceDiversity(t *testing.T) {
	engine := NewEngine()

	event := Event{
		AgentID:    "agent-diversity-resources",
		Action:     "logs.read",
		Resource:   "production/service-8",
		Decision:   "ALLOW",
		RiskScore:  30,
		OccurredAt: time.Now().UTC(),
	}

	snapshot := Snapshot{
		DeniedCount:           0,
		HighRiskCount:         0,
		EventCount:            8,
		DistinctActionCount:   2,
		DistinctResourceCount: 8,
	}

	detections := engine.Evaluate(
		event,
		snapshot,
	)

	found := false

	for _, detection := range detections {
		if detection.Type == "high_resource_diversity" {
			found = true

			if detection.Severity != "high" {
				t.Fatalf(
					"expected high severity, got %s",
					detection.Severity,
				)
			}
		}
	}

	if !found {
		t.Fatal(
			"expected high_resource_diversity detection",
		)
	}
}
