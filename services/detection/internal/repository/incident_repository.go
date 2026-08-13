package repository

import (
	"context"
	"encoding/json"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type IncidentRepository struct {
	DB *pgxpool.Pool
}

type UpsertIncidentParams struct {
	AgentID      string
	SessionID    string
	IncidentType string
	Severity     string
	Title        string
	Description  string
	Metadata     map[string]any
}

func NewIncidentRepository(db *pgxpool.Pool) *IncidentRepository {
	return &IncidentRepository{
		DB: db,
	}
}

func (r *IncidentRepository) UpsertOpenIncident(
	ctx context.Context,
	params UpsertIncidentParams,
) error {

	metadata := params.Metadata
	if metadata == nil {
		metadata = map[string]any{}
	}

	metadataJSON, err := json.Marshal(metadata)
	if err != nil {
		return err
	}

	// Look for a recent open incident of the same type for this agent.
	var existingID string

	err = r.DB.QueryRow(
		ctx,
		`
		SELECT id
		FROM security_incidents
		WHERE agent_id = $1
		  AND incident_type = $2
		  AND status IN ('open', 'investigating')
		  AND last_seen_at >= NOW() - INTERVAL '5 minutes'
		ORDER BY last_seen_at DESC
		LIMIT 1
		`,
		params.AgentID,
		params.IncidentType,
	).Scan(&existingID)

	if err == nil {
		_, err = r.DB.Exec(
			ctx,
			`
			UPDATE security_incidents
			SET
				session_id = $1,
				severity = $2,
				title = $3,
				description = $4,
				event_count = event_count + 1,
				last_seen_at = NOW(),
				metadata = metadata || $5::jsonb
			WHERE id = $6
			`,
			params.SessionID,
			params.Severity,
			params.Title,
			params.Description,
			string(metadataJSON),
			existingID,
		)

		return err
	}

	_, err = r.DB.Exec(
		ctx,
		`
		INSERT INTO security_incidents (
			agent_id,
			session_id,
			incident_type,
			severity,
			title,
			description,
			status,
			event_count,
			first_seen_at,
			last_seen_at,
			metadata
		)
		VALUES (
			$1, $2, $3, $4, $5, $6,
			'open',
			1,
			NOW(),
			NOW(),
			$7::jsonb
		)
		`,
		params.AgentID,
		params.SessionID,
		params.IncidentType,
		params.Severity,
		params.Title,
		params.Description,
		string(metadataJSON),
	)

	return err
}

func (r *IncidentRepository) Resolve(
	ctx context.Context,
	id string,
) error {

	_, err := r.DB.Exec(
		ctx,
		`
		UPDATE security_incidents
		SET
			status = 'resolved',
			resolved_at = NOW(),
			last_seen_at = NOW()
		WHERE id = $1
		`,
		id,
	)

	return err
}

func (r *IncidentRepository) ExpireStaleOpenIncidents(
	ctx context.Context,
	olderThan time.Duration,
) error {

	cutoff := time.Now().UTC().Add(-olderThan)

	_, err := r.DB.Exec(
		ctx,
		`
		UPDATE security_incidents
		SET
			status = 'resolved',
			resolved_at = NOW()
		WHERE status = 'open'
		  AND last_seen_at < $1
		`,
		cutoff,
	)

	return err
}
