package main

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/twmb/franz-go/pkg/kgo"

	"github.com/dhananjay2799/agentshield/services/detection/internal/database"
	"github.com/dhananjay2799/agentshield/services/detection/internal/dlq"
	detectionmetrics "github.com/dhananjay2799/agentshield/services/detection/internal/metrics"
	"github.com/dhananjay2799/agentshield/services/detection/internal/repository"
)

type SecurityEvent struct {
	EventType  string         `json:"event_type"`
	AgentID    string         `json:"agent_id"`
	SessionID  string         `json:"session_id"`
	Action     string         `json:"action"`
	Resource   string         `json:"resource"`
	Decision   string         `json:"decision"`
	RiskScore  int            `json:"risk_score"`
	Metadata   map[string]any `json:"metadata"`
	OccurredAt time.Time      `json:"occurred_at"`
}

type AgentActivity struct {
	DeniedEvents []time.Time
	HighRisk     []time.Time
}

type Detector struct {
	mu                 sync.Mutex
	activity           map[string]*AgentActivity
	IncidentRepository *repository.IncidentRepository
	Metrics            *detectionmetrics.Metrics
}

func NewDetector(
	incidentRepository *repository.IncidentRepository,
	metrics *detectionmetrics.Metrics,
) *Detector {
	return &Detector{
		activity:           make(map[string]*AgentActivity),
		IncidentRepository: incidentRepository,
		Metrics:            metrics,
	}
}

func (d *Detector) Process(
	event SecurityEvent,
) {
	d.mu.Lock()
	defer d.mu.Unlock()

	now := time.Now().UTC()
	windowStart :=
		now.Add(-60 * time.Second)

	activity, exists :=
		d.activity[event.AgentID]

	if !exists {
		activity =
			&AgentActivity{}

		d.activity[event.AgentID] =
			activity
	}

	activity.DeniedEvents =
		pruneOld(
			activity.DeniedEvents,
			windowStart,
		)

		// ---------------------------------------------------------
	// SINGLE-EVENT DETECTIONS
	// ---------------------------------------------------------

	if event.Action == "database.delete" &&
		len(event.Resource) >= len("production/") &&
		event.Resource[:len("production/")] == "production/" {

		log.Printf(
			"CRITICAL: destructive production database action detected agent=%s session=%s resource=%s",
			event.AgentID,
			event.SessionID,
			event.Resource,
		)

		err := d.IncidentRepository.UpsertOpenIncident(
			context.Background(),
			repository.UpsertIncidentParams{
				AgentID:      event.AgentID,
				SessionID:    event.SessionID,
				IncidentType: "destructive_action",
				Severity:     "critical",
				Title:        "Destructive production database action detected",
				Description:  "An autonomous agent attempted a destructive database operation against a production resource.",
				Metadata: map[string]any{
					"event_type": event.EventType,
					"action":     event.Action,
					"resource":   event.Resource,
					"decision":   event.Decision,
					"risk_score": event.RiskScore,
				},
			},
		)

		if err != nil {
			log.Printf(
				"failed to persist destructive-action incident: %v",
				err,
			)
		} else {
			d.Metrics.RecordIncident()
		}
	}

	if event.Action == "secrets.read" &&
		len(event.Resource) >= len("production/") &&
		event.Resource[:len("production/")] == "production/" {

		log.Printf(
			"HIGH: production credential-access activity detected agent=%s session=%s resource=%s",
			event.AgentID,
			event.SessionID,
			event.Resource,
		)

		err := d.IncidentRepository.UpsertOpenIncident(
			context.Background(),
			repository.UpsertIncidentParams{
				AgentID:      event.AgentID,
				SessionID:    event.SessionID,
				IncidentType: "credential_access",
				Severity:     "high",
				Title:        "Production credential access detected",
				Description:  "An autonomous agent attempted to access a production secret or credential resource.",
				Metadata: map[string]any{
					"event_type": event.EventType,
					"action":     event.Action,
					"resource":   event.Resource,
					"decision":   event.Decision,
					"risk_score": event.RiskScore,
				},
			},
		)

		if err != nil {
			log.Printf(
				"failed to persist credential-access incident: %v",
				err,
			)
		} else {
			d.Metrics.RecordIncident()
		}
	}

	if event.Decision == "DENY" &&
		event.RiskScore >= 80 {

		log.Printf(
			"CRITICAL: high-risk denied action detected agent=%s session=%s action=%s risk=%d",
			event.AgentID,
			event.SessionID,
			event.Action,
			event.RiskScore,
		)

		err := d.IncidentRepository.UpsertOpenIncident(
			context.Background(),
			repository.UpsertIncidentParams{
				AgentID:      event.AgentID,
				SessionID:    event.SessionID,
				IncidentType: "high_risk_denied_action",
				Severity:     "critical",
				Title:        "High-risk denied agent action",
				Description:  "AgentShield blocked a high-risk autonomous-agent action.",
				Metadata: map[string]any{
					"event_type": event.EventType,
					"action":     event.Action,
					"resource":   event.Resource,
					"decision":   event.Decision,
					"risk_score": event.RiskScore,
				},
			},
		)

		if err != nil {
			log.Printf(
				"failed to persist high-risk-denied incident: %v",
				err,
			)
		} else {
			d.Metrics.RecordIncident()
		}
	}

	if event.Decision == "REQUIRE_APPROVAL" &&
		event.RiskScore >= 80 {

		log.Printf(
			"HIGH: sensitive action escalated for human approval agent=%s session=%s action=%s risk=%d",
			event.AgentID,
			event.SessionID,
			event.Action,
			event.RiskScore,
		)

		err := d.IncidentRepository.UpsertOpenIncident(
			context.Background(),
			repository.UpsertIncidentParams{
				AgentID:      event.AgentID,
				SessionID:    event.SessionID,
				IncidentType: "sensitive_action_escalation",
				Severity:     "high",
				Title:        "Sensitive action requires human approval",
				Description:  "A high-risk autonomous-agent action was escalated for explicit human authorization.",
				Metadata: map[string]any{
					"event_type": event.EventType,
					"action":     event.Action,
					"resource":   event.Resource,
					"decision":   event.Decision,
					"risk_score": event.RiskScore,
				},
			},
		)

		if err != nil {
			log.Printf(
				"failed to persist sensitive-action escalation incident: %v",
				err,
			)
		} else {
			d.Metrics.RecordIncident()
		}
	}

	activity.HighRisk =
		pruneOld(
			activity.HighRisk,
			windowStart,
		)

	if event.Decision == "DENY" {
		d.Metrics.RecordDenied()

		activity.DeniedEvents =
			append(
				activity.DeniedEvents,
				now,
			)

		log.Printf(
			"ALERT: denied action detected agent=%s action=%s resource=%s",
			event.AgentID,
			event.Action,
			event.Resource,
		)
	}

	if event.RiskScore >= 80 {
		d.Metrics.RecordHighRisk()

		activity.HighRisk =
			append(
				activity.HighRisk,
				now,
			)

		log.Printf(
			"ALERT: high-risk agent behavior agent=%s risk=%d",
			event.AgentID,
			event.RiskScore,
		)
	}

	if len(
		activity.DeniedEvents,
	) >= 5 {
		log.Printf(
			"CRITICAL: repeated denied actions agent=%s count=%d window=60s possible compromise or privilege escalation",
			event.AgentID,
			len(
				activity.DeniedEvents,
			),
		)

		err :=
			d.IncidentRepository.
				UpsertOpenIncident(
					context.Background(),
					repository.UpsertIncidentParams{
						AgentID:      event.AgentID,
						SessionID:    event.SessionID,
						IncidentType: "repeated_denied_actions",
						Severity:     "critical",
						Title:        "Repeated denied actions detected",
						Description:  "Agent generated at least five denied actions within a 60-second window.",
						Metadata: map[string]any{
							"event_count_window": len(
								activity.DeniedEvents,
							),
							"action":     event.Action,
							"resource":   event.Resource,
							"risk_score": event.RiskScore,
						},
					},
				)

		if err != nil {
			log.Printf(
				"failed to persist repeated-denial incident: %v",
				err,
			)
		} else {
			d.Metrics.RecordIncident()
		}
	}

	if len(
		activity.HighRisk,
	) >= 3 {
		log.Printf(
			"CRITICAL: repeated high-risk behavior agent=%s count=%d window=60s",
			event.AgentID,
			len(
				activity.HighRisk,
			),
		)

		err :=
			d.IncidentRepository.
				UpsertOpenIncident(
					context.Background(),
					repository.UpsertIncidentParams{
						AgentID:      event.AgentID,
						SessionID:    event.SessionID,
						IncidentType: "repeated_high_risk_behavior",
						Severity:     "critical",
						Title:        "Repeated high-risk agent behavior",
						Description:  "Agent generated at least three high-risk actions within a 60-second window.",
						Metadata: map[string]any{
							"event_count_window": len(
								activity.HighRisk,
							),
							"action":     event.Action,
							"resource":   event.Resource,
							"risk_score": event.RiskScore,
						},
					},
				)

		if err != nil {
			log.Printf(
				"failed to persist high-risk incident: %v",
				err,
			)
		} else {
			d.Metrics.RecordIncident()
		}
	}
}

func pruneOld(
	timestamps []time.Time,
	windowStart time.Time,
) []time.Time {
	result :=
		timestamps[:0]

	for _, timestamp := range timestamps {

		if timestamp.After(
			windowStart,
		) {
			result =
				append(
					result,
					timestamp,
				)
		}
	}

	return result
}

func main() {
	ctx, stop :=
		signal.NotifyContext(
			context.Background(),
			os.Interrupt,
			syscall.SIGTERM,
		)

	defer stop()

	broker :=
		os.Getenv(
			"KAFKA_BROKER",
		)

	if broker == "" {
		broker =
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

	dlqTopic :=
		os.Getenv(
			"KAFKA_DLQ_TOPIC",
		)

	if dlqTopic == "" {
		dlqTopic =
			"agentshield.security.dlq"
	}

	groupID :=
		os.Getenv(
			"KAFKA_CONSUMER_GROUP",
		)

	if groupID == "" {
		groupID =
			"agentshield-detection"
	}

	httpAddress :=
		os.Getenv(
			"DETECTION_HTTP_ADDRESS",
		)

	if httpAddress == "" {
		httpAddress =
			":8083"
	}

	databaseURL :=
		os.Getenv(
			"DATABASE_URL",
		)

	if databaseURL == "" {
		databaseURL =
			"postgres://agentshield:agentshield_dev_password@localhost:5432/agentshield"
	}

	db, err :=
		database.Connect(
			ctx,
			databaseURL,
		)

	if err != nil {
		log.Fatalf(
			"failed to connect to PostgreSQL: %v",
			err,
		)
	}

	defer db.Close()

	log.Println(
		"Connected to AgentShield PostgreSQL",
	)

	metrics :=
		detectionmetrics.New(
			topic,
		)

	incidentRepository :=
		repository.NewIncidentRepository(
			db,
		)

	detector :=
		NewDetector(
			incidentRepository,
			metrics,
		)

	client, err :=
		kgo.NewClient(
			kgo.SeedBrokers(
				broker,
			),

			kgo.ConsumerGroup(
				groupID,
			),

			kgo.ConsumeTopics(
				topic,
			),

			kgo.DisableAutoCommit(),

			kgo.ConsumeResetOffset(
				kgo.NewOffset().
					AtStart(),
			),
		)

	if err != nil {
		log.Fatalf(
			"failed to create Kafka consumer: %v",
			err,
		)
	}

	defer client.Close()

	dlqPublisher :=
		dlq.NewPublisher(
			client,
			dlqTopic,
		)

	mux :=
		http.NewServeMux()

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
					"service": "agentshield-detection",
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
				client.Ping(
					readyCtx,
				); err != nil {

				writeJSON(
					w,
					http.StatusServiceUnavailable,
					map[string]any{
						"service": "agentshield-detection",

						"status": "not_ready",

						"kafka": "disconnected",

						"reason": "Kafka broker unavailable",
					},
				)

				return
			}

			if err :=
				db.Ping(
					readyCtx,
				); err != nil {

				writeJSON(
					w,
					http.StatusServiceUnavailable,
					map[string]any{
						"service": "agentshield-detection",

						"status": "not_ready",

						"database": "disconnected",

						"reason": "PostgreSQL unavailable",
					},
				)

				return
			}

			writeJSON(
				w,
				http.StatusOK,
				map[string]any{
					"service": "agentshield-detection",

					"status": "ready",

					"kafka": "connected",

					"database": "connected",
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
			w.Header().Set(
				"Content-Type",
				"text/plain; version=0.0.4; charset=utf-8",
			)

			if err := metrics.WritePrometheus(w); err != nil {
				log.Printf(
					"failed to write Prometheus metrics: %v",
					err,
				)
			}
		},
	)

	mux.HandleFunc(
		"GET /debug/metrics",
		func(
			w http.ResponseWriter,
			r *http.Request,
		) {
			writeJSON(
				w,
				http.StatusOK,
				metrics.Snapshot(),
			)
		},
	)

	server :=
		&http.Server{
			Addr: httpAddress,

			Handler: mux,

			ReadHeaderTimeout: 5 * time.Second,
		}

	serverErrors :=
		make(
			chan error,
			1,
		)

	go func() {
		log.Printf(
			"AgentShield Detection HTTP server starting on http://localhost%s",
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
		"AgentShield Detection Service started: broker=%s topic=%s group=%s",
		broker,
		topic,
		groupID,
	)

	log.Println(
		"Listening for security events...",
	)

	consumerErrors :=
		make(
			chan error,
			1,
		)

	go func() {
		consumerErrors <- runConsumer(
			ctx,
			client,
			detector,
			metrics,
			dlqPublisher,
		)
	}()

	select {
	case <-ctx.Done():
		log.Println(
			"Detection Service shutdown requested",
		)

	case err := <-consumerErrors:
		if err != nil &&
			!errors.Is(
				err,
				context.Canceled,
			) {
			log.Printf(
				"detection consumer stopped unexpectedly: %v",
				err,
			)
		}

		stop()

	case err := <-serverErrors:
		if err != nil {
			log.Printf(
				"detection HTTP server stopped unexpectedly: %v",
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

	if err :=
		server.Shutdown(
			shutdownCtx,
		); err != nil {

		log.Printf(
			"failed to gracefully stop detection HTTP server: %v",
			err,
		)
	}

	log.Println(
		"AgentShield Detection Service stopped",
	)
}

func runConsumer(
	ctx context.Context,
	client *kgo.Client,
	detector *Detector,
	metrics *detectionmetrics.Metrics,
	dlqPublisher *dlq.Publisher,
) error {
	for {
		if ctx.Err() != nil {
			return ctx.Err()
		}

		fetches :=
			client.PollFetches(
				ctx,
			)

		if errors.Is(
			fetches.Err(),
			context.Canceled,
		) {
			return context.Canceled
		}

		if err :=
			fetches.Err(); err != nil {

			metrics.RecordFetchError()

			log.Printf(
				"Kafka consumer error: %v",
				err,
			)

			continue
		}

		fetches.EachRecord(
			func(
				record *kgo.Record,
			) {
				var event SecurityEvent

				if err :=
					json.Unmarshal(
						record.Value,
						&event,
					); err != nil {

					metrics.RecordRejected()

					log.Printf(
						"invalid security event partition=%d offset=%d error=%v",
						record.Partition,
						record.Offset,
						err,
					)

					if dlqErr :=
						dlqPublisher.Publish(
							ctx,
							record,
							err.Error(),
						); dlqErr != nil {

						metrics.RecordDLQFailure()

						log.Printf(
							"failed to publish malformed event to DLQ offset=%d: %v",
							record.Offset,
							dlqErr,
						)

						return
					}

					metrics.RecordDLQPublished()

					log.Printf(
						"published malformed event to DLQ offset=%d",
						record.Offset,
					)

					if err :=
						client.CommitRecords(
							ctx,
							record,
						); err != nil {

						metrics.RecordCommitFailure()

						log.Printf(
							"failed to commit malformed event offset=%d: %v",
							record.Offset,
							err,
						)
					}

					return
				}

				if event.EventType == "" ||
					event.AgentID == "" ||
					event.Action == "" ||
					event.Decision == "" {

					metrics.RecordRejected()

					reason :=
						"event failed AgentShield security-event validation"

					log.Printf(
						"invalid AgentShield event partition=%d offset=%d",
						record.Partition,
						record.Offset,
					)

					if dlqErr :=
						dlqPublisher.Publish(
							ctx,
							record,
							reason,
						); dlqErr != nil {

						metrics.RecordDLQFailure()

						log.Printf(
							"failed to publish invalid event to DLQ offset=%d: %v",
							record.Offset,
							dlqErr,
						)

						return
					}

					metrics.RecordDLQPublished()

					log.Printf(
						"published invalid event to DLQ offset=%d",
						record.Offset,
					)

					if err :=
						client.CommitRecords(
							ctx,
							record,
						); err != nil {

						metrics.RecordCommitFailure()

						log.Printf(
							"failed to commit invalid event offset=%d: %v",
							record.Offset,
							err,
						)
					}

					return
				}

				log.Printf(
					"security event agent=%s action=%s resource=%s decision=%s risk=%d partition=%d offset=%d",
					event.AgentID,
					event.Action,
					event.Resource,
					event.Decision,
					event.RiskScore,
					record.Partition,
					record.Offset,
				)

				detector.Process(
					event,
				)

				if err :=
					client.CommitRecords(
						ctx,
						record,
					); err != nil {

					metrics.RecordCommitFailure()

					log.Printf(
						"failed to commit detection event offset=%d: %v",
						record.Offset,
						err,
					)

					return
				}

				metrics.RecordProcessed(
					record.Partition,
					record.Offset,
				)

				log.Printf(
					"committed detection event partition=%d offset=%d",
					record.Partition,
					record.Offset,
				)
			},
		)
	}
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

	w.WriteHeader(
		status,
	)

	_ =
		json.NewEncoder(
			w,
		).Encode(
			value,
		)
}
