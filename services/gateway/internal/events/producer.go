package events

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/twmb/franz-go/pkg/kgo"
)

type SecurityEvent struct {
	EventType  string         `json:"event_type"`
	AgentID    string         `json:"agent_id"`
	SessionID  string         `json:"session_id"`
	Action     string         `json:"action"`
	Resource   string         `json:"resource"`
	Decision   string         `json:"decision"`
	RiskScore  int            `json:"risk_score"`
	Metadata   map[string]any `json:"metadata,omitempty"`
	OccurredAt time.Time      `json:"occurred_at"`
}

type Producer struct {
	Client *kgo.Client
	Topic  string
	Hub    *Hub
}

func NewProducer(brokers []string, topic string) (*Producer, error) {
	client, err := kgo.NewClient(
		kgo.SeedBrokers(brokers...),
		kgo.ClientID("agentshield-gateway"),
	)

	if err != nil {
		return nil, fmt.Errorf("create kafka producer: %w", err)
	}

	return &Producer{
		Client: client,
		Topic:  topic,
		Hub:    NewHub(),
	}, nil
}

func (p *Producer) Close() {
	if p.Client != nil {
		p.Client.Close()
	}
}

func (p *Producer) Publish(
	ctx context.Context,
	event SecurityEvent,
) error {

	payload, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("marshal security event: %w", err)
	}

	record := &kgo.Record{
		Topic: p.Topic,
		Key:   []byte(event.AgentID),
		Value: payload,
	}

	result := p.Client.ProduceSync(ctx, record)

	if err := result.FirstErr(); err != nil {
		return fmt.Errorf("publish security event: %w", err)
	}

	if p.Hub != nil {
		p.Hub.Broadcast(event)
	}

	return nil
}
