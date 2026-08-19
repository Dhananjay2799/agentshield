package baseline

import (
	"math"
	"sync"
)

type Store struct {
	mu       sync.RWMutex
	profiles map[string]*Profile
}

func NewStore() *Store {
	return &Store{
		profiles: make(
			map[string]*Profile,
		),
	}
}

func (s *Store) Observe(
	agentID string,
	observation Observation,
) Profile {
	s.mu.Lock()
	defer s.mu.Unlock()

	profile, exists :=
		s.profiles[agentID]

	if !exists {
		profile =
			&Profile{
				AgentID: agentID,
			}

		s.profiles[agentID] =
			profile
	}

	profile.SampleCount++

	count :=
		float64(
			profile.SampleCount,
		)

	profile.Mean.EventCount =
		updateMean(
			profile.Mean.EventCount,
			observation.EventCount,
			count,
		)

	profile.Mean.DenyRatio =
		updateMean(
			profile.Mean.DenyRatio,
			observation.DenyRatio,
			count,
		)

	profile.Mean.HighRiskRatio =
		updateMean(
			profile.Mean.HighRiskRatio,
			observation.HighRiskRatio,
			count,
		)

	profile.Mean.ActionDiversityRatio =
		updateMean(
			profile.Mean.ActionDiversityRatio,
			observation.ActionDiversityRatio,
			count,
		)

	profile.Mean.ResourceDiversityRatio =
		updateMean(
			profile.Mean.ResourceDiversityRatio,
			observation.ResourceDiversityRatio,
			count,
		)

	profile.Mean.AverageRiskScore =
		updateMean(
			profile.Mean.AverageRiskScore,
			observation.AverageRiskScore,
			count,
		)

	profile.Mean.ProductionAccessRatio =
		updateMean(
			profile.Mean.ProductionAccessRatio,
			observation.ProductionAccessRatio,
			count,
		)

	profile.Mean.SensitiveActionRatio =
		updateMean(
			profile.Mean.SensitiveActionRatio,
			observation.SensitiveActionRatio,
			count,
		)

	return *profile
}

func (s *Store) Get(
	agentID string,
) (Profile, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	profile, exists :=
		s.profiles[agentID]

	if !exists {
		return Profile{}, false
	}

	return *profile, true
}

func (s *Store) Restore(
	profile Profile,
) {
	s.mu.Lock()
	defer s.mu.Unlock()

	copyProfile := profile

	s.profiles[profile.AgentID] =
		&copyProfile
}

func Compare(
	profile Profile,
	observation Observation,
) Deviation {
	return Deviation{
		EventCount: absoluteDifference(
			observation.EventCount,
			profile.Mean.EventCount,
		),

		DenyRatio: absoluteDifference(
			observation.DenyRatio,
			profile.Mean.DenyRatio,
		),

		HighRiskRatio: absoluteDifference(
			observation.HighRiskRatio,
			profile.Mean.HighRiskRatio,
		),

		ActionDiversityRatio: absoluteDifference(
			observation.ActionDiversityRatio,
			profile.Mean.ActionDiversityRatio,
		),

		ResourceDiversityRatio: absoluteDifference(
			observation.ResourceDiversityRatio,
			profile.Mean.ResourceDiversityRatio,
		),

		AverageRiskScore: absoluteDifference(
			observation.AverageRiskScore,
			profile.Mean.AverageRiskScore,
		),

		ProductionAccessRatio: absoluteDifference(
			observation.ProductionAccessRatio,
			profile.Mean.ProductionAccessRatio,
		),

		SensitiveActionRatio: absoluteDifference(
			observation.SensitiveActionRatio,
			profile.Mean.SensitiveActionRatio,
		),
	}
}

func updateMean(
	current float64,
	value float64,
	count float64,
) float64 {
	return current +
		(value-current)/count
}

func absoluteDifference(
	value float64,
	baseline float64,
) float64 {
	return math.Abs(
		value - baseline,
	)
}

func ScoreObservation(
	profile Profile,
	observation Observation,
	minSamples int64,
) Score {
	score := Score{
		SampleCount: profile.SampleCount,
		WarmedUp:    profile.SampleCount >= minSamples,
		Explanation: make(
			map[string]float64,
		),
	}

	if !score.WarmedUp {
		return score
	}

	deviation :=
		Compare(
			profile,
			observation,
		)

	normalizedEventCount :=
		normalizeDifference(
			deviation.EventCount,
			profile.Mean.EventCount,
		)

	normalizedAverageRisk :=
		normalizeDifference(
			deviation.AverageRiskScore,
			profile.Mean.AverageRiskScore,
		)

	score.Explanation["event_count"] =
		normalizedEventCount

	score.Explanation["deny_ratio"] =
		deviation.DenyRatio

	score.Explanation["high_risk_ratio"] =
		deviation.HighRiskRatio

	score.Explanation["action_diversity_ratio"] =
		deviation.ActionDiversityRatio

	score.Explanation["resource_diversity_ratio"] =
		deviation.ResourceDiversityRatio

	score.Explanation["average_risk_score"] =
		normalizedAverageRisk

	score.Explanation["production_access_ratio"] =
		deviation.ProductionAccessRatio

	score.Explanation["sensitive_action_ratio"] =
		deviation.SensitiveActionRatio

	score.Value =
		normalizedEventCount*0.15 +
			deviation.DenyRatio*0.20 +
			deviation.HighRiskRatio*0.20 +
			deviation.ActionDiversityRatio*0.10 +
			deviation.ResourceDiversityRatio*0.10 +
			normalizedAverageRisk*0.10 +
			deviation.ProductionAccessRatio*0.05 +
			deviation.SensitiveActionRatio*0.10

	score.IsAnomalous =
		score.Value >= 0.50

	return score
}

func normalizeDifference(
	difference float64,
	baseline float64,
) float64 {
	if baseline <= 0 {
		if difference <= 0 {
			return 0
		}

		return 1
	}

	value :=
		difference / baseline

	if value > 1 {
		return 1
	}

	return value
}
