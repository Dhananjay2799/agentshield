package metrics

import (
	"fmt"
	"io"
	"sync"
	"time"
)

type Snapshot struct {
	Topic                       string     `json:"topic"`
	ProcessedEvents             int64      `json:"processed_events"`
	DeniedEvents                int64      `json:"denied_events"`
	HighRiskEvents              int64      `json:"high_risk_events"`
	IncidentsTriggered          int64      `json:"incidents_triggered"`
	BehavioralDetections        int64      `json:"behavioral_detections"`
	RepeatedDeniedDetections    int64      `json:"repeated_denied_detections"`
	RepeatedHighRiskDetections  int64      `json:"repeated_high_risk_detections"`
	ActionBurstDetections       int64      `json:"action_burst_detections"`
	ActionDiversityDetections   int64      `json:"action_diversity_detections"`
	ResourceDiversityDetections int64      `json:"resource_diversity_detections"`
	RejectedEvents              int64      `json:"rejected_events"`
	FetchErrors                 int64      `json:"fetch_errors"`
	CommitFailures              int64      `json:"commit_failures"`
	LastOffset                  int64      `json:"last_offset"`
	LastPartition               int32      `json:"last_partition"`
	DLQPublished                int64      `json:"dlq_published"`
	DLQFailures                 int64      `json:"dlq_failures"`
	AnomalyEvaluations          int64      `json:"anomaly_evaluations"`
	AnomalyDetections           int64      `json:"anomaly_detections"`
	BaselineWarmups             int64      `json:"baseline_warmups"`
	BaselineUpdatesSkipped      int64      `json:"baseline_updates_skipped"`
	LastProcessedAt             *time.Time `json:"last_processed_at,omitempty"`
}

type Metrics struct {
	mu sync.RWMutex

	topic                       string
	processedEvents             int64
	deniedEvents                int64
	highRiskEvents              int64
	incidentsTriggered          int64
	behavioralDetections        int64
	repeatedDeniedDetections    int64
	repeatedHighRiskDetections  int64
	actionBurstDetections       int64
	actionDiversityDetections   int64
	resourceDiversityDetections int64
	rejectedEvents              int64
	fetchErrors                 int64
	commitFailures              int64
	lastOffset                  int64
	lastPartition               int32
	dlqPublished                int64
	dlqFailures                 int64
	anomalyEvaluations          int64
	anomalyDetections           int64
	baselineWarmups             int64
	baselineUpdatesSkipped      int64
	lastProcessedAt             *time.Time
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

func (m *Metrics) RecordBehavioralDetection(
	detectionType string,
) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.behavioralDetections++

	switch detectionType {
	case "repeated_denied_actions":
		m.repeatedDeniedDetections++

	case "repeated_high_risk_behavior":
		m.repeatedHighRiskDetections++

	case "agent_action_burst":
		m.actionBurstDetections++

	case "high_action_diversity":
		m.actionDiversityDetections++

	case "high_resource_diversity":
		m.resourceDiversityDetections++
	}
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

func (m *Metrics) RecordAnomalyEvaluation() {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.anomalyEvaluations++
}

func (m *Metrics) RecordAnomalyDetection() {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.anomalyDetections++
}

func (m *Metrics) RecordBaselineWarmup() {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.baselineWarmups++
}

func (m *Metrics) RecordBaselineUpdateSkipped() {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.baselineUpdatesSkipped++
}

func (m *Metrics) Snapshot() Snapshot {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var lastProcessedAt *time.Time
	if m.lastProcessedAt != nil {
		value := *m.lastProcessedAt
		lastProcessedAt = &value
	}

	return Snapshot{
		Topic:                       m.topic,
		ProcessedEvents:             m.processedEvents,
		DeniedEvents:                m.deniedEvents,
		HighRiskEvents:              m.highRiskEvents,
		IncidentsTriggered:          m.incidentsTriggered,
		BehavioralDetections:        m.behavioralDetections,
		RepeatedDeniedDetections:    m.repeatedDeniedDetections,
		RepeatedHighRiskDetections:  m.repeatedHighRiskDetections,
		ActionBurstDetections:       m.actionBurstDetections,
		ActionDiversityDetections:   m.actionDiversityDetections,
		ResourceDiversityDetections: m.resourceDiversityDetections,
		RejectedEvents:              m.rejectedEvents,
		FetchErrors:                 m.fetchErrors,
		CommitFailures:              m.commitFailures,
		DLQPublished:                m.dlqPublished,
		DLQFailures:                 m.dlqFailures,
		LastOffset:                  m.lastOffset,
		LastPartition:               m.lastPartition,
		AnomalyEvaluations:          m.anomalyEvaluations,
		AnomalyDetections:           m.anomalyDetections,
		BaselineWarmups:             m.baselineWarmups,
		BaselineUpdatesSkipped:      m.baselineUpdatesSkipped,
		LastProcessedAt:             lastProcessedAt,
	}
}

func (m *Metrics) WritePrometheus(w io.Writer) error {
	snapshot := m.Snapshot()

	metrics := []struct {
		help  string
		name  string
		mtype string
		value any
	}{
		{
			"Total number of valid security events processed by the detection service.",
			"agentshield_detection_processed_events_total",
			"counter",
			snapshot.ProcessedEvents,
		},
		{
			"Total number of denied security events observed by the detection service.",
			"agentshield_detection_denied_events_total",
			"counter",
			snapshot.DeniedEvents,
		},
		{
			"Total number of high-risk security events observed by the detection service.",
			"agentshield_detection_high_risk_events_total",
			"counter",
			snapshot.HighRiskEvents,
		},
		{
			"Total number of security incidents triggered by the detection service.",
			"agentshield_detection_incidents_triggered_total",
			"counter",
			snapshot.IncidentsTriggered,
		},
		{
			"Total number of behavioral detections generated by the detection service.",
			"agentshield_detection_behavioral_detections_total",
			"counter",
			snapshot.BehavioralDetections,
		},
		{
			"Total number of repeated-denied-action behavioral detections.",
			"agentshield_detection_repeated_denied_total",
			"counter",
			snapshot.RepeatedDeniedDetections,
		},
		{
			"Total number of repeated high-risk behavioral detections.",
			"agentshield_detection_repeated_high_risk_total",
			"counter",
			snapshot.RepeatedHighRiskDetections,
		},
		{
			"Total number of agent action-burst behavioral detections.",
			"agentshield_detection_action_bursts_total",
			"counter",
			snapshot.ActionBurstDetections,
		},
		{
			"Total number of high action-diversity behavioral detections.",
			"agentshield_detection_action_diversity_total",
			"counter",
			snapshot.ActionDiversityDetections,
		},
		{
			"Total number of high resource-diversity behavioral detections.",
			"agentshield_detection_resource_diversity_total",
			"counter",
			snapshot.ResourceDiversityDetections,
		},
		{
			"Total number of malformed or invalid security events rejected by the detection service.",
			"agentshield_detection_rejected_events_total",
			"counter",
			snapshot.RejectedEvents,
		},
		{
			"Total number of Kafka fetch errors encountered by the detection service.",
			"agentshield_detection_fetch_errors_total",
			"counter",
			snapshot.FetchErrors,
		},
		{
			"Total number of Kafka offset commit failures encountered by the detection service.",
			"agentshield_detection_commit_failures_total",
			"counter",
			snapshot.CommitFailures,
		},
		{
			"Total number of malformed security events successfully published to the dead-letter queue.",
			"agentshield_detection_dlq_published_total",
			"counter",
			snapshot.DLQPublished,
		},
		{
			"Total number of failures while publishing security events to the dead-letter queue.",
			"agentshield_detection_dlq_failures_total",
			"counter",
			snapshot.DLQFailures,
		},
		{
			"Last successfully processed Kafka record offset.",
			"agentshield_detection_last_offset",
			"gauge",
			snapshot.LastOffset,
		},
		{
			"Last successfully processed Kafka partition.",
			"agentshield_detection_last_partition",
			"gauge",
			snapshot.LastPartition,
		},
		{
			"Total number of agent baseline anomaly evaluations.",
			"agentshield_detection_anomaly_evaluations_total",
			"counter",
			snapshot.AnomalyEvaluations,
		},
		{
			"Total number of agent behavior anomalies detected.",
			"agentshield_detection_anomaly_detections_total",
			"counter",
			snapshot.AnomalyDetections,
		},
		{
			"Total number of times an agent baseline reached the minimum warm-up sample count.",
			"agentshield_detection_baseline_warmups_total",
			"counter",
			snapshot.BaselineWarmups,
		},
		{
			"Total number of baseline updates skipped because the observation was anomalous.",
			"agentshield_detection_baseline_updates_skipped_total",
			"counter",
			snapshot.BaselineUpdatesSkipped,
		},
	}

	for _, metric := range metrics {
		if _, err := fmt.Fprintf(
			w,
			"# HELP %s %s\n# TYPE %s %s\n%s %v\n",
			metric.name,
			metric.help,
			metric.name,
			metric.mtype,
			metric.name,
			metric.value,
		); err != nil {
			return err
		}
	}

	return nil
}
