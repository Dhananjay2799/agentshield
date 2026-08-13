package handlers

import (
	"errors"
	"net/http"

	"github.com/dhananjay2799/agentshield/services/gateway/internal/repository"
)

type IncidentHandler struct {
	Repository *repository.IncidentRepository
}

func NewIncidentHandler(
	repo *repository.IncidentRepository,
) *IncidentHandler {
	return &IncidentHandler{
		Repository: repo,
	}
}

func (h *IncidentHandler) List(
	w http.ResponseWriter,
	r *http.Request,
) {
	incidents, err := h.Repository.List(r.Context())
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{
			"error": "failed to list incidents",
		})
		return
	}

	writeJSON(w, http.StatusOK, incidents)
}

func (h *IncidentHandler) GetByID(
	w http.ResponseWriter,
	r *http.Request,
) {
	id := r.PathValue("id")

	if id == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{
			"error": "incident id is required",
		})
		return
	}

	incident, err := h.Repository.GetByID(
		r.Context(),
		id,
	)

	if err != nil {
		if errors.Is(err, repository.ErrIncidentNotFound) {
			writeJSON(w, http.StatusNotFound, map[string]string{
				"error": "incident not found",
			})
			return
		}

		writeJSON(w, http.StatusInternalServerError, map[string]string{
			"error": "failed to retrieve incident",
		})
		return
	}

	writeJSON(w, http.StatusOK, incident)
}

func (h *IncidentHandler) Investigate(
	w http.ResponseWriter,
	r *http.Request,
) {
	id := r.PathValue("id")

	incident, err := h.Repository.MarkInvestigating(
		r.Context(),
		id,
	)

	if err != nil {
		h.handleTransitionError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, incident)
}

func (h *IncidentHandler) Resolve(
	w http.ResponseWriter,
	r *http.Request,
) {
	id := r.PathValue("id")

	incident, err := h.Repository.Resolve(
		r.Context(),
		id,
	)

	if err != nil {
		h.handleTransitionError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, incident)
}

func (h *IncidentHandler) Dismiss(
	w http.ResponseWriter,
	r *http.Request,
) {
	id := r.PathValue("id")

	incident, err := h.Repository.Dismiss(
		r.Context(),
		id,
	)

	if err != nil {
		h.handleTransitionError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, incident)
}

func (h *IncidentHandler) handleTransitionError(
	w http.ResponseWriter,
	err error,
) {
	if errors.Is(err, repository.ErrIncidentNotFound) {
		writeJSON(w, http.StatusNotFound, map[string]string{
			"error": "incident not found or transition is not allowed",
		})
		return
	}

	writeJSON(w, http.StatusInternalServerError, map[string]string{
		"error": "failed to update incident",
	})
}
