package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/dhananjay2799/agentshield/services/gateway/internal/database"
	"github.com/dhananjay2799/agentshield/services/gateway/internal/events"
	"github.com/dhananjay2799/agentshield/services/gateway/internal/handlers"
	gatewaymetrics "github.com/dhananjay2799/agentshield/services/gateway/internal/metrics"
	"github.com/dhananjay2799/agentshield/services/gateway/internal/middleware"
	"github.com/dhananjay2799/agentshield/services/gateway/internal/opa"
	"github.com/dhananjay2799/agentshield/services/gateway/internal/repository"
)

func main() {
	// ============================================================
	// DATABASE
	// ============================================================

	databaseURL := os.Getenv("DATABASE_URL")

	if databaseURL == "" {
		databaseURL =
			"postgres://agentshield:agentshield_dev_password@localhost:5432/agentshield"
	}

	ctx := context.Background()

	db, err := database.Connect(
		ctx,
		databaseURL,
	)
	if err != nil {
		log.Fatalf(
			"database connection failed: %v",
			err,
		)
	}

	defer db.Close()

	log.Println(
		"Connected to AgentShield PostgreSQL",
	)

	if err := database.EnsurePolicySchema(
		ctx,
		db,
	); err != nil {
		log.Fatalf(
			"policy schema initialization failed: %v",
			err,
		)
	}

	// ============================================================
	// REPOSITORIES
	// ============================================================

	agentRepository :=
		repository.NewAgentRepository(db)

	sessionRepository :=
		repository.NewSessionRepository(db)

	auditRepository :=
		repository.NewAuditRepository(db)

	approvalRepository :=
		repository.NewApprovalRepository(db)

	grantRepository :=
		repository.NewGrantRepository(db)

	incidentRepository :=
		repository.NewIncidentRepository(db)

	policyRepository :=
		repository.NewPolicyRepository(db)

	containmentRepository :=
		repository.NewContainmentRepository(db)

	// ============================================================
	// METRICS
	// ============================================================

	gatewayMetrics :=
		gatewaymetrics.New(
			incidentRepository,
		)

	// ============================================================
	// OPA
	// ============================================================

	opaURL := os.Getenv("OPA_URL")

	if opaURL == "" {
		opaURL =
			"http://localhost:8181"
	}

	opaClient :=
		opa.NewClient(
			opaURL,
		)

	activePolicies, err :=
		policyRepository.ListActive(
			ctx,
		)

	if err != nil {
		log.Fatalf(
			"failed to load active policies for OPA reconciliation: %v",
			err,
		)
	}

	for i := range activePolicies {
		activePolicy :=
			&activePolicies[i]

		if err :=
			opaClient.PutManagedPolicy(
				ctx,
				activePolicy,
			); err != nil {

			log.Fatalf(
				"failed to synchronize active policy %s with OPA: %v",
				activePolicy.ID,
				err,
			)
		}

		log.Printf(
			"synchronized active policy with OPA: id=%s name=%s",
			activePolicy.ID,
			activePolicy.Name,
		)
	}

	// ============================================================
	// GRANT EXPIRATION
	// ============================================================

	expiredCount, err :=
		grantRepository.ExpireOldGrants(
			ctx,
		)

	if err != nil {
		log.Printf(
			"failed to expire old grants: %v",
			err,
		)
	} else if expiredCount > 0 {
		log.Printf(
			"expired %d old authorization grants",
			expiredCount,
		)
	}

	// ============================================================
	// KAFKA / REDPANDA
	// ============================================================

	kafkaBroker :=
		os.Getenv(
			"KAFKA_BROKER",
		)

	if kafkaBroker == "" {
		kafkaBroker =
			"localhost:19092"
	}

	eventProducer, err :=
		events.NewProducer(
			[]string{
				kafkaBroker,
			},
			"agentshield.security.events",
		)

	if err != nil {
		log.Fatalf(
			"failed to create Kafka producer: %v",
			err,
		)
	}

	defer eventProducer.Close()

	log.Println(
		"Connected to AgentShield event broker",
	)

	// ============================================================
	// HANDLERS
	// ============================================================

	agentHandler :=
		handlers.NewAgentHandler(
			agentRepository,
		)

	sessionHandler :=
		handlers.NewSessionHandler(
			sessionRepository,
		)

	incidentHandler :=
		handlers.NewIncidentHandler(
			incidentRepository,
		)

	policyHandler :=
		handlers.NewPolicyHandler(
			policyRepository,
			auditRepository,
			opaClient,
			eventProducer,
		)

	eventHandler :=
		handlers.NewEventHandler(
			auditRepository,
		)

	eventStreamHandler :=
		handlers.NewEventStreamHandler(
			eventProducer.Hub,
		)

	sessionActivityHandler :=
		handlers.NewSessionActivityHandler(
			auditRepository,
			approvalRepository,
		)

	agentActivityHandler :=
		handlers.NewAgentActivityHandler(
			sessionRepository,
			auditRepository,
			approvalRepository,
		)

	actionHandler :=
		handlers.NewActionHandler(
			agentRepository,
			auditRepository,
			approvalRepository,
			grantRepository,
			opaClient,
			eventProducer,
			gatewayMetrics,
		)

	approvalHandler :=
		handlers.NewApprovalHandler(
			approvalRepository,
			grantRepository,
		)

	grantHandler :=
		handlers.NewGrantHandler(
			grantRepository,
		)

	containmentHandler :=
		handlers.NewContainmentHandler(
			containmentRepository,
			auditRepository,
			eventProducer,
			gatewayMetrics,
		)

	// ============================================================
	// AUTHENTICATION
	// ============================================================

	adminAPIKey :=
		os.Getenv(
			"AGENTSHIELD_ADMIN_API_KEY",
		)

	analystAPIKey :=
		os.Getenv(
			"AGENTSHIELD_ANALYST_API_KEY",
		)

	credentialBrokerAPIKey :=
		os.Getenv(
			"AGENTSHIELD_CREDENTIAL_BROKER_API_KEY",
		)

	if adminAPIKey == "" {
		log.Fatal(
			"AGENTSHIELD_ADMIN_API_KEY is required",
		)
	}

	if analystAPIKey == "" {
		log.Fatal(
			"AGENTSHIELD_ANALYST_API_KEY is required",
		)
	}

	if credentialBrokerAPIKey == "" {
		log.Fatal(
			"AGENTSHIELD_CREDENTIAL_BROKER_API_KEY is required",
		)
	}

	operatorAuth :=
		middleware.NewAPIKeyAuth(
			adminAPIKey,
			analystAPIKey,
			credentialBrokerAPIKey,
		)

	// ============================================================
	// RATE LIMITERS
	// ============================================================

	// Autonomous-agent action evaluations.
	// 120 sustained requests/minute with burst capacity of 20.
	actionLimiter :=
		middleware.NewRateLimiter(
			120,
			20,
			middleware.SessionRateLimitKey,
		)

	// Human analyst/admin control-plane operations.
	// 300 sustained requests/minute with burst capacity of 40.
	operatorLimiter :=
		middleware.NewRateLimiter(
			300,
			40,
			middleware.PrincipalRateLimitKey,
		)

	// Credential Broker service calls.
	// More restrictive because grant claims are privileged.
	serviceLimiter :=
		middleware.NewRateLimiter(
			60,
			10,
			middleware.PrincipalRateLimitKey,
		)

	// Public health endpoint.
	healthLimiter :=
		middleware.NewRateLimiter(
			120,
			20,
			middleware.IPRateLimitKey,
		)

	// ============================================================
	// AUTHORIZATION + RATE-LIMIT WRAPPERS
	// ============================================================

	limitedAnalystOrAdmin :=
		func(
			handler http.HandlerFunc,
		) http.HandlerFunc {
			return operatorAuth.Required(
				middleware.RequireRole(
					"analyst",
					"admin",
				)(
					operatorLimiter.Middleware(
						handler,
					),
				),
			)
		}

	limitedAdminOnly :=
		func(
			handler http.HandlerFunc,
		) http.HandlerFunc {
			return operatorAuth.Required(
				middleware.RequireRole(
					"admin",
				)(
					operatorLimiter.Middleware(
						handler,
					),
				),
			)
		}

	limitedServiceOnly :=
		func(
			handler http.HandlerFunc,
		) http.HandlerFunc {
			return operatorAuth.Required(
				middleware.RequireRole(
					"service",
				)(
					serviceLimiter.Middleware(
						handler,
					),
				),
			)
		}

	json64KB :=
		func(
			handler http.HandlerFunc,
		) http.HandlerFunc {
			return middleware.RequireJSON(
				middleware.LimitBody(
					64*1024,
					handler,
				),
			)
		}

	// ============================================================
	// ROUTER
	// ============================================================

	mux :=
		http.NewServeMux()

	// ------------------------------------------------------------
	// PUBLIC HEALTH & METRICS
	// ------------------------------------------------------------

	mux.HandleFunc(
		"GET /health",
		healthLimiter.Middleware(
			func(
				w http.ResponseWriter,
				r *http.Request,
			) {
				w.Header().Set(
					"Content-Type",
					"application/json",
				)

				w.WriteHeader(
					http.StatusOK,
				)

				_, _ = w.Write(
					[]byte(
						`{"service":"agentshield-gateway","status":"healthy"}`,
					),
				)
			},
		),
	)

	mux.HandleFunc(
		"GET /metrics",
		gatewayMetrics.Handler,
	)

	mux.HandleFunc(
		"GET /debug/metrics",
		gatewayMetrics.DebugHandler,
	)

	// ------------------------------------------------------------
	// AUTHENTICATED IDENTITY TEST
	// ------------------------------------------------------------

	mux.HandleFunc(
		"GET /v1/auth/whoami",
		operatorAuth.Required(
			func(
				w http.ResponseWriter,
				r *http.Request,
			) {
				principal, ok :=
					middleware.PrincipalFromContext(
						r.Context(),
					)

				if !ok {
					w.Header().Set(
						"Content-Type",
						"application/json",
					)

					w.WriteHeader(
						http.StatusInternalServerError,
					)

					_ = json.NewEncoder(
						w,
					).Encode(
						map[string]string{
							"error": "authenticated principal unavailable",
						},
					)

					return
				}

				w.Header().Set(
					"Content-Type",
					"application/json",
				)

				_ = json.NewEncoder(
					w,
				).Encode(
					principal,
				)
			},
		),
	)

	// ------------------------------------------------------------
	// AGENTS
	//
	// These routes are not yet operator-RBAC protected because
	// agent registration / identity hardening will be handled
	// separately.
	// ------------------------------------------------------------

	mux.HandleFunc(
		"POST /v1/agents",
		json64KB(
			agentHandler.Create,
		),
	)

	mux.HandleFunc(
		"POST /v1/agents/{id}/sessions",
		json64KB(
			sessionHandler.Create,
		),
	)

	// ------------------------------------------------------------
	// EMERGENCY AGENT CONTAINMENT
	//
	// Admin-only destructive security operation.
	// Suspends the agent and revokes active sessions and grants.
	// ------------------------------------------------------------

	mux.HandleFunc(
		"POST /v1/agents/{id}/contain",
		limitedAdminOnly(
			containmentHandler.Contain,
		),
	)

	mux.HandleFunc(
		"POST /v1/actions/evaluate",
		middleware.SessionRequired(
			sessionRepository,
			actionLimiter.Middleware(
				json64KB(
					actionHandler.Evaluate,
				),
			),
		),
	)

	mux.HandleFunc(
		"POST /v1/grants/{id}/claim",
		limitedServiceOnly(
			json64KB(
				grantHandler.ClaimForCredential,
			),
		),
	)

	mux.HandleFunc(
		"POST /v1/policies",
		limitedAdminOnly(
			json64KB(
				policyHandler.Create,
			),
		),
	)

	mux.HandleFunc(
		"GET /v1/agents",
		agentHandler.List,
	)

	mux.HandleFunc(
		"GET /v1/agents/{id}",
		agentHandler.GetByID,
	)

	// ------------------------------------------------------------
	// SESSIONS
	// ------------------------------------------------------------

	mux.HandleFunc(
		"GET /v1/sessions/{id}",
		sessionHandler.GetByID,
	)

	mux.HandleFunc(
		"POST /v1/sessions/{id}/revoke",
		sessionHandler.Revoke,
	)

	// ------------------------------------------------------------
	// AGENT / SESSION SECURITY ACTIVITY
	// Analyst or admin + operator rate limiting.
	// ------------------------------------------------------------

	mux.HandleFunc(
		"GET /v1/agents/{id}/sessions",
		limitedAnalystOrAdmin(
			agentActivityHandler.ListSessions,
		),
	)

	mux.HandleFunc(
		"GET /v1/agents/{id}/sessions/security",
		limitedAnalystOrAdmin(
			agentActivityHandler.ListSessionSecurity,
		),
	)

	mux.HandleFunc(
		"GET /v1/agents/{id}/actions",
		limitedAnalystOrAdmin(
			agentActivityHandler.ListActions,
		),
	)

	mux.HandleFunc(
		"GET /v1/agents/{id}/approvals",
		limitedAnalystOrAdmin(
			agentActivityHandler.ListApprovals,
		),
	)

	mux.HandleFunc(
		"GET /v1/sessions/{id}/actions",
		limitedAnalystOrAdmin(
			sessionActivityHandler.ListActions,
		),
	)

	mux.HandleFunc(
		"GET /v1/sessions/{id}/approvals",
		limitedAnalystOrAdmin(
			sessionActivityHandler.ListApprovals,
		),
	)

	// ------------------------------------------------------------
	// SECURITY EVENTS
	// ------------------------------------------------------------

	mux.HandleFunc(
		"GET /v1/events",
		limitedAnalystOrAdmin(
			eventHandler.ListRecent,
		),
	)

	mux.HandleFunc(
		"GET /v1/events/stream",
		limitedAnalystOrAdmin(
			eventStreamHandler.Stream,
		),
	)

	// ------------------------------------------------------------
	// SESSION AUTHENTICATION TEST
	// ------------------------------------------------------------

	mux.HandleFunc(
		"GET /v1/protected/test",
		middleware.SessionRequired(
			sessionRepository,
			func(
				w http.ResponseWriter,
				r *http.Request,
			) {
				w.Header().Set(
					"Content-Type",
					"application/json",
				)

				w.WriteHeader(
					http.StatusOK,
				)

				_, _ = w.Write(
					[]byte(
						`{"status":"allowed","message":"valid AgentShield session"}`,
					),
				)
			},
		),
	)

	// ------------------------------------------------------------
	// INCIDENT MANAGEMENT
	// ------------------------------------------------------------

	mux.HandleFunc(
		"GET /v1/incidents",
		limitedAnalystOrAdmin(
			incidentHandler.List,
		),
	)

	mux.HandleFunc(
		"GET /v1/incidents/{id}",
		limitedAnalystOrAdmin(
			incidentHandler.GetByID,
		),
	)

	mux.HandleFunc(
		"POST /v1/incidents/{id}/investigate",
		limitedAnalystOrAdmin(
			incidentHandler.Investigate,
		),
	)

	mux.HandleFunc(
		"POST /v1/incidents/{id}/resolve",
		limitedAnalystOrAdmin(
			incidentHandler.Resolve,
		),
	)

	mux.HandleFunc(
		"POST /v1/incidents/{id}/dismiss",
		limitedAnalystOrAdmin(
			incidentHandler.Dismiss,
		),
	)

	// ------------------------------------------------------------
	// HUMAN APPROVAL WORKFLOW
	// ------------------------------------------------------------

	mux.HandleFunc(
		"GET /v1/approvals",
		limitedAnalystOrAdmin(
			approvalHandler.ListPending,
		),
	)

	mux.HandleFunc(
		"GET /v1/approvals/{id}",
		limitedAnalystOrAdmin(
			approvalHandler.GetByID,
		),
	)

	mux.HandleFunc(
		"GET /v1/approvals/{id}/lineage",
		limitedAnalystOrAdmin(
			approvalHandler.GetLineage,
		),
	)

	mux.HandleFunc(
		"POST /v1/approvals/{id}/approve",
		limitedAnalystOrAdmin(
			approvalHandler.Approve,
		),
	)

	mux.HandleFunc(
		"POST /v1/approvals/{id}/deny",
		limitedAnalystOrAdmin(
			approvalHandler.Deny,
		),
	)

	// ------------------------------------------------------------
	// AUTHORIZATION GRANTS
	//
	// GET/verify:
	// security administrator only.
	//
	// claim:
	// Credential Broker workload identity only.
	// ------------------------------------------------------------

	mux.HandleFunc(
		"GET /v1/grants/{id}",
		limitedAdminOnly(
			grantHandler.GetByID,
		),
	)

	mux.HandleFunc(
		"GET /v1/grants/{id}/verify",
		limitedAdminOnly(
			grantHandler.Verify,
		),
	)

	// ------------------------------------------------------------
	// POLICY CONTROL PLANE
	//
	// Analysts:
	// read-only.
	//
	// Admin:
	// read + mutation.
	// ------------------------------------------------------------

	mux.HandleFunc(
		"GET /v1/policies",
		limitedAnalystOrAdmin(
			policyHandler.List,
		),
	)

	mux.HandleFunc(
		"GET /v1/policies/{id}",
		limitedAnalystOrAdmin(
			policyHandler.GetByID,
		),
	)

	mux.HandleFunc(
		"POST /v1/policies/{id}/validate",
		limitedAdminOnly(
			policyHandler.Validate,
		),
	)

	mux.HandleFunc(
		"POST /v1/policies/{id}/activate",
		limitedAdminOnly(
			policyHandler.Activate,
		),
	)

	mux.HandleFunc(
		"POST /v1/policies/{id}/deactivate",
		limitedAdminOnly(
			policyHandler.Deactivate,
		),
	)

	// ============================================================
	// HTTP SERVER
	// ============================================================

	server :=
		&http.Server{
			Addr: ":8080",

			Handler: middleware.SecurityHeaders(
				mux,
			),

			ReadHeaderTimeout: 5 * time.Second,

			ReadTimeout: 10 * time.Second,

			// Keep unlimited for SSE/event streaming.
			WriteTimeout: 0,

			IdleTimeout: 60 * time.Second,

			MaxHeaderBytes: 1 << 20,
		}

	log.Println(
		"AgentShield Gateway starting on http://localhost:8080",
	)

	if err :=
		server.ListenAndServe(); err != nil &&
		err != http.ErrServerClosed {

		log.Fatal(err)
	}
}
