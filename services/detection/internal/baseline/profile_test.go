package baseline

import (
	"math"
	"testing"
)

func TestStoreBuildsAgentBaseline(
	t *testing.T,
) {
	store := NewStore()

	store.Observe(
		"agent-1",
		Observation{
			EventCount:       10,
			DenyRatio:        0.10,
			HighRiskRatio:    0.10,
			AverageRiskScore: 20,
		},
	)

	profile :=
		store.Observe(
			"agent-1",
			Observation{
				EventCount:       20,
				DenyRatio:        0.30,
				HighRiskRatio:    0.20,
				AverageRiskScore: 40,
			},
		)

	if profile.SampleCount != 2 {
		t.Fatalf(
			"expected 2 samples, got %d",
			profile.SampleCount,
		)
	}

	if profile.Mean.EventCount != 15 {
		t.Fatalf(
			"expected mean event count 15, got %.2f",
			profile.Mean.EventCount,
		)
	}

	if math.Abs(
		profile.Mean.DenyRatio-0.20,
	) > 0.0001 {
		t.Fatalf(
			"expected mean deny ratio 0.20, got %.4f",
			profile.Mean.DenyRatio,
		)
	}

	if profile.Mean.AverageRiskScore != 30 {
		t.Fatalf(
			"expected average risk 30, got %.2f",
			profile.Mean.AverageRiskScore,
		)
	}
}

func TestStoreKeepsAgentsIndependent(
	t *testing.T,
) {
	store := NewStore()

	store.Observe(
		"agent-a",
		Observation{
			EventCount: 10,
		},
	)

	store.Observe(
		"agent-b",
		Observation{
			EventCount: 50,
		},
	)

	a, ok :=
		store.Get(
			"agent-a",
		)

	if !ok {
		t.Fatal(
			"expected agent-a profile",
		)
	}

	b, ok :=
		store.Get(
			"agent-b",
		)

	if !ok {
		t.Fatal(
			"expected agent-b profile",
		)
	}

	if a.Mean.EventCount == b.Mean.EventCount {
		t.Fatal(
			"expected independent baselines",
		)
	}
}

func TestCompareReturnsDeviation(
	t *testing.T,
) {
	profile :=
		Profile{
			AgentID:     "agent-1",
			SampleCount: 10,

			Mean: Observation{
				EventCount:       10,
				DenyRatio:        0.10,
				AverageRiskScore: 20,
			},
		}

	observation :=
		Observation{
			EventCount:       25,
			DenyRatio:        0.60,
			AverageRiskScore: 70,
		}

	deviation :=
		Compare(
			profile,
			observation,
		)

	if deviation.EventCount != 15 {
		t.Fatalf(
			"expected event-count deviation 15, got %.2f",
			deviation.EventCount,
		)
	}

	if math.Abs(
		deviation.DenyRatio-0.50,
	) > 0.0001 {
		t.Fatalf(
			"expected deny-ratio deviation 0.50, got %.4f",
			deviation.DenyRatio,
		)
	}

	if deviation.AverageRiskScore != 50 {
		t.Fatalf(
			"expected risk deviation 50, got %.2f",
			deviation.AverageRiskScore,
		)
	}
}

func TestScoreObservationRequiresWarmup(
	t *testing.T,
) {
	profile :=
		Profile{
			AgentID:     "agent-1",
			SampleCount: 3,
			Mean: Observation{
				EventCount: 10,
				DenyRatio:  0.10,
			},
		}

	score :=
		ScoreObservation(
			profile,
			Observation{
				EventCount: 30,
				DenyRatio:  0.90,
			},
			5,
		)

	if score.WarmedUp {
		t.Fatal(
			"expected profile not to be warmed up",
		)
	}

	if score.IsAnomalous {
		t.Fatal(
			"expected no anomaly before warmup",
		)
	}
}

func TestScoreObservationDetectsAnomaly(
	t *testing.T,
) {
	profile :=
		Profile{
			AgentID:     "agent-2",
			SampleCount: 10,

			Mean: Observation{
				EventCount:             10,
				DenyRatio:              0.10,
				HighRiskRatio:          0.05,
				ActionDiversityRatio:   0.20,
				ResourceDiversityRatio: 0.20,
				AverageRiskScore:       20,
				ProductionAccessRatio:  0.10,
				SensitiveActionRatio:   0.00,
			},
		}

	score :=
		ScoreObservation(
			profile,
			Observation{
				EventCount:             30,
				DenyRatio:              0.90,
				HighRiskRatio:          0.80,
				ActionDiversityRatio:   0.90,
				ResourceDiversityRatio: 0.90,
				AverageRiskScore:       85,
				ProductionAccessRatio:  1.00,
				SensitiveActionRatio:   0.70,
			},
			5,
		)

	if !score.WarmedUp {
		t.Fatal(
			"expected profile to be warmed up",
		)
	}

	if !score.IsAnomalous {
		t.Fatalf(
			"expected anomaly, score was %.4f",
			score.Value,
		)
	}

	if score.Value < 0.50 {
		t.Fatalf(
			"expected score >= 0.50, got %.4f",
			score.Value,
		)
	}
}

func TestScoreObservationNormalBehavior(
	t *testing.T,
) {
	profile :=
		Profile{
			AgentID:     "agent-3",
			SampleCount: 10,

			Mean: Observation{
				EventCount:             10,
				DenyRatio:              0.10,
				HighRiskRatio:          0.10,
				ActionDiversityRatio:   0.30,
				ResourceDiversityRatio: 0.30,
				AverageRiskScore:       25,
				ProductionAccessRatio:  0.20,
				SensitiveActionRatio:   0.05,
			},
		}

	score :=
		ScoreObservation(
			profile,
			Observation{
				EventCount:             11,
				DenyRatio:              0.12,
				HighRiskRatio:          0.08,
				ActionDiversityRatio:   0.32,
				ResourceDiversityRatio: 0.28,
				AverageRiskScore:       27,
				ProductionAccessRatio:  0.22,
				SensitiveActionRatio:   0.04,
			},
			5,
		)

	if !score.WarmedUp {
		t.Fatal(
			"expected warmed-up profile",
		)
	}

	if score.IsAnomalous {
		t.Fatalf(
			"expected normal behavior, score was %.4f",
			score.Value,
		)
	}
}
