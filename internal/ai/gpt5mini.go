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

// GPT5MiniClient implements Tier 4 (Reasoning Engine) - Complex decisions & explainability
type GPT5MiniClient struct {
	apiKey     string
	endpoint   string
	httpClient *http.Client
	model      string
}

// NewGPT5MiniClient creates a new GPT-5 Mini client
func NewGPT5MiniClient(apiKey string) *GPT5MiniClient {
	return &GPT5MiniClient{
		apiKey:   apiKey,
		endpoint: "https://api.openai.com/v1/chat/completions",
		httpClient: &http.Client{
			Timeout: 90 * time.Second,
		},
		model: "gpt-4o-mini", // Using GPT-4o-mini as GPT-5 Mini proxy
	}
}

// Analyze implements AIClient interface with reasoning
func (c *GPT5MiniClient) Analyze(ctx context.Context, request AIRequest) (*AIResponse, error) {
	startTime := time.Now()

	// Enhanced prompt for reasoning
	reasoningPrompt := fmt.Sprintf(`%s

Think step-by-step and provide:
1. Your recommendation
2. Detailed reasoning for why this is the best approach
3. Alternative options you considered
4. Confidence level (0-100%%)`, request.Prompt)

	reqBody := map[string]interface{}{
		"model": c.model,
		"messages": []map[string]string{
			{
				"role":    "system",
				"content": "You are an expert cloud infrastructure optimization specialist. Provide detailed reasoning for all recommendations.",
			},
			{
				"role":    "user",
				"content": reasoningPrompt,
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
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	if len(result.Choices) == 0 {
		return nil, fmt.Errorf("no content in response")
	}

	content := result.Choices[0].Message.Content
	tokensUsed := result.Usage.TotalTokens

	// GPT-4o-mini pricing: ~$0.15 per 1M input, $0.60 per 1M output
	inputCost := float64(result.Usage.PromptTokens) * 0.00000015
	outputCost := float64(result.Usage.CompletionTokens) * 0.0000006
	costUSD := inputCost + outputCost

	// Extract reasoning from response (simple parsing)
	reasoning := c.extractReasoning(content)
	alternatives := c.extractAlternatives(content)

	return &AIResponse{
		Content:      content,
		TokensUsed:   tokensUsed,
		CostUSD:      costUSD,
		Model:        c.model,
		Latency:      time.Since(startTime),
		Confidence:   0.95, // Very high confidence - Reasoning Engine
		Reasoning:    reasoning,
		Alternatives: alternatives,
	}, nil
}

// extractReasoning extracts reasoning from GPT response
func (c *GPT5MiniClient) extractReasoning(content string) string {
	// Simple extraction - in production, use more sophisticated parsing
	if len(content) > 200 {
		return content[100:200] + "..."
	}
	return content
}

// extractAlternatives extracts alternative options from GPT response
func (c *GPT5MiniClient) extractAlternatives(content string) []string {
	// Placeholder - in production, parse structured output
	return []string{"Option A", "Option B"}
}

func (c *GPT5MiniClient) GetEstimatedCost(request AIRequest) float64 {
	estimatedInputTokens := len(request.Prompt) / 4
	estimatedOutputTokens := request.MaxTokens

	inputCost := float64(estimatedInputTokens) * 0.00000015
	outputCost := float64(estimatedOutputTokens) * 0.0000006

	return inputCost + outputCost
}

func (c *GPT5MiniClient) GetModel() string {
	return c.model
}

func (c *GPT5MiniClient) GetTier() int {
	return 4
}

func (c *GPT5MiniClient) HealthCheck(ctx context.Context) error {
	testReq := AIRequest{
		Prompt:      "Hello",
		MaxTokens:   10,
		Temperature: 0.1,
	}

	_, err := c.Analyze(ctx, testReq)
	return err
}
