package ai

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/project-atlas/atlas/internal/analytics"
	"github.com/project-atlas/atlas/internal/cloud"
	"github.com/project-atlas/atlas/internal/logger"
	"go.uber.org/zap"
)

// TOPAZOrchestrator extends the UnifiedOrchestrator with ROSES/T.O.P.A.Z. capabilities
type TOPAZOrchestrator struct {
	*UnifiedOrchestrator
	rosesFramework *ROSESFramework
	topazLogic     *TOPAZLogic
}

// NewTOPAZOrchestrator creates a new orchestrator with ROSES/T.O.P.A.Z. capabilities and zap logger
func NewTOPAZOrchestrator(config *Config, tracker *analytics.TokenTracker, l *zap.Logger) (*TOPAZOrchestrator, error) {
	if l == nil {
		l = logger.GetLogger()
	}

	// Create base orchestrator
	baseOrchestrator, err := NewUnifiedOrchestrator(config, tracker, l)
	if err != nil {
		return nil, fmt.Errorf("failed to create base orchestrator: %w", err)
	}

	return &TOPAZOrchestrator{
		UnifiedOrchestrator: baseOrchestrator,
		rosesFramework:      NewROSESFramework(),
		topazLogic:          NewTOPAZLogic(),
	}, nil
}

// AnalyzeWithROSES performs analysis using the ROSES framework
func (to *TOPAZOrchestrator) AnalyzeWithROSES(ctx context.Context, resource *cloud.ResourceV2, contextData map[string]interface{}) (*TOPAZDecision, error) {
	// Generate ROSES prompt
	prompt := to.rosesFramework.GenerateROSESPrompt(resource, contextData)

	// Create AI request
	request := AIRequest{
		Prompt:       prompt,
		ResourceType: resource.Type,
		RiskScore:    to.calculateInitialRisk(resource),
		MaxTokens:    2000, // Larger for detailed analysis
		Temperature:  0.3,  // Lower for more consistent results
		Metadata: map[string]interface{}{
			"framework":   "roses",
			"topaz_logic": true,
			"resource_id": resource.ID,
			"timestamp":   time.Now(),
		},
	}

	// Get AI response using base orchestrator, but with custom request parameters
	client := to.UnifiedOrchestrator.GetFactory().GetClientForRisk(request.RiskScore)
	response, err := to.UnifiedOrchestrator.AnalyzeWithRetry(ctx, client, request, 3)
	if err != nil {
		return nil, fmt.Errorf("AI analysis failed: %w", err)
	}

	// Parse AI response and apply T.O.P.A.Z. logic
	topazDecision, err := to.parseAndApplyTOPAZ(ctx, resource, response, contextData)
	if err != nil {
		return nil, fmt.Errorf("failed to apply T.O.P.A.Z. logic: %w", err)
	}

	// Record decision for learning
	outcome := DecisionOutcome{
		ResourceID:    resource.ID,
		Decision:      topazDecision.Recommendation,
		RiskScore:     topazDecision.RiskScore,
		ActualSavings: topazDecision.ExpectedSavings,
		ImpactScore:   topazDecision.AntiFragileScore,
		Timestamp:     time.Now(),
		Success:       topazDecision.GoNoGo == "Go",
	}

	to.topazLogic.RecordDecision(outcome)

	return topazDecision, nil
}

// BatchAnalyzeWithROSES performs batch analysis using ROSES framework
func (to *TOPAZOrchestrator) BatchAnalyzeWithROSES(ctx context.Context, resources []*cloud.ResourceV2) ([]*TOPAZDecision, error) {
	decisions := make([]*TOPAZDecision, len(resources))

	// Process resources in parallel with concurrency control
	semaphore := make(chan struct{}, 5) // Max 5 concurrent analyses

	for i, resource := range resources {
		semaphore <- struct{}{} // Acquire

		go func(idx int, res *cloud.ResourceV2) {
			defer func() { <-semaphore }() // Release

			contextData := map[string]interface{}{
				"batch_mode":      true,
				"batch_index":     idx,
				"total_resources": len(resources),
			}

			decision, err := to.AnalyzeWithROSES(ctx, res, contextData)
			if err != nil {
				to.logger.Error("Failed to analyze resource", zap.String("resource_id", res.ID), zap.Error(err))
				// Create a safe fallback decision
				decision = &TOPAZDecision{
					ResourceID:     res.ID,
					Recommendation: "MONITOR - Analysis failed",
					RiskScore:      100.0, // High risk to prevent action
					GoNoGo:         "No-Go",
					Reasoning:      []string{fmt.Sprintf("Analysis error: %v", err)},
				}
			}

			decisions[idx] = decision
		}(i, resource)
	}

	// Wait for all analyses to complete
	for i := 0; i < cap(semaphore); i++ {
		semaphore <- struct{}{}
	}

	return decisions, nil
}

// parseAndApplyTOPAZ parses AI response and applies T.O.P.A.Z. logic
func (to *TOPAZOrchestrator) parseAndApplyTOPAZ(ctx context.Context, resource *cloud.ResourceV2, aiResponse *AIResponse, _ map[string]interface{}) (*TOPAZDecision, error) {
	// First, apply T.O.P.A.Z. logic
	topazDecision, err := to.topazLogic.AnalyzeWithTOPAZ(ctx, resource, aiResponse.Content)
	if err != nil {
		return nil, err
	}

	// Enhance with AI insights
	if err := to.enhanceWithAIInsights(topazDecision, aiResponse); err != nil {
		to.logger.Warn("Failed to enhance with AI insights", zap.Error(err))
	}

	// Add context metadata
	topazDecision.Metadata["ai_confidence"] = aiResponse.Confidence
	topazDecision.Metadata["ai_tokens_used"] = aiResponse.TokensUsed
	topazDecision.Metadata["analysis_timestamp"] = time.Now()

	return topazDecision, nil
}

// enhanceWithAIInsights enhances the T.O.P.A.Z. decision with AI insights
func (to *TOPAZOrchestrator) enhanceWithAIInsights(decision *TOPAZDecision, aiResponse *AIResponse) error {
	// Try to parse AI response as JSON for additional insights
	var aiInsights map[string]interface{}
	if err := json.Unmarshal([]byte(aiResponse.Content), &aiInsights); err == nil {
		if riskScore, ok := aiInsights["risk_score"].(float64); ok {
			// Blend AI risk score with T.O.P.A.Z. risk score
			decision.RiskScore = (decision.RiskScore + riskScore) / 2
		}

		if confidence, ok := aiInsights["confidence"].(float64); ok {
			decision.Confidence = (decision.Confidence + confidence) / 2
		}

		if reasoning, ok := aiInsights["reasoning"].([]interface{}); ok {
			for _, reason := range reasoning {
				if reasonStr, ok := reason.(string); ok {
					decision.Reasoning = append(decision.Reasoning, fmt.Sprintf("AI Insight: %s", reasonStr))
				}
			}
		}
	} else {
		// If not JSON, add as reasoning
		decision.Reasoning = append(decision.Reasoning, fmt.Sprintf("AI Analysis: %s", aiResponse.Content))
	}

	return nil
}

// calculateInitialRisk calculates initial risk score for resource selection
func (to *TOPAZOrchestrator) calculateInitialRisk(resource *cloud.ResourceV2) float64 {
	// Simple risk calculation for AI model selection
	risk := 0.0

	if resource.CPUUsage > 80 {
		risk += 30
	}
	if resource.MemoryUsage > 80 {
		risk += 30
	}
	if resource.CostPerMonth > 1000 {
		risk += 20
	}
	if isProductionResource(resource) {
		risk += 20
	}

	return risk
}

// GetLearningInsights returns insights from the learning engine (thread-safe)
func (to *TOPAZOrchestrator) GetLearningInsights() map[string]interface{} {
	insights := make(map[string]interface{})

	successCount, failureCount := to.topazLogic.GetPatternsCount()
	historicalDecisions := to.topazLogic.GetHistoricalDecisions()

	insights["total_decisions"] = len(historicalDecisions)
	insights["success_patterns"] = successCount
	insights["failure_patterns"] = failureCount

	// Calculate success rate
	total := successCount + failureCount
	if total > 0 {
		insights["success_rate"] = float64(successCount) / float64(total)
	}

	return insights
}

// OptimizePromptBasedOnHistory optimizes prompts based on historical success (thread-safe)
func (to *TOPAZOrchestrator) OptimizePromptBasedOnHistory(resource *cloud.ResourceV2) string {
	// Analyze historical patterns for similar resources
	promptOptimizations := ""
	historicalDecisions := to.topazLogic.GetHistoricalDecisions()

	for _, outcome := range historicalDecisions {
		if to.isSimilarResourceType(resource, outcome.ResourceID) {
			if outcome.Success {
				promptOptimizations += "\nNote: Similar resources have responded well to optimization in the past."
			} else {
				promptOptimizations += "\nCaution: Similar resources have had issues with optimization. Be extra conservative."
			}
		}
	}

	return promptOptimizations
}

// isSimilarResourceType checks if resources are of similar type
func (to *TOPAZOrchestrator) isSimilarResourceType(resource1 *cloud.ResourceV2, resource2ID string) bool {
	// Extract resource type from ID (simplified)
	resource1Type := extractResourceType(resource1.ID)
	resource2Type := extractResourceType(resource2ID)

	return resource1Type == resource2Type
}

// extractResourceType extracts resource type from ID
func extractResourceType(resourceID string) string {
	// Simple extraction - in production, this would be more sophisticated
	parts := strings.Split(resourceID, "-")
	if len(parts) > 1 {
		return parts[0]
	}
	return "unknown"
}

// ExportLearningData exports learning data for analysis (thread-safe)
func (to *TOPAZOrchestrator) ExportLearningData() ([]byte, error) {
	to.topazLogic.learning.mu.RLock()
	defer to.topazLogic.learning.mu.RUnlock()

	data := map[string]interface{}{
		"historical_decisions": to.topazLogic.learning.historicalDecisions,
		"success_patterns":     to.topazLogic.learning.successPatterns,
		"failure_patterns":     to.topazLogic.learning.failurePatterns,
		"topaz_thresholds":     to.topazLogic.thresholds,
		"antifragile_rules":    to.topazLogic.antifragile,
		"export_timestamp":     time.Now(),
	}

	return json.MarshalIndent(data, "", "  ")
}

// ImportLearningData imports learning data from external source (thread-safe)
func (to *TOPAZOrchestrator) ImportLearningData(data []byte) error {
	var imported map[string]interface{}
	if err := json.Unmarshal(data, &imported); err != nil {
		return fmt.Errorf("failed to import learning data: %w", err)
	}

	to.topazLogic.learning.mu.Lock()
	defer to.topazLogic.learning.mu.Unlock()

	// Import historical decisions
	if decisions, ok := imported["historical_decisions"].(map[string]interface{}); ok {
		for id, outcome := range decisions {
			outcomeData, err := json.Marshal(outcome)
			if err != nil {
				continue
			}
			var decision DecisionOutcome
			if err := json.Unmarshal(outcomeData, &decision); err == nil {
				to.topazLogic.learning.historicalDecisions[id] = decision
			}
		}
	}

	return nil
}
