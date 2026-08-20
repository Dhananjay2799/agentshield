package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type ContainmentResult struct {
	AgentID         string `json:"agent_id"`
	AgentStatus     string `json:"agent_status"`
	SessionsRevoked int64  `json:"sessions_revoked"`
	GrantsRevoked   int64  `json:"grants_revoked"`
}

type ContainmentRepository struct {
	DB *pgxpool.Pool
}

func NewContainmentRepository(
	db *pgxpool.Pool,
) *ContainmentRepository {
	return &ContainmentRepository{
		DB: db,
	}
}

// ContainAgent atomically suspends an agent and invalidates
// its active execution and authorization capabilities.
func (r *ContainmentRepository) ContainAgent(
	ctx context.Context,
	agentID string,
) (*ContainmentResult, error) {
	tx, err := r.DB.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf(
			"begin containment transaction: %w",
			err,
		)
	}
	defer tx.Rollback(ctx)

	var status string

	err = tx.QueryRow(
		ctx,
		`
UPDATE agents
SET
status = 'suspended',
updated_at = NOW()
WHERE id = $1
RETURNING status
`,
		agentID,
	).Scan(&status)

	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrAgentNotFound
	}

	if err != nil {
		return nil, fmt.Errorf(
			"suspend agent: %w",
			err,
		)
	}

	sessionTag, err := tx.Exec(
		ctx,
		`
UPDATE agent_sessions
SET
status = 'revoked',
ended_at = COALESCE(ended_at, NOW())
WHERE agent_id = $1
  AND status = 'active'
`,
		agentID,
	)
	if err != nil {
		return nil, fmt.Errorf(
			"revoke agent sessions: %w",
			err,
		)
	}

	grantTag, err := tx.Exec(
		ctx,
		`
UPDATE authorization_grants
SET
status = 'revoked'
WHERE agent_id = $1
  AND status = 'active'
`,
		agentID,
	)
	if err != nil {
		return nil, fmt.Errorf(
			"revoke authorization grants: %w",
			err,
		)
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf(
			"commit containment transaction: %w",
			err,
		)
	}

	return &ContainmentResult{
		AgentID:         agentID,
		AgentStatus:     status,
		SessionsRevoked: sessionTag.RowsAffected(),
		GrantsRevoked:   grantTag.RowsAffected(),
	}, nil
}
