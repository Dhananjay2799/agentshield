package metrics

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sync/atomic"

	"github.com/dhananjay2799/agentshield/services/gateway/internal/models"
)

type IncidentMetricsRepository interface {
	MetricsSummary(
		ctx context.Context,
	) (models.IncidentMetricsSummary, error)
}

type Metrics struct {
	ActionEvaluations  atomic.Uint64
	Allowed            atomic.Uint64
	Denied             atomic.Uint64
	RequireApproval    atomic.Uint64
	HighRiskActions    atomic.Uint64
	OPAErrors          atomic.Uint64
	KafkaFailures      atomic.Uint64
	incidentRepository IncidentMetricsRepository
}

func New(
	incidentRepository IncidentMetricsRepository,
) *Metrics {
	return &Metrics{
		incidentRepository: incidentRepository,
	}
}

func (m *Metrics) RecordDecision(
	decision string,
	riskScore int,
) {
	m.ActionEvaluations.Add(1)

	switch decision {
	case "ALLOW":
		m.Allowed.Add(1)

	case "DENY":
		m.Denied.Add(1)

	case "REQUIRE_APPROVAL":
		m.RequireApproval.Add(1)
	}

	// AgentShield currently treats scores >= 70 as high-risk
	// for observability purposes.
	if riskScore >= 70 {
		m.HighRiskActions.Add(1)
	}
}

func (m *Metrics) RecordOPAError() {
	m.OPAErrors.Add(1)
}

func (m *Metrics) RecordKafkaFailure() {
	m.KafkaFailures.Add(1)
}

func (m *Metrics) Handler(
	w http.ResponseWriter,
	r *http.Request,
) {
	w.Header().Set(
		"Content-Type",
		"text/plain; version=0.0.4; charset=utf-8",
	)

	summary := models.IncidentMetricsSummary{}

	if m.incidentRepository != nil {
		result, err := m.incidentRepository.MetricsSummary(
			r.Context(),
		)
		if err != nil {
			http.Error(
				w,
				"failed to collect incident metrics",
				http.StatusInternalServerError,
			)
			return
		}
		summary = result
	}

	_, _ = fmt.Fprintf(
		w,
		`# HELP agentshield_gateway_action_evaluations_total Total number of completed AgentShield action evaluations.
# TYPE agentshield_gateway_action_evaluations_total counter
agentshield_gateway_action_evaluations_total %d
# HELP agentshield_gateway_allow_total Total number of final ALLOW decisions.
# TYPE agentshield_gateway_allow_total counter
agentshield_gateway_allow_total %d
# HELP agentshield_gateway_deny_total Total number of final DENY decisions.
# TYPE agentshield_gateway_deny_total counter
agentshield_gateway_deny_total %d
# HELP agentshield_gateway_require_approval_total Total number of final REQUIRE_APPROVAL decisions.
# TYPE agentshield_gateway_require_approval_total counter
agentshield_gateway_require_approval_total %d
# HELP agentshield_gateway_high_risk_actions_total Total number of high-risk evaluated actions.
# TYPE agentshield_gateway_high_risk_actions_total counter
agentshield_gateway_high_risk_actions_total %d
# HELP agentshield_gateway_opa_errors_total Total number of OPA evaluation failures.
# TYPE agentshield_gateway_opa_errors_total counter
agentshield_gateway_opa_errors_total %d
# HELP agentshield_gateway_kafka_publish_failures_total Total number of Kafka security-event publication failures.
# TYPE agentshield_gateway_kafka_publish_failures_total counter
agentshield_gateway_kafka_publish_failures_total %d
# HELP agentshield_incidents_open Current number of open AgentShield security incidents.
# TYPE agentshield_incidents_open gauge
agentshield_incidents_open %d
# HELP agentshield_incidents_investigating Current number of AgentShield security incidents under investigation.
# TYPE agentshield_incidents_investigating gauge
agentshield_incidents_investigating %d
# HELP agentshield_incidents_resolved Current number of resolved AgentShield security incidents.
# TYPE agentshield_incidents_resolved gauge
agentshield_incidents_resolved %d
# HELP agentshield_incidents_dismissed Current number of dismissed AgentShield security incidents.
# TYPE agentshield_incidents_dismissed gauge
agentshield_incidents_dismissed %d
# HELP agentshield_incidents_critical_open Current number of open critical AgentShield security incidents.
# TYPE agentshield_incidents_critical_open gauge
agentshield_incidents_critical_open %d
# HELP agentshield_incidents_unassigned_open Current number of open AgentShield security incidents without an analyst assignment.
# TYPE agentshield_incidents_unassigned_open gauge
agentshield_incidents_unassigned_open %d
# HELP agentshield_incidents_total Current total number of AgentShield security incident records.
# TYPE agentshield_incidents_total gauge
agentshield_incidents_total %d
`,
		m.ActionEvaluations.Load(),
		m.Allowed.Load(),
		m.Denied.Load(),
		m.RequireApproval.Load(),
		m.HighRiskActions.Load(),
		m.OPAErrors.Load(),
		m.KafkaFailures.Load(),
		summary.Open,
		summary.Investigating,
		summary.Resolved,
		summary.Dismissed,
		summary.CriticalOpen,
		summary.UnassignedOpen,
		summary.Total,
	)
}

func (m *Metrics) DebugHandler(
	w http.ResponseWriter,
	r *http.Request,
) {
	w.Header().Set(
		"Content-Type",
		"application/json",
	)

	summary := models.IncidentMetricsSummary{}

	if m.incidentRepository != nil {
		result, err := m.incidentRepository.MetricsSummary(
			r.Context(),
		)
		if err != nil {
			http.Error(
				w,
				"failed to collect incident metrics",
				http.StatusInternalServerError,
			)
			return
		}
		summary = result
	}

	_ = json.NewEncoder(w).Encode(
		map[string]uint64{
			"action_evaluations":        m.ActionEvaluations.Load(),
			"allow":                     m.Allowed.Load(),
			"deny":                      m.Denied.Load(),
			"require_approval":          m.RequireApproval.Load(),
			"high_risk_actions":         m.HighRiskActions.Load(),
			"opa_errors":                m.OPAErrors.Load(),
			"kafka_failures":            m.KafkaFailures.Load(),
			"incidents_open":            summary.Open,
			"incidents_investigating":   summary.Investigating,
			"incidents_resolved":        summary.Resolved,
			"incidents_dismissed":       summary.Dismissed,
			"incidents_critical_open":   summary.CriticalOpen,
			"incidents_unassigned_open": summary.UnassignedOpen,
			"incidents_total":           summary.Total,
		},
	)
}
