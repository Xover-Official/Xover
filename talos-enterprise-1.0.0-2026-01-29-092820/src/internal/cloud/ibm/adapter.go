package ibm

import (
	"context"
	"fmt"

	"github.com/IBM/go-sdk-core/v5/core"
	"github.com/IBM/vpc-go-sdk/vpcv1"
)

// IBMCloudAdapter provides IBM Cloud integration
type IBMCloudAdapter struct {
	VPCService    *vpcv1.VpcV1
	ResourceGroup string
	Region        string
}

// NewIBMCloudAdapter creates a new IBM Cloud adapter
func NewIBMCloudAdapter(apiKey, resourceGroup, region string) (*IBMCloudAdapter, error) {
	authenticator := &core.IamAuthenticator{
		ApiKey: apiKey,
	}

	vpcService, err := vpcv1.NewVpcV1(&vpcv1.VpcV1Options{
		Authenticator: authenticator,
		URL:           fmt.Sprintf("https://%s.iaas.cloud.ibm.com/v1", region),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create VPC service: %w", err)
	}

	return &IBMCloudAdapter{
		VPCService:    vpcService,
		ResourceGroup: resourceGroup,
		Region:        region,
	}, nil
}

// DiscoverResources discovers IBM Cloud resources
func (a *IBMCloudAdapter) DiscoverResources(ctx context.Context) ([]Resource, error) {
	resources := make([]Resource, 0)

	// Discover virtual server instances
	instances, err := a.listInstances(ctx)
	if err != nil {
		return nil, err
	}

	for _, instance := range instances {
		resource := Resource{
			ID:       *instance.ID,
			Name:     *instance.Name,
			Type:     "virtual-server",
			Provider: "ibm",
			Region:   a.Region,
			State:    *instance.Status,
			Tags:     make(map[string]string), // TODO: Extract actual tags when IBM VPC API is available
		}

		// Estimate cost based on profile
		if instance.Profile != nil && instance.Profile.Name != nil {
			resource.CostPerMonth = a.estimateInstanceCost(*instance.Profile.Name)
		}

		resources = append(resources, resource)
	}

	// Discover block storage volumes
	volumes, err := a.listVolumes(ctx)
	if err != nil {
		return nil, err
	}

	for _, volume := range volumes {
		resource := Resource{
			ID:       *volume.ID,
			Name:     *volume.Name,
			Type:     "block-storage",
			Provider: "ibm",
			Region:   a.Region,
			State:    *volume.Status,
			Tags:     extractTags(volume.UserTags),
		}

		// IBM Cloud Block Storage: ~$0.13/GB/month (general purpose)
		if volume.Capacity != nil {
			resource.CostPerMonth = float64(*volume.Capacity) * 0.13
		}

		resources = append(resources, resource)
	}

	return resources, nil
}

// listInstances lists virtual server instances
func (a *IBMCloudAdapter) listInstances(ctx context.Context) ([]vpcv1.Instance, error) {
	options := &vpcv1.ListInstancesOptions{}

	result, _, err := a.VPCService.ListInstances(options)
	if err != nil {
		return nil, err
	}

	return result.Instances, nil
}

// listVolumes lists block storage volumes
func (a *IBMCloudAdapter) listVolumes(ctx context.Context) ([]vpcv1.Volume, error) {
	options := &vpcv1.ListVolumesOptions{}

	result, _, err := a.VPCService.ListVolumes(options)
	if err != nil {
		return nil, err
	}

	return result.Volumes, nil
}

// estimateInstanceCost estimates monthly cost based on profile
func (a *IBMCloudAdapter) estimateInstanceCost(profile string) float64 {
	// Simplified pricing - actual varies by region
	pricing := map[string]float64{
		"cx2-2x4":  73.00,  // 2 vCPU, 4GB RAM
		"cx2-4x8":  146.00, // 4 vCPU, 8GB RAM
		"cx2-8x16": 292.00, // 8 vCPU, 16GB RAM
		"bx2-2x8":  88.00,  // 2 vCPU, 8GB RAM (balanced)
		"mx2-4x32": 233.00, // 4 vCPU, 32GB RAM (memory)
	}

	if cost, ok := pricing[profile]; ok {
		return cost
	}

	return 100.00 // Default estimate
}

// OptimizeInstance provides optimization recommendations
func (a *IBMCloudAdapter) OptimizeInstance(ctx context.Context, instanceID string) (*Optimization, error) {
	options := &vpcv1.GetInstanceOptions{
		ID: &instanceID,
	}

	instance, _, err := a.VPCService.GetInstance(options)
	if err != nil {
		return nil, err
	}

	optimization := &Optimization{
		ResourceID:   instanceID,
		ResourceType: "virtual-server",
		Recommendation: fmt.Sprintf("Consider rightsizing from profile %s",
			*instance.Profile.Name),
		EstimatedSavings: 30.00,
		RiskScore:        3.5,
		Provider:         "ibm",
	}

	return optimization, nil
}

// Helper types
type Resource struct {
	ID           string
	Name         string
	Type         string
	Provider     string
	Region       string
	State        string
	CostPerMonth float64
	Tags         map[string]string
}

type Optimization struct {
	ResourceID       string
	ResourceType     string
	Recommendation   string
	EstimatedSavings float64
	RiskScore        float64
	Provider         string
}

func extractTags(tags []string) map[string]string {
	tagMap := make(map[string]string)
	for _, tag := range tags {
		tagMap[tag] = "true"
	}
	return tagMap
}
