package ai

import (
	"context"
	"fmt"
	"strings"
)

// PromptChain represents a multi-tier AI prompt chain
type PromptChain struct {
	steps []PromptStep
}

// PromptStep represents one step in the chain
type PromptStep struct {
	Tier        int
	Instruction string
	InputKey    string // Key from previous step's output
	OutputKey   string // Key to store this step's output
	Required    bool
}

// NewPromptChain creates a new prompt chain
func NewPromptChain() *PromptChain {
	return &PromptChain{
		steps: make([]PromptStep, 0),
	}
}

// AddStep adds a step to the chain
func (p *PromptChain) AddStep(tier int, instruction, inputKey, outputKey string, required bool) *PromptChain {
	p.steps = append(p.steps, PromptStep{
		Tier:        tier,
		Instruction: instruction,
		InputKey:    inputKey,
		OutputKey:   outputKey,
		Required:    required,
	})
	return p
}

// Execute executes the prompt chain
func (p *PromptChain) Execute(ctx context.Context, orchestrator *UnifiedOrchestrator, initialInput string, riskScore float64) (map[string]string, error) {
	results := make(map[string]string)
	results["initial_input"] = initialInput

	for i, step := range p.steps {
		// Build prompt from previous results
		prompt := p.buildPrompt(step, results)

		// Get model for this tier
		model, useDevin := orchestrator.factory.GetClientForRisk(float64(step.Tier))

		// Make AI call
		request := AIRequest{
			Context:      ctx,
			Prompt:       prompt,
			ResourceType: "chain_step",
			RiskScore:    riskScore,
			MaxTokens:    800,
			Temperature:  0.3,
		}

		var response *AIResponse
		var err error

		if useDevin {
			response, err = orchestrator.factory.devinClient.Analyze(request)
		} else {
			response, err = orchestrator.factory.openRouter.Analyze(request, model)
		}

		if err != nil {
			if step.Required {
				return nil, fmt.Errorf("required step %d failed: %w", i+1, err)
			}
			// Optional step failed, continue
			results[step.OutputKey] = fmt.Sprintf("[Step failed: %v]", err)
			continue
		}

		// Store result
		results[step.OutputKey] = response.Content
	}

	return results, nil
}

// buildPrompt constructs the prompt for this step using previous results
func (p *PromptChain) buildPrompt(step PromptStep, results map[string]string) string {
	prompt := step.Instruction

	// Replace placeholders with actual values
	if step.InputKey != "" {
		if input, exists := results[step.InputKey]; exists {
			prompt = strings.ReplaceAll(prompt, "{input}", input)
			prompt = strings.ReplaceAll(prompt, "{"+step.InputKey+"}", input)
		}
	}

	// Also make all results available
	for key, value := range results {
		placeholder := "{" + key + "}"
		prompt = strings.ReplaceAll(prompt, placeholder, value)
	}

	return prompt
}

// Common prompt chains

// CreateOptimizationChain creates a multi-tier optimization analysis chain
func CreateOptimizationChain() *PromptChain {
	return NewPromptChain().
		// Step 1: Quick analysis (Tier 1 - Sentinel)
		AddStep(1,
			"Quickly analyze this resource: {initial_input}. Summarize key stats in 2-3 sentences.",
			"initial_input",
			"quick_analysis",
			true).
		// Step 2: Deep analysis (Tier 2 - Strategist)
		AddStep(2,
			"Based on this quick analysis: {quick_analysis}\n\nProvide detailed cost-benefit analysis for optimization options.",
			"quick_analysis",
			"detailed_analysis",
			true).
		// Step 3: Safety check (Tier 3 - Arbiter)
		AddStep(3,
			"Review this optimization plan: {detailed_analysis}\n\nIdentify any safety concerns or risks.",
			"detailed_analysis",
			"safety_review",
			true).
		// Step 4: Final recommendation (Tier 2 - Strategist)
		AddStep(2,
			"Given the analysis: {detailed_analysis}\nAnd safety review: {safety_review}\n\nProvide final actionable recommendation.",
			"safety_review",
			"final_recommendation",
			true)
}

// CreateComplianceChain creates a compliance validation chain
func CreateComplianceChain() *PromptChain {
	return NewPromptChain().
		// Step 1: Resource inventory (Tier 1)
		AddStep(1,
			"List all compliance-relevant attributes of: {initial_input}",
			"initial_input",
			"inventory",
			true).
		// Step 2: Compliance check (Tier 3)
		AddStep(3,
			"Check these attributes against SOC2, GDPR, HIPAA requirements: {inventory}",
			"inventory",
			"compliance_status",
			true).
		// Step 3: Remediation (Tier 2)
		AddStep(2,
			"Based on compliance gaps: {compliance_status}\nProvide step-by-step remediation plan.",
			"compliance_status",
			"remediation_plan",
			false)
}

// CreateExplainabilityChain creates a chain for explaining AI decisions
func CreateExplainabilityChain() *PromptChain {
	return NewPromptChain().
		// Step 1: Extract decision (Tier 1)
		AddStep(1,
			"Summarize the key decision made: {initial_input}",
			"initial_input",
			"decision_summary",
			true).
		// Step 2: Explain reasoning (Tier 4 - Reasoning)
		AddStep(4,
			"Explain WHY this decision makes sense: {decision_summary}\nProvide step-by-step reasoning.",
			"decision_summary",
			"reasoning",
			true).
		// Step 3: Alternative options (Tier 4)
		AddStep(4,
			"What are 2-3 alternative approaches to: {decision_summary}\nExplain trade-offs.",
			"decision_summary",
			"alternatives",
			false)
}
