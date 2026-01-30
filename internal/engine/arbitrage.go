package engine

import (
	"fmt"
	"github.com/project-atlas/atlas/internal/cloud"
	"github.com/project-atlas/atlas/internal/logger"
)

// ActionPlan represents a proposed optimization action
type ActionPlan struct {
	Type        string
	Description string
	RiskScore   float64
	ImpactScore float64
	EstSavings  float64
}

// ArbitrageEngine hunts for price discrepancies between Availability Zones
type ArbitrageEngine struct {
	Adapter cloud.CloudAdapter
}

func NewArbitrageEngine(adapter cloud.CloudAdapter) *ArbitrageEngine {
	return &ArbitrageEngine{Adapter: adapter}
}

// FindArbitrageOpportunity checks if moving a workload can save > 20%
func (e *ArbitrageEngine) FindArbitrageOpportunity(currentZone string, instanceType string) (*ActionPlan, error) {
	// 1. Get current price
	currentPrice, err := e.Adapter.GetSpotPrice(currentZone, instanceType)
	if err != nil {
		return nil, err
	}

	// 2. Scan all other zones in the region
	allZones, err := e.Adapter.ListZones()
	if err != nil {
		return nil, err
	}

	var bestZone string
	minPrice := currentPrice

	for _, zone := range allZones {
		if zone == currentZone {
			continue
		}
		price, err := e.Adapter.GetSpotPrice(zone, instanceType)
		if err != nil {
			continue
		}
		if price < minPrice {
			minPrice = price
			bestZone = zone
		}
	}

	// 3. Calculate detailed ROI
	if bestZone == "" {
		return nil, nil // No cheaper zone found
	}

	savings := currentPrice - minPrice
	percentSaving := (savings / currentPrice) * 100

	// Threshold: Only move if savings > 20% to account for migration overhead
	if percentSaving > 20.0 {
		logger.LogAction(logger.Architect, "ArbitrageHunt", "SUCCESS", 
			fmt.Sprintf("Found %.1f%% savings: %s -> %s", percentSaving, currentZone, bestZone))

		return &ActionPlan{
			Type:        "MIGRATE_ZONE",
			Description: fmt.Sprintf("Move %s from %s ($%.3f) to %s ($%.3f)", instanceType, currentZone, currentPrice, bestZone, minPrice),
			RiskScore:   4.5, // Moderate risk due to service interruption during drain
			ImpactScore: 8.0, // High financial impact
			EstSavings:  savings * 730, // Monthly savings (approx hours/month)
		}, nil
	}

	return nil, nil // No worthwhile arbitrage found
}
