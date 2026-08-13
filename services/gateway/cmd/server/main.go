package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/dhananjay2799/agentshield/services/gateway/internal/database"
	"github.com/dhananjay2799/agentshield/services/gateway/internal/events"
	"github.com/dhananjay2799/agentshield/services/gateway/internal/handlers"
	"github.com/dhananjay2799/agentshield/services/gateway/internal/middleware"
	"github.com/dhananjay2799/agentshield/services/gateway/internal/opa"
	"github.com/dhananjay2799/agentshield/services/gateway/internal/repository"
)

func main() {
	databaseURL := os.Getenv("DATABASE_URL")

	if databaseURL == "" {
		databaseURL = "postgres://agentshield:agentshield_dev_password@localhost:5432/agentshield"
	}

	ctx := context.Background()

	db, err := database.Connect(ctx, databaseURL)
	if err != nil {
		log.Fatalf("database connection failed: %v", err)
	}
	defer db.Close()

	log.Println("Connected to AgentShield PostgreSQL")

	if err := database.EnsurePolicySchema(
		ctx,
		db,
	); err != nil {
		log.Fatalf(
			"policy schema initialization failed: %v",
			err,
		)
	}

	agentRepository := repository.NewAgentRepository(db)
	sessionRepository := repository.NewSessionRepository(db)
	auditRepository := repository.NewAuditRepository(db)
	approvalRepository := repository.NewApprovalRepository(db)
	grantRepository := repository.NewGrantRepository(db)
	incidentRepository := repository.NewIncidentRepository(db)
	policyRepository := repository.NewPolicyRepository(db)

	opaURL := os.Getenv("OPA_URL")

	if opaURL == "" {
		opaURL = "http://localhost:8181"
	}

	opaClient := opa.NewClient(opaURL)

	expiredCount, err := grantRepository.ExpireOldGrants(ctx)
	if err != nil {
		log.Printf("failed to expire old grants: %v", err)
	} else if expiredCount > 0 {
		log.Printf("expired %d old authorization grants", expiredCount)
	}

	kafkaBroker := os.Getenv("KAFKA_BROKER")

	if kafkaBroker == "" {
		kafkaBroker = "localhost:19092"
	}

	eventProducer, err := events.NewProducer(
		[]string{kafkaBroker},
		"agentshield.security.events",
	)

	if err != nil {
		log.Fatalf("failed to create Kafka producer: %v", err)
	}

	defer eventProducer.Close()

	log.Println("Connected to AgentShield event broker")

	agentHandler := handlers.NewAgentHandler(agentRepository)
	sessionHandler := handlers.NewSessionHandler(sessionRepository)
	incidentHandler := handlers.NewIncidentHandler(incidentRepository)

	policyHandler :=
		handlers.NewPolicyHandler(
			policyRepository,
		)

	eventHandler := handlers.NewEventHandler(
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

	agentActivityHandler := handlers.NewAgentActivityHandler(
		sessionRepository,
		auditRepository,
		approvalRepository,
	)

	actionHandler := handlers.NewActionHandler(
		agentRepository,
		auditRepository,
		approvalRepository,
		grantRepository,
		opaClient,
		eventProducer,
	)
	approvalHandler := handlers.NewApprovalHandler(
		approvalRepository,
		grantRepository,
	)

	mux := http.NewServeMux()

	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"service":"agentshield-gateway","status":"healthy"}`))
	})

	mux.HandleFunc("POST /v1/agents", agentHandler.Create)
	mux.HandleFunc("GET /v1/agents", agentHandler.List)
	mux.HandleFunc("GET /v1/agents/{id}", agentHandler.GetByID)

	mux.HandleFunc("POST /v1/agents/{id}/sessions", sessionHandler.Create)
	mux.HandleFunc("GET /v1/sessions/{id}", sessionHandler.GetByID)
	mux.HandleFunc("POST /v1/sessions/{id}/revoke", sessionHandler.Revoke)

	mux.HandleFunc(
		"GET /v1/agents/{id}/sessions",
		agentActivityHandler.ListSessions,
	)

	mux.HandleFunc(
		"GET /v1/agents/{id}/sessions/security",
		agentActivityHandler.ListSessionSecurity,
	)

	mux.HandleFunc(
		"GET /v1/agents/{id}/actions",
		agentActivityHandler.ListActions,
	)

	mux.HandleFunc(
		"GET /v1/agents/{id}/approvals",
		agentActivityHandler.ListApprovals,
	)

	mux.HandleFunc(
		"GET /v1/sessions/{id}/actions",
		sessionActivityHandler.ListActions,
	)

	mux.HandleFunc(
		"GET /v1/sessions/{id}/approvals",
		sessionActivityHandler.ListApprovals,
	)

	mux.HandleFunc(
		"GET /v1/events",
		eventHandler.ListRecent,
	)

	mux.HandleFunc(
		"GET /v1/events/stream",
		eventStreamHandler.Stream,
	)

	mux.HandleFunc(
		"GET /v1/protected/test",
		middleware.SessionRequired(
			sessionRepository,
			func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusOK)
				w.Write([]byte(`{"status":"allowed","message":"valid AgentShield session"}`))
			},
		),
	)

	mux.HandleFunc(
		"POST /v1/actions/evaluate",
		middleware.SessionRequired(
			sessionRepository,
			actionHandler.Evaluate,
		),
	)

	mux.HandleFunc(
		"POST /v1/incidents/{id}/investigate",
		incidentHandler.Investigate,
	)

	mux.HandleFunc(
		"POST /v1/incidents/{id}/resolve",
		incidentHandler.Resolve,
	)

	mux.HandleFunc(
		"POST /v1/incidents/{id}/dismiss",
		incidentHandler.Dismiss,
	)

	mux.HandleFunc("GET /v1/approvals", approvalHandler.ListPending)
	mux.HandleFunc("GET /v1/approvals/{id}", approvalHandler.GetByID)
	mux.HandleFunc("POST /v1/approvals/{id}/approve", approvalHandler.Approve)
	mux.HandleFunc("POST /v1/approvals/{id}/deny", approvalHandler.Deny)
	mux.HandleFunc("GET /v1/incidents", incidentHandler.List)
	mux.HandleFunc("GET /v1/incidents/{id}", incidentHandler.GetByID)

	mux.HandleFunc(
		"GET /v1/policies",
		policyHandler.List,
	)

	mux.HandleFunc(
		"GET /v1/policies/{id}",
		policyHandler.GetByID,
	)

	mux.HandleFunc(
		"POST /v1/policies",
		policyHandler.Create,
	)

	mux.HandleFunc(
		"POST /v1/policies/{id}/validate",
		policyHandler.Validate,
	)

	server := &http.Server{
		Addr:         ":8080",
		Handler:      mux,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 0,
		IdleTimeout:  60 * time.Second,
	}

	log.Println("AgentShield Gateway starting on http://localhost:8080")

	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatal(err)
	}

}
