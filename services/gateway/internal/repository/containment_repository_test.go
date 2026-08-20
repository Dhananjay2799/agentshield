package repository

import (
	"context"
	"errors"
	"os"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/dhananjay2799/agentshield/services/gateway/internal/models"
)

func openContainmentTestDB(t *testing.T) *pgxpool.Pool {
	t.Helper()

	databaseURL := os.Getenv("AGENTSHIELD_TEST_DATABASE_URL")

	if databaseURL == "" {
		t.Skip(
			"AGENTSHIELD_TEST_DATABASE_URL is not configured",
		)
	}

	ctx, cancel := context.WithTimeout(
		context.Background(),
		10*time.Second,
	)
	defer cancel()

	db, err := pgxpool.New(
		ctx,
		databaseURL,
	)
	if err != nil {
		t.Fatalf(
			"create test database pool: %v",
			err,
		)
	}

	if err := db.Ping(ctx); err != nil {
		db.Close()

		t.Fatalf(
			"ping test database: %v",
			err,
		)
	}

	t.Cleanup(func() {
		db.Close()
	})

	return db
}

func TestContainAgentSuspendsAgentAndRevokesActiveSessions(
	t *testing.T,
) {
	db := openContainmentTestDB(t)

	ctx, cancel := context.WithTimeout(
		context.Background(),
		15*time.Second,
	)
	defer cancel()

	var agentID string

	err := db.QueryRow(
		ctx,
		`
INSERT INTO agents (
	name,
	agent_type,
	owner,
	framework,
	model,
	environment,
	status
)
VALUES (
	'containment-integration-test',
	'security-test',
	'phase38',
	'test',
	'test-model',
	'development',
	'active'
)
RETURNING id
`,
	).Scan(&agentID)

	if err != nil {
		t.Fatalf(
			"create test agent: %v",
			err,
		)
	}

	t.Cleanup(func() {
		cleanupCtx, cleanupCancel :=
			context.WithTimeout(
				context.Background(),
				5*time.Second,
			)
		defer cleanupCancel()

		_, _ = db.Exec(
			cleanupCtx,
			`
DELETE FROM audit_events
WHERE agent_id = $1
`,
			agentID,
		)

		_, _ = db.Exec(
			cleanupCtx,
			`
DELETE FROM authorization_grants
WHERE agent_id = $1
`,
			agentID,
		)

		_, _ = db.Exec(
			cleanupCtx,
			`
DELETE FROM agent_sessions
WHERE agent_id = $1
`,
			agentID,
		)

		_, _ = db.Exec(
			cleanupCtx,
			`
DELETE FROM agents
WHERE id = $1
`,
			agentID,
		)
	})

	for i := 0; i < 3; i++ {
		_, err := db.Exec(
			ctx,
			`
INSERT INTO agent_sessions (
	agent_id,
	task_id,
	status,
	expires_at
)
VALUES (
	$1,
	$2,
	'active',
	NOW() + INTERVAL '30 minutes'
)
`,
			agentID,
			"phase38-containment-test",
		)

		if err != nil {
			t.Fatalf(
				"create active session %d: %v",
				i+1,
				err,
			)
		}
	}

	repo := NewContainmentRepository(db)

	result, err := repo.ContainAgent(
		ctx,
		agentID,
	)
	if err != nil {
		t.Fatalf(
			"contain agent: %v",
			err,
		)
	}

	if result.AgentID != agentID {
		t.Fatalf(
			"unexpected agent id: got %s want %s",
			result.AgentID,
			agentID,
		)
	}

	if result.AgentStatus != "suspended" {
		t.Fatalf(
			"unexpected agent status: got %s want suspended",
			result.AgentStatus,
		)
	}

	if result.SessionsRevoked != 3 {
		t.Fatalf(
			"unexpected revoked session count: got %d want 3",
			result.SessionsRevoked,
		)
	}

	var agentStatus string

	err = db.QueryRow(
		ctx,
		`
SELECT status
FROM agents
WHERE id = $1
`,
		agentID,
	).Scan(&agentStatus)

	if err != nil {
		t.Fatalf(
			"load contained agent: %v",
			err,
		)
	}

	if agentStatus != "suspended" {
		t.Fatalf(
			"database agent status: got %s want suspended",
			agentStatus,
		)
	}

	var activeSessions int

	err = db.QueryRow(
		ctx,
		`
SELECT COUNT(*)
FROM agent_sessions
WHERE agent_id = $1
  AND status = 'active'
`,
		agentID,
	).Scan(&activeSessions)

	if err != nil {
		t.Fatalf(
			"count active sessions: %v",
			err,
		)
	}

	if activeSessions != 0 {
		t.Fatalf(
			"active sessions remain after containment: %d",
			activeSessions,
		)
	}

	var revokedSessions int

	err = db.QueryRow(
		ctx,
		`
SELECT COUNT(*)
FROM agent_sessions
WHERE agent_id = $1
  AND status = 'revoked'
  AND ended_at IS NOT NULL
`,
		agentID,
	).Scan(&revokedSessions)

	if err != nil {
		t.Fatalf(
			"count revoked sessions: %v",
			err,
		)
	}

	if revokedSessions != 3 {
		t.Fatalf(
			"revoked sessions: got %d want 3",
			revokedSessions,
		)
	}
}

func TestContainAgentIsIdempotent(
	t *testing.T,
) {
	db := openContainmentTestDB(t)

	ctx, cancel := context.WithTimeout(
		context.Background(),
		15*time.Second,
	)
	defer cancel()

	var agentID string

	err := db.QueryRow(
		ctx,
		`
INSERT INTO agents (
	name,
	agent_type,
	owner,
	framework,
	model,
	environment,
	status
)
VALUES (
	'containment-idempotency-test',
	'security-test',
	'phase38',
	'test',
	'test-model',
	'development',
	'active'
)
RETURNING id
`,
	).Scan(&agentID)

	if err != nil {
		t.Fatalf(
			"create test agent: %v",
			err,
		)
	}

	t.Cleanup(func() {
		cleanupCtx, cleanupCancel :=
			context.WithTimeout(
				context.Background(),
				5*time.Second,
			)
		defer cleanupCancel()

		_, _ = db.Exec(
			cleanupCtx,
			`DELETE FROM audit_events WHERE agent_id = $1`,
			agentID,
		)

		_, _ = db.Exec(
			cleanupCtx,
			`DELETE FROM authorization_grants WHERE agent_id = $1`,
			agentID,
		)

		_, _ = db.Exec(
			cleanupCtx,
			`DELETE FROM agent_sessions WHERE agent_id = $1`,
			agentID,
		)

		_, _ = db.Exec(
			cleanupCtx,
			`DELETE FROM agents WHERE id = $1`,
			agentID,
		)
	})

	for i := 0; i < 2; i++ {
		_, err := db.Exec(
			ctx,
			`
INSERT INTO agent_sessions (
	agent_id,
	task_id,
	status,
	expires_at
)
VALUES (
	$1,
	$2,
	'active',
	NOW() + INTERVAL '30 minutes'
)
`,
			agentID,
			"phase38-idempotency",
		)

		if err != nil {
			t.Fatalf(
				"create active session %d: %v",
				i+1,
				err,
			)
		}
	}

	repo := NewContainmentRepository(db)

	first, err := repo.ContainAgent(
		ctx,
		agentID,
	)
	if err != nil {
		t.Fatalf(
			"first containment: %v",
			err,
		)
	}

	if first.AgentStatus != "suspended" {
		t.Fatalf(
			"first containment status: got %s want suspended",
			first.AgentStatus,
		)
	}

	if first.SessionsRevoked != 2 {
		t.Fatalf(
			"first containment revoked sessions: got %d want 2",
			first.SessionsRevoked,
		)
	}

	second, err := repo.ContainAgent(
		ctx,
		agentID,
	)
	if err != nil {
		t.Fatalf(
			"second containment: %v",
			err,
		)
	}

	if second.AgentStatus != "suspended" {
		t.Fatalf(
			"second containment status: got %s want suspended",
			second.AgentStatus,
		)
	}

	if second.SessionsRevoked != 0 {
		t.Fatalf(
			"second containment revoked sessions: got %d want 0",
			second.SessionsRevoked,
		)
	}

	if second.GrantsRevoked != 0 {
		t.Fatalf(
			"second containment revoked grants: got %d want 0",
			second.GrantsRevoked,
		)
	}

	var activeSessions int

	err = db.QueryRow(
		ctx,
		`
SELECT COUNT(*)
FROM agent_sessions
WHERE agent_id = $1
  AND status = 'active'
`,
		agentID,
	).Scan(&activeSessions)

	if err != nil {
		t.Fatalf(
			"count active sessions: %v",
			err,
		)
	}

	if activeSessions != 0 {
		t.Fatalf(
			"active sessions remain after repeated containment: %d",
			activeSessions,
		)
	}
}

func TestContainAgentReturnsNotFoundForUnknownAgent(
	t *testing.T,
) {
	db := openContainmentTestDB(t)

	ctx, cancel := context.WithTimeout(
		context.Background(),
		10*time.Second,
	)
	defer cancel()

	repo := NewContainmentRepository(db)

	_, err := repo.ContainAgent(
		ctx,
		"00000000-0000-0000-0000-000000000001",
	)

	if !errors.Is(
		err,
		ErrAgentNotFound,
	) {
		t.Fatalf(
			"expected ErrAgentNotFound, got %v",
			err,
		)
	}
}

func TestSessionCreationRejectsSuspendedAgent(
	t *testing.T,
) {
	db := openContainmentTestDB(t)

	ctx, cancel := context.WithTimeout(
		context.Background(),
		15*time.Second,
	)
	defer cancel()

	var agentID string

	err := db.QueryRow(
		ctx,
		`
INSERT INTO agents (
	name,
	agent_type,
	owner,
	framework,
	model,
	environment,
	status
)
VALUES (
	'suspended-agent-session-test',
	'security-test',
	'phase38',
	'test',
	'test-model',
	'development',
	'suspended'
)
RETURNING id
`,
	).Scan(&agentID)

	if err != nil {
		t.Fatalf(
			"create suspended test agent: %v",
			err,
		)
	}

	t.Cleanup(func() {
		cleanupCtx, cleanupCancel :=
			context.WithTimeout(
				context.Background(),
				5*time.Second,
			)
		defer cleanupCancel()

		_, _ = db.Exec(
			cleanupCtx,
			`DELETE FROM agent_sessions WHERE agent_id = $1`,
			agentID,
		)

		_, _ = db.Exec(
			cleanupCtx,
			`DELETE FROM agents WHERE id = $1`,
			agentID,
		)
	})

	repo := NewSessionRepository(db)

	session, err := repo.Create(
		ctx,
		agentID,
		models.CreateSessionRequest{
			TaskID:     "phase38-suspended-agent-test",
			TTLMinutes: 30,
		},
	)

	if session != nil {
		t.Fatalf(
			"expected no session, got %+v",
			session,
		)
	}

	if !errors.Is(
		err,
		ErrAgentNotActive,
	) {
		t.Fatalf(
			"expected ErrAgentNotActive, got %v",
			err,
		)
	}

	var sessionCount int

	err = db.QueryRow(
		ctx,
		`
SELECT COUNT(*)
FROM agent_sessions
WHERE agent_id = $1
`,
		agentID,
	).Scan(&sessionCount)

	if err != nil {
		t.Fatalf(
			"count sessions: %v",
			err,
		)
	}

	if sessionCount != 0 {
		t.Fatalf(
			"suspended agent created %d sessions",
			sessionCount,
		)
	}
}
