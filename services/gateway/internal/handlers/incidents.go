package handlers

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"

	"github.com/dhananjay2799/agentshield/services/gateway/internal/models"
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
	query := r.URL.Query()

	status :=
		strings.TrimSpace(
			query.Get("status"),
		)

	severity :=
		strings.TrimSpace(
			query.Get("severity"),
		)

	assignedTo :=
		strings.TrimSpace(
			query.Get("assigned_to"),
		)

	limit := 25
	offset := 0

	if rawLimit :=
		strings.TrimSpace(
			query.Get("limit"),
		); rawLimit != "" {

		parsed, err :=
			strconv.Atoi(
				rawLimit,
			)

		if err != nil ||
			parsed < 1 ||
			parsed > 100 {
			writeJSON(
				w,
				http.StatusBadRequest,
				map[string]string{
					"error": "limit must be between 1 and 100",
				},
			)
			return
		}

		limit = parsed
	}

	if rawOffset :=
		strings.TrimSpace(
			query.Get("offset"),
		); rawOffset != "" {

		parsed, err :=
			strconv.Atoi(
				rawOffset,
			)

		if err != nil ||
			parsed < 0 {
			writeJSON(
				w,
				http.StatusBadRequest,
				map[string]string{
					"error": "offset must be zero or greater",
				},
			)
			return
		}

		offset = parsed
	}

	filter :=
		models.IncidentListFilter{
			Status:     status,
			Severity:   severity,
			AssignedTo: assignedTo,
			Limit:      limit,
			Offset:     offset,
		}

	incidents, total, err :=
		h.Repository.List(
			r.Context(),
			filter,
		)

	if err != nil {
		writeJSON(
			w,
			http.StatusInternalServerError,
			map[string]string{
				"error": "failed to list incidents",
			},
		)
		return
	}

	response :=
		models.IncidentListResponse{
			Incidents: incidents,

			Pagination: models.IncidentPagination{
				Limit:  limit,
				Offset: offset,
				Total:  total,
			},
		}

	writeJSON(
		w,
		http.StatusOK,
		response,
	)
}

func (h *IncidentHandler) GetByID(
	w http.ResponseWriter,
	r *http.Request,
) {
	id := strings.TrimSpace(
		r.PathValue("id"),
	)

	if id == "" {
		writeJSON(
			w,
			http.StatusBadRequest,
			map[string]string{
				"error": "incident id is required",
			},
		)
		return
	}

	incident, err :=
		h.Repository.GetByID(
			r.Context(),
			id,
		)

	if err != nil {
		if errors.Is(
			err,
			repository.ErrIncidentNotFound,
		) {
			writeJSON(
				w,
				http.StatusNotFound,
				map[string]string{
					"error": "incident not found",
				},
			)
			return
		}

		writeJSON(
			w,
			http.StatusInternalServerError,
			map[string]string{
				"error": "failed to retrieve incident",
			},
		)
		return
	}

	writeJSON(
		w,
		http.StatusOK,
		incident,
	)
}

func (h *IncidentHandler) Investigate(
	w http.ResponseWriter,
	r *http.Request,
) {
	id := strings.TrimSpace(
		r.PathValue("id"),
	)

	if id == "" {
		writeJSON(
			w,
			http.StatusBadRequest,
			map[string]string{
				"error": "incident id is required",
			},
		)
		return
	}

	var req models.InvestigateIncidentRequest

	if err := json.NewDecoder(
		r.Body,
	).Decode(&req); err != nil {
		writeJSON(
			w,
			http.StatusBadRequest,
			map[string]string{
				"error": "invalid JSON body",
			},
		)
		return
	}

	req.AssignedTo =
		strings.TrimSpace(
			req.AssignedTo,
		)

	req.InvestigationNote =
		strings.TrimSpace(
			req.InvestigationNote,
		)

	if req.AssignedTo == "" {
		writeJSON(
			w,
			http.StatusBadRequest,
			map[string]string{
				"error": "assigned_to is required",
			},
		)
		return
	}

	incident, err :=
		h.Repository.MarkInvestigating(
			r.Context(),
			id,
			req.AssignedTo,
			req.InvestigationNote,
		)

	if err != nil {
		h.handleTransitionError(
			w,
			err,
		)
		return
	}

	writeJSON(
		w,
		http.StatusOK,
		incident,
	)
}

func (h *IncidentHandler) Resolve(
	w http.ResponseWriter,
	r *http.Request,
) {
	id := strings.TrimSpace(
		r.PathValue("id"),
	)

	if id == "" {
		writeJSON(
			w,
			http.StatusBadRequest,
			map[string]string{
				"error": "incident id is required",
			},
		)
		return
	}

	var req models.ResolveIncidentRequest

	if err := json.NewDecoder(
		r.Body,
	).Decode(&req); err != nil {
		writeJSON(
			w,
			http.StatusBadRequest,
			map[string]string{
				"error": "invalid JSON body",
			},
		)
		return
	}

	req.Resolution =
		strings.TrimSpace(
			req.Resolution,
		)

	if req.Resolution == "" {
		writeJSON(
			w,
			http.StatusBadRequest,
			map[string]string{
				"error": "resolution is required",
			},
		)
		return
	}

	incident, err :=
		h.Repository.Resolve(
			r.Context(),
			id,
			req.Resolution,
		)

	if err != nil {
		h.handleTransitionError(
			w,
			err,
		)
		return
	}

	writeJSON(
		w,
		http.StatusOK,
		incident,
	)
}

func (h *IncidentHandler) Dismiss(
	w http.ResponseWriter,
	r *http.Request,
) {
	id := strings.TrimSpace(
		r.PathValue("id"),
	)

	if id == "" {
		writeJSON(
			w,
			http.StatusBadRequest,
			map[string]string{
				"error": "incident id is required",
			},
		)
		return
	}

	incident, err :=
		h.Repository.Dismiss(
			r.Context(),
			id,
		)

	if err != nil {
		h.handleTransitionError(
			w,
			err,
		)
		return
	}

	writeJSON(
		w,
		http.StatusOK,
		incident,
	)
}

func (h *IncidentHandler) handleTransitionError(
	w http.ResponseWriter,
	err error,
) {
	if errors.Is(
		err,
		repository.ErrIncidentNotFound,
	) {
		writeJSON(
			w,
			http.StatusNotFound,
			map[string]string{
				"error": "incident not found or transition is not allowed",
			},
		)
		return
	}

	writeJSON(
		w,
		http.StatusInternalServerError,
		map[string]string{
			"error": "failed to update incident",
		},
	)
}
