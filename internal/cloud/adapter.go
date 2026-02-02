package cloud

import (
	"context"
)

// Provider constants
const (
	ProviderAWS   = "aws"
	ProviderAzure = "azure"
	ProviderGCP   = "gcp"
)

// Resource type constants
const (
	ResourceTypeEC2     = "ec2"
	ResourceTypeRDS     = "rds"
	ResourceTypeVM      = "vm"
	ResourceTypeStorage = "storage"
	ResourceTypeNetwork = "network"
)

// CloudConfig defines the configuration for a cloud provider adapter.
type CloudConfig struct {
	Provider string
	Region   string
	APIKey   string
	DryRun   bool
}

// CloudAdapter is the interface that all cloud providers must implement.
// It uses the canonical ResourceV2 model for all operations.
type CloudAdapter interface {
	FetchResources(ctx context.Context) ([]*ResourceV2, error)
	GetResource(ctx context.Context, id string) (*ResourceV2, error)
	ApplyOptimization(ctx context.Context, resource *ResourceV2, action string) (float64, error)
	GetSpotPrice(zone, instanceType string) (float64, error)
	ListZones() ([]string, error)
}
