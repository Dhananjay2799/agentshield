package dlq

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/twmb/franz-go/pkg/kgo"
)

type FailedEvent struct {
	SourceTopic     string    `json:"source_topic"`
	SourcePartition int32     `json:"source_partition"`
	SourceOffset    int64     `json:"source_offset"`
	Key             string    `json:"key,omitempty"`
	OriginalValue   string    `json:"original_value"`
	Error           string    `json:"error"`
	FailedAt        time.Time `json:"failed_at"`
}

type Publisher struct {
	Client *kgo.Client
	Topic  string
}

func NewPublisher(
	client *kgo.Client,
	topic string,
) *Publisher {
	return &Publisher{
		Client: client,
		Topic:  topic,
	}
}

func (p *Publisher) Publish(
	ctx context.Context,
	record *kgo.Record,
	reason string,
) error {
	event := FailedEvent{
		SourceTopic:     record.Topic,
		SourcePartition: record.Partition,
		SourceOffset:    record.Offset,
		Key:             string(record.Key),
		OriginalValue:   string(record.Value),
		Error:           reason,
		FailedAt:        time.Now().UTC(),
	}

	payload, err :=
		json.Marshal(event)

	if err != nil {
		return fmt.Errorf(
			"marshal DLQ event: %w",
			err,
		)
	}

	dlqRecord :=
		&kgo.Record{
			Topic: p.Topic,
			Key:   record.Key,
			Value: payload,
		}

	result :=
		p.Client.ProduceSync(
			ctx,
			dlqRecord,
		)

	if err :=
		result.FirstErr(); err != nil {
		return fmt.Errorf(
			"publish DLQ event: %w",
			err,
		)
	}

	return nil
}
