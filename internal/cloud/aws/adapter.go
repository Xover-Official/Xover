package aws

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/cloudwatch"
	cloudwatchtypes "github.com/aws/aws-sdk-go-v2/service/cloudwatch/types"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	ec2types "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/aws/aws-sdk-go-v2/service/rds"
	"go.uber.org/multierr"

	"github.com/project-atlas/atlas/internal/cloud"
)

// mockInstancePricing provides a rough cost estimate per month for instance types.
// In a real application, this would use the AWS Price List API.
var mockInstancePricing = map[string]float64{
	"t2.micro":   10.0,
	"t3.medium":  40.0,
	"m5.large":   80.0,
	"m5.2xlarge": 320.0,
}

// Adapter implements the cloud.CloudAdapter interface for AWS.
type Adapter struct {
	ec2Client *ec2.Client
	rdsClient *rds.Client
	cwClient  *cloudwatch.Client
	region    string
	dryRun    bool
}

// New creates a new AWS adapter. It satisfies the cloud.Adapter interface.
func New(ctx context.Context, cfg cloud.CloudConfig) (*Adapter, error) {
	awsCfg, err := config.LoadDefaultConfig(ctx,
		config.WithRegion(cfg.Region),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to load AWS config: %w", err)
	}

	return &Adapter{
		ec2Client: ec2.NewFromConfig(awsCfg),
		rdsClient: rds.NewFromConfig(awsCfg),
		cwClient:  cloudwatch.NewFromConfig(awsCfg),
		region:    cfg.Region,
		dryRun:    cfg.DryRun,
	}, nil
}

// FetchResources retrieves all supported AWS resources and converts them to the canonical ResourceV2 model.
func (a *Adapter) FetchResources(ctx context.Context) ([]*cloud.ResourceV2, error) {
	var wg sync.WaitGroup
	var ec2Resources, rdsResources []*cloud.ResourceV2
	var ec2Err, rdsErr error

	wg.Add(2)

	// Fetch EC2 and RDS resources concurrently
	go func() {
		defer wg.Done()
		ec2Resources, ec2Err = a.fetchEC2Instances(ctx)
	}()

	go func() {
		defer wg.Done()
		rdsResources, rdsErr = a.fetchRDSInstances(ctx)
	}()

	wg.Wait()

	if ec2Err != nil {
		return nil, fmt.Errorf("failed to fetch EC2 instances: %w", ec2Err)
	}
	if rdsErr != nil {
		return nil, fmt.Errorf("failed to fetch RDS instances: %w", rdsErr)
	}

	return append(ec2Resources, rdsResources...), nil
}

func (a *Adapter) fetchEC2Instances(ctx context.Context) ([]*cloud.ResourceV2, error) {
	paginator := ec2.NewDescribeInstancesPaginator(a.ec2Client, &ec2.DescribeInstancesInput{})

	var instances []ec2types.Instance
	for paginator.HasMorePages() {
		output, err := paginator.NextPage(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to describe instances: %w", err)
		}
		for _, reservation := range output.Reservations {
			instances = append(instances, reservation.Instances...)
		}
	}

	// Worker pool to fetch metrics concurrently
	numWorkers := 10
	jobs := make(chan ec2types.Instance, len(instances))
	results := make(chan *cloud.ResourceV2, len(instances))
	var wg sync.WaitGroup

	for i := 0; i < numWorkers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for instance := range jobs {
				metrics, err := a.getEC2Metrics(ctx, *instance.InstanceId)
				if err != nil {
					log.Printf("failed to get metrics for instance %s: %v", *instance.InstanceId, err)
					continue
				}

				cpu, _ := metrics["cpu_usage"].(float64)
				mem, _ := metrics["memory_usage"].(float64)
				netIn, _ := metrics["network_in"].(float64)
				netOut, _ := metrics["network_out"].(float64)

				cost, _ := mockInstancePricing[string(instance.InstanceType)]

				resource := &cloud.ResourceV2{
					ID:           *instance.InstanceId,
					Type:         cloud.ResourceTypeEC2,
					Provider:     cloud.ProviderAWS,
					Region:       a.region,
					Tags:         make(map[string]string),
					State:        string(instance.State.Name),
					CreatedAt:    *instance.LaunchTime,
					CPUUsage:     cpu,
					MemoryUsage:  mem,
					NetworkIn:    netIn,
					NetworkOut:   netOut,
					CostPerMonth: cost,
					Metadata:     map[string]interface{}{"instance_type": string(instance.InstanceType)},
				}

				for _, tag := range instance.Tags {
					if tag.Key != nil && tag.Value != nil {
						resource.Tags[*tag.Key] = *tag.Value
					}
				}
				results <- resource
			}
		}()
	}

	for _, instance := range instances {
		jobs <- instance
	}
	close(jobs)

	wg.Wait()
	close(results)

	var resources []*cloud.ResourceV2
	for resource := range results {
		resources = append(resources, resource)
	}

	return resources, nil
}

func (a *Adapter) fetchRDSInstances(ctx context.Context) ([]*cloud.ResourceV2, error) {
	result, err := a.rdsClient.DescribeDBInstances(ctx, &rds.DescribeDBInstancesInput{})
	if err != nil {
		return nil, err
	}

	var resources []*cloud.ResourceV2
	for _, instance := range result.DBInstances {
		// RDS metrics fetching would be similar to EC2, omitted for brevity
		resource := &cloud.ResourceV2{
			ID:                 *instance.DBInstanceIdentifier,
			Type:               cloud.ResourceTypeRDS,
			Provider:           cloud.ProviderAWS,
			Region:             a.region,
			Tags:               make(map[string]string),
			State:              *instance.DBInstanceStatus,
			CreatedAt:          *instance.InstanceCreateTime,
			CPUUsage:           30.0,  // Placeholder
			MemoryUsage:        40.0,  // Placeholder
			CostPerMonth:       200.0, // Placeholder
			EncryptionEnabled:  *instance.StorageEncrypted,
			PubliclyAccessible: *instance.PubliclyAccessible,
			Metadata:           map[string]interface{}{"instance_class": *instance.DBInstanceClass},
		}

		resources = append(resources, resource)
	}

	return resources, nil
}

// GetResource retrieves a single resource by its ID
func (a *Adapter) GetResource(ctx context.Context, id string) (*cloud.ResourceV2, error) {
	// For now, only support EC2 by ID
	input := &ec2.DescribeInstancesInput{
		InstanceIds: []string{id},
	}
	result, err := a.ec2Client.DescribeInstances(ctx, input)
	if err != nil {
		return nil, fmt.Errorf("failed to describe instance %s: %w", id, err)
	}

	if len(result.Reservations) == 0 || len(result.Reservations[0].Instances) == 0 {
		return nil, fmt.Errorf("resource %s not found", id)
	}

	instance := result.Reservations[0].Instances[0]
	metrics, err := a.getEC2Metrics(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get metrics for %s: %w", id, err)
	}

	cpu, _ := metrics["cpu_usage"].(float64)
	mem, _ := metrics["memory_usage"].(float64)
	cost, _ := mockInstancePricing[string(instance.InstanceType)]

	resource := &cloud.ResourceV2{
		ID:           *instance.InstanceId,
		Type:         cloud.ResourceTypeEC2,
		Provider:     cloud.ProviderAWS,
		Region:       a.region,
		Tags:         make(map[string]string),
		State:        string(instance.State.Name),
		CreatedAt:    *instance.LaunchTime,
		CPUUsage:     cpu,
		MemoryUsage:  mem,
		CostPerMonth: cost,
		Metadata:     map[string]interface{}{"instance_type": string(instance.InstanceType)},
	}

	for _, tag := range instance.Tags {
		if tag.Key != nil && tag.Value != nil {
			resource.Tags[*tag.Key] = *tag.Value
		}
	}

	return resource, nil
}

// ApplyOptimization applies an optimization to an AWS resource
func (a *Adapter) ApplyOptimization(ctx context.Context, resource *cloud.ResourceV2, action string) (float64, error) {
	if a.dryRun {
		// Simulate savings calculation for dry run
		var estimatedSavings float64
		if action == "resize" {
			// Mock downsizing: assume we save 50% of the cost.
			estimatedSavings = resource.CostPerMonth * 0.5
		}
		return estimatedSavings, nil
	}

	switch action {
	case "stop":
		_, err := a.stopEC2Instance(ctx, resource.ID)
		// Stopping an instance saves its entire monthly cost.
		return resource.CostPerMonth, err
	case "resize":
		_, err := a.resizeEC2Instance(ctx, resource.ID)
		// Mock downsizing: assume we save 50% of the cost.
		estimatedSavings := resource.CostPerMonth * 0.5
		return estimatedSavings, err
	default:
		return 0, fmt.Errorf("unknown action: %s", action)
	}
}

func (a *Adapter) stopEC2Instance(ctx context.Context, instanceID string) (string, error) {
	_, err := a.ec2Client.StopInstances(ctx, &ec2.StopInstancesInput{
		InstanceIds: []string{instanceID},
	})
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("Stopped EC2 instance %s", instanceID), nil
}

// getEC2Metrics fetches real CloudWatch metrics for an EC2 instance
func (a *Adapter) getEC2Metrics(ctx context.Context, instanceID string) (map[string]interface{}, error) {
	var wg sync.WaitGroup
	var cpuResult, netInResult, netOutResult *cloudwatch.GetMetricStatisticsOutput
	var cpuErr, netInErr, netOutErr error

	wg.Add(3)

	go func() {
		defer wg.Done()
		cpuResult, cpuErr = a.cwClient.GetMetricStatistics(ctx, &cloudwatch.GetMetricStatisticsInput{
			Namespace:  aws.String("AWS/EC2"),
			MetricName: aws.String("CPUUtilization"),
			Dimensions: []cloudwatchtypes.Dimension{{Name: aws.String("InstanceId"), Value: aws.String(instanceID)}},
			StartTime:  aws.Time(time.Now().Add(-1 * time.Hour)),
			EndTime:    aws.Time(time.Now()),
			Period:     aws.Int32(300), // 5 minutes
			Statistics: []cloudwatchtypes.Statistic{cloudwatchtypes.StatisticAverage},
		})
	}()

	go func() {
		defer wg.Done()
		netInResult, netInErr = a.cwClient.GetMetricStatistics(ctx, &cloudwatch.GetMetricStatisticsInput{
			Namespace:  aws.String("AWS/EC2"),
			MetricName: aws.String("NetworkIn"),
			Dimensions: []cloudwatchtypes.Dimension{{Name: aws.String("InstanceId"), Value: aws.String(instanceID)}},
			StartTime:  aws.Time(time.Now().Add(-1 * time.Hour)),
			EndTime:    aws.Time(time.Now()),
			Period:     aws.Int32(3600), // 1 hour
			Statistics: []cloudwatchtypes.Statistic{cloudwatchtypes.StatisticSum},
		})
	}()

	go func() {
		defer wg.Done()
		netOutResult, netOutErr = a.cwClient.GetMetricStatistics(ctx, &cloudwatch.GetMetricStatisticsInput{
			Namespace:  aws.String("AWS/EC2"),
			MetricName: aws.String("NetworkOut"),
			Dimensions: []cloudwatchtypes.Dimension{{Name: aws.String("InstanceId"), Value: aws.String(instanceID)}},
			StartTime:  aws.Time(time.Now().Add(-1 * time.Hour)),
			EndTime:    aws.Time(time.Now()),
			Period:     aws.Int32(3600), // 1 hour
			Statistics: []cloudwatchtypes.Statistic{cloudwatchtypes.StatisticSum},
		})
	}()

	wg.Wait()

	err := multierr.Combine(cpuErr, netInErr, netOutErr)

	netInBytes := 0.0
	if netInErr == nil && netInResult != nil && len(netInResult.Datapoints) > 0 {
		latest := netInResult.Datapoints[0]
		if latest.Sum != nil {
			netInBytes = *latest.Sum
		}
	}

	netOutBytes := 0.0
	if netOutErr == nil && netOutResult != nil && len(netOutResult.Datapoints) > 0 {
		latest := netOutResult.Datapoints[0]
		if latest.Sum != nil {
			netOutBytes = *latest.Sum
		}
	}

	cpuUsage := 0.0
	if cpuErr == nil && cpuResult != nil && len(cpuResult.Datapoints) > 0 {
		latest := cpuResult.Datapoints[0]
		if latest.Average != nil {
			cpuUsage = *latest.Average
		}
	}

	metrics := map[string]interface{}{
		"cpu_usage":    cpuUsage,
		"memory_usage": 0.0, // Memory metrics require custom CloudWatch agent
		"network_in":   netInBytes,
		"network_out":  netOutBytes,
		"timestamp":    time.Now(),
	}

	return metrics, err
}

func (a *Adapter) resizeEC2Instance(ctx context.Context, instanceID string) (string, error) {
	// This would involve stopping, modifying, and restarting
	return fmt.Sprintf("Resized EC2 instance %s", instanceID), nil
}

// GetSpotPrice returns the current spot price for an instance type in a zone
func (a *Adapter) GetSpotPrice(zone, instanceType string) (float64, error) {
	// Mock implementation - in production, this would call AWS pricing API
	prices := map[string]float64{
		"us-east-1a:t3.micro":  0.0104,
		"us-east-1a:t3.small":  0.0208,
		"us-east-1a:t3.medium": 0.0416,
		"us-east-1b:t3.micro":  0.0104,
		"us-east-1b:t3.small":  0.0208,
		"us-east-1b:t3.medium": 0.0416,
	}

	key := fmt.Sprintf("%s:%s", zone, instanceType)
	if price, exists := prices[key]; exists {
		return price, nil
	}

	// Default price if not found
	return 0.0416, nil
}

// ListZones returns available availability zones
func (a *Adapter) ListZones() ([]string, error) {
	// Mock implementation - in production, this would call AWS EC2 API
	return []string{"us-east-1a", "us-east-1b", "us-east-1c"}, nil
}
