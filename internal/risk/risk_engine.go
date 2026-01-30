package risk

import (
	"fmt"
	"math"

	"github.com/project-atlas/atlas/internal/logger"
)

type ScoreResult struct {
	Score      float64 `json:"score"`
	Impact     float64 `json:"impact"`     // Monthly savings in USD
	Risk       float64 `json:"risk"`       // 1-100 logic (Higher is riskier)
	Confidence float64 `json:"confidence"` // 0-1
}

type Engine struct {
	DefaultConfidence float64
}

func NewEngine() *Engine {
	return &Engine{
		DefaultConfidence: 0.8,
	}
}

// CalculateScore implements $Score = (Impact / Risk) \times Confidence$
func (e *Engine) CalculateScore(impact float64, metrics CloudMetrics) ScoreResult {
	// Base risk calculation: Higher usage = Higher risk
	// We use max of CPU and Memory as primary risk factor
	usageRisk := math.Max(metrics.CPUUsage, metrics.MemoryUsage)
	
	// Normalize risk to 1-100, ensuring it's never 0 to avoid division by zero
	risk := math.Max(1.0, usageRisk)
	
	score := (impact / risk) * e.DefaultConfidence

	logger.LogAction(logger.Architect, "RiskCalculation", "COMPLETED", 
		fmt.Sprintf("Impact: %.2f, Risk: %.2f, Score: %.4f", impact, risk, score))

	return ScoreResult{
		Score:      score,
		Impact:     impact,
		Risk:       risk,
		Confidence: e.DefaultConfidence,
	}
}
