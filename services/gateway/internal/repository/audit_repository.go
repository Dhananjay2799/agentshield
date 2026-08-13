package repository

import (
	"context"
	"encoding/json"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/dhananjay2799/agentshield/services/gateway/internal/models"
)

type AuditRepository struct {
	DB *pgxpool.Pool
}

func NewAuditRepository(db *pgxpool.Pool) *AuditRepository {
	return &AuditRepository{
		DB: db,
	}
}

type CreateAuditEventParams struct {
	AgentID   string
	SessionID string
	EventType string
	Action    string
	Resource  string
	Decision  string
	RiskScore int
	Metadata  map[string]any
}

func (r *AuditRepository) Create(
	ctx context.Context,
	params CreateAuditEventParams,
) error {
	metadata := params.Metadata

	if metadata == nil {
		metadata = map[string]any{}
	}

	metadataJSON, err := json.Marshal(metadata)
	if err != nil {
		return err
	}

	_, err = r.DB.Exec(
		ctx,
		`
		INSERT INTO audit_events (
			agent_id,
			session_id,
			event_type,
			action,
			resource,
			decision,
			risk_score,
			metadata
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		`,
		params.AgentID,
		params.SessionID,
		params.EventType,
		params.Action,
		params.Resource,
		params.Decision,
		params.RiskScore,
		metadataJSON,
	)

	return err
}

func (r *AuditRepository) ListByAgentID(
	ctx context.Context,
	agentID string,
) ([]models.AuditEvent, error) {

	rows, err := r.DB.Query(
		ctx,
		`
		SELECT
			id,
			agent_id,
			session_id,
			event_type,
			action,
			resource,
			decision,
			risk_score,
			metadata,
			created_at
		FROM audit_events
		WHERE agent_id = $1
		ORDER BY created_at DESC
		LIMIT 200
		`,
		agentID,
	)

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	events := make([]models.AuditEvent, 0)

	for rows.Next() {
		var event models.AuditEvent

		if err := rows.Scan(
			&event.ID,
			&event.AgentID,
			&event.SessionID,
			&event.EventType,
			&event.Action,
			&event.Resource,
			&event.Decision,
			&event.RiskScore,
			&event.Metadata,
			&event.CreatedAt,
		); err != nil {
			return nil, err
		}

		events = append(events, event)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return events, nil
}

func (r *AuditRepository) ListBySessionID(
	ctx context.Context,
	sessionID string,
) ([]models.AuditEvent, error) {

	rows, err := r.DB.Query(
		ctx,
		`
		SELECT
			id,
			agent_id,
			session_id,
			event_type,
			action,
			resource,
			decision,
			risk_score,
			metadata,
			created_at
		FROM audit_events
		WHERE session_id = $1
		ORDER BY created_at ASC
		`,
		sessionID,
	)

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	events := make([]models.AuditEvent, 0)

	for rows.Next() {
		var event models.AuditEvent

		if err := rows.Scan(
			&event.ID,
			&event.AgentID,
			&event.SessionID,
			&event.EventType,
			&event.Action,
			&event.Resource,
			&event.Decision,
			&event.RiskScore,
			&event.Metadata,
			&event.CreatedAt,
		); err != nil {
			return nil, err
		}

		events = append(events, event)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return events, nil
}

func (r *AuditRepository) ListRecent(
	ctx context.Context,
	limit int,
) ([]models.AuditEvent, error) {

	if limit <= 0 {
		limit = 100
	}

	if limit > 500 {
		limit = 500
	}

	rows, err := r.DB.Query(
		ctx,
		`
		SELECT
			id,
			agent_id,
			session_id,
			event_type,
			action,
			resource,
			decision,
			risk_score,
			metadata,
			created_at
		FROM audit_events
		ORDER BY created_at DESC
		LIMIT $1
		`,
		limit,
	)

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	events := make([]models.AuditEvent, 0)

	for rows.Next() {
		var event models.AuditEvent

		if err := rows.Scan(
			&event.ID,
			&event.AgentID,
			&event.SessionID,
			&event.EventType,
			&event.Action,
			&event.Resource,
			&event.Decision,
			&event.RiskScore,
			&event.Metadata,
			&event.CreatedAt,
		); err != nil {
			return nil, err
		}

		events = append(events, event)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return events, nil
}
