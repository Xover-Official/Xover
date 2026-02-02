package cloud

import (
	"context"
	"fmt"
	"time"
)

// Simulator implements the CloudAdapter interface for testing and simulation.
type Simulator struct {
	MockResources []*ResourceV2
}

func NewSimulator() *Simulator {
	return &Simulator{
		MockResources: []*ResourceV2{
			{
				ID:           "db-prod-01",
				Type:         ResourceTypeRDS,
				Provider:     ProviderAWS,
				Region:       "us-east-1",
				State:        "running",
				CPUUsage:     0.155,
				MemoryUsage:  0.220,
				CostPerMonth: 450.00,
				CreatedAt:    time.Now().Add(-30 * 24 * time.Hour),
			},
			{
				ID:           "web-prod-01",
				Type:         ResourceTypeEC2,
				Provider:     ProviderAWS,
				Region:       "us-east-1",
				State:        "running",
				CPUUsage:     0.452,
				MemoryUsage:  0.678,
				CostPerMonth: 125.00,
				CreatedAt:    time.Now().Add(-15 * 24 * time.Hour),
			},
		},
	}
}

func (s *Simulator) FetchResources(ctx context.Context) ([]*ResourceV2, error) {
	return s.MockResources, nil
}

func (s *Simulator) GetResource(ctx context.Context, id string) (*ResourceV2, error) {
	for _, r := range s.MockResources {
		if r.ID == id {
			return r, nil
		}
	}
	return nil, fmt.Errorf("resource not found: %s", id)
}

func (s *Simulator) ApplyOptimization(ctx context.Context, resource *ResourceV2, action string) (float64, error) {
	// Simulate savings: 50% for resize/optimize, 100% for stop/terminate
	switch action {
	case "stop", "terminate":
		return resource.CostPerMonth, nil
	case "resize", "optimize":
		return resource.CostPerMonth * 0.5, nil
	default:
		return 0, nil
	}
}

func (s *Simulator) GetSpotPrice(zone, instanceType string) (float64, error) {
	return 0.05, nil // Static mock price
}

func (s *Simulator) ListZones() ([]string, error) {
	return []string{"us-east-1a", "us-east-1b", "us-east-1c"}, nil
}
