package handlers

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"time"

	"github.com/dhananjay2799/agentshield/services/gateway/internal/events"
	"github.com/dhananjay2799/agentshield/services/gateway/internal/middleware"
	"github.com/dhananjay2799/agentshield/services/gateway/internal/models"
	"github.com/dhananjay2799/agentshield/services/gateway/internal/opa"
	"github.com/dhananjay2799/agentshield/services/gateway/internal/repository"
	"github.com/dhananjay2799/agentshield/services/gateway/internal/risk"
)

type ActionHandler struct {
	AgentRepository    *repository.AgentRepository
	AuditRepository    *repository.AuditRepository
	ApprovalRepository *repository.ApprovalRepository
	GrantRepository    *repository.GrantRepository
	OPAClient          *opa.Client
	EventProducer      *events.Producer
}

func NewActionHandler(
	agentRepository *repository.AgentRepository,
	auditRepository *repository.AuditRepository,
	approvalRepository *repository.ApprovalRepository,
	grantRepository *repository.GrantRepository,
	opaClient *opa.Client,
	eventProducer *events.Producer,
) *ActionHandler {
	return &ActionHandler{
		AgentRepository:    agentRepository,
		AuditRepository:    auditRepository,
		ApprovalRepository: approvalRepository,
		GrantRepository:    grantRepository,
		OPAClient:          opaClient,
		EventProducer:      eventProducer,
	}
}

func (h *ActionHandler) Evaluate(w http.ResponseWriter, r *http.Request) {
	var req models.ActionEvaluationRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{
			"error": "invalid JSON body",
		})
		return
	}

	if req.Action == "" || req.Resource == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{
			"error": "action and resource are required",
		})
		return
	}

	// SessionRequired middleware already validated this session.
	session, ok := middleware.SessionFromContext(r.Context())
	if !ok {
		writeJSON(w, http.StatusUnauthorized, map[string]string{
			"error": "AgentShield session identity unavailable",
		})
		return
	}

	// Load the registered agent so OPA can evaluate identity attributes.
	agent, err := h.AgentRepository.GetByID(
		r.Context(),
		session.AgentID,
	)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{
			"error": "failed to retrieve agent identity",
		})
		return
	}

	// Calculate contextual risk independently of policy.
	riskResult := risk.Evaluate(
		req.Action,
		req.Resource,
	)

	// Ask OPA/Rego for the policy decision.
	opaDecision, err := h.OPAClient.Evaluate(
		r.Context(),
		opa.Input{
			AgentID:     agent.ID,
			AgentType:   agent.AgentType,
			SessionID:   session.ID,
			Action:      req.Action,
			Resource:    req.Resource,
			Environment: agent.Environment,
			RiskScore:   riskResult.Score,
		},
	)

	if err != nil {
		log.Printf("OPA evaluation failed: %v", err)

		// Fail closed: AgentShield does not authorize an action
		// when its policy engine cannot be reached.
		writeJSON(w, http.StatusServiceUnavailable, map[string]string{
			"error": "policy engine unavailable",
		})
		return
	}

	decision := opaDecision.Decision
	policyReason := opaDecision.Reason
	var approvalID string
	var grantID string

	switch decision {

	case "ALLOW":
		// No additional workflow required.

	case "DENY":
		// Explicit policy denial.

	case "REQUIRE_APPROVAL":

		// Before creating a new approval, check whether the human
		// already approved this exact action for this exact session.
		grant, grantErr := h.GrantRepository.FindActiveGrant(
			r.Context(),
			session.AgentID,
			session.ID,
			req.Action,
			req.Resource,
		)

		if grantErr == nil {
			if err := h.GrantRepository.Consume(
				r.Context(),
				grant.ID,
			); err != nil {
				writeJSON(w, http.StatusInternalServerError, map[string]string{
					"error": "failed to consume authorization grant",
				})
				return
			}

			grantID = grant.ID
			decision = "ALLOW"
			policyReason = "human-approved temporary authorization grant"

		} else if errors.Is(grantErr, repository.ErrGrantNotFound) {

			approval, err := h.ApprovalRepository.Create(
				r.Context(),
				session.AgentID,
				session.ID,
				req.Action,
				req.Resource,
				req.Reason,
				riskResult.Score,
			)

			if err != nil {
				writeJSON(w, http.StatusInternalServerError, map[string]string{
					"error": "failed to create approval request",
				})
				return
			}

			approvalID = approval.ID

		} else {
			log.Printf("failed to check authorization grant: %v", grantErr)

			writeJSON(w, http.StatusInternalServerError, map[string]string{
				"error": "failed to evaluate authorization grant",
			})
			return
		}

	default:
		log.Printf("OPA returned unsupported decision: %s", decision)

		writeJSON(w, http.StatusServiceUnavailable, map[string]string{
			"error": "policy engine returned invalid decision",
		})
		return
	}

	response := models.ActionEvaluationResponse{
		Decision:   decision,
		RiskScore:  riskResult.Score,
		Reason:     riskResult.Reason + "; " + policyReason,
		Action:     req.Action,
		Resource:   req.Resource,
		AgentID:    session.AgentID,
		SessionID:  session.ID,
		ApprovalID: approvalID,
		GrantID:    grantID,
		Timestamp:  time.Now().UTC(),
	}

	// Every final decision becomes a permanent security audit event.
	err = h.AuditRepository.Create(
		r.Context(),
		repository.CreateAuditEventParams{
			AgentID:   session.AgentID,
			SessionID: session.ID,
			EventType: "action_evaluation",
			Action:    req.Action,
			Resource:  req.Resource,
			Decision:  decision,
			RiskScore: riskResult.Score,
			Metadata: map[string]any{
				"request_reason": req.Reason,
				"risk_reason":    riskResult.Reason,
				"policy_reason":  policyReason,
				"policy_engine":  "opa",
				"agent_type":     agent.AgentType,
				"environment":    agent.Environment,
				"approval_id":    approvalID,
				"grant_id":       grantID,
			},
		},
	)

	event := events.SecurityEvent{
		EventType: "action_evaluation",
		AgentID:   session.AgentID,
		SessionID: session.ID,
		Action:    req.Action,
		Resource:  req.Resource,
		Decision:  decision,
		RiskScore: riskResult.Score,
		Metadata: map[string]any{
			"request_reason": req.Reason,
			"policy_reason":  policyReason,
			"approval_id":    approvalID,
			"grant_id":       grantID,
			"agent_type":     agent.AgentType,
			"environment":    agent.Environment,
		},
		OccurredAt: time.Now().UTC(),
	}

	log.Printf(
		"publishing security event to Kafka: action=%s decision=%s agent=%s",
		req.Action,
		decision,
		session.AgentID,
	)

	if err := h.EventProducer.Publish(r.Context(), event); err != nil {
		log.Printf("failed to publish security event: %v", err)
	} else {
		log.Printf("security event published successfully")
	}

	writeJSON(w, http.StatusOK, response)
}
