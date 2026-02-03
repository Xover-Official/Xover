package governance

import (
	"context"
	"fmt"

	"github.com/Xover-Official/Xover/internal/learning"
)

// AdaptiveTrainer optimizes policies based on outcomes
type AdaptiveTrainer struct {
	policyEngine   *PolicyEngine
	learningEngine *learning.LearningEngine
}

func NewAdaptiveTrainer(pe *PolicyEngine, le *learning.LearningEngine) *AdaptiveTrainer {
	return &AdaptiveTrainer{
		policyEngine:   pe,
		learningEngine: le,
	}
}

// ProcessAuditOutcome feeds compliance results back into learning
func (t *AdaptiveTrainer) ProcessAuditOutcome(ctx context.Context, violation Violation, enforcementSuccess bool) {
	// If a policy violation occurred and auto-remediation failed or caused issues,
	// we need to adjust the "Trust" score of that policy.

	actionType := fmt.Sprintf("policy_enforcement:%s", violation.PolicyID)

	feedback := learning.FeedbackEvent{
		ActionID: actionType,
		Type:     learning.FeedbackApprove,
	}

	if !enforcementSuccess {
		feedback.Type = learning.FeedbackReject
		fmt.Printf("📉 ADAPTIVE: Penalizing policy %s due to failed enforcement.\n", violation.PolicyID)
	} else {
		fmt.Printf("📈 ADAPTIVE: Reinforcing policy %s (Successful remediation).\n", violation.PolicyID)
	}

	t.learningEngine.RecordFeedback(ctx, feedback)
}

// ProposePolicyAdjustments suggests changes to strictness
func (t *AdaptiveTrainer) ProposePolicyAdjustments(ctx context.Context) {
	// Logic to analyze recent violations
	// If "No Public S3" is violated 100 times/day, propose making it "CRITICAL" (blocking) instead of "HIGH" (alerting)

	fmt.Println("🛡️ ADAPTIVE: Analyzing policy effectiveness...")
	// Logic stub
}
