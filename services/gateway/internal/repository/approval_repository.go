package repository

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/dhananjay2799/agentshield/services/gateway/internal/models"
)

var ErrApprovalNotFound = errors.New("approval request not found")

type ApprovalRepository struct {
	DB *pgxpool.Pool
}

func NewApprovalRepository(db *pgxpool.Pool) *ApprovalRepository {
	return &ApprovalRepository{
		DB: db,
	}
}

func (r *ApprovalRepository) Create(
	ctx context.Context,
	agentID string,
	sessionID string,
	action string,
	resource string,
	reason string,
	riskScore int,
) (*models.ApprovalRequest, error) {

	expiresAt := time.Now().UTC().Add(10 * time.Minute)

	var approval models.ApprovalRequest

	err := r.DB.QueryRow(
		ctx,
		`
		INSERT INTO approval_requests (
			agent_id,
			session_id,
			action,
			resource,
			reason,
			risk_score,
			status,
			expires_at
		)
		VALUES ($1, $2, $3, $4, $5, $6, 'pending', $7)
		RETURNING
			id,
			agent_id,
			session_id,
			action,
			resource,
			reason,
			risk_score,
			status,
			requested_at,
			approved_at,
			denied_at,
			expires_at
		`,
		agentID,
		sessionID,
		action,
		resource,
		reason,
		riskScore,
		expiresAt,
	).Scan(
		&approval.ID,
		&approval.AgentID,
		&approval.SessionID,
		&approval.Action,
		&approval.Resource,
		&approval.Reason,
		&approval.RiskScore,
		&approval.Status,
		&approval.RequestedAt,
		&approval.ApprovedAt,
		&approval.DeniedAt,
		&approval.ExpiresAt,
	)

	if err != nil {
		return nil, err
	}

	return &approval, nil
}

func (r *ApprovalRepository) GetByID(
	ctx context.Context,
	id string,
) (*models.ApprovalRequest, error) {

	var approval models.ApprovalRequest

	err := r.DB.QueryRow(
		ctx,
		`
		SELECT
			id,
			agent_id,
			session_id,
			action,
			resource,
			reason,
			risk_score,
			status,
			requested_at,
			approved_at,
			denied_at,
			expires_at
		FROM approval_requests
		WHERE id = $1
		`,
		id,
	).Scan(
		&approval.ID,
		&approval.AgentID,
		&approval.SessionID,
		&approval.Action,
		&approval.Resource,
		&approval.Reason,
		&approval.RiskScore,
		&approval.Status,
		&approval.RequestedAt,
		&approval.ApprovedAt,
		&approval.DeniedAt,
		&approval.ExpiresAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrApprovalNotFound
		}
		return nil, err
	}

	return &approval, nil
}

func (r *ApprovalRepository) ListPending(
	ctx context.Context,
) ([]models.ApprovalRequest, error) {

	rows, err := r.DB.Query(
		ctx,
		`
		SELECT
			id,
			agent_id,
			session_id,
			action,
			resource,
			reason,
			risk_score,
			status,
			requested_at,
			approved_at,
			denied_at,
			expires_at
		FROM approval_requests
		WHERE status = 'pending'
		ORDER BY requested_at DESC
		`,
	)

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	approvals := make([]models.ApprovalRequest, 0)

	for rows.Next() {
		var approval models.ApprovalRequest

		if err := rows.Scan(
			&approval.ID,
			&approval.AgentID,
			&approval.SessionID,
			&approval.Action,
			&approval.Resource,
			&approval.Reason,
			&approval.RiskScore,
			&approval.Status,
			&approval.RequestedAt,
			&approval.ApprovedAt,
			&approval.DeniedAt,
			&approval.ExpiresAt,
		); err != nil {
			return nil, err
		}

		approvals = append(approvals, approval)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return approvals, nil
}

func (r *ApprovalRepository) Approve(
	ctx context.Context,
	id string,
) error {

	tag, err := r.DB.Exec(
		ctx,
		`
		UPDATE approval_requests
		SET
			status = 'approved',
			approved_at = NOW()
		WHERE id = $1
		  AND status = 'pending'
		  AND (expires_at IS NULL OR expires_at > NOW())
		`,
		id,
	)

	if err != nil {
		return err
	}

	if tag.RowsAffected() == 0 {
		return ErrApprovalNotFound
	}

	return nil
}

func (r *ApprovalRepository) Deny(
	ctx context.Context,
	id string,
) error {

	tag, err := r.DB.Exec(
		ctx,
		`
		UPDATE approval_requests
		SET
			status = 'denied',
			denied_at = NOW()
		WHERE id = $1
		  AND status = 'pending'
		`,
		id,
	)

	if err != nil {
		return err
	}

	if tag.RowsAffected() == 0 {
		return ErrApprovalNotFound
	}

	return nil
}

func (r *ApprovalRepository) ApproveAndCreateGrant(
	ctx context.Context,
	id string,
) (*models.ApprovalRequest, *models.AuthorizationGrant, error) {

	tx, err := r.DB.Begin(ctx)
	if err != nil {
		return nil, nil, err
	}

	defer tx.Rollback(ctx)

	var approval models.ApprovalRequest

	err = tx.QueryRow(
		ctx,
		`
		SELECT
			id,
			agent_id,
			session_id,
			action,
			resource,
			reason,
			risk_score,
			status,
			requested_at,
			approved_at,
			denied_at,
			expires_at
		FROM approval_requests
		WHERE id = $1
		FOR UPDATE
		`,
		id,
	).Scan(
		&approval.ID,
		&approval.AgentID,
		&approval.SessionID,
		&approval.Action,
		&approval.Resource,
		&approval.Reason,
		&approval.RiskScore,
		&approval.Status,
		&approval.RequestedAt,
		&approval.ApprovedAt,
		&approval.DeniedAt,
		&approval.ExpiresAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil, ErrApprovalNotFound
		}

		return nil, nil, err
	}

	if approval.Status != "pending" {
		return nil, nil, ErrApprovalNotFound
	}

	if approval.ExpiresAt != nil &&
		time.Now().UTC().After(*approval.ExpiresAt) {

		_, err = tx.Exec(
			ctx,
			`
			UPDATE approval_requests
			SET status = 'expired'
			WHERE id = $1
			`,
			id,
		)

		if err != nil {
			return nil, nil, err
		}

		if err := tx.Commit(ctx); err != nil {
			return nil, nil, err
		}

		return nil, nil, ErrApprovalNotFound
	}

	err = tx.QueryRow(
		ctx,
		`
		UPDATE approval_requests
		SET
			status = 'approved',
			approved_at = NOW()
		WHERE id = $1
		RETURNING approved_at
		`,
		id,
	).Scan(&approval.ApprovedAt)

	if err != nil {
		return nil, nil, err
	}

	approval.Status = "approved"

	grantExpiresAt := time.Now().UTC().Add(5 * time.Minute)

	var grant models.AuthorizationGrant

	err = tx.QueryRow(
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
		grantExpiresAt,
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
		return nil, nil, err
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, nil, err
	}

	return &approval, &grant, nil
}
