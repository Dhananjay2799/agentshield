package handlers

import (
	"errors"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/dhananjay2799/agentshield/services/gateway/internal/events"
	"github.com/dhananjay2799/agentshield/services/gateway/internal/repository"
)

type ContainmentHandler struct {
	Repository      *repository.ContainmentRepository
	AuditRepository *repository.AuditRepository
	EventProducer   *events.Producer
}

func NewContainmentHandler(
	repository *repository.ContainmentRepository,
	auditRepository *repository.AuditRepository,
	eventProducer *events.Producer,
) *ContainmentHandler {
	return &ContainmentHandler{
		Repository:      repository,
		AuditRepository: auditRepository,
		EventProducer:   eventProducer,
	}
}

func (h *ContainmentHandler) Contain(
	w http.ResponseWriter,
	r *http.Request,
) {
	agentID := strings.TrimSpace(
		r.PathValue("id"),
	)

	if agentID == "" {
		writeJSON(
			w,
			http.StatusBadRequest,
			map[string]string{
				"error": "agent id is required",
			},
		)
		return
	}

	result, err :=
		h.Repository.ContainAgent(
			r.Context(),
			agentID,
		)

	if err != nil {
		if errors.Is(
			err,
			repository.ErrAgentNotFound,
		) {
			writeJSON(
				w,
				http.StatusNotFound,
				map[string]string{
					"error": "agent not found",
				},
			)
			return
		}

		writeJSON(
			w,
			http.StatusInternalServerError,
			map[string]string{
				"error": "failed to contain agent",
			},
		)
		return
	}

	metadata := map[string]any{
		"agent_status":     result.AgentStatus,
		"sessions_revoked": result.SessionsRevoked,
		"grants_revoked":   result.GrantsRevoked,
		"containment":      true,
		"source":           "gateway-admin-api",
	}

	// Persist the containment operation in the permanent audit trail.
	if h.AuditRepository != nil {
		if err := h.AuditRepository.Create(
			r.Context(),
			repository.CreateAuditEventParams{
				AgentID:   result.AgentID,
				EventType: "agent.contained",
				Action:    "agent.contain",
				Resource:  "agent/" + result.AgentID,
				Decision:  "SUCCESS",
				RiskScore: 0,
				Metadata:  metadata,
			},
		); err != nil {
			log.Printf(
				"failed to persist containment audit event: agent=%s error=%v",
				result.AgentID,
				err,
			)
		}
	}

	// Publish containment into the security-event stream.
	//
	// Kafka failure deliberately does not roll back containment.
	// Security containment has already succeeded and must remain effective.
	if h.EventProducer != nil {
		event := events.SecurityEvent{
			EventType:  "agent.contained",
			AgentID:    result.AgentID,
			SessionID:  "",
			Action:     "agent.contain",
			Resource:   "agent/" + result.AgentID,
			Decision:   "SUCCESS",
			RiskScore:  0,
			Metadata:   metadata,
			OccurredAt: time.Now().UTC(),
		}

		if err := h.EventProducer.Publish(
			r.Context(),
			event,
		); err != nil {
			log.Printf(
				"failed to publish containment security event: agent=%s error=%v",
				result.AgentID,
				err,
			)
		} else {
			log.Printf(
				"containment security event published successfully: agent=%s sessions_revoked=%d grants_revoked=%d",
				result.AgentID,
				result.SessionsRevoked,
				result.GrantsRevoked,
			)
		}
	}

	writeJSON(
		w,
		http.StatusOK,
		map[string]any{
			"status":           "contained",
			"agent_id":         result.AgentID,
			"agent_status":     result.AgentStatus,
			"sessions_revoked": result.SessionsRevoked,
			"grants_revoked":   result.GrantsRevoked,
		},
	)
}
