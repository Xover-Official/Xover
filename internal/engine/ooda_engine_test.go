package engine

import (
	"context"
	"testing"
	"time"

	"github.com/project-atlas/atlas/internal/ai"
	"github.com/project-atlas/atlas/internal/cloud"
	"github.com/project-atlas/atlas/internal/database"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
)

// --- Mocks ---

type MockCloudAdapter struct {
	mock.Mock
}

func (m *MockCloudAdapter) FetchResources(ctx context.Context) ([]*cloud.ResourceV2, error) {
	args := m.Called(ctx)
	return args.Get(0).([]*cloud.ResourceV2), args.Error(1)
}

func (m *MockCloudAdapter) GetResource(ctx context.Context, id string) (*cloud.ResourceV2, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(*cloud.ResourceV2), args.Error(1)
}

func (m *MockCloudAdapter) ApplyOptimization(ctx context.Context, resource *cloud.ResourceV2, action string) (float64, error) {
	args := m.Called(ctx, resource, action)
	return args.Get(0).(float64), args.Error(1)
}

func (m *MockCloudAdapter) GetSpotPrice(zone, instanceType string) (float64, error) {
	args := m.Called(zone, instanceType)
	return args.Get(0).(float64), args.Error(1)
}

func (m *MockCloudAdapter) ListZones() ([]string, error) {
	args := m.Called()
	return args.Get(0).([]string), args.Error(1)
}

type MockRepository struct {
	mock.Mock
}

func (m *MockRepository) CreateAction(ctx context.Context, action *database.Action) error {
	args := m.Called(ctx, action)
	return args.Error(0)
}

func (m *MockRepository) UpdateActionStatus(ctx context.Context, id string, status string, startedAt *time.Time, completedAt *time.Time, errorMsg *string) error {
	args := m.Called(ctx, id, status, startedAt, completedAt, errorMsg)
	return args.Error(0)
}

func (m *MockRepository) CreateSavingsEvent(ctx context.Context, event *database.SavingsEvent) error {
	args := m.Called(ctx, event)
	return args.Error(0)
}

type MockAIClient struct {
	mock.Mock
}

func (m *MockAIClient) Analyze(ctx context.Context, request ai.AIRequest) (*ai.AIResponse, error) {
	args := m.Called(ctx, request)
	return args.Get(0).(*ai.AIResponse), args.Error(1)
}

func (m *MockAIClient) GetEstimatedCost(request ai.AIRequest) float64 { return 0.0 }
func (m *MockAIClient) GetModel() string                              { return "mock-model" }
func (m *MockAIClient) GetTier() int                                  { return 1 }
func (m *MockAIClient) HealthCheck(ctx context.Context) error         { return nil }

// --- Tests ---

func TestOODAEngine_Observe(t *testing.T) {
	// Setup
	mockAdapter := new(MockCloudAdapter)
	mockRepo := new(MockRepository)
	logger := zap.NewNop()
	tracer := trace.NewNoopTracerProvider().Tracer("")

	expectedResources := []*cloud.ResourceV2{
		{ID: "res-1", Type: "ec2", CPUUsage: 0.1},
		{ID: "res-2", Type: "rds", CPUUsage: 0.8},
	}

	mockAdapter.On("FetchResources", mock.Anything).Return(expectedResources, nil)

	engine := NewOODAEngine(
		nil, // Orchestrator not needed for observe
		mockAdapter,
		mockRepo,
		nil,
		logger,
		tracer,
		DefaultEngineConfig(),
	)

	// Execute
	resources, err := engine.observe(context.Background())

	// Verify
	assert.NoError(t, err)
	assert.Len(t, resources, 2)
	assert.Equal(t, "res-1", resources[0].ID)
	mockAdapter.AssertExpectations(t)
}

func TestOODAEngine_Orient(t *testing.T) {
	// Setup
	mockAdapter := new(MockCloudAdapter)
	mockRepo := new(MockRepository)
	mockAIClient := new(MockAIClient)
	logger := zap.NewNop()
	tracer := trace.NewNoopTracerProvider().Tracer("")

	// Setup AI Factory with Mock Client
	aiConfig := &ai.Config{}

	orchestrator, err := ai.NewUnifiedOrchestrator(aiConfig, nil, logger)
	assert.NoError(t, err)

	orchestrator.GetFactory().SetClient("sentinel", mockAIClient)
	orchestrator.GetFactory().SetClient("strategist", mockAIClient)

	engine := NewOODAEngine(
		orchestrator,
		mockAdapter,
		mockRepo,
		nil,
		logger,
		tracer,
		DefaultEngineConfig(),
	)

	resources := []*cloud.ResourceV2{
		{ID: "res-1", Type: "ec2", CPUUsage: 0.05, MemoryUsage: 0.1, CostPerMonth: 50},
	}

	// Mock AI Response
	mockAIResponse := &ai.AIResponse{
		Content:    "- Recommendation: Downsize to t3.micro\n- Risk: Low",
		Confidence: 0.95,
	}

	// The engine calls Analyze on the orchestrator, which calls the client
	mockAIClient.On("Analyze", mock.Anything, mock.Anything).Return(mockAIResponse, nil)

	// Execute
	opportunities, err := engine.orient(context.Background(), resources)

	// Verify
	assert.NoError(t, err)
	assert.NotEmpty(t, opportunities)
	opp := opportunities[0]
	assert.Equal(t, "res-1", opp.Resource.ID)

	// Verify Analysis Vectors
	// Rightsizing vector should have high score due to low CPU (0.05)
	var rightsizingScore float64
	for _, v := range opp.AnalysisVectors {
		if v.Name == "rightsizing" {
			rightsizingScore = v.Score
		}
	}
	assert.Greater(t, rightsizingScore, 0.7, "Rightsizing score should be high for underutilized resource")
}
