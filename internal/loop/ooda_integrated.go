package loop

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/project-atlas/atlas/internal/ai"
	"github.com/project-atlas/atlas/internal/analytics"
	"github.com/project-atlas/atlas/internal/cloud"
	"github.com/project-atlas/atlas/internal/config"
	"github.com/project-atlas/atlas/internal/persistence"
)

// OODALoop implements the Observe-Orient-Decide-Act cycle
type OODALoop struct {
	config       *config.Config
	ledger       persistence.Ledger
	orchestrator *ai.UnifiedOrchestrator
	tokenTracker *analytics.TokenTracker
	stopChan     chan struct{}
}

// NewOODALoop creates a new OODA loop
func NewOODALoop(cfg *config.Config, ledger persistence.Ledger, orchestrator *ai.UnifiedOrchestrator, tracker *analytics.TokenTracker) *OODALoop {
	return &OODALoop{
		config:       cfg,
		ledger:       ledger,
		orchestrator: orchestrator,
		tokenTracker: tracker,
		stopChan:     make(chan struct{}),
	}
}

// Start begins the OODA loop
func (o *OODALoop) Start() error {
	log.Println("🔄 OODA Loop started")

	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	// Run immediately on start
	if err := o.runCycle(); err != nil {
		log.Printf("Initial cycle error: %v", err)
	}

	for {
		select {
		case <-ticker.C:
			if err := o.runCycle(); err != nil {
				log.Printf("Cycle error: %v", err)
			}
		case <-o.stopChan:
			log.Println("🛑 OODA Loop stopped")
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
	ctx := context.Background()

	log.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	log.Println("🔄 Starting new OODA cycle")
	log.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")

	// 1. OBSERVE: Discover cloud resources
	log.Println("👁️  OBSERVE: Discovering resources...")
	resources, err := o.observe(ctx)
	if err != nil {
		return fmt.Errorf("observe failed: %w", err)
	}
	log.Printf("   Found %d resources\n", len(resources))

	// 2. ORIENT: Analyze and calculate risk
	log.Println("🧭 ORIENT: Analyzing resources...")
	analyses := o.orient(ctx, resources)
	log.Printf("   Analyzed %d resources\n", len(analyses))

	// 3. DECIDE: Use AI to determine optimizations
	log.Println("🤔 DECIDE: Consulting AI swarm...")
	decisions := o.decide(ctx, analyses)
	log.Printf("   Made %d optimization decisions\n", len(decisions))

	// 4. ACT: Apply optimizations
	log.Println("⚡ ACT: Applying optimizations...")
	applied := o.act(ctx, decisions)
	log.Printf("   Applied %d optimizations\n", applied)

	// Print cycle summary
	stats := o.tokenTracker.GetBreakdown()
	log.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	log.Printf("✅ Cycle complete - Cost: $%.4f, Savings: $%.2f, ROI: %.1fx\n",
		stats["total_cost"], stats["projected_savings"], stats["roi"])
	log.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n")

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

	for _, analysis := range analyses {
		// Build prompt for AI
		prompt := fmt.Sprintf(`Analyze this cloud resource:
Resource: %s (%s)
Type: %s
CPU Usage: %.1f%%
Memory Usage: %.1f%%
Current Cost: $%.2f/month
Risk Score: %.1f/10

Should we optimize this resource? If yes, what specific action should we take?
Provide a concise recommendation.`,
			analysis.Resource.ID,
			analysis.Resource.Provider,
			analysis.Resource.Type,
			analysis.Resource.CPUUsage,
			analysis.Resource.MemoryUsage,
			analysis.Resource.CostPerMonth,
			analysis.RiskScore,
		)

		// Call AI orchestrator
		response, err := o.orchestrator.Analyze(
			ctx,
			prompt,
			analysis.RiskScore,
			analysis.Resource,
		)

		if err != nil {
			log.Printf("   ⚠️  AI analysis failed for %s: %v", analysis.Resource.ID, err)
			continue
		}

		log.Printf("   🤖 Tier %s recommendation: %s",
			response.Model,
			truncate(response.Content, 80))

		// Create decision
		decision := Decision{
			ResourceID:       analysis.Resource.ID,
			Action:           analysis.RecommendedAction,
			Reasoning:        response.Content,
			Confidence:       response.Confidence,
			RiskScore:        analysis.RiskScore,
			EstimatedSavings: analysis.SavingsEstimate,
			AIModel:          response.Model,
		}

		decisions = append(decisions, decision)
	}

	return decisions
}

// act applies the optimization decisions
func (o *OODALoop) act(ctx context.Context, decisions []Decision) int {
	applied := 0

	for _, decision := range decisions {
		// Skip if in dry-run mode
		if o.config.Cloud.DryRun {
			log.Printf("   [DRY RUN] Would apply: %s to %s (saves $%.2f/mo)",
				decision.Action, decision.ResourceID, decision.EstimatedSavings)
			continue
		}

		// Skip low-confidence decisions
		if decision.Confidence < 0.8 {
			log.Printf("   ⏭️  Skipping %s (confidence %.2f < 0.8)",
				decision.ResourceID, decision.Confidence)
			continue
		}

		// Record in ledger
		action := persistence.Action{
			ResourceID:       decision.ResourceID,
			ActionType:       decision.Action,
			Status:           "pending",
			RiskScore:        decision.RiskScore,
			EstimatedSavings: decision.EstimatedSavings,
			Reasoning:        decision.Reasoning,
		}

		if err := o.ledger.RecordAction(action); err != nil {
			log.Printf("   ❌ Failed to record action: %v", err)
			continue
		}

		log.Printf("   ✅ Applied %s to %s (saves $%.2f/mo)",
			decision.Action, decision.ResourceID, decision.EstimatedSavings)

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
