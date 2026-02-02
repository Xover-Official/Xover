package tests

import (
	"context"
	"testing"
	"time"

	"github.com/project-atlas/atlas/internal/cloud"
	"github.com/project-atlas/atlas/internal/cloud/aws"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestAWSAdapterIntegration tests real AWS integration
func TestAWSAdapterIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	ctx := context.Background()

	// Test configuration
	cfg := cloud.CloudConfig{
		Region: "us-east-1",
		DryRun: false, // Use real API calls
	}

	// Create adapter
	adapter, err := aws.New(ctx, cfg)
	require.NoError(t, err)
	require.NotNil(t, adapter)

	// Test FetchResources
	resources, err := adapter.FetchResources(ctx)
	require.NoError(t, err)
	assert.NotEmpty(t, resources)

	// Verify resource structure
	for _, res := range resources {
		assert.NotEmpty(t, res.ID)
		assert.NotEmpty(t, res.Type)
		assert.NotEmpty(t, res.Provider)
		assert.GreaterOrEqual(t, res.CPUUsage, 0.0)
		assert.GreaterOrEqual(t, res.MemoryUsage, 0.0)
		assert.Greater(t, res.CostPerMonth, 0.0)
	}

	t.Logf("Successfully fetched %d resources from AWS", len(resources))
}

// TestCloudAdapterInterface ensures all adapters implement the interface correctly
func TestCloudAdapterInterface(t *testing.T) {
	ctx := context.Background()

	cfg := cloud.CloudConfig{
		Region: "us-east-1",
		DryRun: true,
	}

	adapter, err := aws.New(ctx, cfg)
	require.NoError(t, err)

	// Test interface methods
	t.Run("GetSpotPrice", func(t *testing.T) {
		price, err := adapter.GetSpotPrice("us-east-1a", "t3.micro")
		assert.NoError(t, err)
		assert.Greater(t, price, 0.0)
	})

	t.Run("ListZones", func(t *testing.T) {
		zones, err := adapter.ListZones()
		assert.NoError(t, err)
		assert.NotEmpty(t, zones)
	})

	t.Run("ApplyOptimization", func(t *testing.T) {
		resource := &cloud.ResourceV2{
			ID:           "i-test123",
			Type:         "ec2",
			Provider:     "aws",
			CostPerMonth: 100.0,
		}

		savings, err := adapter.ApplyOptimization(ctx, resource, "stop")
		assert.NoError(t, err)
		assert.GreaterOrEqual(t, savings, 0.0)
	})
}

// BenchmarkAWSFetchResources benchmarks resource fetching performance
func BenchmarkAWSFetchResources(b *testing.B) {
	ctx := context.Background()
	cfg := cloud.CloudConfig{
		Region: "us-east-1",
		DryRun: true,
	}

	adapter, err := aws.New(ctx, cfg)
	if err != nil {
		b.Fatalf("Failed to create adapter: %v", err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := adapter.FetchResources(ctx)
		if err != nil {
			b.Fatalf("Failed to fetch resources: %v", err)
		}
	}
}

// TestResourceValidation tests resource data integrity
func TestResourceValidation(t *testing.T) {
	ctx := context.Background()
	cfg := cloud.CloudConfig{
		Region: "us-east-1",
		DryRun: true,
	}

	adapter, err := aws.New(ctx, cfg)
	require.NoError(t, err)

	resources, err := adapter.FetchResources(ctx)
	require.NoError(t, err)

	for _, res := range resources {
		// Validate required fields
		assert.NotEmpty(t, res.ID, "Resource ID should not be empty")
		assert.NotEmpty(t, res.Type, "Resource type should not be empty")
		assert.NotEmpty(t, res.Provider, "Provider should not be empty")
		assert.NotEmpty(t, res.Region, "Region should not be empty")

		// Validate metrics
		assert.GreaterOrEqual(t, res.CPUUsage, 0.0, "CPU usage should be non-negative")
		assert.LessOrEqual(t, res.CPUUsage, 100.0, "CPU usage should not exceed 100")
		assert.GreaterOrEqual(t, res.MemoryUsage, 0.0, "Memory usage should be non-negative")
		assert.LessOrEqual(t, res.MemoryUsage, 100.0, "Memory usage should not exceed 100")

		// Validate cost
		assert.Greater(t, res.CostPerMonth, 0.0, "Monthly cost should be positive")

		// Validate tags
		assert.NotNil(t, res.Tags, "Tags should not be nil")
	}
}

// TestConcurrentAccess tests thread safety of adapter operations
func TestConcurrentAccess(t *testing.T) {
	ctx := context.Background()
	cfg := cloud.CloudConfig{
		Region: "us-east-1",
		DryRun: true,
	}

	adapter, err := aws.New(ctx, cfg)
	require.NoError(t, err)

	const numGoroutines = 10
	const numIterations = 5

	done := make(chan bool, numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		go func(id int) {
			for j := 0; j < numIterations; j++ {
				_, err := adapter.FetchResources(ctx)
				assert.NoError(t, err)

				price, err := adapter.GetSpotPrice("us-east-1a", "t3.micro")
				assert.NoError(t, err)
				assert.Greater(t, price, 0.0)
			}
			done <- true
		}(i)
	}

	// Wait for all goroutines to complete
	for i := 0; i < numGoroutines; i++ {
		select {
		case <-done:
		case <-time.After(30 * time.Second):
			t.Fatal("Test timed out waiting for goroutines")
		}
	}
}
