package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/dhananjay2799/agentshield/services/credential-broker/internal/gateway"
	"github.com/dhananjay2799/agentshield/services/credential-broker/internal/handlers"
	brokermetrics "github.com/dhananjay2799/agentshield/services/credential-broker/internal/metrics"
	"github.com/dhananjay2799/agentshield/services/credential-broker/internal/token"
)

func main() {
	signingSecret :=
		os.Getenv(
			"AGENTSHIELD_CREDENTIAL_SIGNING_SECRET",
		)

	if signingSecret == "" {
		log.Fatal(
			"AGENTSHIELD_CREDENTIAL_SIGNING_SECRET is required",
		)
	}

	issuer, err := token.NewIssuer(
		signingSecret,
		5*time.Minute,
	)

	if err != nil {
		log.Fatal(err)
	}

	gatewayURL :=
		os.Getenv(
			"AGENTSHIELD_GATEWAY_URL",
		)

	if gatewayURL == "" {
		gatewayURL =
			"http://localhost:8080"
	}

	gatewayAPIKey :=
		os.Getenv(
			"AGENTSHIELD_CREDENTIAL_BROKER_API_KEY",
		)

	if gatewayAPIKey == "" {
		log.Fatal(
			"AGENTSHIELD_CREDENTIAL_BROKER_API_KEY is required",
		)
	}

	gatewayClient :=
		gateway.NewClient(
			gatewayURL,
			gatewayAPIKey,
		)

	brokerMetrics :=
		brokermetrics.New()

	credentialHandler :=
		handlers.NewCredentialHandler(
			issuer,
			gatewayClient,
			brokerMetrics,
		)

	mux := http.NewServeMux()

	mux.HandleFunc(
		"GET /health",
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

			_ = json.NewEncoder(
				w,
			).Encode(
				map[string]string{
					"service": "agentshield-credential-broker",
					"status":  "healthy",
				},
			)
		},
	)

	mux.HandleFunc(
		"GET /metrics",
		brokerMetrics.Handler,
	)

	mux.HandleFunc(
		"GET /debug/metrics",
		brokerMetrics.DebugHandler,
	)

	mux.HandleFunc(
		"POST /v1/credentials/issue",
		credentialHandler.Issue,
	)

	address := ":8081"

	log.Printf(
		"AgentShield Credential Broker starting on http://localhost%s",
		address,
	)

	if err := http.ListenAndServe(
		address,
		mux,
	); err != nil {
		log.Fatal(err)
	}
}
