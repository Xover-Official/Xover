package main

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/Xover-Official/Xover/internal/ai"
	"github.com/Xover-Official/Xover/internal/cloud"
)

func main() {
	fmt.Println("🌹 ROSES/T.O.P.A.Z. Framework Demo")
	fmt.Println("=====================================")

	// Create T.O.P.A.Z. Orchestrator
	config := &ai.Config{
		GeminiAPIKey: "your-gemini-api-key",
		ClaudeAPIKey: "your-claude-api-key",
		CacheEnabled: true,
	}

	orchestrator, err := ai.NewTOPAZOrchestrator(config, nil, nil)
	if err != nil {
		log.Fatalf("Failed to create T.O.P.A.Z. orchestrator: %v", err)
	}

	// Example resources to analyze
	resources := []*cloud.ResourceV2{
		{
			ID:           "db-prod-01",
			Type:         "r6g.large",
			Provider:     "aws",
			Region:       "us-east-1",
			CPUUsage:     15.2,
			MemoryUsage:  22.8,
			CostPerMonth: 150.0,
			Tags: map[string]string{
				"environment":  "production",
				"anti-fragile": "true",
				"auto-scaling": "enabled",
				"redundancy":   "high",
			},
		},
		{
			ID:           "web-server-03",
			Type:         "t3.medium",
			Provider:     "aws",
			Region:       "us-east-1",
			CPUUsage:     85.7,
			MemoryUsage:  78.3,
			CostPerMonth: 75.0,
			Tags: map[string]string{
				"environment":   "production",
				"load-balancer": "true",
			},
		},
		{
			ID:           "batch-worker-07",
			Type:         "c5.large",
			Provider:     "aws",
			Region:       "us-east-1",
			CPUUsage:     8.1,
			MemoryUsage:  12.4,
			CostPerMonth: 85.0,
			Tags: map[string]string{
				"environment": "staging",
				"scheduled":   "true",
			},
		},
	}

	ctx := context.Background()

	// Analyze each resource using ROSES/T.O.P.A.Z.
	for i, resource := range resources {
		fmt.Printf("\n🔍 Analyzing Resource %d: %s\n", i+1, resource.ID)
		fmt.Println(strings.Repeat("=", 50))

		// Prepare context data
		contextData := map[string]interface{}{
			"current_time":    time.Now().Format("2006-01-02 15:04:05"),
			"analysis_type":   "comprehensive",
			"sla_requirement": "99.9%",
			"business_impact": "high",
		}

		// Perform ROSES analysis
		decision, err := orchestrator.AnalyzeWithROSES(ctx, resource, contextData)
		if err != nil {
			log.Printf("Failed to analyze resource %s: %v", resource.ID, err)
			continue
		}

		// Display results
		displayROSESResults(resource, decision)
	}

	// Batch analysis example
	fmt.Println("\n🚀 Batch Analysis Example")
	fmt.Println("=========================")

	decisions, err := orchestrator.BatchAnalyzeWithROSES(ctx, resources)
	if err != nil {
		log.Printf("Batch analysis failed: %v", err)
		return
	}

	// Display batch summary
	displayBatchSummary(decisions)

	// Learning insights
	fmt.Println("\n📊 Learning Insights")
	fmt.Println("===================")

	insights := orchestrator.GetLearningInsights()
	for key, value := range insights {
		fmt.Printf("%s: %v\n", key, value)
	}
}

func displayROSESResults(resource *cloud.ResourceV2, decision *ai.TOPAZDecision) {
	fmt.Printf("📋 Resource: %s (%s)\n", resource.ID, resource.Type)
	fmt.Printf("💰 Monthly Cost: $%.2f\n", resource.CostPerMonth)
	fmt.Printf("📊 CPU: %.1f%% | Memory: %.1f%%\n", resource.CPUUsage, resource.MemoryUsage)

	fmt.Printf("\n🎯 ROSES/T.O.P.A.Z. Analysis:\n")
	fmt.Printf("📈 Risk Score: %.1f/100\n", decision.RiskScore)
	fmt.Printf("🎪 Recommendation: %s\n", decision.Recommendation)
	fmt.Printf("✅ Go/No-Go: %s\n", decision.GoNoGo)
	fmt.Printf("💡 Confidence: %.1f%%\n", decision.Confidence*100)
	fmt.Printf("💸 Expected Savings: $%.2f\n", decision.ExpectedSavings)
	fmt.Printf("🛡️ Anti-Fragile Score: %.1f/100\n", decision.AntiFragileScore)

	fmt.Printf("\n🧠 Reasoning:\n")
	for i, reason := range decision.Reasoning {
		fmt.Printf("  %d. %s\n", i+1, reason)
	}

	if len(decision.Metadata) > 0 {
		fmt.Printf("\n📋 Metadata:\n")
		for key, value := range decision.Metadata {
			fmt.Printf("  %s: %v\n", key, value)
		}
	}

	fmt.Println()
}

func displayBatchSummary(decisions []*ai.TOPAZDecision) {
	totalResources := len(decisions)
	goDecisions := 0
	noGoDecisions := 0
	totalSavings := 0.0
	avgRisk := 0.0

	for _, decision := range decisions {
		if decision.GoNoGo == "Go" {
			goDecisions++
		} else {
			noGoDecisions++
		}
		totalSavings += decision.ExpectedSavings
		avgRisk += decision.RiskScore
	}

	if totalResources > 0 {
		avgRisk /= float64(totalResources)
	}

	fmt.Printf("📊 Batch Analysis Summary:\n")
	fmt.Printf("  Total Resources: %d\n", totalResources)
	fmt.Printf("  Go Decisions: %d\n", goDecisions)
	fmt.Printf("  No-Go Decisions: %d\n", noGoDecisions)
	fmt.Printf("  Total Expected Savings: $%.2f\n", totalSavings)
	fmt.Printf("  Average Risk Score: %.1f\n", avgRisk)
	fmt.Printf("  Success Rate: %.1f%%\n", float64(goDecisions)/float64(totalResources)*100)
}

// Example of how to generate a ROSES prompt manually
func demonstrateROSESPromptGeneration() {
	fmt.Println("\n🌹 ROSES Prompt Generation Example")
	fmt.Println("=================================")

	roses := ai.NewROSESFramework()

	resource := &cloud.ResourceV2{
		ID:           "db-prod-01",
		Type:         "r6g.large",
		Provider:     "aws",
		Region:       "us-east-1",
		CPUUsage:     15.2,
		MemoryUsage:  22.8,
		CostPerMonth: 150.0,
		Tags: map[string]string{
			"environment":  "production",
			"anti-fragile": "true",
		},
	}

	contextData := map[string]interface{}{
		"sla_requirement": "99.9%",
		"business_impact": "high",
		"current_time":    time.Now().Format("2006-01-02 15:04:05"),
	}

	prompt := roses.GenerateROSESPrompt(resource, contextData)

	fmt.Println("Generated ROSES Prompt:")
	fmt.Println("========================")
	fmt.Println(prompt)
}
