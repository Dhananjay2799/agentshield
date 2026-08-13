package repository

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/dhananjay2799/agentshield/services/gateway/internal/models"
)

var ErrSessionNotFound = errors.New("session not found")

type SessionRepository struct {
	DB *pgxpool.Pool
}

func NewSessionRepository(db *pgxpool.Pool) *SessionRepository {
	return &SessionRepository{
		DB: db,
	}
}

func (r *SessionRepository) Create(
	ctx context.Context,
	agentID string,
	req models.CreateSessionRequest,
) (*models.AgentSession, error) {

	ttl := req.TTLMinutes

	if ttl <= 0 {
		ttl = 15
	}

	expiresAt := time.Now().UTC().Add(time.Duration(ttl) * time.Minute)

	var session models.AgentSession

	err := r.DB.QueryRow(
		ctx,
		`
		INSERT INTO agent_sessions (
			agent_id,
			task_id,
			status,
			expires_at
		)
		VALUES ($1, $2, 'active', $3)
		RETURNING
			id,
			agent_id,
			task_id,
			status,
			started_at,
			ended_at,
			expires_at
		`,
		agentID,
		req.TaskID,
		expiresAt,
	).Scan(
		&session.ID,
		&session.AgentID,
		&session.TaskID,
		&session.Status,
		&session.StartedAt,
		&session.EndedAt,
		&session.ExpiresAt,
	)

	if err != nil {
		return nil, err
	}

	return &session, nil
}

func (r *SessionRepository) GetByID(
	ctx context.Context,
	id string,
) (*models.AgentSession, error) {

	var session models.AgentSession

	err := r.DB.QueryRow(
		ctx,
		`
		SELECT
			id,
			agent_id,
			task_id,
			status,
			started_at,
			ended_at,
			expires_at
		FROM agent_sessions
		WHERE id = $1
		`,
		id,
	).Scan(
		&session.ID,
		&session.AgentID,
		&session.TaskID,
		&session.Status,
		&session.StartedAt,
		&session.EndedAt,
		&session.ExpiresAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrSessionNotFound
		}

		return nil, err
	}

	return &session, nil
}

func (r *SessionRepository) Revoke(
	ctx context.Context,
	id string,
) error {

	commandTag, err := r.DB.Exec(
		ctx,
		`
		UPDATE agent_sessions
		SET
			status = 'revoked',
			ended_at = NOW()
		WHERE id = $1
		  AND status = 'active'
		`,
		id,
	)

	if err != nil {
		return err
	}

	if commandTag.RowsAffected() == 0 {
		return ErrSessionNotFound
	}

	return nil
}

func (r *SessionRepository) ValidateActive(
	ctx context.Context,
	id string,
) (*models.AgentSession, error) {

	session, err := r.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if session.Status != "active" {
		return nil, errors.New("session is not active")
	}

	if session.ExpiresAt != nil && time.Now().UTC().After(*session.ExpiresAt) {
		return nil, errors.New("session has expired")
	}

	return session, nil
}

func (r *SessionRepository) ListSecurityByAgentID(
	ctx context.Context,
	agentID string,
) ([]models.SessionSecuritySummary, error) {

	rows, err := r.DB.Query(
		ctx,
		`
		SELECT
			s.id,
			s.agent_id,
			s.task_id,
			s.status,
			s.started_at,
			s.ended_at,
			s.expires_at,

			COUNT(a.id)::int AS action_count,

			COUNT(a.id) FILTER (
				WHERE a.decision = 'ALLOW'
			)::int AS allowed_count,

			COUNT(a.id) FILTER (
				WHERE a.decision = 'DENY'
			)::int AS denied_count,

			COUNT(a.id) FILTER (
				WHERE a.decision = 'REQUIRE_APPROVAL'
			)::int AS approval_count,

			COALESCE(MAX(a.risk_score), 0)::int
				AS highest_risk_score,

			MAX(a.created_at) AS last_action_at

		FROM agent_sessions s

		LEFT JOIN audit_events a
			ON a.session_id = s.id

		WHERE s.agent_id = $1

		GROUP BY
			s.id,
			s.agent_id,
			s.task_id,
			s.status,
			s.started_at,
			s.ended_at,
			s.expires_at

		ORDER BY s.started_at DESC

		LIMIT 100
		`,
		agentID,
	)

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	sessions := make(
		[]models.SessionSecuritySummary,
		0,
	)

	for rows.Next() {
		var session models.SessionSecuritySummary

		if err := rows.Scan(
			&session.ID,
			&session.AgentID,
			&session.TaskID,
			&session.Status,
			&session.StartedAt,
			&session.EndedAt,
			&session.ExpiresAt,
			&session.ActionCount,
			&session.AllowedCount,
			&session.DeniedCount,
			&session.ApprovalCount,
			&session.HighestRiskScore,
			&session.LastActionAt,
		); err != nil {
			return nil, err
		}

		sessions = append(
			sessions,
			session,
		)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return sessions, nil
}

func (r *SessionRepository) ListByAgentID(
	ctx context.Context,
	agentID string,
) ([]models.AgentSession, error) {

	rows, err := r.DB.Query(
		ctx,
		`
		SELECT
			id,
			agent_id,
			task_id,
			status,
			started_at,
			ended_at,
			expires_at
		FROM agent_sessions
		WHERE agent_id = $1
		ORDER BY started_at DESC
		LIMIT 100
		`,
		agentID,
	)

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	sessions := make([]models.AgentSession, 0)

	for rows.Next() {
		var session models.AgentSession

		if err := rows.Scan(
			&session.ID,
			&session.AgentID,
			&session.TaskID,
			&session.Status,
			&session.StartedAt,
			&session.EndedAt,
			&session.ExpiresAt,
		); err != nil {
			return nil, err
		}

		sessions = append(sessions, session)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return sessions, nil
}
