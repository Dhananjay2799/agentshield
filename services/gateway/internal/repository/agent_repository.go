package repository

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/dhananjay2799/agentshield/services/gateway/internal/models"
)

var ErrAgentNotFound = errors.New("agent not found")

type AgentRepository struct {
	DB *pgxpool.Pool
}

func NewAgentRepository(db *pgxpool.Pool) *AgentRepository {
	return &AgentRepository{
		DB: db,
	}
}

func (r *AgentRepository) Create(
	ctx context.Context,
	req models.CreateAgentRequest,
) (*models.Agent, error) {

	if req.Environment == "" {
		req.Environment = "development"
	}

	var agent models.Agent

	err := r.DB.QueryRow(
		ctx,
		`
		INSERT INTO agents (
			name,
			agent_type,
			owner,
			framework,
			model,
			environment
		)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING
			id,
			name,
			agent_type,
			owner,
			framework,
			model,
			environment,
			status,
			created_at,
			updated_at
		`,
		req.Name,
		req.AgentType,
		req.Owner,
		req.Framework,
		req.Model,
		req.Environment,
	).Scan(
		&agent.ID,
		&agent.Name,
		&agent.AgentType,
		&agent.Owner,
		&agent.Framework,
		&agent.Model,
		&agent.Environment,
		&agent.Status,
		&agent.CreatedAt,
		&agent.UpdatedAt,
	)

	if err != nil {
		return nil, err
	}

	return &agent, nil
}

func (r *AgentRepository) List(ctx context.Context) ([]models.Agent, error) {
	rows, err := r.DB.Query(
		ctx,
		`
		SELECT
			id,
			name,
			agent_type,
			owner,
			framework,
			model,
			environment,
			status,
			created_at,
			updated_at
		FROM agents
		ORDER BY created_at DESC
		`,
	)

	if err != nil {
		return nil, err
	}

	defer rows.Close()

	agents := make([]models.Agent, 0)

	for rows.Next() {
		var agent models.Agent

		if err := rows.Scan(
			&agent.ID,
			&agent.Name,
			&agent.AgentType,
			&agent.Owner,
			&agent.Framework,
			&agent.Model,
			&agent.Environment,
			&agent.Status,
			&agent.CreatedAt,
			&agent.UpdatedAt,
		); err != nil {
			return nil, err
		}

		agents = append(agents, agent)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return agents, nil
}

func (r *AgentRepository) GetByID(
	ctx context.Context,
	id string,
) (*models.Agent, error) {

	var agent models.Agent

	err := r.DB.QueryRow(
		ctx,
		`
		SELECT
			id,
			name,
			agent_type,
			owner,
			framework,
			model,
			environment,
			status,
			created_at,
			updated_at
		FROM agents
		WHERE id = $1
		`,
		id,
	).Scan(
		&agent.ID,
		&agent.Name,
		&agent.AgentType,
		&agent.Owner,
		&agent.Framework,
		&agent.Model,
		&agent.Environment,
		&agent.Status,
		&agent.CreatedAt,
		&agent.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrAgentNotFound
		}

		return nil, err
	}

	return &agent, nil
}
