package handlers

import (
	"errors"
	"net/http"

	"github.com/dhananjay2799/agentshield/services/gateway/internal/models"
	"github.com/dhananjay2799/agentshield/services/gateway/internal/repository"
)

type SessionHandler struct {
	Repository *repository.SessionRepository
}

func NewSessionHandler(repo *repository.SessionRepository) *SessionHandler {
	return &SessionHandler{
		Repository: repo,
	}
}

func (h *SessionHandler) Create(w http.ResponseWriter, r *http.Request) {
	agentID := r.PathValue("id")

	if agentID == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{
			"error": "agent id is required",
		})
		return
	}

	var req models.CreateSessionRequest

	if !decodeJSONBody(
		w,
		r,
		&req,
	) {
		return
	}

	if req.TaskID == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{
			"error": "task_id is required",
		})
		return
	}

	session, err := h.Repository.Create(
		r.Context(),
		agentID,
		req,
	)

	if err != nil {
		switch {
		case errors.Is(
			err,
			repository.ErrAgentNotFound,
		):
			writeJSON(
				w,
				http.StatusNotFound,
				map[string]string{
					"error": "agent not found",
				},
			)

		case errors.Is(
			err,
			repository.ErrAgentNotActive,
		):
			writeJSON(
				w,
				http.StatusConflict,
				map[string]string{
					"error": "agent is not active",
				},
			)

		default:
			writeJSON(
				w,
				http.StatusInternalServerError,
				map[string]string{
					"error": "failed to create session",
				},
			)
		}

		return
	}

	writeJSON(w, http.StatusCreated, session)
}

func (h *SessionHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	sessionID := r.PathValue("id")

	if sessionID == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{
			"error": "session id is required",
		})
		return
	}

	session, err := h.Repository.GetByID(
		r.Context(),
		sessionID,
	)

	if err != nil {
		if errors.Is(err, repository.ErrSessionNotFound) {
			writeJSON(w, http.StatusNotFound, map[string]string{
				"error": "session not found",
			})
			return
		}

		writeJSON(w, http.StatusInternalServerError, map[string]string{
			"error": "failed to retrieve session",
		})
		return
	}

	writeJSON(w, http.StatusOK, session)
}

func (h *SessionHandler) Revoke(w http.ResponseWriter, r *http.Request) {
	sessionID := r.PathValue("id")

	if sessionID == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{
			"error": "session id is required",
		})
		return
	}

	err := h.Repository.Revoke(
		r.Context(),
		sessionID,
	)

	if err != nil {
		if errors.Is(err, repository.ErrSessionNotFound) {
			writeJSON(w, http.StatusNotFound, map[string]string{
				"error": "active session not found",
			})
			return
		}

		writeJSON(w, http.StatusInternalServerError, map[string]string{
			"error": "failed to revoke session",
		})
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{
		"status":  "revoked",
		"session": sessionID,
	})
}
