package behavior

import (
	"sync"
	"time"
)

type Tracker struct {
	mu       sync.Mutex
	window   time.Duration
	activity map[string]*AgentActivity
}

func NewTracker(
	window time.Duration,
) *Tracker {
	return &Tracker{
		window:   window,
		activity: make(map[string]*AgentActivity),
	}
}

func (t *Tracker) Record(
	event Event,
) Snapshot {
	t.mu.Lock()
	defer t.mu.Unlock()

	now := event.OccurredAt

	if now.IsZero() {
		now = time.Now().UTC()
	}

	windowStart :=
		now.Add(-t.window)

	activity, exists :=
		t.activity[event.AgentID]

	if !exists {
		activity =
			&AgentActivity{}

		t.activity[event.AgentID] =
			activity
	}

	activity.DeniedEvents =
		pruneOld(
			activity.DeniedEvents,
			windowStart,
		)

	activity.HighRisk =
		pruneOld(
			activity.HighRisk,
			windowStart,
		)

	activity.AllEvents =
		pruneOld(
			activity.AllEvents,
			windowStart,
		)
	event.OccurredAt = now

	activity.Events =
		append(
			activity.Events,
			event,
		)

	activity.Actions =
		pruneTimedValues(
			activity.Actions,
			windowStart,
		)

	activity.Resources =
		pruneTimedValues(
			activity.Resources,
			windowStart,
		)

	activity.Events =
		pruneEvents(
			activity.Events,
			windowStart,
		)

	activity.AllEvents =
		append(
			activity.AllEvents,
			now,
		)

	if event.Action != "" {
		activity.Actions =
			append(
				activity.Actions,
				TimedValue{
					Value: event.Action,
					At:    now,
				},
			)
	}

	if event.Resource != "" {
		activity.Resources =
			append(
				activity.Resources,
				TimedValue{
					Value: event.Resource,
					At:    now,
				},
			)
	}

	if event.Decision == "DENY" {
		activity.DeniedEvents =
			append(
				activity.DeniedEvents,
				now,
			)
	}

	if event.RiskScore >= 80 {
		activity.HighRisk =
			append(
				activity.HighRisk,
				now,
			)
	}

	return Snapshot{
		DeniedCount: len(
			activity.DeniedEvents,
		),

		HighRiskCount: len(
			activity.HighRisk,
		),

		EventCount: len(
			activity.AllEvents,
		),

		DistinctActionCount: countDistinct(
			activity.Actions,
		),

		DistinctResourceCount: countDistinct(
			activity.Resources,
		),
	}
}

func pruneOld(
	timestamps []time.Time,
	windowStart time.Time,
) []time.Time {
	result :=
		timestamps[:0]

	for _, timestamp := range timestamps {

		if timestamp.After(
			windowStart,
		) {
			result =
				append(
					result,
					timestamp,
				)
		}
	}

	return result
}

func pruneTimedValues(
	values []TimedValue,
	windowStart time.Time,
) []TimedValue {
	result := values[:0]

	for _, value := range values {
		if value.At.After(windowStart) {
			result = append(
				result,
				value,
			)
		}
	}

	return result
}

func countDistinct(
	values []TimedValue,
) int {
	seen := make(
		map[string]struct{},
	)

	for _, value := range values {
		if value.Value == "" {
			continue
		}

		seen[value.Value] =
			struct{}{}
	}

	return len(seen)
}

func pruneEvents(
	events []Event,
	windowStart time.Time,
) []Event {
	result := events[:0]

	for _, event := range events {
		if event.OccurredAt.After(windowStart) {
			result = append(
				result,
				event,
			)
		}
	}

	return result
}

func (t *Tracker) Window(
	agentID string,
) []Event {
	t.mu.Lock()
	defer t.mu.Unlock()

	activity, exists :=
		t.activity[agentID]

	if !exists ||
		len(activity.Events) == 0 {

		return []Event{}
	}

	events :=
		make(
			[]Event,
			len(activity.Events),
		)

	copy(
		events,
		activity.Events,
	)

	return events
}
