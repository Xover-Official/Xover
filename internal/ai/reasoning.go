package ai

import (
	"context"
	"fmt"
	"strings"
)

// ReasoningEngine explains the "Why" behind AI decisions
type ReasoningEngine struct {
	// dependencies like cost calculator, risk model
}

// Explanation provides the rationale for a decision
type Explanation struct {
	DecisionID   string   `json:"decision_id"`
	Action       string   `json:"action"`
	Outcome      string   `json:"outcome"`   // Proposed outcome (e.g., Save $50/mo)
	Reasoning    []string `json:"reasoning"` // Bullet points
	Tradeoffs    []string `json:"tradeoffs"` // Cons
	Confidence   float64  `json:"confidence"`
	Alternatives []string `json:"alternatives"` // What else was considered
}

func NewReasoningEngine() *ReasoningEngine {
	return &ReasoningEngine{}
}

// ExplainDecision generates a structured explanation for an action
func (e *ReasoningEngine) ExplainDecision(ctx context.Context, actionType string, params map[string]interface{}) *Explanation {
	explanation := &Explanation{
		Action: params["resource_id"].(string), // Simplified
	}

	switch actionType {
	case "resize_instance":
		explanation.Outcome = "Reduce monthly cost by $45 (30% savings)"
		explanation.Reasoning = []string{
			"CPU utilization has been < 10% for the last 14 days.",
			"Memory usage peaks at 4GB, but current instance has 16GB.",
			"No traffic spikes correlated with business hours.",
		}
		explanation.Tradeoffs = []string{
			"Downsizing requires a reboot (approx. 2 mins downtime).",
			"Burst capacity will be reduced by 50%.",
		}
		explanation.Alternatives = []string{
			"Convert to Spot Instance (Riskier, 60% savings)",
			"Buy Reserved Instance (Commitment required, 40% savings)",
		}
		explanation.Confidence = 0.92

	case "block_ip":
		explanation.Outcome = "Prevent potential brute-force attack"
		explanation.Reasoning = []string{
			"IP 192.168.1.1 attempted 500 logins in 1 minute.",
			"Geo-location resolves to a high-risk region (North Korea).",
			"Pattern matches known botnet signature 'Mirai-Variant'.",
		}
		explanation.Tradeoffs = []string{
			"False positive could block a legitimate user behind a VPN.",
		}
		explanation.Confidence = 0.99

	default:
		explanation.Reasoning = []string{"Automated decision based on standard heuristics."}
		explanation.Confidence = 0.5
	}

	return explanation
}

// GenerateNaturalLanguageSummary creates a readable summary
func (e *ReasoningEngine) GenerateNaturalLanguageSummary(expl *Explanation) string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("I recommend **%s** to **%s**.\n\n", expl.Action, expl.Outcome))

	sb.WriteString("**Why?**\n")
	for _, r := range expl.Reasoning {
		sb.WriteString(fmt.Sprintf("- %s\n", r))
	}

	if len(expl.Tradeoffs) > 0 {
		sb.WriteString("\n**Risks:**\n")
		for _, t := range expl.Tradeoffs {
			sb.WriteString(fmt.Sprintf("- %s\n", t))
		}
	}

	return sb.String()
}
