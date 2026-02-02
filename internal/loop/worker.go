package loop

import (
	"context"
	"fmt"
	"time"

	"github.com/project-atlas/atlas/internal/ai"
	"github.com/project-atlas/atlas/internal/cloud"
	"github.com/project-atlas/atlas/internal/engine"
	"github.com/project-atlas/atlas/internal/idempotency"
	"github.com/project-atlas/atlas/internal/logger"
	"github.com/project-atlas/atlas/internal/risk"
)

// IsIndieForceWindow checks if current time is in the Indie-Force shutdown window (12 AM - 6 AM)
func IsIndieForceWindow() bool {
	hour := time.Now().Hour()
	return hour >= 0 && hour < 6
}

type Worker struct {
	Provider    cloud.CloudAdapter
	RiskEngine  *risk.Engine
	IdempEngine *idempotency.Engine
	ArbEngine   *engine.ArbitrageEngine
	AIClient    *ai.UnifiedOrchestrator
	DryRun      bool
}

func NewWorker(p cloud.CloudAdapter, r *risk.Engine, i *idempotency.Engine, a *engine.ArbitrageEngine, aiClient *ai.UnifiedOrchestrator, dryRun bool) *Worker {
	return &Worker{
		Provider:    p,
		RiskEngine:  r,
		IdempEngine: i,
		ArbEngine:   a,
		AIClient:    aiClient,
		DryRun:      dryRun,
	}
}

func (w *Worker) RunCycle() error {
	logger.LogAction(logger.Architect, "LoopCycle", "STARTED", "Phase 1: Observation & Analysis")

	// 1. Observe: List resources
	resources, err := w.Provider.FetchResources(context.Background())
	if err != nil {
		return fmt.Errorf("observe failed: %w", err)
	}

	for _, res := range resources {
		// 2. Orient: Multi-Vector Analysis

		// Vector A: Standard Rightsizing
		projectedImpact := res.CostPerMonth * 0.25
		metrics := risk.CloudMetrics{
			CPUUsage:    res.CPUUsage,
			MemoryUsage: res.MemoryUsage,
			MeasuredAt:  time.Now(),
		}
		analysis := w.RiskEngine.CalculateScore(projectedImpact, metrics)

		// Vector B: Spot Arbitrage
		arbPlan, _ := w.ArbEngine.FindArbitrageOpportunity("us-east-1a", res.Type)

		// Vector C: Off-Peak Scheduling
		sched := &engine.Scheduler{} // Simple init
		schedPlan, _ := sched.GenerateSchedulePlan(res)

		// Vector D: Intelligent AI Analysis (For high-value targets)
		if res.CostPerMonth > 500 && w.AIClient != nil {
			logger.LogAction(logger.Architect, "AIAnalysis", "STARTED", fmt.Sprintf("Consulting AI for %s", res.ID))

			// Use the unified orchestrator instead of direct client methods
			prompt := fmt.Sprintf("Resource: %s, Cost: %.2f, CPU: %.1f, Memory: %.1f", res.ID, res.CostPerMonth, res.CPUUsage, res.MemoryUsage)
			response, err := w.AIClient.Analyze(context.Background(), prompt, analysis.Risk, res)

			if err == nil {
				logger.LogAction(logger.Architect, "AIAnalysis", "COMPLETED", response.Content)
			}
		}

		// 3. Decide & Act

		// market/usability: AI Explainability (The "Why")
		if (w.DryRun || analysis.Risk > 4.0) && w.AIClient != nil {
			prompt := fmt.Sprintf("Explain this optimization decision: Risk: %.1f, Cost: %.2f, Resource: %s", analysis.Risk, res.CostPerMonth, res.ID)
			explanation, _ := w.AIClient.Analyze(context.Background(), prompt, analysis.Risk, res)
			logger.LogAction(logger.Strategist, "Explainability", "AI-EXPLANATION", explanation.Content)
		}

		if w.DryRun {
			logger.LogAction(logger.Sentinel, "Simulation", "DRY-RUN", fmt.Sprintf("Would execute optimization for %s. Risk: %.1f", res.ID, analysis.Risk))
			continue
		}

		// Priority 1: High-savings Arbitrage
		if arbPlan != nil && arbPlan.RiskScore < 5.0 {
			logger.LogAction(logger.Auditor, "Decision", "MATCH", "Arbitrage opportunity found")
			w.IdempEngine.ExecuteGuarded(logger.Builder, "MigrateZone", res, func() (string, error) {
				_, err := w.Provider.ApplyOptimization(context.Background(), res, "spot-migrate")
				return "spot-migration completed", err
			})
		}

		// Priority 2: Scheduling (if peak/off-peak matches)
		if schedPlan != nil {
			w.IdempEngine.ExecuteGuarded(logger.Builder, "ScheduleStop", res, func() (string, error) {
				_, err := w.Provider.ApplyOptimization(context.Background(), res, "stopped")
				return "schedule-stop completed", err
			})
		}

		// Priority 3: Rightsizing
		if analysis.Score > 5.0 && res.CPUUsage < 40 && res.MemoryUsage < 50 {
			w.IdempEngine.ExecuteGuarded(logger.Builder, "Rightsize", res, func() (string, error) {
				_, err := w.Provider.ApplyOptimization(context.Background(), res, "resize")
				return "rightsize completed", err
			})
		}
	}

	logger.LogAction(logger.Architect, "LoopCycle", "COMPLETED", "Phase 4: Autonomous Verification (Delayed)")
	return nil
}
