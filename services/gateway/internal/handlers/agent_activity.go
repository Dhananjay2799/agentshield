package handlers

import (
	"net/http"

	"github.com/dhananjay2799/agentshield/services/gateway/internal/repository"
)

type AgentActivityHandler struct {
	SessionRepository  *repository.SessionRepository
	AuditRepository    *repository.AuditRepository
	ApprovalRepository *repository.ApprovalRepository
}

func NewAgentActivityHandler(
	sessionRepository *repository.SessionRepository,
	auditRepository *repository.AuditRepository,
	approvalRepository *repository.ApprovalRepository,
) *AgentActivityHandler {
	return &AgentActivityHandler{
		SessionRepository:  sessionRepository,
		AuditRepository:    auditRepository,
		ApprovalRepository: approvalRepository,
	}
}

func (h *AgentActivityHandler) ListSessions(
	w http.ResponseWriter,
	r *http.Request,
) {
	agentID := r.PathValue("id")

	if agentID == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{
			"error": "agent id is required",
		})
		return
	}

	sessions, err := h.SessionRepository.ListByAgentID(
		r.Context(),
		agentID,
	)

	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{
			"error": "failed to list agent sessions",
		})
		return
	}

	writeJSON(w, http.StatusOK, sessions)
}

func (h *AgentActivityHandler) ListActions(
	w http.ResponseWriter,
	r *http.Request,
) {
	agentID := r.PathValue("id")

	if agentID == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{
			"error": "agent id is required",
		})
		return
	}

	events, err := h.AuditRepository.ListByAgentID(
		r.Context(),
		agentID,
	)

	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{
			"error": "failed to list agent actions",
		})
		return
	}

	writeJSON(w, http.StatusOK, events)
}

func (h *AgentActivityHandler) ListApprovals(
	w http.ResponseWriter,
	r *http.Request,
) {
	agentID := r.PathValue("id")

	if agentID == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{
			"error": "agent id is required",
		})
		return
	}

	approvals, err := h.ApprovalRepository.ListByAgentID(
		r.Context(),
		agentID,
	)

	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{
			"error": "failed to list agent approvals",
		})
		return
	}

	writeJSON(w, http.StatusOK, approvals)
}

func (h *AgentActivityHandler) ListSessionSecurity(
	w http.ResponseWriter,
	r *http.Request,
) {
	agentID := r.PathValue("id")

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

	sessions, err :=
		h.SessionRepository.ListSecurityByAgentID(
			r.Context(),
			agentID,
		)

	if err != nil {
		writeJSON(
			w,
			http.StatusInternalServerError,
			map[string]string{
				"error": "failed to list session security activity",
			},
		)
		return
	}

	writeJSON(
		w,
		http.StatusOK,
		sessions,
	)
}
