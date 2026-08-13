package repository

import (
	"context"
	"errors"

	"github.com/dhananjay2799/agentshield/services/gateway/internal/models"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var ErrIncidentNotFound = errors.New("incident not found")

type IncidentRepository struct {
	db *pgxpool.Pool
}

func NewIncidentRepository(db *pgxpool.Pool) *IncidentRepository {
	return &IncidentRepository{
		db: db,
	}
}

func (r *IncidentRepository) List(
	ctx context.Context,
) ([]models.SecurityIncident, error) {

	rows, err := r.db.Query(
		ctx,
		`
		SELECT
			id,
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
			created_at,
			resolved_at,
			metadata
		FROM security_incidents
		ORDER BY last_seen_at DESC
		`,
	)

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	incidents := make([]models.SecurityIncident, 0)

	for rows.Next() {
		var incident models.SecurityIncident

		err := rows.Scan(
			&incident.ID,
			&incident.AgentID,
			&incident.SessionID,
			&incident.IncidentType,
			&incident.Severity,
			&incident.Title,
			&incident.Description,
			&incident.Status,
			&incident.EventCount,
			&incident.FirstSeenAt,
			&incident.LastSeenAt,
			&incident.CreatedAt,
			&incident.ResolvedAt,
			&incident.Metadata,
		)

		if err != nil {
			return nil, err
		}

		incidents = append(incidents, incident)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return incidents, nil
}

func (r *IncidentRepository) GetByID(
	ctx context.Context,
	id string,
) (*models.SecurityIncident, error) {

	var incident models.SecurityIncident

	err := r.db.QueryRow(
		ctx,
		`
		SELECT
			id,
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
			created_at,
			resolved_at,
			metadata
		FROM security_incidents
		WHERE id = $1
		`,
		id,
	).Scan(
		&incident.ID,
		&incident.AgentID,
		&incident.SessionID,
		&incident.IncidentType,
		&incident.Severity,
		&incident.Title,
		&incident.Description,
		&incident.Status,
		&incident.EventCount,
		&incident.FirstSeenAt,
		&incident.LastSeenAt,
		&incident.CreatedAt,
		&incident.ResolvedAt,
		&incident.Metadata,
	)

	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrIncidentNotFound
	}

	if err != nil {
		return nil, err
	}

	return &incident, nil
}

func (r *IncidentRepository) MarkInvestigating(
	ctx context.Context,
	id string,
) (*models.SecurityIncident, error) {

	result, err := r.db.Exec(
		ctx,
		`
		UPDATE security_incidents
		SET status = 'investigating'
		WHERE id = $1
		  AND status = 'open'
		`,
		id,
	)

	if err != nil {
		return nil, err
	}

	if result.RowsAffected() == 0 {
		return nil, ErrIncidentNotFound
	}

	return r.GetByID(ctx, id)
}

func (r *IncidentRepository) Resolve(
	ctx context.Context,
	id string,
) (*models.SecurityIncident, error) {

	result, err := r.db.Exec(
		ctx,
		`
		UPDATE security_incidents
		SET
			status = 'resolved',
			resolved_at = NOW()
		WHERE id = $1
		  AND status IN ('open', 'investigating')
		`,
		id,
	)

	if err != nil {
		return nil, err
	}

	if result.RowsAffected() == 0 {
		return nil, ErrIncidentNotFound
	}

	return r.GetByID(ctx, id)
}

func (r *IncidentRepository) Dismiss(
	ctx context.Context,
	id string,
) (*models.SecurityIncident, error) {

	result, err := r.db.Exec(
		ctx,
		`
		UPDATE security_incidents
		SET
			status = 'dismissed',
			resolved_at = NOW()
		WHERE id = $1
		  AND status IN ('open', 'investigating')
		`,
		id,
	)

	if err != nil {
		return nil, err
	}

	if result.RowsAffected() == 0 {
		return nil, ErrIncidentNotFound
	}

	return r.GetByID(ctx, id)
}
