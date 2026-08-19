package features

import (
	"math"
	"testing"
	"time"
)

func TestExtractorBuildsBehaviorVector(
	t *testing.T,
) {
	extractor :=
		NewExtractor()

	base :=
		time.Date(
			2026,
			time.August,
			19,
			12,
			0,
			0,
			0,
			time.UTC,
		)

	events := []Event{
		{
			AgentID:    "agent-1",
			Action:     "logs.read",
			Resource:   "production/api",
			Decision:   "ALLOW",
			RiskScore:  10,
			OccurredAt: base,
		},
		{
			AgentID:    "agent-1",
			Action:     "database.read",
			Resource:   "production/database",
			Decision:   "DENY",
			RiskScore:  90,
			OccurredAt: base.Add(10 * time.Second),
		},
		{
			AgentID:    "agent-1",
			Action:     "secrets.read",
			Resource:   "production/secrets",
			Decision:   "DENY",
			RiskScore:  95,
			OccurredAt: base.Add(20 * time.Second),
		},
	}

	vector :=
		extractor.Extract(
			"agent-1",
			events,
		)

	if vector.EventCount != 3 {
		t.Fatalf(
			"expected 3 events, got %d",
			vector.EventCount,
		)
	}

	if vector.DeniedCount != 2 {
		t.Fatalf(
			"expected 2 denied events, got %d",
			vector.DeniedCount,
		)
	}

	if vector.HighRiskCount != 2 {
		t.Fatalf(
			"expected 2 high-risk events, got %d",
			vector.HighRiskCount,
		)
	}

	if vector.DistinctActionCount != 3 {
		t.Fatalf(
			"expected 3 distinct actions, got %d",
			vector.DistinctActionCount,
		)
	}

	if vector.DistinctResourceCount != 3 {
		t.Fatalf(
			"expected 3 distinct resources, got %d",
			vector.DistinctResourceCount,
		)
	}

	if vector.ProductionAccessCount != 3 {
		t.Fatalf(
			"expected 3 production accesses, got %d",
			vector.ProductionAccessCount,
		)
	}

	if vector.SensitiveActionCount != 1 {
		t.Fatalf(
			"expected 1 sensitive action, got %d",
			vector.SensitiveActionCount,
		)
	}

	expectedAverage :=
		float64(10+90+95) / 3

	if math.Abs(
		vector.AverageRiskScore-
			expectedAverage,
	) > 0.001 {

		t.Fatalf(
			"expected average risk %.2f, got %.2f",
			expectedAverage,
			vector.AverageRiskScore,
		)
	}

	if vector.MaximumRiskScore != 95 {
		t.Fatalf(
			"expected max risk 95, got %d",
			vector.MaximumRiskScore,
		)
	}
}

func TestExtractorHandlesEmptyEvents(
	t *testing.T,
) {
	extractor :=
		NewExtractor()

	vector :=
		extractor.Extract(
			"agent-empty",
			nil,
		)

	if vector.EventCount != 0 {
		t.Fatalf(
			"expected zero events, got %d",
			vector.EventCount,
		)
	}

	if vector.AgentID !=
		"agent-empty" {

		t.Fatalf(
			"unexpected agent ID %s",
			vector.AgentID,
		)
	}
}

func TestWindowDuration(
	t *testing.T,
) {
	start :=
		time.Now().UTC()

	vector :=
		Vector{
			WindowStartedAt: start,

			WindowEndedAt: start.Add(
				45 * time.Second,
			),
		}

	duration :=
		WindowDuration(
			vector,
		)

	if duration !=
		45*time.Second {

		t.Fatalf(
			"expected 45s duration, got %s",
			duration,
		)
	}
}
