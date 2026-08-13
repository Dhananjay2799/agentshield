package handlers

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"github.com/dhananjay2799/agentshield/services/gateway/internal/models"
	"github.com/dhananjay2799/agentshield/services/gateway/internal/policy"
	"github.com/dhananjay2799/agentshield/services/gateway/internal/repository"
)

type PolicyHandler struct {
	Repository *repository.PolicyRepository
}

func NewPolicyHandler(
	repository *repository.PolicyRepository,
) *PolicyHandler {
	return &PolicyHandler{
		Repository: repository,
	}
}

func (h *PolicyHandler) List(
	w http.ResponseWriter,
	r *http.Request,
) {
	policies, err :=
		h.Repository.List(
			r.Context(),
		)

	if err != nil {
		writeJSON(
			w,
			http.StatusInternalServerError,
			map[string]string{
				"error": "failed to list policies",
			},
		)
		return
	}

	writeJSON(
		w,
		http.StatusOK,
		policies,
	)
}

func (h *PolicyHandler) GetByID(
	w http.ResponseWriter,
	r *http.Request,
) {
	id := r.PathValue("id")

	policy, err :=
		h.Repository.GetByID(
			r.Context(),
			id,
		)

	if err != nil {
		if errors.Is(
			err,
			repository.ErrPolicyNotFound,
		) {
			writeJSON(
				w,
				http.StatusNotFound,
				map[string]string{
					"error": "policy not found",
				},
			)
			return
		}

		writeJSON(
			w,
			http.StatusInternalServerError,
			map[string]string{
				"error": "failed to retrieve policy",
			},
		)
		return
	}

	writeJSON(
		w,
		http.StatusOK,
		policy,
	)
}

func (h *PolicyHandler) Create(
	w http.ResponseWriter,
	r *http.Request,
) {
	var req models.CreatePolicyRequest

	if err := json.NewDecoder(
		r.Body,
	).Decode(&req); err != nil {
		writeJSON(
			w,
			http.StatusBadRequest,
			map[string]string{
				"error": "invalid JSON payload",
			},
		)
		return
	}

	req.Name =
		strings.TrimSpace(req.Name)

	req.Description =
		strings.TrimSpace(
			req.Description,
		)

	req.Effect =
		strings.ToUpper(
			strings.TrimSpace(
				req.Effect,
			),
		)

	req.Action =
		strings.TrimSpace(
			req.Action,
		)

	req.ActionMatch =
		strings.ToLower(
			strings.TrimSpace(
				req.ActionMatch,
			),
		)

	req.Resource =
		strings.TrimSpace(
			req.Resource,
		)

	req.ResourceMatch =
		strings.ToLower(
			strings.TrimSpace(
				req.ResourceMatch,
			),
		)

	req.CreatedBy =
		strings.TrimSpace(
			req.CreatedBy,
		)

	if req.Name == "" {
		writeJSON(
			w,
			http.StatusBadRequest,
			map[string]string{
				"error": "name is required",
			},
		)
		return
	}

	if req.Action == "" {
		writeJSON(
			w,
			http.StatusBadRequest,
			map[string]string{
				"error": "action is required",
			},
		)
		return
	}

	if req.Resource == "" {
		writeJSON(
			w,
			http.StatusBadRequest,
			map[string]string{
				"error": "resource is required",
			},
		)
		return
	}

	switch req.Effect {
	case "ALLOW",
		"REQUIRE_APPROVAL",
		"DENY":

	default:
		writeJSON(
			w,
			http.StatusBadRequest,
			map[string]string{
				"error": "effect must be ALLOW, REQUIRE_APPROVAL, or DENY",
			},
		)
		return
	}

	if req.ActionMatch == "" {
		req.ActionMatch = "exact"
	}

	switch req.ActionMatch {
	case "exact",
		"prefix",
		"suffix":

	default:
		writeJSON(
			w,
			http.StatusBadRequest,
			map[string]string{
				"error": "action_match must be exact, prefix, or suffix",
			},
		)
		return
	}

	if req.ResourceMatch == "" {
		req.ResourceMatch = "prefix"
	}

	switch req.ResourceMatch {
	case "exact",
		"prefix",
		"suffix":

	default:
		writeJSON(
			w,
			http.StatusBadRequest,
			map[string]string{
				"error": "resource_match must be exact, prefix, or suffix",
			},
		)
		return
	}

	policy, err :=
		h.Repository.Create(
			r.Context(),
			req,
		)

	if err != nil {
		writeJSON(
			w,
			http.StatusInternalServerError,
			map[string]string{
				"error": "failed to create policy",
			},
		)
		return
	}

	writeJSON(
		w,
		http.StatusCreated,
		policy,
	)
}

func (h *PolicyHandler) Validate(
	w http.ResponseWriter,
	r *http.Request,
) {
	id := r.PathValue("id")

	storedPolicy, err :=
		h.Repository.GetByID(
			r.Context(),
			id,
		)

	if err != nil {
		if errors.Is(
			err,
			repository.ErrPolicyNotFound,
		) {
			writeJSON(
				w,
				http.StatusNotFound,
				map[string]string{
					"error": "policy not found",
				},
			)
			return
		}

		writeJSON(
			w,
			http.StatusInternalServerError,
			map[string]string{
				"error": "failed to retrieve policy",
			},
		)
		return
	}

	result :=
		policy.Validate(
			storedPolicy,
		)

	statusCode :=
		http.StatusOK

	if !result.Valid {
		statusCode =
			http.StatusUnprocessableEntity
	}

	writeJSON(
		w,
		statusCode,
		result,
	)
}
