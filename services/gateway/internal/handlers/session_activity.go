package handlers

import (
	"net/http"

	"github.com/dhananjay2799/agentshield/services/gateway/internal/repository"
)

type SessionActivityHandler struct {
	AuditRepository    *repository.AuditRepository
	ApprovalRepository *repository.ApprovalRepository
}

func NewSessionActivityHandler(
	auditRepository *repository.AuditRepository,
	approvalRepository *repository.ApprovalRepository,
) *SessionActivityHandler {
	return &SessionActivityHandler{
		AuditRepository:    auditRepository,
		ApprovalRepository: approvalRepository,
	}
}

func (h *SessionActivityHandler) ListActions(
	w http.ResponseWriter,
	r *http.Request,
) {
	sessionID := r.PathValue("id")

	if sessionID == "" {
		writeJSON(
			w,
			http.StatusBadRequest,
			map[string]string{
				"error": "session id is required",
			},
		)
		return
	}

	events, err :=
		h.AuditRepository.ListBySessionID(
			r.Context(),
			sessionID,
		)

	if err != nil {
		writeJSON(
			w,
			http.StatusInternalServerError,
			map[string]string{
				"error": "failed to list session actions",
			},
		)
		return
	}

	writeJSON(
		w,
		http.StatusOK,
		events,
	)
}

func (h *SessionActivityHandler) ListApprovals(
	w http.ResponseWriter,
	r *http.Request,
) {
	sessionID := r.PathValue("id")

	if sessionID == "" {
		writeJSON(
			w,
			http.StatusBadRequest,
			map[string]string{
				"error": "session id is required",
			},
		)
		return
	}

	approvals, err :=
		h.ApprovalRepository.ListBySessionID(
			r.Context(),
			sessionID,
		)

	if err != nil {
		writeJSON(
			w,
			http.StatusInternalServerError,
			map[string]string{
				"error": "failed to list session approvals",
			},
		)
		return
	}

	writeJSON(
		w,
		http.StatusOK,
		approvals,
	)
}
