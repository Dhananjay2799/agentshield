package database

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

func EnsurePolicySchema(
	ctx context.Context,
	db *pgxpool.Pool,
) error {

	const schema = `
	CREATE TABLE IF NOT EXISTS policies (
		id UUID PRIMARY KEY DEFAULT gen_random_uuid(),

		name TEXT NOT NULL,
		description TEXT NOT NULL DEFAULT '',

		effect TEXT NOT NULL
			CHECK (
				effect IN (
					'ALLOW',
					'REQUIRE_APPROVAL',
					'DENY'
				)
			),

		status TEXT NOT NULL DEFAULT 'draft'
			CHECK (
				status IN (
					'draft',
					'active',
					'disabled',
					'archived'
				)
			),

		priority INTEGER NOT NULL DEFAULT 100,

		agent_type TEXT NULL,

		action TEXT NOT NULL,

		action_match TEXT NOT NULL DEFAULT 'exact'
			CHECK (
				action_match IN (
					'exact',
					'prefix',
					'suffix'
				)
			),

		resource TEXT NOT NULL,

		resource_match TEXT NOT NULL DEFAULT 'prefix'
			CHECK (
				resource_match IN (
					'exact',
					'prefix',
					'suffix'
				)
			),

		environment TEXT NULL,

		version INTEGER NOT NULL DEFAULT 1,

		source TEXT NOT NULL DEFAULT 'control_plane',

		created_by TEXT NOT NULL DEFAULT 'system',

		created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

		updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
	);

	CREATE INDEX IF NOT EXISTS idx_policies_status
	ON policies(status);

	CREATE INDEX IF NOT EXISTS idx_policies_effect
	ON policies(effect);

	CREATE INDEX IF NOT EXISTS idx_policies_action
	ON policies(action);

	CREATE INDEX IF NOT EXISTS idx_policies_priority
	ON policies(priority);
	`

	if _, err := db.Exec(
		ctx,
		schema,
	); err != nil {
		return fmt.Errorf(
			"ensure policy schema: %w",
			err,
		)
	}

	return nil
}
