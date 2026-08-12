package repository

import (
	"context"
	"encoding/json"

	"github.com/jackc/pgx/v5/pgxpool"
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
