package handlers

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/dhananjay2799/agentshield/services/gateway/internal/models"
	"github.com/dhananjay2799/agentshield/services/gateway/internal/repository"
)

type AgentHandler struct {
	Repository *repository.AgentRepository
}

func NewAgentHandler(repo *repository.AgentRepository) *AgentHandler {
	return &AgentHandler{
		Repository: repo,
	}
}

func (h *AgentHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req models.CreateAgentRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{
			"error": "invalid JSON body",
		})
		return
	}

	if req.Name == "" || req.AgentType == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{
			"error": "name and agent_type are required",
		})
		return
	}

	agent, err := h.Repository.Create(r.Context(), req)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{
			"error": "failed to create agent",
		})
		return
	}

	writeJSON(w, http.StatusCreated, agent)
}

func (h *AgentHandler) List(w http.ResponseWriter, r *http.Request) {
	agents, err := h.Repository.List(r.Context())
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{
			"error": "failed to list agents",
		})
		return
	}

	writeJSON(w, http.StatusOK, agents)
}

func (h *AgentHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")

	if id == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{
			"error": "agent id is required",
		})
		return
	}

	agent, err := h.Repository.GetByID(r.Context(), id)

	if err != nil {
		if errors.Is(err, repository.ErrAgentNotFound) {
			writeJSON(w, http.StatusNotFound, map[string]string{
				"error": "agent not found",
			})
			return
		}

		writeJSON(w, http.StatusInternalServerError, map[string]string{
			"error": "failed to retrieve agent",
		})
		return
	}

	writeJSON(w, http.StatusOK, agent)
}
