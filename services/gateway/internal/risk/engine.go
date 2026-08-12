package risk

import "strings"

type Result struct {
	Score  int
	Reason string
}

// Evaluate calculates a deterministic risk score for an action.
// This is intentionally rule-based for the first version.
// Later we will replace/augment this with behavioral ML.
func Evaluate(action string, resource string) Result {
	action = strings.ToLower(action)
	resource = strings.ToLower(resource)

	score := 10
	reason := "low-risk operation"

	// Destructive actions
	if strings.Contains(action, "delete") ||
		strings.Contains(action, "drop") ||
		strings.Contains(action, "destroy") {
		score += 55
		reason = "destructive operation"
	}

	// Secret or credential access
	if strings.Contains(action, "secret") ||
		strings.Contains(action, "credential") ||
		strings.Contains(action, "token") {
		score += 45
		reason = "sensitive credential access"
	}

	// IAM / privilege modifications
	if strings.Contains(action, "iam") ||
		strings.Contains(action, "admin") ||
		strings.Contains(action, "permission") {
		score += 50
		reason = "privilege or identity modification"
	}

	// Production resources increase risk
	if strings.Contains(resource, "production") ||
		strings.Contains(resource, "prod") {
		score += 25
	}

	// Code changes and deployments carry additional risk.
	if strings.Contains(action, "push") ||
		strings.Contains(action, "deploy") ||
		strings.Contains(action, "merge") {
		score += 20
		reason = "code or deployment modification"
	}

	return Result{
		Score:  score,
		Reason: reason,
	}
}
