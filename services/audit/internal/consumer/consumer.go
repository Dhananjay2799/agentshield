package consumer

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"sync/atomic"
	"time"

	"github.com/twmb/franz-go/pkg/kgo"

	"github.com/dhananjay2799/agentshield/services/audit/internal/models"
)

type Consumer struct {
	Client *kgo.Client
	Topic  string

	processedEvents atomic.Uint64
	rejectedEvents  atomic.Uint64
	fetchErrors     atomic.Uint64
	commitFailures  atomic.Uint64

	lastOffset      atomic.Int64
	lastPartition   atomic.Int64
	lastProcessedAt atomic.Int64
}

type Metrics struct {
	Topic           string `json:"topic"`
	ProcessedEvents uint64 `json:"processed_events"`
	RejectedEvents  uint64 `json:"rejected_events"`
	FetchErrors     uint64 `json:"fetch_errors"`
	CommitFailures  uint64 `json:"commit_failures"`
	LastOffset      int64  `json:"last_offset"`
	LastPartition   int64  `json:"last_partition"`
	LastProcessedAt string `json:"last_processed_at,omitempty"`
}

func New(
	brokers []string,
	topic string,
	groupID string,
) (*Consumer, error) {
	client, err := kgo.NewClient(
		kgo.SeedBrokers(brokers...),
		kgo.ConsumerGroup(groupID),
		kgo.ConsumeTopics(topic),

		// Brand-new consumer groups replay historical
		// events from the beginning. Existing groups
		// resume from their committed offsets.
		kgo.ConsumeResetOffset(
			kgo.NewOffset().AtStart(),
		),
	)

	if err != nil {
		return nil, err
	}

	consumer := &Consumer{
		Client: client,
		Topic:  topic,
	}

	consumer.lastOffset.Store(-1)
	consumer.lastPartition.Store(-1)

	return consumer, nil
}

func (c *Consumer) Close() {
	if c.Client != nil {
		c.Client.Close()
	}
}

func (c *Consumer) Ready(
	ctx context.Context,
) error {
	return c.Client.Ping(ctx)
}

func (c *Consumer) Metrics() Metrics {
	lastProcessedAt := ""

	timestamp :=
		c.lastProcessedAt.Load()

	if timestamp > 0 {
		lastProcessedAt =
			time.Unix(
				0,
				timestamp,
			).UTC().Format(
				time.RFC3339Nano,
			)
	}

	return Metrics{
		Topic:           c.Topic,
		ProcessedEvents: c.processedEvents.Load(),
		RejectedEvents:  c.rejectedEvents.Load(),
		FetchErrors:     c.fetchErrors.Load(),
		CommitFailures:  c.commitFailures.Load(),
		LastOffset:      c.lastOffset.Load(),
		LastPartition:   c.lastPartition.Load(),
		LastProcessedAt: lastProcessedAt,
	}
}

func (c *Consumer) Run(
	ctx context.Context,
) error {
	log.Printf(
		"Audit consumer listening on topic %s",
		c.Topic,
	)

	for {
		if ctx.Err() != nil {
			return ctx.Err()
		}

		fetches :=
			c.Client.PollFetches(ctx)

		if errors.Is(
			fetches.Err(),
			context.Canceled,
		) {
			return context.Canceled
		}

		if err := fetches.Err(); err != nil {
			c.fetchErrors.Add(1)

			log.Printf(
				"Kafka fetch error: %v",
				err,
			)

			continue
		}

		fetches.EachRecord(
			func(record *kgo.Record) {
				c.processRecord(
					ctx,
					record,
				)
			},
		)
	}
}

func (c *Consumer) processRecord(
	ctx context.Context,
	record *kgo.Record,
) {
	var event models.SecurityEvent

	if err := json.Unmarshal(
		record.Value,
		&event,
	); err != nil {
		log.Printf(
			"Rejected malformed Kafka event: topic=%s partition=%d offset=%d error=%v",
			record.Topic,
			record.Partition,
			record.Offset,
			err,
		)

		if c.commitRecord(
			ctx,
			record,
		) {
			c.rejectedEvents.Add(1)
			c.recordProgress(record)
		}

		return
	}

	if !event.Valid() {
		log.Printf(
			"Rejected invalid security event: topic=%s partition=%d offset=%d",
			record.Topic,
			record.Partition,
			record.Offset,
		)

		if c.commitRecord(
			ctx,
			record,
		) {
			c.rejectedEvents.Add(1)
			c.recordProgress(record)
		}

		return
	}

	log.Printf(
		"AUDIT EVENT event_type=%s decision=%s risk=%d agent=%s session=%s action=%s resource=%s partition=%d offset=%d",
		event.EventType,
		event.Decision,
		event.RiskScore,
		event.AgentID,
		event.SessionID,
		event.Action,
		event.Resource,
		record.Partition,
		record.Offset,
	)

	if c.commitRecord(
		ctx,
		record,
	) {
		c.processedEvents.Add(1)
		c.recordProgress(record)
	}
}

func (c *Consumer) commitRecord(
	ctx context.Context,
	record *kgo.Record,
) bool {
	if err := c.Client.CommitRecords(
		ctx,
		record,
	); err != nil {
		c.commitFailures.Add(1)

		log.Printf(
			"Failed to commit Kafka record offset=%d: %v",
			record.Offset,
			err,
		)

		return false
	}

	return true
}

func (c *Consumer) recordProgress(
	record *kgo.Record,
) {
	c.lastOffset.Store(
		record.Offset,
	)

	c.lastPartition.Store(
		int64(record.Partition),
	)

	c.lastProcessedAt.Store(
		time.Now().UTC().UnixNano(),
	)
}
