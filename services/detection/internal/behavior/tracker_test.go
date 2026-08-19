package behavior

import (
	"testing"
	"time"
)

func TestTrackerCountsEventsInsideWindow(
	t *testing.T,
) {
	tracker :=
		NewTracker(
			60 * time.Second,
		)

	base :=
		time.Date(
			2026,
			time.August,
			18,
			12,
			0,
			0,
			0,
			time.UTC,
		)

	event := Event{
		AgentID:    "agent-1",
		Decision:   "ALLOW",
		RiskScore:  10,
		OccurredAt: base,
	}

	first :=
		tracker.Record(
			event,
		)

	if first.EventCount != 1 {
		t.Fatalf(
			"expected event count 1, got %d",
			first.EventCount,
		)
	}

	event.OccurredAt =
		base.Add(
			10 * time.Second,
		)

	second :=
		tracker.Record(
			event,
		)

	if second.EventCount != 2 {
		t.Fatalf(
			"expected event count 2, got %d",
			second.EventCount,
		)
	}
}

func TestTrackerCountsDeniedAndHighRisk(
	t *testing.T,
) {
	tracker :=
		NewTracker(
			60 * time.Second,
		)

	snapshot :=
		tracker.Record(
			Event{
				AgentID:    "agent-2",
				Decision:   "DENY",
				RiskScore:  90,
				OccurredAt: time.Now().UTC(),
			},
		)

	if snapshot.DeniedCount != 1 {
		t.Fatalf(
			"expected denied count 1, got %d",
			snapshot.DeniedCount,
		)
	}

	if snapshot.HighRiskCount != 1 {
		t.Fatalf(
			"expected high-risk count 1, got %d",
			snapshot.HighRiskCount,
		)
	}
}

func TestTrackerPrunesExpiredEvents(
	t *testing.T,
) {
	tracker :=
		NewTracker(
			60 * time.Second,
		)

	base :=
		time.Date(
			2026,
			time.August,
			18,
			12,
			0,
			0,
			0,
			time.UTC,
		)

	tracker.Record(
		Event{
			AgentID:    "agent-3",
			Decision:   "DENY",
			RiskScore:  90,
			OccurredAt: base,
		},
	)

	snapshot :=
		tracker.Record(
			Event{
				AgentID:   "agent-3",
				Decision:  "ALLOW",
				RiskScore: 10,
				OccurredAt: base.Add(
					61 * time.Second,
				),
			},
		)

	if snapshot.EventCount != 1 {
		t.Fatalf(
			"expected 1 event after pruning, got %d",
			snapshot.EventCount,
		)
	}

	if snapshot.DeniedCount != 0 {
		t.Fatalf(
			"expected denied count 0 after pruning, got %d",
			snapshot.DeniedCount,
		)
	}

	if snapshot.HighRiskCount != 0 {
		t.Fatalf(
			"expected high-risk count 0 after pruning, got %d",
			snapshot.HighRiskCount,
		)
	}
}

func TestTrackerKeepsAgentsIndependent(
	t *testing.T,
) {
	tracker :=
		NewTracker(
			60 * time.Second,
		)

	now := time.Now().UTC()

	tracker.Record(
		Event{
			AgentID:    "agent-a",
			Decision:   "DENY",
			RiskScore:  90,
			OccurredAt: now,
		},
	)

	snapshot :=
		tracker.Record(
			Event{
				AgentID:    "agent-b",
				Decision:   "ALLOW",
				RiskScore:  10,
				OccurredAt: now,
			},
		)

	if snapshot.EventCount != 1 {
		t.Fatalf(
			"expected agent-b event count 1, got %d",
			snapshot.EventCount,
		)
	}

	if snapshot.DeniedCount != 0 {
		t.Fatalf(
			"expected agent-b denied count 0, got %d",
			snapshot.DeniedCount,
		)
	}
}

func TestTrackerCountsDistinctActionsAndResources(
	t *testing.T,
) {
	tracker :=
		NewTracker(
			60 * time.Second,
		)

	base := time.Now().UTC()

	events := []Event{
		{
			AgentID:    "agent-diversity",
			Action:     "logs.read",
			Resource:   "production/api",
			Decision:   "ALLOW",
			RiskScore:  10,
			OccurredAt: base,
		},
		{
			AgentID:    "agent-diversity",
			Action:     "database.read",
			Resource:   "production/db",
			Decision:   "ALLOW",
			RiskScore:  20,
			OccurredAt: base.Add(5 * time.Second),
		},
		{
			AgentID:    "agent-diversity",
			Action:     "logs.read",
			Resource:   "production/worker",
			Decision:   "ALLOW",
			RiskScore:  10,
			OccurredAt: base.Add(10 * time.Second),
		},
	}

	var snapshot Snapshot

	for _, event := range events {
		snapshot = tracker.Record(event)
	}

	if snapshot.EventCount != 3 {
		t.Fatalf(
			"expected event count 3, got %d",
			snapshot.EventCount,
		)
	}

	if snapshot.DistinctActionCount != 2 {
		t.Fatalf(
			"expected distinct action count 2, got %d",
			snapshot.DistinctActionCount,
		)
	}

	if snapshot.DistinctResourceCount != 3 {
		t.Fatalf(
			"expected distinct resource count 3, got %d",
			snapshot.DistinctResourceCount,
		)
	}
}
