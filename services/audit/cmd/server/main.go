package main

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/dhananjay2799/agentshield/services/audit/internal/consumer"
)

func main() {
	kafkaBroker :=
		os.Getenv(
			"KAFKA_BROKER",
		)

	if kafkaBroker == "" {
		kafkaBroker =
			"localhost:19092"
	}

	topic :=
		os.Getenv(
			"KAFKA_SECURITY_TOPIC",
		)

	if topic == "" {
		topic =
			"agentshield.security.events"
	}

	groupID :=
		os.Getenv(
			"KAFKA_CONSUMER_GROUP",
		)

	if groupID == "" {
		groupID =
			"agentshield-audit-service"
	}

	httpAddress :=
		os.Getenv(
			"AUDIT_HTTP_ADDRESS",
		)

	if httpAddress == "" {
		httpAddress = ":8082"
	}

	auditConsumer, err :=
		consumer.New(
			[]string{
				kafkaBroker,
			},
			topic,
			groupID,
		)

	if err != nil {
		log.Fatalf(
			"failed to create audit consumer: %v",
			err,
		)
	}

	defer auditConsumer.Close()

	ctx, stop :=
		signal.NotifyContext(
			context.Background(),
			os.Interrupt,
			syscall.SIGTERM,
		)

	defer stop()

	mux := http.NewServeMux()

	mux.HandleFunc(
		"GET /health",
		func(
			w http.ResponseWriter,
			r *http.Request,
		) {
			writeJSON(
				w,
				http.StatusOK,
				map[string]any{
					"service": "agentshield-audit",
					"status":  "healthy",
				},
			)
		},
	)

	mux.HandleFunc(
		"GET /ready",
		func(
			w http.ResponseWriter,
			r *http.Request,
		) {
			readyCtx, cancel :=
				context.WithTimeout(
					r.Context(),
					2*time.Second,
				)

			defer cancel()

			if err :=
				auditConsumer.Ready(
					readyCtx,
				); err != nil {
				writeJSON(
					w,
					http.StatusServiceUnavailable,
					map[string]any{
						"service": "agentshield-audit",
						"status":  "not_ready",
						"reason":  "Kafka broker unavailable",
					},
				)
				return
			}

			writeJSON(
				w,
				http.StatusOK,
				map[string]any{
					"service": "agentshield-audit",
					"status":  "ready",
					"kafka":   "connected",
				},
			)
		},
	)

	mux.HandleFunc(
		"GET /metrics",
		func(
			w http.ResponseWriter,
			r *http.Request,
		) {
			writeJSON(
				w,
				http.StatusOK,
				auditConsumer.Metrics(),
			)
		},
	)

	server := &http.Server{
		Addr:              httpAddress,
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
	}

	serverErrors :=
		make(chan error, 1)

	go func() {
		log.Printf(
			"AgentShield Audit HTTP server starting on http://localhost%s",
			httpAddress,
		)

		err :=
			server.ListenAndServe()

		if err != nil &&
			!errors.Is(
				err,
				http.ErrServerClosed,
			) {
			serverErrors <- err
			return
		}

		serverErrors <- nil
	}()

	log.Printf(
		"AgentShield Audit Service starting: broker=%s topic=%s group=%s",
		kafkaBroker,
		topic,
		groupID,
	)

	consumerErrors :=
		make(chan error, 1)

	go func() {
		consumerErrors <- auditConsumer.Run(ctx)
	}()

	select {
	case <-ctx.Done():
		log.Println(
			"shutdown signal received",
		)

	case err := <-consumerErrors:
		if err != nil &&
			!errors.Is(
				err,
				context.Canceled,
			) {
			log.Printf(
				"audit consumer stopped unexpectedly: %v",
				err,
			)
		}

		stop()

	case err := <-serverErrors:
		if err != nil {
			log.Printf(
				"audit HTTP server stopped unexpectedly: %v",
				err,
			)
		}

		stop()
	}

	shutdownCtx, cancel :=
		context.WithTimeout(
			context.Background(),
			5*time.Second,
		)

	defer cancel()

	if err := server.Shutdown(
		shutdownCtx,
	); err != nil {
		log.Printf(
			"failed to gracefully stop audit HTTP server: %v",
			err,
		)
	}

	log.Println(
		"AgentShield Audit Service stopped",
	)
}

func writeJSON(
	w http.ResponseWriter,
	status int,
	value any,
) {
	w.Header().Set(
		"Content-Type",
		"application/json",
	)

	w.WriteHeader(status)

	_ = json.NewEncoder(
		w,
	).Encode(value)
}
