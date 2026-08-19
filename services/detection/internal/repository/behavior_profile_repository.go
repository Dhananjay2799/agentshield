package repository

import (
	"context"
	"errors"

	"github.com/dhananjay2799/agentshield/services/detection/internal/baseline"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var ErrBehaviorProfileNotFound = errors.New(
	"behavior profile not found",
)

type BehaviorProfileRepository struct {
	db *pgxpool.Pool
}

func NewBehaviorProfileRepository(
	db *pgxpool.Pool,
) *BehaviorProfileRepository {
	return &BehaviorProfileRepository{
		db: db,
	}
}

func (r *BehaviorProfileRepository) Get(
	ctx context.Context,
	agentID string,
) (*baseline.Profile, error) {
	var profile baseline.Profile

	profile.AgentID =
		agentID

	err :=
		r.db.QueryRow(
			ctx,
			`
			SELECT
				sample_count,
				mean_event_count,
				mean_deny_ratio,
				mean_high_risk_ratio,
				mean_action_diversity_ratio,
				mean_resource_diversity_ratio,
				mean_average_risk_score,
				mean_production_access_ratio,
				mean_sensitive_action_ratio
			FROM agent_behavior_profiles
			WHERE agent_id = $1
			`,
			agentID,
		).Scan(
			&profile.SampleCount,
			&profile.Mean.EventCount,
			&profile.Mean.DenyRatio,
			&profile.Mean.HighRiskRatio,
			&profile.Mean.ActionDiversityRatio,
			&profile.Mean.ResourceDiversityRatio,
			&profile.Mean.AverageRiskScore,
			&profile.Mean.ProductionAccessRatio,
			&profile.Mean.SensitiveActionRatio,
		)

	if errors.Is(
		err,
		pgx.ErrNoRows,
	) {
		return nil,
			ErrBehaviorProfileNotFound
	}

	if err != nil {
		return nil, err
	}

	return &profile, nil
}

func (r *BehaviorProfileRepository) Upsert(
	ctx context.Context,
	profile baseline.Profile,
) error {
	_, err :=
		r.db.Exec(
			ctx,
			`
			INSERT INTO agent_behavior_profiles (
				agent_id,
				sample_count,
				mean_event_count,
				mean_deny_ratio,
				mean_high_risk_ratio,
				mean_action_diversity_ratio,
				mean_resource_diversity_ratio,
				mean_average_risk_score,
				mean_production_access_ratio,
				mean_sensitive_action_ratio,
				updated_at
			)
			VALUES (
				$1,
				$2,
				$3,
				$4,
				$5,
				$6,
				$7,
				$8,
				$9,
				$10,
				NOW()
			)
			ON CONFLICT (agent_id)
			DO UPDATE SET
				sample_count =
					EXCLUDED.sample_count,

				mean_event_count =
					EXCLUDED.mean_event_count,

				mean_deny_ratio =
					EXCLUDED.mean_deny_ratio,

				mean_high_risk_ratio =
					EXCLUDED.mean_high_risk_ratio,

				mean_action_diversity_ratio =
					EXCLUDED.mean_action_diversity_ratio,

				mean_resource_diversity_ratio =
					EXCLUDED.mean_resource_diversity_ratio,

				mean_average_risk_score =
					EXCLUDED.mean_average_risk_score,

				mean_production_access_ratio =
					EXCLUDED.mean_production_access_ratio,

				mean_sensitive_action_ratio =
					EXCLUDED.mean_sensitive_action_ratio,

				updated_at =
					NOW()
			`,
			profile.AgentID,
			profile.SampleCount,
			profile.Mean.EventCount,
			profile.Mean.DenyRatio,
			profile.Mean.HighRiskRatio,
			profile.Mean.ActionDiversityRatio,
			profile.Mean.ResourceDiversityRatio,
			profile.Mean.AverageRiskScore,
			profile.Mean.ProductionAccessRatio,
			profile.Mean.SensitiveActionRatio,
		)

	return err
}
