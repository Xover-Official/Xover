package loop

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/project-atlas/atlas/internal/ai"
	"github.com/project-atlas/atlas/internal/analytics"
	"github.com/project-atlas/atlas/internal/cloud"
	"github.com/project-atlas/atlas/internal/config"
	"github.com/project-atlas/atlas/internal/logger"
	"github.com/project-atlas/atlas/internal/persistence"
	"go.uber.org/zap"
)

// OODALoop implements the Observe-Orient-Decide-Act cycle
type OODALoop struct {
	config       *config.Config
	ledger       persistence.Ledger
	orchestrator *ai.UnifiedOrchestrator
	tokenTracker *analytics.TokenTracker
	logger       *zap.Logger
	stopChan     chan struct{}
}

// NewOODALoop creates a new OODA loop with zap logger
func NewOODALoop(cfg *config.Config, ledger persistence.Ledger, orchestrator *ai.UnifiedOrchestrator, tracker *analytics.TokenTracker, l *zap.Logger) *OODALoop {
	if l == nil {
		l = logger.GetLogger()
	}
	return &OODALoop{
		config:       cfg,
		ledger:       ledger,
		orchestrator: orchestrator,
		tokenTracker: tracker,
		logger:       l,
		stopChan:     make(chan struct{}),
	}
}

// Start begins the OODA loop
func (o *OODALoop) Start() error {
	o.logger.Info("🔄 OODA Loop started", zap.String("mode", o.config.Server.Mode))

	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	// Run immediately on start
	if err := o.runCycle(); err != nil {
		o.logger.Error("Initial cycle error", zap.Error(err))
	}

	for {
		select {
		case <-ticker.C:
			if err := o.runCycle(); err != nil {
				o.logger.Error("Cycle error", zap.Error(err))
			}
		case <-o.stopChan:
			o.logger.Info("🛑 OODA Loop stopped")
			return nil
		}
	}
}

// Stop halts the OODA loop
func (o *OODALoop) Stop() {
	close(o.stopChan)
}

// runCycle executes one complete OODA cycle
func (o *OODALoop) runCycle() error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()

	o.logger.Info("🔄 Starting new OODA loop cycle")

	// 1. OBSERVE: Discover cloud resources
	resources, err := o.observe(ctx)
	if err != nil {
		return fmt.Errorf("observe failed: %w", err)
	}
	o.logger.Info("👁️ OBSERVE complete", zap.Int("count", len(resources)))

	// 2. ORIENT: Analyze and calculate risk
	analyses := o.orient(ctx, resources)
	o.logger.Info("🧭 ORIENT complete", zap.Int("analyzed", len(analyses)))

	// 3. DECIDE: Use AI to determine optimizations
	decisions := o.decide(ctx, analyses)
	o.logger.Info("🤔 DECIDE complete", zap.Int("decisions", len(decisions)))

	// 4. ACT: Apply optimizations
	applied := o.act(ctx, decisions)
	o.logger.Info("⚡ ACT complete", zap.Int("applied", applied))

	// Print cycle summary
	stats := o.tokenTracker.GetBreakdown()
	o.logger.Info("✅ Cycle complete",
		zap.Float64("total_cost", stats["total_cost"].(float64)),
		zap.Float64("projected_savings", stats["projected_savings"].(float64)),
		zap.Float64("roi", stats["roi"].(float64)),
	)

	return nil
}

// observe discovers cloud resources
func (o *OODALoop) observe(ctx context.Context) ([]*cloud.ResourceV2, error) {
	// Placeholder - integrate with actual cloud adapters
	resources := []*cloud.ResourceV2{
		{
			ID:           "i-abc123",
			Type:         "ec2",
			Provider:     "aws",
			Region:       "us-east-1",
			State:        "running",
			CPUUsage:     15.5,
			MemoryUsage:  30.2,
			CostPerMonth: 73.00,
		},
		{
			ID:           "db-xyz789",
			Type:         "rds",
			Provider:     "aws",
			Region:       "us-east-1",
			State:        "available",
			CPUUsage:     45.0,
			MemoryUsage:  60.0,
			CostPerMonth: 180.00,
		},
	}

	return resources, nil
}

// orient analyzes resources and calculates risk
func (o *OODALoop) orient(ctx context.Context, resources []*cloud.ResourceV2) []ResourceAnalysis {
	analyses := make([]ResourceAnalysis, 0, len(resources))

	for _, resource := range resources {
		// Calculate risk score based on usage and cost
		riskScore := o.calculateRisk(resource)

		// Estimate potential savings
		savingsEstimate := o.estimateSavings(resource)

		analyses = append(analyses, ResourceAnalysis{
			Resource:          resource,
			RiskScore:         riskScore,
			SavingsEstimate:   savingsEstimate,
			RecommendedAction: o.suggestAction(resource),
		})
	}

	return analyses
}

// decide uses AI to make optimization decisions
func (o *OODALoop) decide(ctx context.Context, analyses []ResourceAnalysis) []Decision {
	decisions := make([]Decision, 0)
	var mu sync.Mutex
	var wg sync.WaitGroup

	// Concurrency control
	semaphore := make(chan struct{}, 10)

	for _, analysis := range analyses {
		wg.Add(1)
		go func(a ResourceAnalysis) {
			defer wg.Done()
			semaphore <- struct{}{}
			defer func() { <-semaphore }()

			prompt := fmt.Sprintf("Analyze resource [%s] for optimization. Risk: %.1f", a.Resource.ID, a.RiskScore)

			response, err := o.orchestrator.Analyze(ctx, prompt, a.RiskScore, a.Resource)
			if err != nil {
				o.logger.Warn("AI analysis failed", zap.String("resource_id", a.Resource.ID), zap.Error(err))
				return
			}

			mu.Lock()
			decisions = append(decisions, Decision{
				ResourceID:       a.Resource.ID,
				Action:           a.RecommendedAction,
				Reasoning:        response.Content,
				Confidence:       response.Confidence,
				RiskScore:        a.RiskScore,
				EstimatedSavings: a.SavingsEstimate,
				AIModel:          response.Model,
			})
			mu.Unlock()
		}(analysis)
	}

	wg.Wait()
	return decisions
}

// act applies the optimization decisions
func (o *OODALoop) act(ctx context.Context, decisions []Decision) int {
	applied := 0

	for _, decision := range decisions {
		// Skip if in dry-run mode
		if o.config.Cloud.DryRun {
			o.logger.Info("[DRY RUN] Optimization proposed",
				zap.String("action", decision.Action),
				zap.String("resource", decision.ResourceID),
				zap.Float64("savings", decision.EstimatedSavings))
			continue
		}

		// Skip low-confidence decisions
		if decision.Confidence < 0.8 {
			o.logger.Debug("Skipping low-confidence decision", zap.String("resource", decision.ResourceID))
			continue
		}

		// Record in ledger
		action := persistence.Action{
			ResourceID:       decision.ResourceID,
			ActionType:       decision.Action,
			Status:           "pending",
			RiskScore:        decision.RiskScore,
			EstimatedSavings: decision.EstimatedSavings,
			Payload:          map[string]interface{}{"reasoning": decision.Reasoning, "model": decision.AIModel},
		}

		if err := o.ledger.RecordAction(ctx, &action); err != nil {
			o.logger.Error("Failed to record action", zap.Error(err))
			continue
		}

		o.logger.Info("Applied optimization",
			zap.String("action", decision.Action),
			zap.String("resource", decision.ResourceID))

		applied++
	}

	return applied
}

// calculateRisk computes a risk score (0-10) for a resource
func (o *OODALoop) calculateRisk(r *cloud.ResourceV2) float64 {
	// Simple heuristic - in production, use more sophisticated logic
	risk := 0.0

	// High cost increases risk
	if r.CostPerMonth > 200 {
		risk += 2.0
	} else if r.CostPerMonth > 100 {
		risk += 1.0
	}

	// Low usage increases risk of optimization
	if r.CPUUsage < 20 {
		risk += 3.0
	} else if r.CPUUsage < 40 {
		risk += 1.5
	}

	// Database resources are higher risk
	if r.Type == "rds" || r.Type == "database" {
		risk += 2.0
	}

	return min(risk, 10.0)
}

// estimateSavings estimates monthly savings from optimization
func (o *OODALoop) estimateSavings(r *cloud.ResourceV2) float64 {
	// Simple estimation - rightsize by ~40% for underutilized resources
	if r.CPUUsage < 30 && r.MemoryUsage < 40 {
		return r.CostPerMonth * 0.4
	}
	if r.CPUUsage < 50 {
		return r.CostPerMonth * 0.25
	}
	return 0
}

// suggestAction suggests an optimization action
func (o *OODALoop) suggestAction(r *cloud.ResourceV2) string {
	if r.CPUUsage < 20 && r.MemoryUsage < 30 {
		return "rightsize_smaller"
	}
	if r.State == "stopped" {
		return "delete_if_unused"
	}
	return "monitor"
}

// Helper types

type ResourceAnalysis struct {
	Resource          *cloud.ResourceV2
	RiskScore         float64
	SavingsEstimate   float64
	RecommendedAction string
}

type Decision struct {
	ResourceID       string
	Action           string
	Reasoning        string
	Confidence       float64
	RiskScore        float64
	EstimatedSavings float64
	AIModel          string
}

func min(a, b float64) float64 {
	if a < b {
		return a
	}
	return b
}

func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}
