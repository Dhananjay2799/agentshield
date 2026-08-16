package metrics

import (
	"sync"
	"time"
)

type Snapshot struct {
	Topic              string     `json:"topic"`
	ProcessedEvents    int64      `json:"processed_events"`
	DeniedEvents       int64      `json:"denied_events"`
	HighRiskEvents     int64      `json:"high_risk_events"`
	IncidentsTriggered int64      `json:"incidents_triggered"`
	RejectedEvents     int64      `json:"rejected_events"`
	FetchErrors        int64      `json:"fetch_errors"`
	CommitFailures     int64      `json:"commit_failures"`
	LastOffset         int64      `json:"last_offset"`
	LastPartition      int32      `json:"last_partition"`
	DLQPublished       int64      `json:"dlq_published"`
	DLQFailures        int64      `json:"dlq_failures"`
	LastProcessedAt    *time.Time `json:"last_processed_at,omitempty"`
}

type Metrics struct {
	mu sync.RWMutex

	topic              string
	processedEvents    int64
	deniedEvents       int64
	highRiskEvents     int64
	incidentsTriggered int64
	rejectedEvents     int64
	fetchErrors        int64
	commitFailures     int64
	lastOffset         int64
	lastPartition      int32
	dlqPublished       int64
	dlqFailures        int64
	lastProcessedAt    *time.Time
}

func New(topic string) *Metrics {
	return &Metrics{
		topic:         topic,
		lastOffset:    -1,
		lastPartition: -1,
	}
}

func (m *Metrics) RecordDLQPublished() {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.dlqPublished++
}

func (m *Metrics) RecordDLQFailure() {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.dlqFailures++
}

func (m *Metrics) RecordProcessed(
	partition int32,
	offset int64,
) {
	m.mu.Lock()
	defer m.mu.Unlock()

	now := time.Now().UTC()

	m.processedEvents++
	m.lastPartition = partition
	m.lastOffset = offset
	m.lastProcessedAt = &now
}

func (m *Metrics) RecordDenied() {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.deniedEvents++
}

func (m *Metrics) RecordHighRisk() {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.highRiskEvents++
}

func (m *Metrics) RecordIncident() {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.incidentsTriggered++
}

func (m *Metrics) RecordRejected() {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.rejectedEvents++
}

func (m *Metrics) RecordFetchError() {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.fetchErrors++
}

func (m *Metrics) RecordCommitFailure() {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.commitFailures++
}

func (m *Metrics) Snapshot() Snapshot {
	m.mu.RLock()
	defer m.mu.RUnlock()

	return Snapshot{
		Topic:              m.topic,
		ProcessedEvents:    m.processedEvents,
		DeniedEvents:       m.deniedEvents,
		HighRiskEvents:     m.highRiskEvents,
		IncidentsTriggered: m.incidentsTriggered,
		RejectedEvents:     m.rejectedEvents,
		FetchErrors:        m.fetchErrors,
		CommitFailures:     m.commitFailures,
		DLQPublished:       m.dlqPublished,
		DLQFailures:        m.dlqFailures,
		LastOffset:         m.lastOffset,
		LastPartition:      m.lastPartition,
		LastProcessedAt:    m.lastProcessedAt,
	}
}
