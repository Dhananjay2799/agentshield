package handlers

import (
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/dhananjay2799/agentshield/services/gateway/internal/models"
	"github.com/dhananjay2799/agentshield/services/gateway/internal/repository"
)

type GrantHandler struct {
	Repository *repository.GrantRepository
}

func NewGrantHandler(
	repo *repository.GrantRepository,
) *GrantHandler {
	return &GrantHandler{
		Repository: repo,
	}
}

func (h *GrantHandler) GetByID(
	w http.ResponseWriter,
	r *http.Request,
) {
	id := r.PathValue("id")

	grant, err := h.Repository.GetByID(
		r.Context(),
		id,
	)

	if err != nil {
		if errors.Is(
			err,
			repository.ErrGrantNotFound,
		) {
			writeJSON(
				w,
				http.StatusNotFound,
				map[string]string{
					"error": "authorization grant not found",
				},
			)
			return
		}

		writeJSON(
			w,
			http.StatusInternalServerError,
			map[string]string{
				"error": "failed to retrieve authorization grant",
			},
		)
		return
	}

	writeJSON(
		w,
		http.StatusOK,
		grant,
	)
}

func (h *GrantHandler) Verify(
	w http.ResponseWriter,
	r *http.Request,
) {
	id := r.PathValue("id")

	grant, err := h.Repository.GetByID(
		r.Context(),
		id,
	)

	if err != nil {
		if errors.Is(
			err,
			repository.ErrGrantNotFound,
		) {
			writeJSON(
				w,
				http.StatusNotFound,
				map[string]any{
					"valid": false,
					"error": "authorization grant not found",
				},
			)
			return
		}

		writeJSON(
			w,
			http.StatusInternalServerError,
			map[string]any{
				"valid": false,
				"error": "failed to verify authorization grant",
			},
		)
		return
	}

	valid :=
		grant.Status == "active" &&
			grant.ExpiresAt.After(
				time.Now().UTC(),
			)

	reason := "authorization grant is valid"

	if !valid {
		switch {
		case grant.Status != "active":
			reason =
				"authorization grant is not active"

		case !grant.ExpiresAt.After(
			time.Now().UTC(),
		):
			reason =
				"authorization grant has expired"

		default:
			reason =
				"authorization grant is invalid"
		}
	}

	writeJSON(
		w,
		http.StatusOK,
		map[string]any{
			"valid":  valid,
			"reason": reason,
			"grant":  grant,
		},
	)
}

func (h *GrantHandler) ClaimForCredential(
	w http.ResponseWriter,
	r *http.Request,
) {
	id := r.PathValue("id")

	var req models.ClaimGrantRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(
			w,
			http.StatusBadRequest,
			map[string]any{
				"claimed": false,
				"error":   "invalid JSON body",
			},
		)
		return
	}

	if req.AgentID == "" ||
		req.SessionID == "" ||
		req.Action == "" ||
		req.Resource == "" {
		writeJSON(
			w,
			http.StatusBadRequest,
			map[string]any{
				"claimed": false,
				"error":   "agent_id, session_id, action, and resource are required",
			},
		)
		return
	}

	grant, err := h.Repository.ConsumeForCredential(
		r.Context(),
		id,
		req.AgentID,
		req.SessionID,
		req.Action,
		req.Resource,
	)

	if err != nil {
		if errors.Is(
			err,
			repository.ErrGrantNotFound,
		) {
			writeJSON(
				w,
				http.StatusConflict,
				map[string]any{
					"claimed": false,
					"error":   "authorization grant is unavailable, expired, already consumed, or does not match the requested scope",
				},
			)
			return
		}

		writeJSON(
			w,
			http.StatusInternalServerError,
			map[string]any{
				"claimed": false,
				"error":   "failed to claim authorization grant",
			},
		)
		return
	}

	writeJSON(
		w,
		http.StatusOK,
		map[string]any{
			"claimed": true,
			"grant":   grant,
		},
	)
}
