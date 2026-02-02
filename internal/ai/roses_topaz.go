package ai

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/project-atlas/atlas/internal/cloud"
	"github.com/project-atlas/atlas/internal/telemetry"
	"go.opentelemetry.io/otel/attribute"
	"sync"
)

// ROSESFramework implements the Role-Objective-Scenario-ExpectedSolution-Steps prompting method
type ROSESFramework struct {
	role              string
	objective         string
	scenario          string
	expectedFormat    string
	steps             []string
	rules             []string
	systemInstruction string
}

// TOPAZLogic implements the T.O.P.A.Z. Zero-Sum Learning framework
type TOPAZLogic struct {
	thresholds  TOPAZThresholds
	antifragile AntifragileRules
	learning    LearningEngine
}

// TOPAZThresholds defines risk thresholds for decision making
type TOPAZThresholds struct {
	MaxRiskScore      float64 `json:"max_risk_score"`
	ConservativeMode  bool    `json:"conservative_mode"`
	WeekendMultiplier float64 `json:"weekend_multiplier"`
	ProductionSLA     float64 `json:"production_sla"`
}

// AntifragileRules defines anti-fragile system rules
type AntifragileRules struct {
	RequireAntiFragileTags bool     `json:"require_anti_fragile_tags"`
	ProtectedResources     []string `json:"protected_resources"`
	MaintenanceWindows     []string `json:"maintenance_windows"`
}

// LearningEngine handles zero-sum learning from past decisions (thread-safe)
type LearningEngine struct {
	mu                  sync.RWMutex
	historicalDecisions map[string]DecisionOutcome
	successPatterns     []string
	failurePatterns     []string
}

// DecisionOutcome tracks the result of past AI decisions
type DecisionOutcome struct {
	ResourceID    string    `json:"resource_id"`
	Decision      string    `json:"decision"`
	RiskScore     float64   `json:"risk_score"`
	ActualSavings float64   `json:"actual_savings"`
	ImpactScore   float64   `json:"impact_score"`
	Timestamp     time.Time `json:"timestamp"`
	Success       bool      `json:"success"`
}

// TOPAZDecision represents a structured AI decision
type TOPAZDecision struct {
	ResourceID       string                 `json:"resource_id"`
	Recommendation   string                 `json:"recommendation"`
	RiskScore        float64                `json:"risk_score"`
	Confidence       float64                `json:"confidence"`
	Reasoning        []string               `json:"reasoning"`
	ExpectedSavings  float64                `json:"expected_savings"`
	AntiFragileScore float64                `json:"anti_fragile_score"`
	GoNoGo           string                 `json:"go_no_go"`
	Metadata         map[string]interface{} `json:"metadata"`
}

// NewROSESFramework creates a new ROSES prompting framework
func NewROSESFramework() *ROSESFramework {
	return &ROSESFramework{
		role:           "You are a Senior Cloud Economics Analyst and Cost Optimization Expert using the T.O.P.A.Z. Zero-Sum Learning framework.",
		objective:      "Analyze cloud resources and provide data-driven optimization recommendations with risk assessment.",
		scenario:       "Enterprise production environment with 99.9% SLA requirements and anti-fragile system design principles.",
		expectedFormat: "JSON response with Risk Score (0-100), Go/No-Go recommendation, and detailed reasoning.",
		steps: []string{
			"1. Analyze current resource utilization patterns",
			"2. Predict future load based on historical data",
			"3. Check for anti-fragile system tags and dependencies",
			"4. Calculate risk score using T.O.P.A.Z. methodology",
			"5. Apply zero-sum learning from similar past decisions",
			"6. Provide recommendation with confidence score",
		},
		rules: []string{
			"Never suggest optimization with Risk Score > 50 in production",
			"Prioritize long-term system stability over short-term cost savings",
			"Always consider anti-fragile system requirements",
			"Apply weekend multiplier for production workloads",
			"Check maintenance windows before suggesting changes",
		},
		systemInstruction: "Apply the T.O.P.A.Z. Zero-Sum Learning logic with strict risk management and anti-fragile system principles.",
	}
}

// NewTOPAZLogic creates a new T.O.P.A.Z. logic engine
func NewTOPAZLogic() *TOPAZLogic {
	return &TOPAZLogic{
		thresholds: TOPAZThresholds{
			MaxRiskScore:      50.0,
			ConservativeMode:  true,
			WeekendMultiplier: 1.5,
			ProductionSLA:     99.9,
		},
		antifragile: AntifragileRules{
			RequireAntiFragileTags: true,
			ProtectedResources:     []string{"db-prod-*", "auth-*", "payment-*"},
			MaintenanceWindows:     []string{"Saturday 02:00-04:00", "Sunday 02:00-04:00"},
		},
		learning: LearningEngine{
			historicalDecisions: make(map[string]DecisionOutcome),
			successPatterns:     []string{},
			failurePatterns:     []string{},
		},
	}
}

// GenerateROSESPrompt creates a structured prompt using the ROSES framework
func (r *ROSESFramework) GenerateROSESPrompt(resource *cloud.ResourceV2, contextData map[string]interface{}) string {
	promptBuilder := strings.Builder{}

	// XML Delimiters for structure
	promptBuilder.WriteString("<System_Instruction>\n")
	promptBuilder.WriteString(fmt.Sprintf("%s\n", r.systemInstruction))
	promptBuilder.WriteString("</System_Instruction>\n\n")

	promptBuilder.WriteString("<Role>\n")
	promptBuilder.WriteString(fmt.Sprintf("%s\n", r.role))
	promptBuilder.WriteString("</Role>\n\n")

	promptBuilder.WriteString("<Objective>\n")
	promptBuilder.WriteString(fmt.Sprintf("%s\n", r.objective))
	promptBuilder.WriteString("</Objective>\n\n")

	promptBuilder.WriteString("<Scenario>\n")
	promptBuilder.WriteString(fmt.Sprintf("%s\n", r.scenario))
	promptBuilder.WriteString(fmt.Sprintf("Current Time: %s\n", time.Now().Format("2006-01-02 15:04:05")))
	if isWeekend() {
		promptBuilder.WriteString("⚠️ WEEKEND MODE: Apply 1.5x risk multiplier\n")
	}
	promptBuilder.WriteString("</Scenario>\n\n")

	promptBuilder.WriteString("<Current_Cloud_Data>\n")
	promptBuilder.WriteString(fmt.Sprintf("Resource ID: %s\n", resource.ID))
	promptBuilder.WriteString(fmt.Sprintf("Resource Type: %s\n", resource.Type))
	promptBuilder.WriteString(fmt.Sprintf("Provider: %s\n", resource.Provider))
	promptBuilder.WriteString(fmt.Sprintf("Region: %s\n", resource.Region))
	promptBuilder.WriteString(fmt.Sprintf("CPU Usage: %.2f%%\n", resource.CPUUsage))
	promptBuilder.WriteString(fmt.Sprintf("Memory Usage: %.2f%%\n", resource.MemoryUsage))
	promptBuilder.WriteString(fmt.Sprintf("Cost Per Month: $%.2f\n", resource.CostPerMonth))

	// Add tags
	if len(resource.Tags) > 0 {
		promptBuilder.WriteString("Tags:\n")
		for key, value := range resource.Tags {
			promptBuilder.WriteString(fmt.Sprintf("  %s: %s\n", key, value))
		}
	}

	// Add context data
	if contextData != nil {
		promptBuilder.WriteString("Additional Context:\n")
		for key, value := range contextData {
			promptBuilder.WriteString(fmt.Sprintf("  %s: %v\n", key, value))
		}
	}

	promptBuilder.WriteString("</Current_Cloud_Data>\n\n")

	promptBuilder.WriteString("<Rules>\n")
	for _, rule := range r.rules {
		promptBuilder.WriteString(fmt.Sprintf("- %s\n", rule))
	}
	promptBuilder.WriteString("</Rules>\n\n")

	promptBuilder.WriteString("<Analysis_Steps>\n")
	for _, step := range r.steps {
		promptBuilder.WriteString(fmt.Sprintf("%s\n", step))
	}
	promptBuilder.WriteString("</Analysis_Steps>\n\n")

	promptBuilder.WriteString("<Expected_Solution>\n")
	promptBuilder.WriteString(fmt.Sprintf("%s\n", r.expectedFormat))
	promptBuilder.WriteString("Let's think step by step to ensure accurate analysis.\n")
	promptBuilder.WriteString("</Expected_Solution>\n")

	return promptBuilder.String()
}

// AnalyzeWithTOPAZ applies T.O.P.A.Z. logic to analyze a resource with distributed tracing
func (t *TOPAZLogic) AnalyzeWithTOPAZ(ctx context.Context, resource *cloud.ResourceV2, prompt string) (*TOPAZDecision, error) {
	ctx, span := telemetry.StartSpan(ctx, "AnalyzeWithTOPAZ")
	defer span.End()

	if resource == nil {
		err := fmt.Errorf("resource is nil")
		telemetry.RecordError(span, err)
		return nil, err
	}

	span.SetAttributes(
		attribute.String("resource.id", resource.ID),
		attribute.String("resource.type", resource.Type),
		attribute.String("resource.provider", resource.Provider),
	)

	decision := &TOPAZDecision{
		ResourceID: resource.ID,
		Metadata:   make(map[string]interface{}),
	}

	// Step 1: Calculate base risk score
	baseRisk := t.calculateBaseRisk(resource)
	span.SetAttributes(attribute.Float64("risk.base", baseRisk))

	// Step 2: Apply weekend multiplier if applicable
	if isWeekend() {
		baseRisk *= t.thresholds.WeekendMultiplier
		decision.Metadata["weekend_mode"] = true
		span.SetAttributes(attribute.Bool("risk.weekend_multiplier_applied", true))
	}

	// Step 3: Check anti-fragile requirements
	antiFragileScore := t.calculateAntiFragileScore(resource)
	decision.AntiFragileScore = antiFragileScore
	span.SetAttributes(attribute.Float64("antifragile.score", antiFragileScore))

	// Step 4: Apply zero-sum learning
	learningAdjustment := t.applyZeroSumLearning(resource)
	baseRisk += learningAdjustment
	span.SetAttributes(attribute.Float64("risk.learning_adjustment", learningAdjustment))

	// Step 5: Final risk assessment
	decision.RiskScore = baseRisk
	decision.Confidence = t.calculateConfidence(resource, antiFragileScore)
	span.SetAttributes(attribute.Float64("risk.final", decision.RiskScore))
	span.SetAttributes(attribute.Float64("confidence.final", decision.Confidence))

	// Step 6: Generate recommendation
	decision.Recommendation = t.generateRecommendation(resource, decision.RiskScore)
	decision.GoNoGo = t.determineGoNoGo(decision.RiskScore)
	decision.ExpectedSavings = t.calculateExpectedSavings(resource)
	decision.Reasoning = t.generateReasoning(resource, decision)

	telemetry.AddEvent(span, "decision_finalized", 
		attribute.String("recommendation", decision.Recommendation),
		attribute.String("go_no_go", decision.GoNoGo),
	)

	return decision, nil
}

// calculateBaseRisk calculates the base risk score
func (t *TOPAZLogic) calculateBaseRisk(resource *cloud.ResourceV2) float64 {
	risk := 0.0

	// CPU utilization risk
	if resource.CPUUsage > 80 {
		risk += 30
	} else if resource.CPUUsage > 60 {
		risk += 15
	}

	// Memory utilization risk
	if resource.MemoryUsage > 85 {
		risk += 25
	} else if resource.MemoryUsage > 70 {
		risk += 12
	}

	// Cost risk (higher cost = higher risk for optimization)
	if resource.CostPerMonth > 1000 {
		risk += 20
	} else if resource.CostPerMonth > 500 {
		risk += 10
	}

	// Production tag risk
	if isProductionResource(resource) {
		risk += 15
	}

	return risk
}

// calculateAntiFragileScore calculates how anti-fragile the resource is
func (t *TOPAZLogic) calculateAntiFragileScore(resource *cloud.ResourceV2) float64 {
	score := 50.0 // Base score

	// Check for anti-fragile tags
	if tags, ok := resource.Tags["anti-fragile"]; ok && tags == "true" {
		score += 30
	}

	if tags, ok := resource.Tags["auto-scaling"]; ok && tags == "enabled" {
		score += 20
	}

	if tags, ok := resource.Tags["redundancy"]; ok && tags == "high" {
		score += 15
	}

	// Check provider-specific anti-fragile features
	if resource.Provider == "aws" {
		score += 10 // AWS has good anti-fragile features
	}

	return score
}

// applyZeroSumLearning applies learning from past decisions (thread-safe)
func (t *TOPAZLogic) applyZeroSumLearning(resource *cloud.ResourceV2) float64 {
	t.learning.mu.RLock()
	defer t.learning.mu.RUnlock()

	adjustment := 0.0

	// Look for similar past decisions
	for _, outcome := range t.learning.historicalDecisions {
		if t.isSimilarResource(resource, outcome.ResourceID) {
			if outcome.Success {
				adjustment -= 5 // Reduce risk for successful patterns
			} else {
				adjustment += 10 // Increase risk for failed patterns
			}
		}
	}

	return adjustment
}

// generateRecommendation generates the optimization recommendation
func (t *TOPAZLogic) generateRecommendation(resource *cloud.ResourceV2, riskScore float64) string {
	if riskScore > t.thresholds.MaxRiskScore {
		return "NO_ACTION - Risk too high for optimization"
	}

	if resource.CPUUsage < 20 && resource.MemoryUsage < 30 {
		return "DOWNSIZE - Resource significantly underutilized"
	}

	if resource.CPUUsage < 50 && resource.MemoryUsage < 60 {
		return "RIGHTSIZE - Moderate underutilization detected"
	}

	if resource.CostPerMonth > 1000 && (resource.CPUUsage < 40 || resource.MemoryUsage < 40) {
		return "OPTIMIZE_INSTANCE_TYPE - High cost with low utilization"
	}

	return "MONITOR - Current utilization acceptable"
}

// determineGoNoGo determines if action should be taken
func (t *TOPAZLogic) determineGoNoGo(riskScore float64) string {
	if riskScore > t.thresholds.MaxRiskScore {
		return "No-Go"
	}
	return "Go"
}

// calculateExpectedSavings calculates potential savings
func (t *TOPAZLogic) calculateExpectedSavings(resource *cloud.ResourceV2) float64 {
	if resource.CPUUsage < 20 && resource.MemoryUsage < 30 {
		return resource.CostPerMonth * 0.6 // 60% savings for downsizing
	}

	if resource.CPUUsage < 50 && resource.MemoryUsage < 60 {
		return resource.CostPerMonth * 0.3 // 30% savings for rightsizing
	}

	return resource.CostPerMonth * 0.1 // 10% savings for optimization
}

// generateReasoning generates detailed reasoning for the decision
func (t *TOPAZLogic) generateReasoning(resource *cloud.ResourceV2, decision *TOPAZDecision) []string {
	reasoning := []string{
		fmt.Sprintf("CPU utilization at %.1f%% indicates %s", resource.CPUUsage, t.getUtilizationLevel(resource.CPUUsage)),
		fmt.Sprintf("Memory utilization at %.1f%% shows %s", resource.MemoryUsage, t.getUtilizationLevel(resource.MemoryUsage)),
		fmt.Sprintf("Risk score of %.1f is %s threshold", decision.RiskScore, t.getRiskLevel(decision.RiskScore)),
		fmt.Sprintf("Anti-fragile score of %.1f suggests %s", decision.AntiFragileScore, t.getAntiFragileLevel(decision.AntiFragileScore)),
	}

	if decision.Metadata["weekend_mode"] == true {
		reasoning = append(reasoning, "Weekend mode applied: 1.5x risk multiplier")
	}

	return reasoning
}

// Helper functions

func isWeekend() bool {
	now := time.Now()
	return now.Weekday() == time.Saturday || now.Weekday() == time.Sunday
}

func isProductionResource(resource *cloud.ResourceV2) bool {
	if tags, ok := resource.Tags["environment"]; ok && tags == "production" {
		return true
	}
	if tags, ok := resource.Tags["env"]; ok && tags == "prod" {
		return true
	}
	return strings.Contains(resource.ID, "prod") || strings.Contains(resource.ID, "production")
}

func (t *TOPAZLogic) isSimilarResource(resource1 *cloud.ResourceV2, resource2ID string) bool {
	// Simple similarity check - in production, this would be more sophisticated
	return strings.HasPrefix(resource1.ID, strings.Split(resource2ID, "-")[0])
}

func (t *TOPAZLogic) calculateConfidence(resource *cloud.ResourceV2, antiFragileScore float64) float64 {
	confidence := 0.8 // Base confidence

	if antiFragileScore > 70 {
		confidence += 0.1
	}

	if resource.CPUUsage < 10 || resource.CPUUsage > 90 {
		confidence += 0.1 // High confidence in extreme cases
	}

	return confidence
}

func (t *TOPAZLogic) getUtilizationLevel(usage float64) string {
	if usage < 20 {
		return "severe underutilization"
	} else if usage < 50 {
		return "moderate underutilization"
	} else if usage < 80 {
		return "acceptable utilization"
	}
	return "high utilization"
}

func (t *TOPAZLogic) getRiskLevel(risk float64) string {
	if risk < 25 {
		return "below"
	} else if risk < 50 {
		return "near"
	}
	return "above"
}

func (t *TOPAZLogic) getAntiFragileLevel(score float64) string {
	if score > 70 {
		return "strong anti-fragile characteristics"
	} else if score > 50 {
		return "moderate anti-fragile characteristics"
	}
	return "limited anti-fragile characteristics"
}

// RecordDecision records a decision outcome for learning (thread-safe)
func (t *TOPAZLogic) RecordDecision(outcome DecisionOutcome) {
	t.learning.mu.Lock()
	defer t.learning.mu.Unlock()

	t.learning.historicalDecisions[outcome.ResourceID] = outcome

	if outcome.Success {
		t.learning.successPatterns = append(t.learning.successPatterns, outcome.ResourceID)
	} else {
		t.learning.failurePatterns = append(t.learning.failurePatterns, outcome.ResourceID)
	}
}

// GetHistoricalDecisions returns a copy of historical decisions for safe access
func (t *TOPAZLogic) GetHistoricalDecisions() map[string]DecisionOutcome {
	t.learning.mu.RLock()
	defer t.learning.mu.RUnlock()

	copy := make(map[string]DecisionOutcome, len(t.learning.historicalDecisions))
	for k, v := range t.learning.historicalDecisions {
		copy[k] = v
	}
	return copy
}

// GetPatternsCount returns the number of success and failure patterns
func (t *TOPAZLogic) GetPatternsCount() (int, int) {
	t.learning.mu.RLock()
	defer t.learning.mu.RUnlock()

	return len(t.learning.successPatterns), len(t.learning.failurePatterns)
}
