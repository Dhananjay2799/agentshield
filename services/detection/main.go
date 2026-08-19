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

	"github.com/twmb/franz-go/pkg/kgo"

	"github.com/dhananjay2799/agentshield/services/detection/internal/baseline"
	"github.com/dhananjay2799/agentshield/services/detection/internal/behavior"
	"github.com/dhananjay2799/agentshield/services/detection/internal/database"
	"github.com/dhananjay2799/agentshield/services/detection/internal/dlq"
	"github.com/dhananjay2799/agentshield/services/detection/internal/features"
	detectionmetrics "github.com/dhananjay2799/agentshield/services/detection/internal/metrics"
	"github.com/dhananjay2799/agentshield/services/detection/internal/mlclient"
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

type Detector struct {
	IncidentRepository        *repository.IncidentRepository
	BehaviorProfileRepository *repository.BehaviorProfileRepository
	MLClient                  *mlclient.Client
	Metrics                   *detectionmetrics.Metrics
	BehaviorTracker           *behavior.Tracker
	BehaviorEngine            *behavior.Engine
	FeatureExtractor          *features.Extractor
	BaselineStore             *baseline.Store
	BaselineMinSamples        int64
	MLMinWindowEvents         int
}

func NewDetector(
	incidentRepository *repository.IncidentRepository,
	behaviorProfileRepository *repository.BehaviorProfileRepository,
	mlClient *mlclient.Client,
	metrics *detectionmetrics.Metrics,
) *Detector {
	return &Detector{
		IncidentRepository:        incidentRepository,
		BehaviorProfileRepository: behaviorProfileRepository,
		MLClient:                  mlClient,
		Metrics:                   metrics,

		BehaviorTracker: behavior.NewTracker(
			60 * time.Second,
		),

		BehaviorEngine: behavior.NewEngine(),

		FeatureExtractor: features.NewExtractor(),

		BaselineStore: baseline.NewStore(),

		BaselineMinSamples: 5,

		MLMinWindowEvents: 3,
	}
}

func (d *Detector) Process(
	event SecurityEvent,
) {
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

	// ---------------------------------------------------------
	// BEHAVIORAL DETECTIONS
	// ---------------------------------------------------------

	behaviorEvent := behavior.Event{
		EventType:  event.EventType,
		AgentID:    event.AgentID,
		SessionID:  event.SessionID,
		Action:     event.Action,
		Resource:   event.Resource,
		Decision:   event.Decision,
		RiskScore:  event.RiskScore,
		OccurredAt: event.OccurredAt,
	}

	snapshot := d.BehaviorTracker.Record(behaviorEvent)

	if event.Decision == "DENY" {
		d.Metrics.RecordDenied()

		log.Printf(
			"ALERT: denied action detected agent=%s action=%s resource=%s",
			event.AgentID,
			event.Action,
			event.Resource,
		)
	}

	if event.RiskScore >= 80 {
		d.Metrics.RecordHighRisk()

		log.Printf(
			"ALERT: high-risk agent behavior agent=%s risk=%d",
			event.AgentID,
			event.RiskScore,
		)
	}

	detections := d.BehaviorEngine.Evaluate(
		behaviorEvent,
		snapshot,
	)

	for _, detection := range detections {
		log.Printf(
			"BEHAVIORAL DETECTION: type=%s severity=%s agent=%s action=%s denied_window=%d high_risk_window=%d event_window=%d",
			detection.Type,
			detection.Severity,
			event.AgentID,
			event.Action,
			snapshot.DeniedCount,
			snapshot.HighRiskCount,
			snapshot.EventCount,
		)

		metadata := detection.Metadata
		if metadata == nil {
			metadata = map[string]any{}
		}

		metadata["event_type"] = event.EventType
		metadata["behavioral_detection"] = true
		metadata["denied_count_window"] = snapshot.DeniedCount
		metadata["high_risk_count_window"] = snapshot.HighRiskCount
		metadata["event_count_window"] = snapshot.EventCount

		err := d.IncidentRepository.UpsertOpenIncident(
			context.Background(),
			repository.UpsertIncidentParams{
				AgentID:      event.AgentID,
				SessionID:    event.SessionID,
				IncidentType: detection.Type,
				Severity:     detection.Severity,
				Title:        detection.Title,
				Description:  detection.Description,
				Metadata:     metadata,
			},
		)

		if err != nil {
			log.Printf(
				"failed to persist behavioral incident type=%s: %v",
				detection.Type,
				err,
			)
			continue
		}

		d.Metrics.RecordIncident()

		d.Metrics.RecordBehavioralDetection(
			detection.Type,
		)
	}

	// ---------------------------------------------------------
	// AGENT BASELINE + ANOMALY DETECTION
	// ---------------------------------------------------------

	window :=
		d.BehaviorTracker.Window(
			event.AgentID,
		)

	featureEvents :=
		make(
			[]features.Event,
			0,
			len(window),
		)

	for _, windowEvent := range window {
		featureEvents =
			append(
				featureEvents,
				features.Event{
					AgentID:    windowEvent.AgentID,
					SessionID:  windowEvent.SessionID,
					Action:     windowEvent.Action,
					Resource:   windowEvent.Resource,
					Decision:   windowEvent.Decision,
					RiskScore:  windowEvent.RiskScore,
					OccurredAt: windowEvent.OccurredAt,
				},
			)
	}

	vector :=
		d.FeatureExtractor.Extract(
			event.AgentID,
			featureEvents,
		)

	observation :=
		baseline.Observation{
			EventCount: float64(
				vector.EventCount,
			),

			DenyRatio: vector.DenyRatio,

			HighRiskRatio: vector.HighRiskRatio,

			ActionDiversityRatio: vector.ActionDiversityRatio,

			ResourceDiversityRatio: vector.ResourceDiversityRatio,

			AverageRiskScore: vector.AverageRiskScore,

			ProductionAccessRatio: vector.ProductionAccessRatio,

			SensitiveActionRatio: vector.SensitiveActionRatio,
		}

	isAnomalous := false

	if d.MLClient != nil &&
		vector.EventCount >= d.MLMinWindowEvents {

		d.Metrics.RecordMLPrediction()

		mlPrediction, mlErr :=
			d.MLClient.Predict(
				context.Background(),
				mlclient.PredictionRequest{
					EventCount:             float64(vector.EventCount),
					DenyRatio:              vector.DenyRatio,
					HighRiskRatio:          vector.HighRiskRatio,
					ActionDiversityRatio:   vector.ActionDiversityRatio,
					ResourceDiversityRatio: vector.ResourceDiversityRatio,
					AverageRiskScore:       vector.AverageRiskScore,
					ProductionAccessRatio:  vector.ProductionAccessRatio,
					SensitiveActionRatio:   vector.SensitiveActionRatio,
				},
			)

		if mlErr != nil {
			d.Metrics.RecordMLFailure()

			log.Printf(
				"ML ANOMALY INFERENCE FAILED: agent=%s error=%v",
				event.AgentID,
				mlErr,
			)
		} else {
			log.Printf(
				"ML ANOMALY SCORE: agent=%s error=%.4f threshold=%.4f score_ratio=%.4f anomaly=%t model=%s",
				event.AgentID,
				mlPrediction.ReconstructionError,
				mlPrediction.Threshold,
				mlPrediction.ScoreRatio,
				mlPrediction.IsAnomaly,
				mlPrediction.Model,
			)

			if mlPrediction.IsAnomaly {
				d.Metrics.RecordMLAnomaly()

				isAnomalous = true

				metadata := map[string]any{
					"ml_anomaly_detection":     true,
					"ml_model":                 mlPrediction.Model,
					"ml_reconstruction_error":  mlPrediction.ReconstructionError,
					"ml_threshold":             mlPrediction.Threshold,
					"ml_score_ratio":           mlPrediction.ScoreRatio,
					"event_count":              vector.EventCount,
					"deny_ratio":               vector.DenyRatio,
					"high_risk_ratio":          vector.HighRiskRatio,
					"action_diversity_ratio":   vector.ActionDiversityRatio,
					"resource_diversity_ratio": vector.ResourceDiversityRatio,
					"average_risk_score":       vector.AverageRiskScore,
					"production_access_ratio":  vector.ProductionAccessRatio,
					"sensitive_action_ratio":   vector.SensitiveActionRatio,
				}

				err := d.IncidentRepository.UpsertOpenIncident(
					context.Background(),
					repository.UpsertIncidentParams{
						AgentID:      event.AgentID,
						SessionID:    event.SessionID,
						IncidentType: "ml_behavior_anomaly",
						Severity:     "high",
						Title:        "Neural behavior anomaly detected",
						Description:  "AgentShield's PyTorch behavioral autoencoder detected activity outside the learned normal behavior distribution.",
						Metadata:     metadata,
					},
				)

				if err != nil {
					log.Printf(
						"failed to persist ML behavior anomaly incident: %v",
						err,
					)
				} else {
					log.Printf(
						"ML ANOMALY DETECTED: agent=%s model=%s score_ratio=%.4f",
						event.AgentID,
						mlPrediction.Model,
						mlPrediction.ScoreRatio,
					)

					d.Metrics.RecordIncident()
				}
			}
		}
	} else if d.MLClient != nil {
		d.Metrics.RecordMLSkippedInsufficientWindow()

		log.Printf(
			"ML ANOMALY SKIPPED: agent=%s events=%d minimum_events=%d reason=insufficient_window",
			event.AgentID,
			vector.EventCount,
			d.MLMinWindowEvents,
		)
	}

	profile, exists :=
		d.BaselineStore.Get(
			event.AgentID,
		)

	profileLoadFailed := false

	if !exists {
		persistedProfile, err :=
			d.BehaviorProfileRepository.Get(
				context.Background(),
				event.AgentID,
			)

		switch {
		case err == nil:
			d.BaselineStore.Restore(
				*persistedProfile,
			)

			profile =
				*persistedProfile

			exists = true

			log.Printf(
				"AGENT BASELINE RESTORED: agent=%s samples=%d mean_events=%.2f mean_risk=%.2f",
				event.AgentID,
				profile.SampleCount,
				profile.Mean.EventCount,
				profile.Mean.AverageRiskScore,
			)

		case errors.Is(
			err,
			repository.ErrBehaviorProfileNotFound,
		):
			// No persisted profile exists yet.
			// The first normal observation will create it.

		default:
			profileLoadFailed = true

			log.Printf(
				"failed to load persisted behavior profile agent=%s: %v",
				event.AgentID,
				err,
			)
		}
	}

	if exists {
		d.Metrics.RecordAnomalyEvaluation()

		score :=
			baseline.ScoreObservation(
				profile,
				observation,
				d.BaselineMinSamples,
			)

		log.Printf(
			"ANOMALY SCORE: agent=%s score=%.4f warmed_up=%t samples=%d events=%d",
			event.AgentID,
			score.Value,
			score.WarmedUp,
			score.SampleCount,
			vector.EventCount,
		)

		if score.IsAnomalous {
			isAnomalous = true
			d.Metrics.RecordAnomalyDetection()

			metadata :=
				map[string]any{
					"anomaly_detection": true,
					"anomaly_score":     score.Value,
					"baseline_samples":  score.SampleCount,

					"event_count": vector.EventCount,

					"deny_ratio": vector.DenyRatio,

					"high_risk_ratio": vector.HighRiskRatio,

					"action_diversity_ratio": vector.ActionDiversityRatio,

					"resource_diversity_ratio": vector.ResourceDiversityRatio,

					"average_risk_score": vector.AverageRiskScore,

					"production_access_ratio": vector.ProductionAccessRatio,

					"sensitive_action_ratio": vector.SensitiveActionRatio,

					"explanation": score.Explanation,
				}

			err :=
				d.IncidentRepository.UpsertOpenIncident(
					context.Background(),
					repository.UpsertIncidentParams{
						AgentID: event.AgentID,

						SessionID: event.SessionID,

						IncidentType: "agent_behavior_anomaly",

						Severity: "critical",

						Title: "Autonomous agent behavior anomaly detected",

						Description: "AgentShield detected behavior that significantly deviates from the agent's learned baseline.",

						Metadata: metadata,
					},
				)

			if err != nil {
				log.Printf(
					"failed to persist anomaly incident agent=%s score=%.4f: %v",
					event.AgentID,
					score.Value,
					err,
				)
			} else {
				d.Metrics.RecordIncident()

				log.Printf(
					"AGENT ANOMALY DETECTED: agent=%s score=%.4f samples=%d",
					event.AgentID,
					score.Value,
					score.SampleCount,
				)
			}
		}
	}

	if isAnomalous {
		d.Metrics.RecordBaselineUpdateSkipped()

		log.Printf(
			"BASELINE UPDATE SKIPPED: agent=%s reason=anomalous_observation",
			event.AgentID,
		)
	} else if profileLoadFailed {
		d.Metrics.RecordBaselineUpdateSkipped()

		log.Printf(
			"BASELINE UPDATE SKIPPED: agent=%s reason=persisted_profile_load_failed",
			event.AgentID,
		)
	} else {
		updatedProfile :=
			d.BaselineStore.Observe(
				event.AgentID,
				observation,
			)

		if updatedProfile.SampleCount ==
			d.BaselineMinSamples {

			d.Metrics.RecordBaselineWarmup()
		}

		err :=
			d.BehaviorProfileRepository.Upsert(
				context.Background(),
				updatedProfile,
			)

		if err != nil {
			log.Printf(
				"failed to persist behavior profile agent=%s samples=%d: %v",
				event.AgentID,
				updatedProfile.SampleCount,
				err,
			)
		} else {
			log.Printf(
				"AGENT BASELINE PERSISTED: agent=%s samples=%d",
				event.AgentID,
				updatedProfile.SampleCount,
			)
		}

		log.Printf(
			"AGENT BASELINE: agent=%s samples=%d mean_events=%.2f mean_risk=%.2f",
			event.AgentID,
			updatedProfile.SampleCount,
			updatedProfile.Mean.EventCount,
			updatedProfile.Mean.AverageRiskScore,
		)
	}
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

	mlAnomalyURL :=
		os.Getenv(
			"ML_ANOMALY_URL",
		)

	if mlAnomalyURL == "" {
		mlAnomalyURL =
			"http://localhost:18085"
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

	behaviorProfileRepository :=
		repository.NewBehaviorProfileRepository(
			db,
		)

	mlClient :=
		mlclient.New(
			mlAnomalyURL,
			2*time.Second,
		)

	detector :=
		NewDetector(
			incidentRepository,
			behaviorProfileRepository,
			mlClient,
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
