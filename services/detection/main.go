package main

import (
	"context"
	"encoding/json"
	"log"
	"os"
	"sync"
	"time"

	"github.com/twmb/franz-go/pkg/kgo"

	"github.com/dhananjay2799/agentshield/services/detection/internal/database"
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
}

func NewDetector(
	incidentRepository *repository.IncidentRepository,
) *Detector {
	return &Detector{
		activity:           make(map[string]*AgentActivity),
		IncidentRepository: incidentRepository,
	}
}

func (d *Detector) Process(event SecurityEvent) {
	d.mu.Lock()
	defer d.mu.Unlock()

	now := time.Now().UTC()
	windowStart := now.Add(-60 * time.Second)

	activity, exists := d.activity[event.AgentID]
	if !exists {
		activity = &AgentActivity{}
		d.activity[event.AgentID] = activity
	}

	activity.DeniedEvents = pruneOld(
		activity.DeniedEvents,
		windowStart,
	)

	activity.HighRisk = pruneOld(
		activity.HighRisk,
		windowStart,
	)

	if event.Decision == "DENY" {
		activity.DeniedEvents = append(
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
		activity.HighRisk = append(
			activity.HighRisk,
			now,
		)

		log.Printf(
			"ALERT: high-risk agent behavior agent=%s risk=%d",
			event.AgentID,
			event.RiskScore,
		)
	}

	if len(activity.DeniedEvents) >= 5 {
		log.Printf(
			"CRITICAL: repeated denied actions agent=%s count=%d window=60s possible compromise or privilege escalation",
			event.AgentID,
			len(activity.DeniedEvents),
		)

		err := d.IncidentRepository.UpsertOpenIncident(
			context.Background(),
			repository.UpsertIncidentParams{
				AgentID:      event.AgentID,
				SessionID:    event.SessionID,
				IncidentType: "repeated_denied_actions",
				Severity:     "critical",
				Title:        "Repeated denied actions detected",
				Description:  "Agent generated at least five denied actions within a 60-second window.",
				Metadata: map[string]any{
					"event_count_window": len(activity.DeniedEvents),
					"action":             event.Action,
					"resource":           event.Resource,
					"risk_score":         event.RiskScore,
				},
			},
		)

		if err != nil {
			log.Printf(
				"failed to persist repeated-denial incident: %v",
				err,
			)
		}
	}

	if len(activity.HighRisk) >= 3 {
		log.Printf(
			"CRITICAL: repeated high-risk behavior agent=%s count=%d window=60s",
			event.AgentID,
			len(activity.HighRisk),
		)

		err := d.IncidentRepository.UpsertOpenIncident(
			context.Background(),
			repository.UpsertIncidentParams{
				AgentID:      event.AgentID,
				SessionID:    event.SessionID,
				IncidentType: "repeated_high_risk_behavior",
				Severity:     "critical",
				Title:        "Repeated high-risk agent behavior",
				Description:  "Agent generated at least three high-risk actions within a 60-second window.",
				Metadata: map[string]any{
					"event_count_window": len(activity.HighRisk),
					"action":             event.Action,
					"resource":           event.Resource,
					"risk_score":         event.RiskScore,
				},
			},
		)

		if err != nil {
			log.Printf(
				"failed to persist high-risk incident: %v",
				err,
			)
		}
	}
}

func pruneOld(
	timestamps []time.Time,
	windowStart time.Time,
) []time.Time {

	result := timestamps[:0]

	for _, timestamp := range timestamps {
		if timestamp.After(windowStart) {
			result = append(result, timestamp)
		}
	}

	return result
}

func main() {
	ctx := context.Background()

	broker := os.Getenv("KAFKA_BROKER")
	if broker == "" {
		broker = "localhost:19092"
	}

	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		databaseURL = "postgres://agentshield:agentshield_dev_password@localhost:5432/agentshield"
	}

	db, err := database.Connect(ctx, databaseURL)
	if err != nil {
		log.Fatalf(
			"failed to connect to PostgreSQL: %v",
			err,
		)
	}
	defer db.Close()

	log.Println("Connected to AgentShield PostgreSQL")

	incidentRepository := repository.NewIncidentRepository(db)
	detector := NewDetector(incidentRepository)

	client, err := kgo.NewClient(
		kgo.SeedBrokers(broker),
		kgo.ConsumerGroup("agentshield-detection"),
		kgo.ConsumeTopics("agentshield.security.events"),
	)

	if err != nil {
		log.Fatalf(
			"failed to create Kafka consumer: %v",
			err,
		)
	}

	defer client.Close()

	log.Println("AgentShield Detection Service started")
	log.Println("Listening for security events...")

	for {
		fetches := client.PollFetches(ctx)

		if errs := fetches.Errors(); len(errs) > 0 {
			for _, fetchErr := range errs {
				log.Printf(
					"Kafka consumer error: %v",
					fetchErr,
				)
			}
			continue
		}

		fetches.EachRecord(func(record *kgo.Record) {
			var event SecurityEvent

			if err := json.Unmarshal(record.Value, &event); err != nil {
				log.Printf(
					"invalid security event: %v",
					err,
				)
				return
			}

			if event.EventType == "" ||
				event.AgentID == "" ||
				event.Action == "" ||
				event.Decision == "" {

				log.Printf(
					"ignoring malformed or non-AgentShield event at partition=%d offset=%d",
					record.Partition,
					record.Offset,
				)

				return
			}

			log.Printf(
				"security event agent=%s action=%s resource=%s decision=%s risk=%d",
				event.AgentID,
				event.Action,
				event.Resource,
				event.Decision,
				event.RiskScore,
			)

			detector.Process(event)
		})
	}
}
