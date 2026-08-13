package handlers

import (
	"net/http"
	"strconv"

	"github.com/dhananjay2799/agentshield/services/gateway/internal/repository"
)

type EventHandler struct {
	AuditRepository *repository.AuditRepository
}

func NewEventHandler(
	auditRepository *repository.AuditRepository,
) *EventHandler {
	return &EventHandler{
		AuditRepository: auditRepository,
	}
}

func (h *EventHandler) ListRecent(
	w http.ResponseWriter,
	r *http.Request,
) {
	limit := 100

	if value := r.URL.Query().Get("limit"); value != "" {
		parsed, err := strconv.Atoi(value)

		if err != nil || parsed <= 0 {
			writeJSON(
				w,
				http.StatusBadRequest,
				map[string]string{
					"error": "limit must be a positive integer",
				},
			)
			return
		}

		limit = parsed
	}

	events, err := h.AuditRepository.ListRecent(
		r.Context(),
		limit,
	)

	if err != nil {
		writeJSON(
			w,
			http.StatusInternalServerError,
			map[string]string{
				"error": "failed to list security events",
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
