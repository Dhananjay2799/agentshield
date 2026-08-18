package metrics

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync/atomic"
)

type Metrics struct {
	IssueRequests   atomic.Uint64
	Issued          atomic.Uint64
	Rejected        atomic.Uint64
	GatewayErrors   atomic.Uint64
	SigningErrors   atomic.Uint64
	ScopeMismatches atomic.Uint64
}

func New() *Metrics {
	return &Metrics{}
}

func (m *Metrics) RecordRequest() {
	m.IssueRequests.Add(1)
}

func (m *Metrics) RecordIssued() {
	m.Issued.Add(1)
}

func (m *Metrics) RecordRejected() {
	m.Rejected.Add(1)
}

func (m *Metrics) RecordGatewayError() {
	m.GatewayErrors.Add(1)
}

func (m *Metrics) RecordSigningError() {
	m.SigningErrors.Add(1)
}

func (m *Metrics) RecordScopeMismatch() {
	m.ScopeMismatches.Add(1)
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
		`# HELP agentshield_credential_broker_issue_requests_total Total number of credential issuance requests received.
# TYPE agentshield_credential_broker_issue_requests_total counter
agentshield_credential_broker_issue_requests_total %d
# HELP agentshield_credential_broker_issued_total Total number of short-lived credentials successfully issued.
# TYPE agentshield_credential_broker_issued_total counter
agentshield_credential_broker_issued_total %d
# HELP agentshield_credential_broker_rejected_total Total number of credential issuance requests rejected by the broker.
# TYPE agentshield_credential_broker_rejected_total counter
agentshield_credential_broker_rejected_total %d
# HELP agentshield_credential_broker_gateway_errors_total Total number of failures communicating with AgentShield Gateway during grant claims.
# TYPE agentshield_credential_broker_gateway_errors_total counter
agentshield_credential_broker_gateway_errors_total %d
# HELP agentshield_credential_broker_signing_errors_total Total number of token signing or credential issuance failures.
# TYPE agentshield_credential_broker_signing_errors_total counter
agentshield_credential_broker_signing_errors_total %d
# HELP agentshield_credential_broker_scope_mismatches_total Total number of requests rejected because the requested scope did not match the authorization grant.
# TYPE agentshield_credential_broker_scope_mismatches_total counter
agentshield_credential_broker_scope_mismatches_total %d
`,
		m.IssueRequests.Load(),
		m.Issued.Load(),
		m.Rejected.Load(),
		m.GatewayErrors.Load(),
		m.SigningErrors.Load(),
		m.ScopeMismatches.Load(),
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
			"issue_requests":   m.IssueRequests.Load(),
			"issued":           m.Issued.Load(),
			"rejected":         m.Rejected.Load(),
			"gateway_errors":   m.GatewayErrors.Load(),
			"signing_errors":   m.SigningErrors.Load(),
			"scope_mismatches": m.ScopeMismatches.Load(),
		},
	)
}
