package ai

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/project-atlas/atlas/internal/engine"
	"github.com/project-atlas/atlas/internal/memory"
)

// ScalingEngine handles proactive scaling based on predictions
type ScalingEngine struct {
	oracle *engine.Oracle
	memory *memory.MemoryStore
}

func NewScalingEngine(oracle *engine.Oracle, mem *memory.MemoryStore) *ScalingEngine {
	return &ScalingEngine{
		oracle: oracle,
		memory: mem,
	}
}

// EvaluateScalingNeeds checks if resources need proactive scaling
func (e *ScalingEngine) EvaluateScalingNeeds(ctx context.Context, resourceID string) error {
	// 1. Get Forecast
	// Predict 1 hour ahead
	forecast, err := e.oracle.ForecastMetric(ctx, "cpu_usage:"+resourceID, 1*time.Hour)
	if err != nil {
		return err
	}

	// 2. Log Prediction for accuracy tracking
	err = e.memory.RecordPrediction(ctx, memory.Prediction{
		ResourceID:    resourceID,
		PredictedAt:   forecast.Timestamp,
		ForecastValue: forecast.Value,
		Confidence:    0.85, // Mocked confidence
		ModelUsed:     "linear_regression_v1",
	})
	if err != nil {
		log.Printf("Failed to record prediction: %v", err)
	}

	// 3. make Decision
	// If predicted load > 80%, scale UP now to be ready
	if forecast.Value > 80.0 {
		return e.triggerScaleUp(ctx, resourceID, "Proactive: Predicted load > 80% in 1h")
	}

	// If predicted load < 30% for 4 hours, scale DOWN
	// (Simplified logic)
	if forecast.Value < 30.0 {
		return e.triggerScaleDown(ctx, resourceID, "Proactive: Predicted load < 30% in 1h")
	}

	return nil
}

func (e *ScalingEngine) triggerScaleUp(_ context.Context, resourceID, reason string) error {
	fmt.Printf("🚀 PROACTIVE SCALE UP: %s. Reason: %s\n", resourceID, reason)
	// Call cloud adapter to resize
	return nil
}

func (e *ScalingEngine) triggerScaleDown(_ context.Context, resourceID, reason string) error {
	fmt.Printf("📉 PROACTIVE SCALE DOWN: %s. Reason: %s\n", resourceID, reason)
	return nil
}
