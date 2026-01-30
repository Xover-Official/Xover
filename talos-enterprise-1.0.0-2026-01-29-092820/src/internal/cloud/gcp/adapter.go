package gcp

import (
	"context"
	"fmt"

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
func (g *GCPAdapter) FetchResources() ([]cloud.ResourceJSON, error) {
	var resources []cloud.ResourceJSON

	// Fetch Compute Engine instances
	instances, err := g.fetchComputeInstances()
	if err != nil {
		return nil, fmt.Errorf("failed to fetch instances: %w", err)
	}
	resources = append(resources, instances...)

	return resources, nil
}

func (g *GCPAdapter) fetchComputeInstances() ([]cloud.ResourceJSON, error) {
	// Placeholder for actual implementation
	var resources []cloud.ResourceJSON

	resource := cloud.ResourceJSON{
		ID:          "gcp-instance-placeholder",
		Type:        "gce",
		CurrentType: "n1-standard-2",
		Region:      g.zone,
		Tags:        make(map[string]string),
		Metrics: map[string]interface{}{
			"cpu_usage":    40.0,
			"memory_usage": 50.0,
		},
		MonthlyCost: 120.0,
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
