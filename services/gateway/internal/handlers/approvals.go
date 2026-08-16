package handlers

import (
	"errors"
	"net/http"

	"github.com/dhananjay2799/agentshield/services/gateway/internal/repository"
)

type ApprovalHandler struct {
	Repository      *repository.ApprovalRepository
	GrantRepository *repository.GrantRepository
}

func NewApprovalHandler(
	repo *repository.ApprovalRepository,
	grantRepo *repository.GrantRepository,
) *ApprovalHandler {
	return &ApprovalHandler{
		Repository:      repo,
		GrantRepository: grantRepo,
	}
}

func (h *ApprovalHandler) ListPending(w http.ResponseWriter, r *http.Request) {
	approvals, err := h.Repository.ListPending(r.Context())
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{
			"error": "failed to list approval requests",
		})
		return
	}

	writeJSON(w, http.StatusOK, approvals)
}

func (h *ApprovalHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")

	approval, err := h.Repository.GetByID(r.Context(), id)
	if err != nil {
		if errors.Is(err, repository.ErrApprovalNotFound) {
			writeJSON(w, http.StatusNotFound, map[string]string{
				"error": "approval request not found",
			})
			return
		}

		writeJSON(w, http.StatusInternalServerError, map[string]string{
			"error": "failed to retrieve approval request",
		})
		return
	}

	writeJSON(w, http.StatusOK, approval)
}

func (h *ApprovalHandler) GetLineage(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")

	approval, err := h.Repository.GetByID(r.Context(), id)
	if err != nil {
		if errors.Is(err, repository.ErrApprovalNotFound) {
			writeJSON(w, http.StatusNotFound, map[string]string{
				"error": "approval request not found",
			})
			return
		}

		writeJSON(w, http.StatusInternalServerError, map[string]string{
			"error": "failed to retrieve approval lineage",
		})
		return
	}

	grant, err := h.GrantRepository.GetByApprovalID(
		r.Context(),
		approval.ID,
	)

	if err != nil && !errors.Is(err, repository.ErrGrantNotFound) {
		writeJSON(w, http.StatusInternalServerError, map[string]string{
			"error": "failed to retrieve authorization grant lineage",
		})
		return
	}

	state := "unknown"
	finalDecision := ""

	switch approval.Status {
	case "pending":
		state = "awaiting_human_approval"

	case "denied":
		state = "denied"
		finalDecision = "DENY"

	case "expired":
		state = "approval_expired"

	case "cancelled":
		state = "cancelled"

	case "approved":
		if grant == nil {
			state = "approved_without_grant"
		}
	}

	if grant != nil {
		switch grant.Status {
		case "active":
			state = "grant_active"

		case "used":
			state = "consumed"
			finalDecision = "ALLOW"

		case "expired":
			state = "grant_expired"

		case "revoked":
			state = "grant_revoked"
		}
	}

	lineage := map[string]any{
		"state": state,
	}

	if finalDecision != "" {
		lineage["final_decision"] = finalDecision
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"approval": approval,
		"grant":    grant,
		"lineage":  lineage,
	})
}

func (h *ApprovalHandler) Approve(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")

	err := h.Repository.Approve(r.Context(), id)
	if err != nil {
		if errors.Is(err, repository.ErrApprovalNotFound) {
			writeJSON(w, http.StatusNotFound, map[string]string{
				"error": "pending approval request not found or expired",
			})
			return
		}

		writeJSON(w, http.StatusInternalServerError, map[string]string{
			"error": "failed to approve request",
		})
		return
	}

	approval, err := h.Repository.GetByID(r.Context(), id)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{
			"error": "approval updated but could not be retrieved",
		})
		return
	}

	grant, err := h.GrantRepository.CreateFromApproval(
		r.Context(),
		approval,
	)

	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{
			"error": "approval succeeded but authorization grant creation failed",
		})
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"approval": approval,
		"grant":    grant,
	})
}

func (h *ApprovalHandler) Deny(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")

	err := h.Repository.Deny(r.Context(), id)
	if err != nil {
		if errors.Is(err, repository.ErrApprovalNotFound) {
			writeJSON(w, http.StatusNotFound, map[string]string{
				"error": "pending approval request not found",
			})
			return
		}

		writeJSON(w, http.StatusInternalServerError, map[string]string{
			"error": "failed to deny request",
		})
		return
	}

	approval, err := h.Repository.GetByID(r.Context(), id)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{
			"error": "approval updated but could not be retrieved",
		})
		return
	}

	writeJSON(w, http.StatusOK, approval)
}
