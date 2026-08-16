package repository

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/dhananjay2799/agentshield/services/gateway/internal/models"
)

var ErrPolicyNotFound = errors.New("policy not found")

type PolicyRepository struct {
	DB *pgxpool.Pool
}

func NewPolicyRepository(
	db *pgxpool.Pool,
) *PolicyRepository {
	return &PolicyRepository{
		DB: db,
	}
}

func (r *PolicyRepository) Create(
	ctx context.Context,
	req models.CreatePolicyRequest,
) (*models.Policy, error) {

	if req.Priority <= 0 {
		req.Priority = 100
	}

	if req.ActionMatch == "" {
		req.ActionMatch = "exact"
	}

	if req.ResourceMatch == "" {
		req.ResourceMatch = "prefix"
	}

	if req.CreatedBy == "" {
		req.CreatedBy = "soc-analyst"
	}

	var policy models.Policy

	err := r.DB.QueryRow(
		ctx,
		`
		INSERT INTO policies (
			name,
			description,
			effect,
			status,
			priority,
			agent_type,
			action,
			action_match,
			resource,
			resource_match,
			environment,
			version,
			source,
			created_by
		)
		VALUES (
			$1,
			$2,
			$3,
			'draft',
			$4,
			$5,
			$6,
			$7,
			$8,
			$9,
			$10,
			1,
			'control_plane',
			$11
		)
		RETURNING
			id,
			name,
			description,
			effect,
			status,
			priority,
			agent_type,
			action,
			action_match,
			resource,
			resource_match,
			environment,
			version,
			source,
			created_by,
			created_at,
			updated_at
		`,
		req.Name,
		req.Description,
		req.Effect,
		req.Priority,
		req.AgentType,
		req.Action,
		req.ActionMatch,
		req.Resource,
		req.ResourceMatch,
		req.Environment,
		req.CreatedBy,
	).Scan(
		&policy.ID,
		&policy.Name,
		&policy.Description,
		&policy.Effect,
		&policy.Status,
		&policy.Priority,
		&policy.AgentType,
		&policy.Action,
		&policy.ActionMatch,
		&policy.Resource,
		&policy.ResourceMatch,
		&policy.Environment,
		&policy.Version,
		&policy.Source,
		&policy.CreatedBy,
		&policy.CreatedAt,
		&policy.UpdatedAt,
	)

	if err != nil {
		return nil, err
	}

	return &policy, nil
}

func (r *PolicyRepository) List(
	ctx context.Context,
) ([]models.Policy, error) {

	rows, err := r.DB.Query(
		ctx,
		`
		SELECT
			id,
			name,
			description,
			effect,
			status,
			priority,
			agent_type,
			action,
			action_match,
			resource,
			resource_match,
			environment,
			version,
			source,
			created_by,
			created_at,
			updated_at
		FROM policies
		ORDER BY
			priority ASC,
			created_at DESC
		`,
	)

	if err != nil {
		return nil, err
	}

	defer rows.Close()

	policies :=
		make([]models.Policy, 0)

	for rows.Next() {
		var policy models.Policy

		if err := rows.Scan(
			&policy.ID,
			&policy.Name,
			&policy.Description,
			&policy.Effect,
			&policy.Status,
			&policy.Priority,
			&policy.AgentType,
			&policy.Action,
			&policy.ActionMatch,
			&policy.Resource,
			&policy.ResourceMatch,
			&policy.Environment,
			&policy.Version,
			&policy.Source,
			&policy.CreatedBy,
			&policy.CreatedAt,
			&policy.UpdatedAt,
		); err != nil {
			return nil, err
		}

		policies =
			append(
				policies,
				policy,
			)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return policies, nil
}

func (r *PolicyRepository) GetByID(
	ctx context.Context,
	id string,
) (*models.Policy, error) {

	var policy models.Policy

	err := r.DB.QueryRow(
		ctx,
		`
		SELECT
			id,
			name,
			description,
			effect,
			status,
			priority,
			agent_type,
			action,
			action_match,
			resource,
			resource_match,
			environment,
			version,
			source,
			created_by,
			created_at,
			updated_at
		FROM policies
		WHERE id = $1
		`,
		id,
	).Scan(
		&policy.ID,
		&policy.Name,
		&policy.Description,
		&policy.Effect,
		&policy.Status,
		&policy.Priority,
		&policy.AgentType,
		&policy.Action,
		&policy.ActionMatch,
		&policy.Resource,
		&policy.ResourceMatch,
		&policy.Environment,
		&policy.Version,
		&policy.Source,
		&policy.CreatedBy,
		&policy.CreatedAt,
		&policy.UpdatedAt,
	)

	if err != nil {
		if errors.Is(
			err,
			pgx.ErrNoRows,
		) {
			return nil, ErrPolicyNotFound
		}

		return nil, err
	}

	return &policy, nil
}

func (r *PolicyRepository) Activate(
	ctx context.Context,
	id string,
) (*models.Policy, error) {

	commandTag, err :=
		r.DB.Exec(
			ctx,
			`
			UPDATE policies
			SET
				status = 'active',
				updated_at = NOW()
			WHERE id = $1
			  AND status IN ('draft', 'disabled')
			`,
			id,
		)

	if err != nil {
		return nil, err
	}

	if commandTag.RowsAffected() == 0 {
		return nil, ErrPolicyNotFound
	}

	return r.GetByID(
		ctx,
		id,
	)
}

func (r *PolicyRepository) Deactivate(
	ctx context.Context,
	id string,
) (*models.Policy, error) {

	commandTag, err :=
		r.DB.Exec(
			ctx,
			`
			UPDATE policies
			SET
				status = 'disabled',
				updated_at = NOW()
			WHERE id = $1
			  AND status = 'active'
			`,
			id,
		)

	if err != nil {
		return nil, err
	}

	if commandTag.RowsAffected() == 0 {
		return nil, ErrPolicyNotFound
	}

	return r.GetByID(
		ctx,
		id,
	)
}

func (r *PolicyRepository) ListActive(
	ctx context.Context,
) ([]models.Policy, error) {

	rows, err := r.DB.Query(
		ctx,
		`
		SELECT
			id,
			name,
			description,
			effect,
			status,
			priority,
			agent_type,
			action,
			action_match,
			resource,
			resource_match,
			environment,
			version,
			source,
			created_by,
			created_at,
			updated_at
		FROM policies
		WHERE status = 'active'
		ORDER BY
			priority ASC,
			created_at ASC
		`,
	)

	if err != nil {
		return nil, err
	}

	defer rows.Close()

	policies := make([]models.Policy, 0)

	for rows.Next() {
		var policy models.Policy

		if err := rows.Scan(
			&policy.ID,
			&policy.Name,
			&policy.Description,
			&policy.Effect,
			&policy.Status,
			&policy.Priority,
			&policy.AgentType,
			&policy.Action,
			&policy.ActionMatch,
			&policy.Resource,
			&policy.ResourceMatch,
			&policy.Environment,
			&policy.Version,
			&policy.Source,
			&policy.CreatedBy,
			&policy.CreatedAt,
			&policy.UpdatedAt,
		); err != nil {
			return nil, err
		}

		policies = append(
			policies,
			policy,
		)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return policies, nil
}
