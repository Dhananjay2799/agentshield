package handlers

import (
	"errors"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/dhananjay2799/agentshield/services/gateway/internal/events"
	"github.com/dhananjay2799/agentshield/services/gateway/internal/models"
	"github.com/dhananjay2799/agentshield/services/gateway/internal/opa"
	"github.com/dhananjay2799/agentshield/services/gateway/internal/policy"
	"github.com/dhananjay2799/agentshield/services/gateway/internal/repository"
)

type PolicyHandler struct {
	Repository      *repository.PolicyRepository
	AuditRepository *repository.AuditRepository
	OPAClient       *opa.Client
	EventProducer   *events.Producer
}

func NewPolicyHandler(
	policyRepository *repository.PolicyRepository,
	auditRepository *repository.AuditRepository,
	opaClient *opa.Client,
	eventProducer *events.Producer,
) *PolicyHandler {
	return &PolicyHandler{
		Repository:      policyRepository,
		AuditRepository: auditRepository,
		OPAClient:       opaClient,
		EventProducer:   eventProducer,
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

	if !decodeJSONBody(
		w,
		r,
		&req,
	) {
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
	_ = h.AuditRepository.Create(
		r.Context(),
		repository.CreateAuditEventParams{
			EventType: "policy.created",
			Action:    "policy.create",
			Resource:  "policy/" + policy.ID,
			Decision:  "SUCCESS",
			RiskScore: 0,
			Metadata: map[string]any{
				"policy_id":   policy.ID,
				"policy_name": policy.Name,
				"effect":      policy.Effect,
				"status":      policy.Status,
				"version":     policy.Version,
				"created_by":  policy.CreatedBy,
				"source":      policy.Source,
			},
		},
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

func (h *PolicyHandler) Activate(
	w http.ResponseWriter,
	r *http.Request,
) {
	ctx := r.Context()
	id := r.PathValue("id")

	storedPolicy, err :=
		h.Repository.GetByID(
			ctx,
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

	if storedPolicy.Status != "draft" &&
		storedPolicy.Status != "disabled" {

		writeJSON(
			w,
			http.StatusConflict,
			map[string]string{
				"error": "only draft or disabled policies can be activated",
			},
		)
		return
	}

	validation :=
		policy.Validate(
			storedPolicy,
		)

	if !validation.Valid {
		writeJSON(
			w,
			http.StatusUnprocessableEntity,
			validation,
		)
		return
	}

	// Synchronize with OPA first.
	if err :=
		h.OPAClient.PutManagedPolicy(
			ctx,
			storedPolicy,
		); err != nil {

		writeJSON(
			w,
			http.StatusBadGateway,
			map[string]string{
				"error":  "failed to synchronize policy with OPA",
				"detail": err.Error(),
			},
		)
		return
	}

	// Update PostgreSQL.
	activePolicy, err :=
		h.Repository.Activate(
			ctx,
			id,
		)

	if err != nil {
		// PostgreSQL failed after OPA synchronization.
		// Remove the policy from OPA so runtime and
		// control-plane state remain consistent.
		rollbackErr :=
			h.OPAClient.DeleteManagedPolicy(
				ctx,
				id,
			)

		if rollbackErr != nil {
			writeJSON(
				w,
				http.StatusInternalServerError,
				map[string]string{
					"error":  "database activation failed and OPA rollback also failed",
					"detail": rollbackErr.Error(),
				},
			)
			return
		}

		writeJSON(
			w,
			http.StatusInternalServerError,
			map[string]string{
				"error": "failed to activate policy; OPA synchronization was rolled back",
			},
		)
		return
	}

	metadata := map[string]any{
		"policy_id":   activePolicy.ID,
		"policy_name": activePolicy.Name,
		"effect":      activePolicy.Effect,
		"status":      activePolicy.Status,
		"version":     activePolicy.Version,
		"created_by":  activePolicy.CreatedBy,
		"source":      activePolicy.Source,
		"opa_sync":    "synchronized",
	}

	// Persist the lifecycle event to the audit timeline.
	if err := h.AuditRepository.Create(
		ctx,
		repository.CreateAuditEventParams{
			EventType: "policy.activated",
			Action:    "policy.activate",
			Resource:  "policy/" + activePolicy.ID,
			Decision:  "SUCCESS",
			RiskScore: 0,
			Metadata:  metadata,
		},
	); err != nil {
		log.Printf(
			"failed to persist policy activation audit event: %v",
			err,
		)
	}

	// Publish through the same Kafka + SSE pipeline used
	// by autonomous-agent action events.
	if h.EventProducer != nil {
		event := events.SecurityEvent{
			EventType:  "policy.activated",
			AgentID:    "",
			SessionID:  "",
			Action:     "policy.activate",
			Resource:   "policy/" + activePolicy.ID,
			Decision:   "SUCCESS",
			RiskScore:  0,
			Metadata:   metadata,
			OccurredAt: time.Now().UTC(),
		}

		if err :=
			h.EventProducer.Publish(
				ctx,
				event,
			); err != nil {

			log.Printf(
				"failed to publish policy activation event: %v",
				err,
			)
		}
	}

	writeJSON(
		w,
		http.StatusOK,
		map[string]any{
			"policy": activePolicy,
			"sync": map[string]string{
				"opa": "synchronized",
			},
		},
	)
}

func (h *PolicyHandler) Deactivate(
	w http.ResponseWriter,
	r *http.Request,
) {
	ctx := r.Context()
	id := r.PathValue("id")

	storedPolicy, err :=
		h.Repository.GetByID(
			ctx,
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

	if storedPolicy.Status != "active" {
		writeJSON(
			w,
			http.StatusConflict,
			map[string]string{
				"error": "only active policies can be deactivated",
			},
		)
		return
	}

	// Remove the runtime policy from OPA first.
	if err :=
		h.OPAClient.DeleteManagedPolicy(
			ctx,
			id,
		); err != nil {

		writeJSON(
			w,
			http.StatusBadGateway,
			map[string]string{
				"error":  "failed to remove policy from OPA",
				"detail": err.Error(),
			},
		)
		return
	}

	// Update PostgreSQL.
	inactivePolicy, err :=
		h.Repository.Deactivate(
			ctx,
			id,
		)

	if err != nil {
		// PostgreSQL failed after OPA deletion.
		// Restore the policy because PostgreSQL still
		// considers it active.
		rollbackErr :=
			h.OPAClient.PutManagedPolicy(
				ctx,
				storedPolicy,
			)

		if rollbackErr != nil {
			writeJSON(
				w,
				http.StatusInternalServerError,
				map[string]string{
					"error":  "database deactivation failed and OPA rollback also failed",
					"detail": rollbackErr.Error(),
				},
			)
			return
		}

		writeJSON(
			w,
			http.StatusInternalServerError,
			map[string]string{
				"error": "failed to deactivate policy; OPA policy was restored",
			},
		)
		return
	}

	metadata := map[string]any{
		"policy_id":   inactivePolicy.ID,
		"policy_name": inactivePolicy.Name,
		"effect":      inactivePolicy.Effect,
		"status":      inactivePolicy.Status,
		"version":     inactivePolicy.Version,
		"created_by":  inactivePolicy.CreatedBy,
		"source":      inactivePolicy.Source,
		"opa_sync":    "removed",
	}

	// Persist lifecycle event.
	if err := h.AuditRepository.Create(
		ctx,
		repository.CreateAuditEventParams{
			EventType: "policy.deactivated",
			Action:    "policy.deactivate",
			Resource:  "policy/" + inactivePolicy.ID,
			Decision:  "SUCCESS",
			RiskScore: 0,
			Metadata:  metadata,
		},
	); err != nil {
		log.Printf(
			"failed to persist policy deactivation audit event: %v",
			err,
		)
	}

	// Publish to Kafka and the live SSE hub.
	if h.EventProducer != nil {
		event := events.SecurityEvent{
			EventType:  "policy.deactivated",
			AgentID:    "",
			SessionID:  "",
			Action:     "policy.deactivate",
			Resource:   "policy/" + inactivePolicy.ID,
			Decision:   "SUCCESS",
			RiskScore:  0,
			Metadata:   metadata,
			OccurredAt: time.Now().UTC(),
		}

		if err :=
			h.EventProducer.Publish(
				ctx,
				event,
			); err != nil {

			log.Printf(
				"failed to publish policy deactivation event: %v",
				err,
			)
		}
	}

	writeJSON(
		w,
		http.StatusOK,
		map[string]any{
			"policy": inactivePolicy,
			"sync": map[string]string{
				"opa": "removed",
			},
		},
	)
}
