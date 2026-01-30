package ai

import (
	"context"
	"time"
)

// AIRequest represents a request to any AI model
type AIRequest struct {
	Prompt       string
	ResourceType string
	RiskScore    float64
	MaxTokens    int
	Temperature  float64
	Metadata     map[string]interface{}
}

// AIResponse represents the response from any AI model
type AIResponse struct {
	Content      string
	TokensUsed   int
	CostUSD      float64
	Model        string
	Latency      time.Duration
	Confidence   float64  // 0.0 to 1.0
	Reasoning    string   // Explanation of the decision
	Alternatives []string // Alternative recommendations considered
}

// AIClient is the interface all AI tier implementations must satisfy
type AIClient interface {
	// Analyze processes a request and returns a response
	Analyze(request AIRequest) (*AIResponse, error)

	// GetEstimatedCost estimates cost before making the call
	GetEstimatedCost(request AIRequest) float64

	// GetModel returns the model identifier
	GetModel() string

	// GetTier returns the tier level (1-5)
	GetTier() int

	// HealthCheck verifies the API is accessible
	HealthCheck(ctx context.Context) error
}

// AIClientFactory creates the appropriate AI client based on tier
type AIClientFactory struct {
	geminiFlashClient *GeminiFlashClient
	geminiProClient   *GeminiProClient
	claudeClient      *ClaudeClient
	gpt5MiniClient    *GPT5MiniClient
	devinClient       *DevinClient
}

// NewAIClientFactory creates a new factory with all clients initialized
func NewAIClientFactory(config *Config) (*AIClientFactory, error) {
	factory := &AIClientFactory{
		geminiFlashClient: NewGeminiFlashClient(config.GeminiAPIKey),
		geminiProClient:   NewGeminiProClient(config.GeminiAPIKey),
		claudeClient:      NewClaudeClient(config.ClaudeAPIKey),
		gpt5MiniClient:    NewGPT5MiniClient(config.GPT5APIKey),
		devinClient:       NewDevinClient(config.DevinAPIKey),
	}

	return factory, nil
}

// GetClientForRisk returns the appropriate AI client based on risk score
func (f *AIClientFactory) GetClientForRisk(riskScore float64) AIClient {
	switch {
	case riskScore < 3.0:
		return f.geminiFlashClient // Tier 1: Sentinel
	case riskScore < 5.0:
		return f.geminiProClient // Tier 2: Strategist
	case riskScore < 7.0:
		return f.claudeClient // Tier 3: Arbiter
	case riskScore < 9.0:
		return f.gpt5MiniClient // Tier 4: Reasoning Engine
	default:
		return f.devinClient // Tier 5: Oracle
	}
}

// GetClientByName returns a specific client by name
func (f *AIClientFactory) GetClientByName(name string) AIClient {
	switch name {
	case "gemini-flash", "sentinel":
		return f.geminiFlashClient
	case "gemini-pro", "strategist":
		return f.geminiProClient
	case "claude", "arbiter":
		return f.claudeClient
	case "gpt5-mini", "reasoning":
		return f.gpt5MiniClient
	case "devin", "oracle":
		return f.devinClient
	default:
		return f.geminiFlashClient // Safe default
	}
}

// Config holds API configuration
type Config struct {
	GeminiAPIKey string
	ClaudeAPIKey string
	GPT5APIKey   string
	DevinAPIKey  string
	CacheEnabled bool
	CacheAddr    string
}
