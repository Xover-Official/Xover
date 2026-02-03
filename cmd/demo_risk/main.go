package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/Xover-Official/Xover/internal/cloud"
	"github.com/Xover-Official/Xover/internal/risk"
)

func main() {
	fmt.Println("=== Project Atlas: Risk Analysis Demo ===")

	// Setup
	riskEngine := risk.NewEngine()
	simulator := cloud.NewSimulator()

	// List mock resources
	resources, err := simulator.FetchResources(context.Background())
	if err != nil {
		log.Fatalf("Failed to list resources: %v", err)
	}

	if len(resources) == 0 {
		fmt.Println("No resources found for analysis")
		return
	}

	for _, res := range resources {
		fmt.Printf("\nAnalyzing Resource: %s (%s)\n", res.ID, res.Type)
		fmt.Printf("Monthly Cost: $%.2f | CPU: %.1f%% | Mem: %.1f%%\n",
			res.CostPerMonth, res.CPUUsage, res.MemoryUsage)

		// Calculate impact (mock: 30% savings if downsized)
		projectedImpact := res.CostPerMonth * 0.30

		// Analyze risk
		metrics := risk.CloudMetrics{
			CPUUsage:    res.CPUUsage,
			MemoryUsage: res.MemoryUsage,
			MeasuredAt:  time.Now(),
		}

		result := riskEngine.CalculateScore(projectedImpact, metrics)

		fmt.Printf("--- Recommendation ---\n")
		fmt.Printf("Status: ")
		if res.CPUUsage < 40 && res.MemoryUsage < 50 {
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
