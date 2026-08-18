package handlers

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/dhananjay2799/agentshield/services/credential-broker/internal/gateway"
	brokermetrics "github.com/dhananjay2799/agentshield/services/credential-broker/internal/metrics"
	"github.com/dhananjay2799/agentshield/services/credential-broker/internal/models"
	"github.com/dhananjay2799/agentshield/services/credential-broker/internal/token"
)

type CredentialHandler struct {
	Issuer        *token.Issuer
	GatewayClient *gateway.Client
	Metrics       *brokermetrics.Metrics
}

func NewCredentialHandler(
	issuer *token.Issuer,
	gatewayClient *gateway.Client,
	metrics *brokermetrics.Metrics,
) *CredentialHandler {
	return &CredentialHandler{
		Issuer:        issuer,
		GatewayClient: gatewayClient,
		Metrics:       metrics,
	}
}

func (h *CredentialHandler) Issue(
	w http.ResponseWriter,
	r *http.Request,
) {
	h.Metrics.RecordRequest()

	var req models.IssueCredentialRequest

	if err := json.NewDecoder(
		r.Body,
	).Decode(&req); err != nil {
		h.Metrics.RecordRejected()
		writeJSON(
			w,
			http.StatusBadRequest,
			map[string]string{
				"error": "invalid JSON body",
			},
		)
		return
	}

	if req.GrantID == "" ||
		req.AgentID == "" ||
		req.SessionID == "" ||
		req.Action == "" ||
		req.Resource == "" {
		h.Metrics.RecordRejected()
		writeJSON(
			w,
			http.StatusBadRequest,
			map[string]string{
				"error": "grant_id, agent_id, session_id, action, and resource are required",
			},
		)
		return
	}

	claim, err :=
		h.GatewayClient.ClaimGrant(
			r.Context(),
			req.GrantID,
			req.AgentID,
			req.SessionID,
			req.Action,
			req.Resource,
		)

	if err != nil {
		h.Metrics.RecordGatewayError()
		writeJSON(
			w,
			http.StatusBadGateway,
			map[string]string{
				"error": "unable to claim authorization grant from AgentShield Gateway",
			},
		)
		return
	}

	if !claim.Claimed {
		h.Metrics.RecordRejected()
		writeJSON(
			w,
			http.StatusForbidden,
			map[string]any{
				"error":  "authorization grant is unavailable for credential issuance",
				"reason": claim.Error,
			},
		)
		return
	}

	grant := claim.Grant

	// Never trust caller-supplied authorization context.
	// Every scoped field must exactly match the authoritative
	// grant returned by AgentShield Gateway.
	if grant.AgentID != req.AgentID ||
		grant.SessionID != req.SessionID ||
		grant.Action != req.Action ||
		grant.Resource != req.Resource {
		h.Metrics.RecordRejected()
		h.Metrics.RecordScopeMismatch()
		writeJSON(
			w,
			http.StatusForbidden,
			map[string]any{
				"error": "credential request does not match authorization grant scope",
				"grant_scope": map[string]string{
					"agent_id":   grant.AgentID,
					"session_id": grant.SessionID,
					"action":     grant.Action,
					"resource":   grant.Resource,
				},
			},
		)
		return
	}

	signedToken, claims, err :=
		h.Issuer.Issue(
			grant.ID,
			grant.AgentID,
			grant.SessionID,
			grant.Action,
			grant.Resource,
		)

	if err != nil {
		h.Metrics.RecordSigningError()
		writeJSON(
			w,
			http.StatusInternalServerError,
			map[string]string{
				"error": "failed to issue credential",
			},
		)
		return
	}

	issuedAt := time.Unix(
		claims.IssuedAt,
		0,
	).UTC()

	expiresAt := time.Unix(
		claims.ExpiresAt,
		0,
	).UTC()

	response :=
		models.IssuedCredential{
			ID:        claims.CredentialID,
			GrantID:   grant.ID,
			AgentID:   grant.AgentID,
			SessionID: grant.SessionID,

			Scope: models.CredentialScope{
				Action:   grant.Action,
				Resource: grant.Resource,
			},

			Token:     signedToken,
			IssuedAt:  issuedAt,
			ExpiresAt: expiresAt,
			Status:    "active",
		}

	h.Metrics.RecordIssued()

	writeJSON(
		w,
		http.StatusCreated,
		response,
	)
}

func writeJSON(
	w http.ResponseWriter,
	status int,
	value any,
) {
	w.Header().Set(
		"Content-Type",
		"application/json",
	)

	w.WriteHeader(status)

	_ = json.NewEncoder(
		w,
	).Encode(value)
}
