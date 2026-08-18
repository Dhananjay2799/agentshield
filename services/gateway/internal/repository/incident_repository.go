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

func NewIncidentRepository(
	db *pgxpool.Pool,
) *IncidentRepository {
	return &IncidentRepository{
		db: db,
	}
}

func (r *IncidentRepository) List(
	ctx context.Context,
	filter models.IncidentListFilter,
) ([]models.SecurityIncident, int, error) {

	var total int

	err := r.db.QueryRow(
		ctx,
		`
		SELECT COUNT(*)
		FROM security_incidents
		WHERE ($1 = '' OR status = $1)
		  AND ($2 = '' OR severity = $2)
		  AND ($3 = '' OR assigned_to = $3)
		`,
		filter.Status,
		filter.Severity,
		filter.AssignedTo,
	).Scan(&total)

	if err != nil {
		return nil, 0, err
	}

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
			assigned_to,
			investigation_note,
			resolution,
			investigating_at,
			first_seen_at,
			last_seen_at,
			created_at,
			updated_at,
			resolved_at,
			metadata
		FROM security_incidents
		WHERE ($1 = '' OR status = $1)
		  AND ($2 = '' OR severity = $2)
		  AND ($3 = '' OR assigned_to = $3)
		ORDER BY last_seen_at DESC
		LIMIT $4
		OFFSET $5
		`,
		filter.Status,
		filter.Severity,
		filter.AssignedTo,
		filter.Limit,
		filter.Offset,
	)

	if err != nil {
		return nil, 0, err
	}

	defer rows.Close()

	incidents := make(
		[]models.SecurityIncident,
		0,
	)

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
			&incident.AssignedTo,
			&incident.InvestigationNote,
			&incident.Resolution,
			&incident.InvestigatingAt,
			&incident.FirstSeenAt,
			&incident.LastSeenAt,
			&incident.CreatedAt,
			&incident.UpdatedAt,
			&incident.ResolvedAt,
			&incident.Metadata,
		)

		if err != nil {
			return nil, 0, err
		}

		incidents = append(
			incidents,
			incident,
		)
	}

	if err := rows.Err(); err != nil {
		return nil, 0, err
	}

	return incidents, total, nil
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
			assigned_to,
			investigation_note,
			resolution,
			investigating_at,
			first_seen_at,
			last_seen_at,
			created_at,
			updated_at,
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
		&incident.AssignedTo,
		&incident.InvestigationNote,
		&incident.Resolution,
		&incident.InvestigatingAt,
		&incident.FirstSeenAt,
		&incident.LastSeenAt,
		&incident.CreatedAt,
		&incident.UpdatedAt,
		&incident.ResolvedAt,
		&incident.Metadata,
	)

	if errors.Is(
		err,
		pgx.ErrNoRows,
	) {
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
	assignedTo string,
	investigationNote string,
) (*models.SecurityIncident, error) {

	result, err := r.db.Exec(
		ctx,
		`
		UPDATE security_incidents
		SET
			status = 'investigating',
			assigned_to = $2,
			investigation_note = NULLIF($3, ''),
			investigating_at = NOW(),
			updated_at = NOW()
		WHERE id = $1
		  AND status = 'open'
		`,
		id,
		assignedTo,
		investigationNote,
	)

	if err != nil {
		return nil, err
	}

	if result.RowsAffected() == 0 {
		return nil, ErrIncidentNotFound
	}

	return r.GetByID(
		ctx,
		id,
	)
}

func (r *IncidentRepository) Resolve(
	ctx context.Context,
	id string,
	resolution string,
) (*models.SecurityIncident, error) {

	result, err := r.db.Exec(
		ctx,
		`
		UPDATE security_incidents
		SET
			status = 'resolved',
			resolution = $2,
			resolved_at = NOW(),
			updated_at = NOW()
		WHERE id = $1
		  AND status IN ('open', 'investigating')
		`,
		id,
		resolution,
	)

	if err != nil {
		return nil, err
	}

	if result.RowsAffected() == 0 {
		return nil, ErrIncidentNotFound
	}

	return r.GetByID(
		ctx,
		id,
	)
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
			resolution = 'dismissed by SOC analyst',
			resolved_at = NOW(),
			updated_at = NOW()
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

	return r.GetByID(
		ctx,
		id,
	)
}

func (r *IncidentRepository) MetricsSummary(
	ctx context.Context,
) (models.IncidentMetricsSummary, error) {

	var summary models.IncidentMetricsSummary

	err := r.db.QueryRow(
		ctx,
		`
		SELECT
			COUNT(*) FILTER (
				WHERE status = 'open'
			),
			COUNT(*) FILTER (
				WHERE status = 'investigating'
			),
			COUNT(*) FILTER (
				WHERE status = 'resolved'
			),
			COUNT(*) FILTER (
				WHERE status = 'dismissed'
			),
			COUNT(*) FILTER (
				WHERE status = 'open'
				  AND severity = 'critical'
			),
			COUNT(*) FILTER (
				WHERE status = 'open'
				  AND assigned_to IS NULL
			),
			COUNT(*)
		FROM security_incidents
		`,
	).Scan(
		&summary.Open,
		&summary.Investigating,
		&summary.Resolved,
		&summary.Dismissed,
		&summary.CriticalOpen,
		&summary.UnassignedOpen,
		&summary.Total,
	)

	if err != nil {
		return models.IncidentMetricsSummary{}, err
	}

	return summary, nil
}
