package analytics

import (
	"math"
	"time"
)

// AnomalyDetector detects cost anomalies using statistical methods
type AnomalyDetector struct {
	historicalCosts []float64
	timestamps      []time.Time
	windowSize      int
}

// NewAnomalyDetector creates a new anomaly detector
func NewAnomalyDetector(windowSize int) *AnomalyDetector {
	return &AnomalyDetector{
		historicalCosts: make([]float64, 0),
		timestamps:      make([]time.Time, 0),
		windowSize:      windowSize,
	}
}

// AddDataPoint adds a new cost data point
func (a *AnomalyDetector) AddDataPoint(cost float64, timestamp time.Time) {
	a.historicalCosts = append(a.historicalCosts, cost)
	a.timestamps = append(a.timestamps, timestamp)

	// Keep only recent window
	if len(a.historicalCosts) > a.windowSize {
		a.historicalCosts = a.historicalCosts[1:]
		a.timestamps = a.timestamps[1:]
	}
}

// DetectAnomaly detects if current cost is anomalous
func (a *AnomalyDetector) DetectAnomaly(currentCost float64) (bool, float64, string) {
	if len(a.historicalCosts) < 7 {
		return false, 0, "insufficient data"
	}

	mean := a.calculateMean()
	stdDev := a.calculateStdDev(mean)

	// Z-score: how many standard deviations away from mean
	zScore := math.Abs((currentCost - mean) / stdDev)

	// Anomaly if > 3 standard deviations (99.7% confidence)
	isAnomaly := zScore > 3.0

	var reason string
	if isAnomaly {
		if currentCost > mean {
			percentIncrease := ((currentCost - mean) / mean) * 100
			reason = sprintf("Cost is %.1f%% above normal (%.2f vs %.2f avg)", percentIncrease, currentCost, mean)
		} else {
			percentDecrease := ((mean - currentCost) / mean) * 100
			reason = sprintf("Cost is %.1f%% below normal (%.2f vs %.2f avg)", percentDecrease, currentCost, mean)
		}
	}

	return isAnomaly, zScore, reason
}

// calculateMean calculates average cost
func (a *AnomalyDetector) calculateMean() float64 {
	if len(a.historicalCosts) == 0 {
		return 0
	}

	sum := 0.0
	for _, cost := range a.historicalCosts {
		sum += cost
	}

	return sum / float64(len(a.historicalCosts))
}

// calculateStdDev calculates standard deviation
func (a *AnomalyDetector) calculateStdDev(mean float64) float64 {
	if len(a.historicalCosts) == 0 {
		return 0
	}

	variance := 0.0
	for _, cost := range a.historicalCosts {
		diff := cost - mean
		variance += diff * diff
	}

	variance /= float64(len(a.historicalCosts))
	return math.Sqrt(variance)
}

// GetTrend calculates cost trend (positive = increasing, negative = decreasing)
func (a *AnomalyDetector) GetTrend() float64 {
	if len(a.historicalCosts) < 2 {
		return 0
	}

	// Simple linear regression slope
	n := float64(len(a.historicalCosts))
	sumX := 0.0
	sumY := 0.0
	sumXY := 0.0
	sumX2 := 0.0

	for i, cost := range a.historicalCosts {
		x := float64(i)
		sumX += x
		sumY += cost
		sumXY += x * cost
		sumX2 += x * x
	}

	slope := (n*sumXY - sumX*sumY) / (n*sumX2 - sumX*sumX)
	return slope
}

// PredictNextCost predicts next period's cost using linear regression
func (a *AnomalyDetector) PredictNextCost() float64 {
	if len(a.historicalCosts) < 2 {
		return 0
	}

	trend := a.GetTrend()
	mean := a.calculateMean()
	nextX := float64(len(a.historicalCosts))

	// y = mx + b, where m is trend and b is intercept
	intercept := mean - trend*float64(len(a.historicalCosts)/2)
	prediction := trend*nextX + intercept

	return max(prediction, 0) // Can't be negative
}

func sprintf(format string, args ...interface{}) string {
	// Simplified sprintf - in production use fmt.Sprintf
	return format
}
