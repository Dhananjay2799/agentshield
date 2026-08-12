package repository

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/dhananjay2799/agentshield/services/gateway/internal/models"
)

var ErrGrantNotFound = errors.New("authorization grant not found")

type GrantRepository struct {
	DB *pgxpool.Pool
}

func NewGrantRepository(db *pgxpool.Pool) *GrantRepository {
	return &GrantRepository{
		DB: db,
	}
}

// CreateFromApproval creates a short-lived authorization grant
// after a human has approved an action.
func (r *GrantRepository) CreateFromApproval(
	ctx context.Context,
	approval *models.ApprovalRequest,
) (*models.AuthorizationGrant, error) {

	expiresAt := time.Now().UTC().Add(5 * time.Minute)

	var grant models.AuthorizationGrant

	err := r.DB.QueryRow(
		ctx,
		`
		INSERT INTO authorization_grants (
			approval_id,
			agent_id,
			session_id,
			action,
			resource,
			status,
			expires_at
		)
		VALUES ($1, $2, $3, $4, $5, 'active', $6)
		RETURNING
			id,
			approval_id,
			agent_id,
			session_id,
			action,
			resource,
			status,
			issued_at,
			expires_at,
			used_at
		`,
		approval.ID,
		approval.AgentID,
		approval.SessionID,
		approval.Action,
		approval.Resource,
		expiresAt,
	).Scan(
		&grant.ID,
		&grant.ApprovalID,
		&grant.AgentID,
		&grant.SessionID,
		&grant.Action,
		&grant.Resource,
		&grant.Status,
		&grant.IssuedAt,
		&grant.ExpiresAt,
		&grant.UsedAt,
	)

	if err != nil {
		return nil, err
	}

	return &grant, nil
}

// FindActiveGrant looks for an unused, unexpired grant matching
// the exact agent, session, action and resource.
func (r *GrantRepository) FindActiveGrant(
	ctx context.Context,
	agentID string,
	sessionID string,
	action string,
	resource string,
) (*models.AuthorizationGrant, error) {

	var grant models.AuthorizationGrant

	err := r.DB.QueryRow(
		ctx,
		`
		SELECT
			id,
			approval_id,
			agent_id,
			session_id,
			action,
			resource,
			status,
			issued_at,
			expires_at,
			used_at
		FROM authorization_grants
		WHERE agent_id = $1
		  AND session_id = $2
		  AND action = $3
		  AND resource = $4
		  AND status = 'active'
		  AND expires_at > NOW()
		ORDER BY issued_at DESC
		LIMIT 1
		`,
		agentID,
		sessionID,
		action,
		resource,
	).Scan(
		&grant.ID,
		&grant.ApprovalID,
		&grant.AgentID,
		&grant.SessionID,
		&grant.Action,
		&grant.Resource,
		&grant.Status,
		&grant.IssuedAt,
		&grant.ExpiresAt,
		&grant.UsedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrGrantNotFound
		}

		return nil, err
	}

	return &grant, nil
}

// Consume marks a grant as used.
// Our first implementation uses one-time authorization grants.
func (r *GrantRepository) Consume(
	ctx context.Context,
	id string,
) error {

	tag, err := r.DB.Exec(
		ctx,
		`
		UPDATE authorization_grants
		SET
			status = 'used',
			used_at = NOW()
		WHERE id = $1
		  AND status = 'active'
		  AND expires_at > NOW()
		`,
		id,
	)

	if err != nil {
		return err
	}

	if tag.RowsAffected() == 0 {
		return ErrGrantNotFound
	}

	return nil
}

func (r *GrantRepository) ExpireOldGrants(
	ctx context.Context,
) (int64, error) {

	tag, err := r.DB.Exec(
		ctx,
		`
		UPDATE authorization_grants
		SET status = 'expired'
		WHERE status = 'active'
		  AND expires_at <= NOW()
		`,
	)

	if err != nil {
		return 0, err
	}

	return tag.RowsAffected(), nil
}
