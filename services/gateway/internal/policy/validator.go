package policy

import (
	"fmt"
	"strings"

	"github.com/dhananjay2799/agentshield/services/gateway/internal/models"
)

func Validate(
	policy *models.Policy,
) models.PolicyValidationResponse {

	response := models.PolicyValidationResponse{
		Valid:    true,
		PolicyID: policy.ID,
		Status:   policy.Status,
	}

	// ------------------------------------------------------------
	// Schema validation
	// ------------------------------------------------------------

	schemaErrors := make([]string, 0)

	if strings.TrimSpace(policy.Name) == "" {
		schemaErrors = append(
			schemaErrors,
			"name is required",
		)
	}

	switch policy.Effect {
	case "ALLOW",
		"REQUIRE_APPROVAL",
		"DENY":
	default:
		schemaErrors = append(
			schemaErrors,
			"unsupported policy effect",
		)
	}

	if strings.TrimSpace(policy.Action) == "" {
		schemaErrors = append(
			schemaErrors,
			"action is required",
		)
	}

	if strings.TrimSpace(policy.Resource) == "" {
		schemaErrors = append(
			schemaErrors,
			"resource is required",
		)
	}

	if policy.Priority <= 0 {
		schemaErrors = append(
			schemaErrors,
			"priority must be greater than zero",
		)
	}

	if policy.Version <= 0 {
		schemaErrors = append(
			schemaErrors,
			"version must be greater than zero",
		)
	}

	if len(schemaErrors) > 0 {
		response.Valid = false

		response.Checks.Schema = models.PolicyValidationCheck{
			Status:  "failed",
			Message: strings.Join(schemaErrors, "; "),
		}
	} else {
		response.Checks.Schema = models.PolicyValidationCheck{
			Status:  "passed",
			Message: "policy schema is valid",
		}
	}

	// ------------------------------------------------------------
	// Matcher validation
	// ------------------------------------------------------------

	matcherErrors := make([]string, 0)

	if !validMatcher(policy.ActionMatch) {
		matcherErrors = append(
			matcherErrors,
			fmt.Sprintf(
				"unsupported action matcher %q",
				policy.ActionMatch,
			),
		)
	}

	if !validMatcher(policy.ResourceMatch) {
		matcherErrors = append(
			matcherErrors,
			fmt.Sprintf(
				"unsupported resource matcher %q",
				policy.ResourceMatch,
			),
		)
	}

	if len(matcherErrors) > 0 {
		response.Valid = false

		response.Checks.Matchers = models.PolicyValidationCheck{
			Status:  "failed",
			Message: strings.Join(matcherErrors, "; "),
		}
	} else {
		response.Checks.Matchers = models.PolicyValidationCheck{
			Status:  "passed",
			Message: "action and resource matchers are supported",
		}
	}

	// ------------------------------------------------------------
	// OPA compatibility
	//
	// These checks correspond to the fields currently sent to OPA
	// by ActionHandler:
	//
	// agent_type
	// action
	// resource
	// environment
	//
	// A nil agent type/environment means the rule applies
	// regardless of that attribute.
	// ------------------------------------------------------------

	opaErrors := make([]string, 0)

	if strings.ContainsAny(
		policy.Action,
		"\r\n\t",
	) {
		opaErrors = append(
			opaErrors,
			"action contains unsupported control characters",
		)
	}

	if strings.ContainsAny(
		policy.Resource,
		"\r\n\t",
	) {
		opaErrors = append(
			opaErrors,
			"resource contains unsupported control characters",
		)
	}

	if policy.AgentType != nil &&
		strings.TrimSpace(*policy.AgentType) == "" {

		opaErrors = append(
			opaErrors,
			"agent_type cannot be empty when supplied",
		)
	}

	if policy.Environment != nil &&
		strings.TrimSpace(*policy.Environment) == "" {

		opaErrors = append(
			opaErrors,
			"environment cannot be empty when supplied",
		)
	}

	if len(opaErrors) > 0 {
		response.Valid = false

		response.Checks.OPACompatibility =
			models.PolicyValidationCheck{
				Status:  "failed",
				Message: strings.Join(opaErrors, "; "),
			}
	} else {
		response.Checks.OPACompatibility =
			models.PolicyValidationCheck{
				Status:  "passed",
				Message: "policy fields are compatible with the current AgentShield OPA input model",
			}
	}

	return response
}

func validMatcher(
	value string,
) bool {
	switch value {
	case "exact",
		"prefix",
		"suffix":
		return true

	default:
		return false
	}
}
