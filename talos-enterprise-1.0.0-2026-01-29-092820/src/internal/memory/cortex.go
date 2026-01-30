package memory

import (
	"context"
	"fmt"
	"time"
)

// BrainState represents a snapshot of the AI's heuristic weights and logic
type BrainState struct {
	ID        string             `json:"id"` // v1.0.0, v1.0.1-canary
	Timestamp time.Time          `json:"timestamp"`
	Weights   map[string]float64 `json:"weights"`   // Action decision weights
	Rules     []RuleSignature    `json:"rules"`     // Active policies/heuristics
	ParentID  string             `json:"parent_id"` // Lineage tracking
	Metrics   PerformanceMetrics `json:"metrics"`   // Stats for this state
}

type RuleSignature struct {
	ID   string `json:"id"`
	Hash string `json:"hash"` // Content hash
}

type PerformanceMetrics struct {
	TotalActions int     `json:"total_actions"`
	SuccessRate  float64 `json:"success_rate"`
	AvgSavings   float64 `json:"avg_savings"`
}

// Cortex manages the AI's state lineage
type Cortex struct {
	CurrentState *BrainState
	History      []*BrainState
}

func NewCortex() *Cortex {
	return &Cortex{
		CurrentState: &BrainState{
			ID:        "v0.0.1-genesis",
			Timestamp: time.Now(),
			Weights:   make(map[string]float64),
		},
	}
}

// Snapshot creates a new version of the brain
func (c *Cortex) Snapshot(ctx context.Context, reason string) *BrainState {
	newState := &BrainState{
		ID:        fmt.Sprintf("v%d-evo", time.Now().Unix()),
		Timestamp: time.Now(),
		ParentID:  c.CurrentState.ID,
		Weights:   copyMap(c.CurrentState.Weights),
		// Rules would be deep copied here
	}

	c.History = append(c.History, c.CurrentState)
	c.CurrentState = newState

	fmt.Printf("🧠 BRAIN EVOLUTION: Preserved %s, Evolved to %s (%s)\n", c.CurrentState.ParentID, newState.ID, reason)
	return newState
}

// Rollback reverts to a previous stable state
func (c *Cortex) Rollback(ctx context.Context, targetID string) error {
	for _, state := range c.History {
		if state.ID == targetID {
			c.CurrentState = state
			fmt.Printf("⏪ REFLEX REVERT: Rolled back to brain state %s\n", targetID)
			return nil
		}
	}
	return fmt.Errorf("state %s not found in lineage", targetID)
}

func copyMap(original map[string]float64) map[string]float64 {
	dest := make(map[string]float64)
	for k, v := range original {
		dest[k] = v
	}
	return dest
}
