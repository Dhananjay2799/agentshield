package main

import (
	"bufio"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"
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

func main() {
	var (
		broker = flag.String(
			"broker",
			"localhost:19092",
			"Kafka broker address",
		)

		dlqTopic = flag.String(
			"dlq-topic",
			"agentshield.security.dlq",
			"DLQ topic",
		)

		targetTopic = flag.String(
			"target-topic",
			"agentshield.security.events",
			"target replay topic",
		)

		offset = flag.Int64(
			"offset",
			-1,
			"DLQ offset to replay",
		)

		payloadFile = flag.String(
			"payload-file",
			"",
			"optional corrected JSON payload file",
		)

		force = flag.Bool(
			"force",
			false,
			"skip interactive confirmation",
		)
	)

	flag.Parse()

	if *offset < 0 {
		log.Fatal(
			"-offset is required and must be >= 0",
		)
	}

	client, err := kgo.NewClient(
		kgo.SeedBrokers(
			*broker,
		),
	)
	if err != nil {
		log.Fatalf(
			"create Kafka client: %v",
			err,
		)
	}

	defer client.Close()

	ctx, cancel :=
		context.WithTimeout(
			context.Background(),
			15*time.Second,
		)

	defer cancel()

	failedEvent, err :=
		loadDLQEvent(
			ctx,
			*broker,
			*dlqTopic,
			*offset,
		)

	if err != nil {
		log.Fatalf(
			"load DLQ event: %v",
			err,
		)
	}

	replayPayload :=
		[]byte(
			strings.TrimSpace(
				failedEvent.OriginalValue,
			),
		)

	if *payloadFile != "" {
		replayPayload, err =
			os.ReadFile(
				*payloadFile,
			)

		if err != nil {
			log.Fatalf(
				"read corrected payload: %v",
				err,
			)
		}
	}

	if !json.Valid(replayPayload) {
		log.Fatal(
			"replay payload is not valid JSON",
		)
	}

	fmt.Println(
		"AgentShield DLQ Replay",
	)

	fmt.Printf(
		"DLQ offset:      %d\n",
		*offset,
	)

	fmt.Printf(
		"Original source: %s partition=%d offset=%d\n",
		failedEvent.SourceTopic,
		failedEvent.SourcePartition,
		failedEvent.SourceOffset,
	)

	fmt.Printf(
		"Failure reason:  %s\n",
		failedEvent.Error,
	)

	fmt.Printf(
		"Target topic:    %s\n",
		*targetTopic,
	)

	fmt.Printf(
		"Replay payload:  %s\n",
		string(replayPayload),
	)

	if !*force {
		fmt.Print(
			"\nReplay this event? Type YES to continue: ",
		)

		reader :=
			bufio.NewReader(
				os.Stdin,
			)

		answer, _ :=
			reader.ReadString('\n')

		if strings.TrimSpace(answer) != "YES" {
			fmt.Println(
				"Replay cancelled.",
			)
			return
		}
	}

	record :=
		&kgo.Record{
			Topic: *targetTopic,

			Key: []byte(
				failedEvent.Key,
			),

			Value: replayPayload,

			Headers: []kgo.RecordHeader{
				{
					Key: "agentshield-replayed",

					Value: []byte("true"),
				},
				{
					Key: "agentshield-dlq-offset",

					Value: []byte(
						fmt.Sprintf(
							"%d",
							*offset,
						),
					),
				},
			},
		}

	result :=
		client.ProduceSync(
			ctx,
			record,
		)

	if err :=
		result.FirstErr(); err != nil {
		log.Fatalf(
			"replay event: %v",
			err,
		)
	}

	fmt.Println(
		"Replay published successfully.",
	)
}

func loadDLQEvent(
	ctx context.Context,
	broker string,
	topic string,
	offset int64,
) (*FailedEvent, error) {
	replayConsumer, err :=
		kgo.NewClient(
			kgo.SeedBrokers(
				broker,
			),

			kgo.ConsumePartitions(
				map[string]map[int32]kgo.Offset{
					topic: {
						0: kgo.NewOffset().
							At(offset),
					},
				},
			),
		)

	if err != nil {
		return nil, err
	}

	defer replayConsumer.Close()

	for {
		fetches :=
			replayConsumer.PollFetches(
				ctx,
			)

		if err :=
			fetches.Err(); err != nil {
			return nil, err
		}

		var found *FailedEvent

		fetches.EachRecord(
			func(record *kgo.Record) {
				if found != nil ||
					record.Offset != offset {
					return
				}

				var event FailedEvent

				if err :=
					json.Unmarshal(
						record.Value,
						&event,
					); err != nil {
					return
				}

				found =
					&event
			},
		)

		if found != nil {
			return found, nil
		}
	}
}
