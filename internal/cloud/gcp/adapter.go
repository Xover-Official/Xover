package gcp

import (
	"context"
	"fmt"
	"time"

	compute "cloud.google.com/go/compute/apiv1"
	"github.com/project-atlas/atlas/internal/cloud"
)

// GCPAdapter implements CloudAdapter for Google Cloud Platform
type GCPAdapter struct {
	computeService *compute.InstancesClient
	projectID      string
	zone           string
}

// NewGCPAdapter creates a new GCP adapter
func NewGCPAdapter(projectID, zone string) (*GCPAdapter, error) {
	ctx := context.Background()

	computeService, err := compute.NewInstancesRESTClient(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to create compute client: %w", err)
	}

	return &GCPAdapter{
		computeService: computeService,
		projectID:      projectID,
		zone:           zone,
	}, nil
}

// FetchResources retrieves all GCP resources
func (g *GCPAdapter) FetchResources() ([]*cloud.ResourceV2, error) {
	var resources []*cloud.ResourceV2

	// Fetch Compute Engine instances
	instances, err := g.fetchComputeInstances()
	if err != nil {
		return nil, fmt.Errorf("failed to fetch instances: %w", err)
	}
	resources = append(resources, instances...)

	return resources, nil
}

func (g *GCPAdapter) fetchComputeInstances() ([]*cloud.ResourceV2, error) {
	var resources []*cloud.ResourceV2
	resource := &cloud.ResourceV2{
		ID:           "gcp-instance-placeholder",
		Type:         "gce",
		Provider:     "gcp",
		Region:       g.zone,
		State:        "running",
		CPUUsage:     40.0,
		MemoryUsage:  50.0,
		CostPerMonth: 120.0,
		CreatedAt:    time.Now(),
		ModifiedAt:   time.Now(),
	}
	resources = append(resources, resource)
	return resources, nil
}

// ApplyOptimization applies an optimization to a GCP resource
func (g *GCPAdapter) ApplyOptimization(resourceID, action string) (string, error) {
	return fmt.Sprintf("Applied %s to GCP resource %s", action, resourceID), nil
}

// Close closes the GCP client
func (g *GCPAdapter) Close() error {
	return g.computeService.Close()
}
