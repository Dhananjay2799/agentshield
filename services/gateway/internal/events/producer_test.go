package events

import (
	"context"
	"strings"
	"testing"
	"time"
)

func TestPublishHonorsContextDeadline(
	t *testing.T,
) {
	producer, err := NewProducer(
		[]string{
			"127.0.0.1:1",
		},
		"agentshield.test.events",
	)
	if err != nil {
		t.Fatalf(
			"create producer: %v",
			err,
		)
	}
	defer producer.Close()

	ctx, cancel := context.WithTimeout(
		context.Background(),
		250*time.Millisecond,
	)
	defer cancel()

	startedAt := time.Now()

	err = producer.Publish(
		ctx,
		SecurityEvent{
			EventType:  "test.kafka.timeout",
			AgentID:    "phase38-timeout-test",
			Action:     "test.publish",
			Resource:   "test/resource",
			Decision:   "TEST",
			OccurredAt: time.Now().UTC(),
		},
	)

	elapsed := time.Since(startedAt)

	if err == nil {
		t.Fatal(
			"expected publish to fail when broker is unavailable",
		)
	}

	if elapsed > 1500*time.Millisecond {
		t.Fatalf(
			"publish ignored context deadline: elapsed=%s",
			elapsed,
		)
	}

	if !strings.Contains(
		err.Error(),
		"context deadline exceeded",
	) &&
		!strings.Contains(
			err.Error(),
			"context canceled",
		) {
		t.Fatalf(
			"unexpected publish error: %v",
			err,
		)
	}

	t.Logf(
		"unavailable broker publish terminated in %s",
		elapsed,
	)
}
