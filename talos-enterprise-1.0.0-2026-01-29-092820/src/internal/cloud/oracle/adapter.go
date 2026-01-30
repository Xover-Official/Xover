package oracle

import (
	"context"
	"fmt"

	"github.com/oracle/oci-go-sdk/v65/common"
	"github.com/oracle/oci-go-sdk/v65/core"
	"github.com/oracle/oci-go-sdk/v65/identity"
)

// OracleAdapter provides Oracle Cloud Infrastructure integration
type OracleAdapter struct {
	ComputeClient        *core.ComputeClient
	VirtualNetworkClient *core.VirtualNetworkClient
	IdentityClient       *identity.IdentityClient
	TenancyOCID          string
	CompartmentID        string
}

// NewOracleAdapter creates a new Oracle Cloud adapter
func NewOracleAdapter(configPath string, tenancyOCID, compartmentID string) (*OracleAdapter, error) {
	configProvider := common.DefaultConfigProvider()

	computeClient, err := core.NewComputeClientWithConfigurationProvider(configProvider)
	if err != nil {
		return nil, fmt.Errorf("failed to create compute client: %w", err)
	}

	vnClient, err := core.NewVirtualNetworkClientWithConfigurationProvider(configProvider)
	if err != nil {
		return nil, fmt.Errorf("failed to create network client: %w", err)
	}

	identityClient, err := identity.NewIdentityClientWithConfigurationProvider(configProvider)
	if err != nil {
		return nil, fmt.Errorf("failed to create identity client: %w", err)
	}

	return &OracleAdapter{
		ComputeClient:        &computeClient,
		VirtualNetworkClient: &vnClient,
		IdentityClient:       &identityClient,
		TenancyOCID:          tenancyOCID,
		CompartmentID:        compartmentID,
	}, nil
}

// DiscoverResources discovers Oracle Cloud resources
func (a *OracleAdapter) DiscoverResources(ctx context.Context) ([]Resource, error) {
	resources := make([]Resource, 0)

	// Discover compute instances
	instances, err := a.listInstances(ctx)
	if err != nil {
		return nil, err
	}

	for _, instance := range instances {
		resource := Resource{
			ID:       *instance.Id,
			Name:     *instance.DisplayName,
			Type:     "compute",
			Provider: "oracle",
			Region:   *instance.Region,
			State:    string(instance.LifecycleState),
			Tags:     convertTags(instance.FreeformTags),
		}

		// Estimate cost based on shape
		resource.CostPerMonth = a.estimateInstanceCost(*instance.Shape)

		resources = append(resources, resource)
	}

	// Discover block volumes
	volumes, err := a.listBlockVolumes(ctx)
	if err != nil {
		return nil, err
	}

	for _, volume := range volumes {
		resource := Resource{
			ID:       *volume.Id,
			Name:     *volume.DisplayName,
			Type:     "block-storage",
			Provider: "oracle",
			State:    string(volume.LifecycleState),
			Tags:     convertTags(volume.FreeformTags),
		}

		// Cost: ~$0.0255/GB/month
		if volume.SizeInGBs != nil {
			resource.CostPerMonth = float64(*volume.SizeInGBs) * 0.0255
		}

		resources = append(resources, resource)
	}

	return resources, nil
}

// listInstances lists all compute instances
func (a *OracleAdapter) listInstances(ctx context.Context) ([]core.Instance, error) {
	request := core.ListInstancesRequest{
		CompartmentId: &a.CompartmentID,
	}

	response, err := a.ComputeClient.ListInstances(ctx, request)
	if err != nil {
		return nil, err
	}

	return response.Items, nil
}

// listBlockVolumes lists all block volumes
func (a *OracleAdapter) listBlockVolumes(ctx context.Context) ([]core.Volume, error) {
	blockStorageClient, err := core.NewBlockstorageClientWithConfigurationProvider(
		common.DefaultConfigProvider(),
	)
	if err != nil {
		return nil, err
	}

	request := core.ListVolumesRequest{
		CompartmentId: &a.CompartmentID,
	}

	response, err := blockStorageClient.ListVolumes(ctx, request)
	if err != nil {
		return nil, err
	}

	return response.Items, nil
}

// estimateInstanceCost estimates monthly cost based on shape
func (a *OracleAdapter) estimateInstanceCost(shape string) float64 {
	// Simplified pricing - actual pricing varies by region
	pricing := map[string]float64{
		"VM.Standard2.1":      50.00,  // 1 OCPU, 15GB RAM
		"VM.Standard2.2":      100.00, // 2 OCPU, 30GB RAM
		"VM.Standard2.4":      200.00, // 4 OCPU, 60GB RAM
		"VM.Standard.E4.Flex": 30.00,  // Flex shape base
	}

	if cost, ok := pricing[shape]; ok {
		return cost
	}

	return 75.00 // Default estimate
}

// OptimizeInstance provides optimization recommendations
func (a *OracleAdapter) OptimizeInstance(ctx context.Context, instanceID string) (*Optimization, error) {
	// Get instance details
	request := core.GetInstanceRequest{
		InstanceId: &instanceID,
	}

	response, err := a.ComputeClient.GetInstance(ctx, request)
	if err != nil {
		return nil, err
	}

	instance := response.Instance

	// Check if instance is underutilized
	// (In production, fetch actual utilization metrics)
	optimization := &Optimization{
		ResourceID:       instanceID,
		ResourceType:     "compute",
		Recommendation:   fmt.Sprintf("Consider downsizing from %s", *instance.Shape),
		EstimatedSavings: 25.00,
		RiskScore:        4.0,
		Provider:         "oracle",
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

func convertTags(ociTags map[string]string) map[string]string {
	tags := make(map[string]string)
	for k, v := range ociTags {
		tags[k] = v
	}
	return tags
}
