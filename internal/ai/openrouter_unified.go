package ai

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// OpenRouterClient is a unified client that can call any AI model through OpenRouter
type OpenRouterClient struct {
	apiKey     string
	endpoint   string
	httpClient *http.Client
}

// NewOpenRouterClient creates a new OpenRouter client
func NewOpenRouterClient(apiKey string) *OpenRouterClient {
	return &OpenRouterClient{
		apiKey:   apiKey,
		endpoint: "https://openrouter.ai/api/v1/chat/completions",
		httpClient: &http.Client{
			Timeout: 90 * time.Second,
		},
	}
}

// Analyze calls any model through OpenRouter
func (c *OpenRouterClient) Analyze(ctx context.Context, request AIRequest, modelName string) (*AIResponse, error) {
	startTime := time.Now()

	reqBody := map[string]interface{}{
		"model": modelName,
		"messages": []map[string]string{
			{
				"role":    "user",
				"content": request.Prompt,
			},
		},
		"max_tokens":  request.MaxTokens,
		"temperature": request.Temperature,
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", c.endpoint, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+c.apiKey)
	httpReq.Header.Set("HTTP-Referer", "https://github.com/Xover-Official/Xover")
	httpReq.Header.Set("X-Title", "Talos Cloud Guardian")

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("API call failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API returned status %d: %s", resp.StatusCode, string(body))
	}

	var result struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
		Usage struct {
			PromptTokens     int `json:"prompt_tokens"`
			CompletionTokens int `json:"completion_tokens"`
			TotalTokens      int `json:"total_tokens"`
		} `json:"usage"`
		Model string `json:"model"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	if len(result.Choices) == 0 {
		return nil, fmt.Errorf("no content in response")
	}

	content := result.Choices[0].Message.Content
	tokensUsed := result.Usage.TotalTokens

	// Cost calculation based on model
	costUSD := c.calculateCost(result.Model, result.Usage.PromptTokens, result.Usage.CompletionTokens)

	return &AIResponse{
		Content:    content,
		TokensUsed: tokensUsed,
		CostUSD:    costUSD,
		Model:      result.Model,
		Latency:    time.Since(startTime),
		Confidence: 0.90, // Default confidence
	}, nil
}

// calculateCost estimates cost based on OpenRouter pricing
func (c *OpenRouterClient) calculateCost(model string, inputTokens, outputTokens int) float64 {
	// OpenRouter pricing (approximate, per 1M tokens)
	pricing := map[string]struct{ input, output float64 }{
		"google/gemini-2.0-flash-exp": {0.0, 0.0},     // Free tier
		"google/gemini-pro":           {0.125, 0.375}, // $0.125/$0.375 per 1M
		"anthropic/claude-3.5-sonnet": {3.0, 15.0},    // $3/$15 per 1M
		"openai/gpt-4o-mini":          {0.15, 0.60},   // $0.15/$0.60 per 1M
	}

	costs, ok := pricing[model]
	if !ok {
		// Default fallback pricing
		costs = struct{ input, output float64 }{1.0, 3.0}
	}

	inputCost := float64(inputTokens) * costs.input / 1_000_000
	outputCost := float64(outputTokens) * costs.output / 1_000_000

	return inputCost + outputCost
}

// HealthCheck verifies OpenRouter API is accessible
func (c *OpenRouterClient) HealthCheck(ctx context.Context, model string) error {
	testReq := AIRequest{
		Prompt:      "test",
		MaxTokens:   10,
		Temperature: 0.1,
	}

	_, err := c.Analyze(ctx, testReq, model)
	return err
}
