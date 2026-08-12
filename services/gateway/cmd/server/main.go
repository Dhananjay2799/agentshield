package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/dhananjay2799/agentshield/services/gateway/internal/database"
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

	agentRepository := repository.NewAgentRepository(db)
	sessionRepository := repository.NewSessionRepository(db)
	auditRepository := repository.NewAuditRepository(db)
	approvalRepository := repository.NewApprovalRepository(db)
	grantRepository := repository.NewGrantRepository(db)

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

	agentHandler := handlers.NewAgentHandler(agentRepository)
	sessionHandler := handlers.NewSessionHandler(sessionRepository)
	actionHandler := handlers.NewActionHandler(
		agentRepository,
		auditRepository,
		approvalRepository,
		grantRepository,
		opaClient,
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

	mux.HandleFunc("GET /v1/approvals", approvalHandler.ListPending)
	mux.HandleFunc("GET /v1/approvals/{id}", approvalHandler.GetByID)
	mux.HandleFunc("POST /v1/approvals/{id}/approve", approvalHandler.Approve)
	mux.HandleFunc("POST /v1/approvals/{id}/deny", approvalHandler.Deny)

	server := &http.Server{
		Addr:         ":8080",
		Handler:      mux,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	log.Println("AgentShield Gateway starting on http://localhost:8080")

	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatal(err)
	}

}
