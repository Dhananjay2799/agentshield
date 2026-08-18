package metrics

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync/atomic"
)

type Metrics struct {
	ActionEvaluations atomic.Uint64
	Allowed           atomic.Uint64
	Denied            atomic.Uint64
	RequireApproval   atomic.Uint64
	HighRiskActions   atomic.Uint64
	OPAErrors         atomic.Uint64
	KafkaFailures     atomic.Uint64
}

func New() *Metrics {
	return &Metrics{}
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
	_ *http.Request,
) {
	w.Header().Set(
		"Content-Type",
		"text/plain; version=0.0.4; charset=utf-8",
	)

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
`,
		m.ActionEvaluations.Load(),
		m.Allowed.Load(),
		m.Denied.Load(),
		m.RequireApproval.Load(),
		m.HighRiskActions.Load(),
		m.OPAErrors.Load(),
		m.KafkaFailures.Load(),
	)
}

func (m *Metrics) DebugHandler(
	w http.ResponseWriter,
	_ *http.Request,
) {
	w.Header().Set(
		"Content-Type",
		"application/json",
	)

	_ = json.NewEncoder(w).Encode(
		map[string]uint64{
			"action_evaluations": m.ActionEvaluations.Load(),
			"allow":              m.Allowed.Load(),
			"deny":               m.Denied.Load(),
			"require_approval":   m.RequireApproval.Load(),
			"high_risk_actions":  m.HighRiskActions.Load(),
			"opa_errors":         m.OPAErrors.Load(),
			"kafka_failures":     m.KafkaFailures.Load(),
		},
	)
}
