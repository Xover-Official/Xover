// Copyright (c) 2026 Project Atlas (Talos)
// Licensed under the MIT License. See LICENSE in the project root for license information.

package engine

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/Xover-Official/Xover/internal/ai"
	"github.com/Xover-Official/Xover/internal/cloud"
	"github.com/Xover-Official/Xover/internal/database"
	"github.com/Xover-Official/Xover/internal/security"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
)

// OODAState represents the state in the OODA loop
type OODAState int

const (
	StateObserve OODAState = iota
	StateOrient
	StateDecide
	StateAct
	StateComplete
	StateFailed
)

func (s OODAState) String() string {
	return [...]string{"OBSERVE", "ORIENT", "DECIDE", "ACT", "COMPLETE", "FAILED"}[s]
}

// OptimizationOpportunity represents a potential optimization
type OptimizationOpportunity struct {
	Resource         *cloud.ResourceV2
	AnalysisVectors  []AnalysisVector
	RiskScore        float64
	Recommendations  []string
	EstimatedSavings float64
	Confidence       float64
}

// AnalysisVector represents a dimension of analysis
type AnalysisVector struct {
	Name       string
	Score      float64
	Weight     float64
	Findings   []string
	Confidence float64
}

// Repository defines the interface for data persistence required by the engine
type Repository interface {
	CreateAction(ctx context.Context, action *database.Action) error
	UpdateActionStatus(ctx context.Context, id string, status string, startedAt *time.Time, completedAt *time.Time, errorMsg *string) error
	CreateSavingsEvent(ctx context.Context, event *database.SavingsEvent) error
}

// OODAEngine implements the OODA loop for cloud optimization
type OODAEngine struct {
	aiOrchestrator *ai.UnifiedOrchestrator
	cloudAdapter   cloud.CloudAdapter
	repository     Repository
	securityCtrl   *security.SecurityController
	logger         *zap.Logger
	tracer         trace.Tracer
	config         *EngineConfig
}

// EngineConfig holds configuration for the OODA engine
type EngineConfig struct {
	MaxConcurrentCycles   int           `yaml:"max_concurrent_cycles"`
	MaxConcurrentAnalysis int           `yaml:"max_concurrent_analysis"`
	CycleInterval         time.Duration `yaml:"cycle_interval"`
	RiskThreshold         float64       `yaml:"risk_threshold"`
	MinSavingsThreshold   float64       `yaml:"min_savings_threshold"`
	MaxAnalysisTime       time.Duration `yaml:"max_analysis_time"`
	EnableAutoExecution   bool          `yaml:"enable_auto_execution"`
	RequireHumanApproval  bool          `yaml:"require_human_approval"`
	DefaultSavingsRatio   float64       `yaml:"default_savings_ratio"`
}

// NewOODAEngine creates a new OODA engine
func NewOODAEngine(
	aiOrchestrator *ai.UnifiedOrchestrator,
	cloudAdapter cloud.CloudAdapter,
	repository Repository,
	securityCtrl *security.SecurityController,
	logger *zap.Logger,
	tracer trace.Tracer,
	config *EngineConfig,
) *OODAEngine {
	return &OODAEngine{
		aiOrchestrator: aiOrchestrator,
		cloudAdapter:   cloudAdapter,
		repository:     repository,
		securityCtrl:   securityCtrl,
		logger:         logger,
		tracer:         tracer,
		config:         config,
	}
}

// RunCycle executes a complete OODA cycle
func (e *OODAEngine) RunCycle(ctx context.Context) error {
	ctx, span := e.tracer.Start(ctx, "ooda.cycle")
	defer span.End()

	e.logger.Info("Starting OODA cycle")

	// OBSERVE: Scan cloud resources
	resources, err := e.observe(ctx)
	if err != nil {
		span.RecordError(err)
		return fmt.Errorf("observe phase failed: %w", err)
	}

	// ORIENT: Multi-vector analysis
	opportunities, err := e.orient(ctx, resources)
	if err != nil {
		span.RecordError(err)
		return fmt.Errorf("orient phase failed: %w", err)
	}

	// DECIDE: Risk assessment and prioritization
	decisions, err := e.decide(ctx, opportunities)
	if err != nil {
		span.RecordError(err)
		return fmt.Errorf("decide phase failed: %w", err)
	}

	// ACT: Execute optimizations
	results, err := e.act(ctx, decisions)
	if err != nil {
		span.RecordError(err)
		return fmt.Errorf("act phase failed: %w", err)
	}

	e.logger.Info("OODA cycle completed",
		zap.Int("resources_scanned", len(resources)),
		zap.Int("opportunities_found", len(opportunities)),
		zap.Int("decisions_made", len(decisions)),
		zap.Int("actions_executed", len(results)),
	)

	return nil
}

// observe scans and collects cloud resources
func (e *OODAEngine) observe(ctx context.Context) ([]*cloud.ResourceV2, error) {
	ctx, span := e.tracer.Start(ctx, "ooda.observe")
	defer span.End()

	e.logger.Info("Observing cloud resources")

	resources, err := e.cloudAdapter.FetchResources(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch resources: %w", err)
	}

	e.logger.Info("Successfully observed resources", zap.Int("count", len(resources)))
	return resources, nil
}

// orient performs multi-vector analysis on resources concurrently
func (e *OODAEngine) orient(ctx context.Context, resources []*cloud.ResourceV2) ([]*OptimizationOpportunity, error) {
	ctx, span := e.tracer.Start(ctx, "ooda.orient")
	defer span.End()

	e.logger.Info("Orienting - performing concurrent multi-vector analysis", zap.Int("resource_count", len(resources)))

	type result struct {
		opp *OptimizationOpportunity
		err error
	}

	resChan := make(chan result, len(resources))
	workerCount := e.config.MaxConcurrentAnalysis
	if workerCount <= 0 {
		workerCount = 10 // Default safe fallback
	}
	if len(resources) < workerCount {
		workerCount = len(resources)
	}

	jobs := make(chan *cloud.ResourceV2, len(resources))
	var wg sync.WaitGroup

	// Start workers
	for i := 0; i < workerCount; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for r := range jobs {
				select {
				case <-ctx.Done():
					return
				default:
					opp, err := e.analyzeResource(ctx, r)
					resChan <- result{opp, err}
				}
			}
		}()
	}

	// Feed jobs
	go func() {
		for _, r := range resources {
			jobs <- r
		}
		close(jobs)
	}()

	// Wait and close results
	go func() {
		wg.Wait()
		close(resChan)
	}()

	var opportunities []*OptimizationOpportunity
	for res := range resChan {
		if res.err != nil {
			e.logger.Warn("Failed to analyze resource", zap.Error(res.err))
			continue
		}
		if res.opp != nil && res.opp.EstimatedSavings >= e.config.MinSavingsThreshold {
			opportunities = append(opportunities, res.opp)
		}
	}

	e.logger.Info("Orientation completed", zap.Int("opportunities", len(opportunities)))
	return opportunities, nil
}

// analyzeResource performs comprehensive analysis on a single resource
func (e *OODAEngine) analyzeResource(ctx context.Context, resource *cloud.ResourceV2) (*OptimizationOpportunity, error) {
	ctx, span := e.tracer.Start(ctx, "ooda.analyze_resource")
	defer span.End()

	span.SetAttributes(attribute.String("resource.id", resource.ID), attribute.String("resource.type", resource.Type))

	vectors := []AnalysisVector{
		e.analyzeRightsizing(resource),
		e.analyzeSpotArbitrage(resource),
		e.analyzeScheduling(resource),
		e.analyzeCostPatterns(resource),
	}

	// Calculate weighted risk score
	riskScore := e.calculateRiskScore(vectors)

	// Generate AI-powered recommendations
	recommendations, confidence, err := e.generateRecommendations(ctx, resource, vectors)
	if err != nil {
		return nil, fmt.Errorf("failed to generate recommendations: %w", err)
	}

	// Estimate savings
	estimatedSavings := e.estimateSavings(resource, recommendations)

	return &OptimizationOpportunity{
		Resource:         resource,
		AnalysisVectors:  vectors,
		RiskScore:        riskScore,
		Recommendations:  recommendations,
		EstimatedSavings: estimatedSavings,
		Confidence:       confidence,
	}, nil
}

// analyzeRightsizing analyzes CPU/memory utilization patterns
func (e *OODAEngine) analyzeRightsizing(resource *cloud.ResourceV2) AnalysisVector {
	vector := AnalysisVector{
		Name:   "rightsizing",
		Weight: 0.3,
	}

	// CPU utilization analysis
	if resource.CPUUsage < 0.2 {
		vector.Score = 0.8 // High opportunity for rightsizing
		vector.Findings = append(vector.Findings, "Low CPU utilization detected")
	} else if resource.CPUUsage > 0.8 {
		vector.Score = 0.2 // Low opportunity, might need upgrade
		vector.Findings = append(vector.Findings, "High CPU utilization detected")
	} else {
		vector.Score = 0.5 // Optimal range
		vector.Findings = append(vector.Findings, "CPU utilization within optimal range")
	}

	// Memory utilization analysis
	if resource.MemoryUsage < 0.3 {
		vector.Score = (vector.Score + 0.8) / 2
		vector.Findings = append(vector.Findings, "Low memory utilization detected")
	} else if resource.MemoryUsage > 0.9 {
		vector.Score = (vector.Score + 0.1) / 2
		vector.Findings = append(vector.Findings, "High memory utilization detected")
	}

	vector.Confidence = 0.7
	return vector
}

// analyzeSpotArbitrage analyzes spot instance opportunities
func (e *OODAEngine) analyzeSpotArbitrage(resource *cloud.ResourceV2) AnalysisVector {
	vector := AnalysisVector{
		Name:   "spot_arbitrage",
		Weight: 0.25,
	}

	// Check if resource is suitable for spot instances
	if resource.Type == "ec2" && resource.CPUUsage < 0.7 {
		vector.Score = 0.7
		vector.Findings = append(vector.Findings, "Candidate for spot instance optimization")
		vector.Confidence = 0.6
	} else {
		vector.Score = 0.2
		vector.Findings = append(vector.Findings, "Not suitable for spot instances")
		vector.Confidence = 0.8
	}

	return vector
}

// analyzeScheduling analyzes scheduling opportunities
func (e *OODAEngine) analyzeScheduling(resource *cloud.ResourceV2) AnalysisVector {
	vector := AnalysisVector{
		Name:   "scheduling",
		Weight: 0.2,
	}

	// Check for non-production workloads
	if resource.Tags != nil {
		if env, ok := resource.Tags["environment"]; ok && env != "production" {
			vector.Score = 0.6
			vector.Findings = append(vector.Findings, "Non-production workload detected")
			vector.Confidence = 0.5
		} else {
			vector.Score = 0.1
			vector.Findings = append(vector.Findings, "Production workload - scheduling limited")
			vector.Confidence = 0.9
		}
	} else {
		vector.Score = 0.3
		vector.Findings = append(vector.Findings, "No environment tags detected")
		vector.Confidence = 0.3
	}

	return vector
}

// analyzeCostPatterns analyzes cost patterns and trends
func (e *OODAEngine) analyzeCostPatterns(resource *cloud.ResourceV2) AnalysisVector {
	vector := AnalysisVector{
		Name:   "cost_patterns",
		Weight: 0.25,
	}

	// Analyze cost efficiency
	if resource.CostPerMonth > 100 {
		vector.Score = 0.6
		vector.Findings = append(vector.Findings, "High-cost resource identified")
	} else {
		vector.Score = 0.3
		vector.Findings = append(vector.Findings, "Cost within normal range")
	}

	vector.Confidence = 0.4
	return vector
}

// calculateRiskScore calculates overall risk score from analysis vectors
func (e *OODAEngine) calculateRiskScore(vectors []AnalysisVector) float64 {
	var weightedScore float64
	var totalWeight float64

	for _, vector := range vectors {
		weightedScore += vector.Score * vector.Weight
		totalWeight += vector.Weight
	}

	if totalWeight == 0 {
		return 0
	}

	return weightedScore / totalWeight
}

// generateRecommendations uses AI to generate optimization recommendations
func (e *OODAEngine) generateRecommendations(ctx context.Context, resource *cloud.ResourceV2, vectors []AnalysisVector) ([]string, float64, error) {
	ctx, span := e.tracer.Start(ctx, "ooda.generate_recommendations")
	defer span.End()

	// Build analysis context for AI
	analysisContext := e.buildAnalysisContext(resource, vectors)

	// Get AI recommendation
	response, err := e.aiOrchestrator.Analyze(ctx, analysisContext, e.calculateRiskScore(vectors), resource)
	if err != nil {
		return nil, 0, fmt.Errorf("AI analysis failed: %w", err)
	}

	// Parse recommendations from AI response
	recommendations := e.parseRecommendations(response.Content)

	return recommendations, response.Confidence, nil
}

// buildAnalysisContext builds the analysis context for AI
func (e *OODAEngine) buildAnalysisContext(resource *cloud.ResourceV2, vectors []AnalysisVector) string {
	context := fmt.Sprintf(`
Resource Analysis Request:
- ID: %s
- Type: %s
- Provider: %s
- Region: %s
- CPU Usage: %.2f%%
- Memory Usage: %.2f%%
- Monthly Cost: $%.2f

Analysis Vectors:
`, resource.ID, resource.Type, resource.Provider, resource.Region,
		resource.CPUUsage*100, resource.MemoryUsage*100, resource.CostPerMonth)

	for _, vector := range vectors {
		context += fmt.Sprintf("- %s: Score %.2f, Findings: %v\n",
			vector.Name, vector.Score, vector.Findings)
	}

	context += `
Please provide specific optimization recommendations for this resource.
Consider the risk factors and provide actionable steps.
Format as a list of recommendations.
`

	return context
}

// parseRecommendations parses AI response into recommendation list
func (e *OODAEngine) parseRecommendations(aiResponse string) []string {
	// Simple parsing - in production, use more sophisticated parsing
	var recommendations []string

	// Split by lines and filter non-empty
	lines := strings.Split(aiResponse, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		// Filter out empty lines, markdown headers, and list markers
		if line != "" && !strings.HasPrefix(line, "#") && !strings.HasPrefix(line, "Here are") {
			line = strings.TrimPrefix(line, "- ")
			line = strings.TrimPrefix(line, "* ")
			line = strings.TrimPrefix(line, "1. ")
			recommendations = append(recommendations, line)
		}
	}

	return recommendations
}

// estimateSavings estimates potential savings from recommendations
func (e *OODAEngine) estimateSavings(resource *cloud.ResourceV2, recommendations []string) float64 {
	// Simple estimation based on resource cost and recommendation impact
	ratio := e.config.DefaultSavingsRatio
	if ratio <= 0 {
		ratio = 0.2 // Default to 20% if not configured
	}
	baseSavings := resource.CostPerMonth * ratio

	// Adjust based on number of recommendations
	multiplier := 1.0 + (float64(len(recommendations)) * 0.1)

	return baseSavings * multiplier
}

// decide prioritizes and makes decisions on opportunities
func (e *OODAEngine) decide(ctx context.Context, opportunities []*OptimizationOpportunity) ([]*database.Action, error) {
	ctx, span := e.tracer.Start(ctx, "ooda.decide")
	defer span.End()

	e.logger.Info("Deciding - prioritizing optimization opportunities")

	var actions []*database.Action

	for _, opportunity := range opportunities {
		// Check risk threshold
		if opportunity.RiskScore > e.config.RiskThreshold {
			e.logger.Info("Skipping high-risk opportunity",
				zap.String("resource_id", opportunity.Resource.ID),
				zap.Float64("risk_score", opportunity.RiskScore),
				zap.Float64("threshold", e.config.RiskThreshold),
			)
			continue
		}

		// Create action record
		action := &database.Action{
			ID:               e.generateActionID(opportunity),
			ResourceID:       opportunity.Resource.ID,
			ActionType:       "optimize",
			Status:           "PENDING",
			Checksum:         e.generateChecksum(opportunity),
			RiskScore:        opportunity.RiskScore,
			EstimatedSavings: opportunity.EstimatedSavings,
		}

		// Serialize recommendations to payload
		payload := map[string]interface{}{
			"recommendations": opportunity.Recommendations,
			"confidence":      opportunity.Confidence,
			"vectors":         opportunity.AnalysisVectors,
		}
		payloadBytes, _ := json.Marshal(payload)
		action.Payload = string(payloadBytes)

		// Store action in database
		err := e.repository.CreateAction(ctx, action)
		if err != nil {
			e.logger.Error("Failed to create action", zap.Error(err))
			continue
		}

		actions = append(actions, action)
	}

	e.logger.Info("Decision phase completed", zap.Int("actions_created", len(actions)))
	return actions, nil
}

// act executes the optimization actions
func (e *OODAEngine) act(ctx context.Context, actions []*database.Action) ([]*database.SavingsEvent, error) {
	ctx, span := e.tracer.Start(ctx, "ooda.act")
	defer span.End()

	e.logger.Info("Acting - executing optimization actions")

	var results []*database.SavingsEvent

	for _, action := range actions {
		result, err := e.executeAction(ctx, action)
		if err != nil {
			e.logger.Error("Failed to execute action", zap.String("action_id", action.ID), zap.Error(err))
			continue
		}

		if result != nil {
			results = append(results, result)
		}
	}

	e.logger.Info("Act phase completed", zap.Int("savings_recorded", len(results)))
	return results, nil
}

// executeAction executes a single optimization action
func (e *OODAEngine) executeAction(ctx context.Context, action *database.Action) (*database.SavingsEvent, error) {
	ctx, span := e.tracer.Start(ctx, "ooda.execute_action")
	defer span.End()

	// Update action status to in progress
	now := time.Now()
	err := e.repository.UpdateActionStatus(ctx, action.ID, "IN_PROGRESS", &now, nil, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to update action status: %w", err)
	}

	// Get resource details
	resource, err := e.cloudAdapter.GetResource(ctx, action.ResourceID)
	if err != nil {
		return nil, fmt.Errorf("failed to get resource: %w", err)
	}

	// Execute optimization based on action type
	var actualSavings float64
	switch action.ActionType {
	case "optimize":
		actualSavings, err = e.executeOptimization(ctx, resource, action)
	case "terminate":
		actualSavings, err = e.executeTermination(ctx, resource, action)
	default:
		err = fmt.Errorf("unknown action type: %s", action.ActionType)
	}

	if err != nil {
		// Update action status to failed
		errorMsg := err.Error()
		e.repository.UpdateActionStatus(ctx, action.ID, "FAILED", nil, nil, &errorMsg)
		return nil, fmt.Errorf("action execution failed: %w", err)
	}

	// Update action status to completed
	completedAt := time.Now()
	err = e.repository.UpdateActionStatus(ctx, action.ID, "COMPLETED", nil, &completedAt, nil)
	if err != nil {
		e.logger.Warn("Failed to update action completion status", zap.Error(err))
	}

	// Record savings event
	savingsEvent := &database.SavingsEvent{
		ID:               e.generateSavingsEventID(action),
		ActionID:         &action.ID,
		ResourceID:       action.ResourceID,
		OptimizationType: &action.ActionType,
		EstimatedSavings: &action.EstimatedSavings,
		ActualSavings:    &actualSavings,
	}

	err = e.repository.CreateSavingsEvent(ctx, savingsEvent)
	if err != nil {
		e.logger.Warn("Failed to create savings event", zap.Error(err))
	}

	return savingsEvent, nil
}

// executeOptimization executes resource optimization
func (e *OODAEngine) executeOptimization(ctx context.Context, resource *cloud.ResourceV2, action *database.Action) (float64, error) {
	// Parse action payload
	var payload map[string]interface{}
	err := json.Unmarshal([]byte(action.Payload), &payload)
	if err != nil {
		return 0, fmt.Errorf("failed to parse action payload: %w", err)
	}

	// Execute optimization via cloud adapter
	savings, err := e.cloudAdapter.ApplyOptimization(ctx, resource, "optimize")
	if err != nil {
		return 0, fmt.Errorf("cloud optimization failed: %w", err)
	}

	return savings, nil
}

// executeTermination executes resource termination
func (e *OODAEngine) executeTermination(ctx context.Context, resource *cloud.ResourceV2, _ *database.Action) (float64, error) {
	// Execute termination via cloud adapter
	savings, err := e.cloudAdapter.ApplyOptimization(ctx, resource, "terminate")
	if err != nil {
		return 0, fmt.Errorf("cloud termination failed: %w", err)
	}

	return savings, nil
}

// generateActionID generates a unique action ID
func (e *OODAEngine) generateActionID(opportunity *OptimizationOpportunity) string {
	timestamp := time.Now().Unix()
	return fmt.Sprintf("action_%s_%d", opportunity.Resource.ID, timestamp)
}

// generateSavingsEventID generates a unique savings event ID
func (e *OODAEngine) generateSavingsEventID(action *database.Action) string {
	timestamp := time.Now().Unix()
	return fmt.Sprintf("savings_%s_%d", action.ID, timestamp)
}

// generateChecksum generates a checksum for idempotency
func (e *OODAEngine) generateChecksum(opportunity *OptimizationOpportunity) string {
	data := fmt.Sprintf("%s-%s-%v",
		opportunity.Resource.ID,
		opportunity.Resource.Type,
		opportunity.Recommendations)

	hash := sha256.Sum256([]byte(data))
	return hex.EncodeToString(hash[:])
}

// DefaultEngineConfig returns default engine configuration
func DefaultEngineConfig() *EngineConfig {
	return &EngineConfig{
		MaxConcurrentCycles:   3,
		MaxConcurrentAnalysis: 10,
		CycleInterval:         30 * time.Minute,
		RiskThreshold:         7.0,
		MinSavingsThreshold:   10.0,
		MaxAnalysisTime:       5 * time.Minute,
		EnableAutoExecution:   false,
		RequireHumanApproval:  true,
		DefaultSavingsRatio:   0.2,
	}
}

// ProductionEngineConfig returns production engine configuration
func ProductionEngineConfig() *EngineConfig {
	return &EngineConfig{
		MaxConcurrentCycles:   5,
		MaxConcurrentAnalysis: 50,
		CycleInterval:         15 * time.Minute,
		RiskThreshold:         5.0,
		MinSavingsThreshold:   25.0,
		MaxAnalysisTime:       3 * time.Minute,
		EnableAutoExecution:   false,
		RequireHumanApproval:  true,
		DefaultSavingsRatio:   0.2,
	}
}
