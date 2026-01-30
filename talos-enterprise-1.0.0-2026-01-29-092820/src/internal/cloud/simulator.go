package cloud

import (
	"fmt"
	"time"

	"github.com/project-atlas/atlas/internal/risk"
)

type Provider interface {
	ListResources() ([]*ResourceV2, error)
	GetMetrics(resourceID string) (risk.CloudMetrics, error)
	ApplyOptimization(resourceID string, newType string) (string, error)
}

type Simulator struct {
	MockResources []*ResourceV2
}

func NewSimulator() *Simulator {
	return &Simulator{
		MockResources: []*ResourceV2{
			{
				ID:           "db-prod-01",
				Type:         "rds",
				Provider:     "aws",
				Region:       "us-east-1",
				State:        "running",
				CPUUsage:     15.5,
				MemoryUsage:  22.0,
				CostPerMonth: 450.00,
				CreatedAt:    time.Now().Add(-30 * 24 * time.Hour),
				ModifiedAt:   time.Now(),
			},
			{
				ID:           "web-prod-01",
				Type:         "ec2",
				Provider:     "aws",
				Region:       "us-east-1",
				State:        "running",
				CPUUsage:     45.2,
				MemoryUsage:  67.8,
				CostPerMonth: 125.00,
				CreatedAt:    time.Now().Add(-15 * 24 * time.Hour),
				ModifiedAt:   time.Now(),
			},
		},
	}
}

func (s *Simulator) ListResources() ([]*ResourceV2, error) {
	return s.MockResources, nil
}

func (s *Simulator) GetMetrics(resourceID string) (risk.CloudMetrics, error) {
	for _, r := range s.MockResources {
		if r.ID == resourceID {
			return risk.CloudMetrics{
				CPUUsage:    r.CPUUsage,
				MemoryUsage: r.MemoryUsage,
				MeasuredAt:  time.Now(),
			}, nil
		}
	}
	return risk.CloudMetrics{}, fmt.Errorf("resource not found: %s", resourceID)
}

func (s *Simulator) ApplyOptimization(resourceID string, newType string) (string, error) {
	// Simulate an optimization action
	return fmt.Sprintf("sim-opt-%s-to-%s", resourceID, newType), nil
}
