package features

import (
	"strings"
	"time"
)

type Extractor struct {
	HighRiskThreshold int
}

func NewExtractor() *Extractor {
	return &Extractor{
		HighRiskThreshold: 80,
	}
}

func (e *Extractor) Extract(
	agentID string,
	events []Event,
) Vector {
	vector := Vector{
		AgentID: agentID,
	}

	if len(events) == 0 {
		return vector
	}

	actions :=
		make(
			map[string]struct{},
		)

	resources :=
		make(
			map[string]struct{},
		)

	totalRisk := 0

	windowStart :=
		events[0].OccurredAt

	windowEnd :=
		events[0].OccurredAt

	for _, event := range events {
		vector.EventCount++

		totalRisk += event.RiskScore

		if event.RiskScore >
			vector.MaximumRiskScore {

			vector.MaximumRiskScore =
				event.RiskScore
		}

		if event.Decision == "DENY" {
			vector.DeniedCount++
		}

		if event.RiskScore >=
			e.HighRiskThreshold {

			vector.HighRiskCount++
		}

		if event.Action != "" {
			actions[event.Action] =
				struct{}{}
		}

		if event.Resource != "" {
			resources[event.Resource] =
				struct{}{}
		}

		if strings.HasPrefix(
			event.Resource,
			"production/",
		) {
			vector.ProductionAccessCount++
		}

		if isSensitiveAction(
			event.Action,
		) {
			vector.SensitiveActionCount++
		}

		if event.OccurredAt.Before(
			windowStart,
		) {
			windowStart =
				event.OccurredAt
		}

		if event.OccurredAt.After(
			windowEnd,
		) {
			windowEnd =
				event.OccurredAt
		}
	}

	vector.DistinctActionCount =
		len(actions)

	vector.DistinctResourceCount =
		len(resources)

	count :=
		float64(
			vector.EventCount,
		)

	vector.AverageRiskScore =
		float64(totalRisk) /
			count

	vector.DenyRatio =
		float64(
			vector.DeniedCount,
		) / count

	vector.HighRiskRatio =
		float64(
			vector.HighRiskCount,
		) / count

	vector.ActionDiversityRatio =
		float64(
			vector.DistinctActionCount,
		) / count

	vector.ResourceDiversityRatio =
		float64(
			vector.DistinctResourceCount,
		) / count

	vector.ProductionAccessRatio =
		float64(
			vector.ProductionAccessCount,
		) / count

	vector.SensitiveActionRatio =
		float64(
			vector.SensitiveActionCount,
		) / count

	vector.WindowStartedAt =
		windowStart

	vector.WindowEndedAt =
		windowEnd

	return vector
}

func isSensitiveAction(
	action string,
) bool {
	switch action {
	case "database.delete",
		"database.drop",
		"secrets.read",
		"iam.modify",
		"iam.admin",
		"kubernetes.delete",
		"github.push":
		return true

	default:
		return false
	}
}

func WindowDuration(
	vector Vector,
) time.Duration {
	if vector.WindowStartedAt.IsZero() ||
		vector.WindowEndedAt.IsZero() {

		return 0
	}

	return vector.WindowEndedAt.Sub(
		vector.WindowStartedAt,
	)
}
