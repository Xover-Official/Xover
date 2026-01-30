package main

import (
	"fmt"
	"log"

	"github.com/project-atlas/atlas/internal/cloud"
	"github.com/project-atlas/atlas/internal/risk"
)

func main() {
	fmt.Println("=== Project Atlas: Risk Analysis Demo ===")

	// Setup
	riskEngine := risk.NewEngine()
	simulator := cloud.NewSimulator()

	// List mock resources
	resources, err := simulator.ListResources()
	if err != nil {
		log.Fatalf("Failed to list resources: %v", err)
	}

	for _, res := range resources {
		fmt.Printf("\nAnalyzing Resource: %s (%s)\n", res.ID, res.CurrentType)
		fmt.Printf("Monthly Cost: $%.2f | CPU: %.1f%% | Mem: %.1f%%\n", 
			res.MonthlyCost, res.Metrics.CPUUsage, res.Metrics.MemoryUsage)

		// Calculate impact (mock: 30% savings if downsized)
		projectedImpact := res.MonthlyCost * 0.30

		// Analyze risk
		result := riskEngine.CalculateScore(projectedImpact, res.Metrics)

		fmt.Printf("--- Recommendation ---\n")
		fmt.Printf("Status: ")
		if res.Metrics.IsSafeForSizing(40, 50) {
			fmt.Println("✅ RECOMMENDED")
		} else {
			fmt.Println("❌ SKIP (Too High Risk)")
		}

		fmt.Printf("Projected Impact: $%.2f/mo\n", result.Impact)
		fmt.Printf("Risk Factor: %.2f\n", result.Risk)
		fmt.Printf("Confidence: %.2f\n", result.Confidence)
		fmt.Printf("Final Atlas Score: %.4f\n", result.Score)
	}
}
