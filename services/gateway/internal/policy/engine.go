package policy

type Decision string

const (
	DecisionAllow           Decision = "ALLOW"
	DecisionRequireApproval Decision = "REQUIRE_APPROVAL"
	DecisionDeny            Decision = "DENY"
)

type Result struct {
	Decision Decision
	Reason   string
}

func Evaluate(riskScore int) Result {
	switch {
	case riskScore >= 70:
		return Result{
			Decision: DecisionDeny,
			Reason:   "risk score exceeds deny threshold",
		}

	case riskScore >= 40:
		return Result{
			Decision: DecisionRequireApproval,
			Reason:   "risk score requires human approval",
		}

	default:
		return Result{
			Decision: DecisionAllow,
			Reason:   "risk score is within allowed threshold",
		}
	}
}
