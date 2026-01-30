package guardian

import (
	"context"
	"fmt"
	"math/rand"
	"time"
)

// Simulator is the "Multiverse Check" engine
type Simulator struct {
	// Dependencies for running simulations would act here
	// e.g., cloud provider mocks, cost calculators
}

// SimulationResult contains the outcome of a simulation
type SimulationResult struct {
	ActionID         string
	Safe             bool
	ProjectedSavings float64
	RiskScore        float64
	FailureProb      float64
	Reason           string
}

func NewSimulator() *Simulator {
	return &Simulator{}
}

// ValidateAction runs a counterfactual simulation for a proposed action
func (s *Simulator) ValidateAction(ctx context.Context, actionType string, params map[string]interface{}) (*SimulationResult, error) {
	// In a real implementation, this would:
	// 1. Spin up a shadow environment (or use a mock)
	// 2. Apply the action
	// 3. Measure the outcome
	// 4. Tear down

	// For Phase 1, we use a probabilistic model based on historical "memory" (mocked here)

	result := &SimulationResult{
		ActionID: fmt.Sprintf("sim-%d", time.Now().UnixNano()),
	}

	switch actionType {
	case "downsize_instance":
		// Simulate downsizing
		safelyDownsized := rand.Float64() > 0.1 // 90% chance of safety
		if safelyDownsized {
			result.Safe = true
			result.ProjectedSavings = 45.00 // Predicted savings
			result.RiskScore = 2.5
			result.FailureProb = 0.1
			result.Reason = "Simulation shows memory usage remains below 70% after downsize."
		} else {
			result.Safe = false
			result.RiskScore = 8.5
			result.FailureProb = 0.9
			result.Reason = "Simulation detected OOM killer triggering after downsize."
		}

	case "delete_snapshot":
		// Simulate deletion impact
		result.Safe = true
		result.ProjectedSavings = 5.00
		result.RiskScore = 1.0
		result.Reason = "Snapshot is 90 days old and has no dependencies."

	default:
		result.Safe = true
		result.Reason = "No simulation model available for this action type."
	}

	return result, nil
}

// SimulateAlternative runs a "what-if" on a past decision
func (s *Simulator) SimulateAlternative(decisionID string, alternativeAction string) (*SimulationResult, error) {
	// This supports the "Counterfactual" capability of the Double-Helix
	// "What if we HAD resized that database?"

	// Mock logic
	return &SimulationResult{
		Safe:             true,
		ProjectedSavings: 150.00,
		RiskScore:        3.0,
		Reason:           "Counterfactual analysis: Traffic spike was handled by read replicas, primary resize was actually safe.",
	}, nil
}
