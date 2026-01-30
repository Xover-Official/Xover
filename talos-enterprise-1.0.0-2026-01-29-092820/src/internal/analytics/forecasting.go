package analytics

import (
	"math"
	"time"
)

// Forecaster provides cost forecasting and anomaly detection
type Forecaster struct {
	historicalData []DataPoint
}

// DataPoint represents a single cost data point
type DataPoint struct {
	Timestamp time.Time
	Cost      float64
	Savings   float64
}

// Anomaly represents a detected cost anomaly
type Anomaly struct {
	Timestamp    time.Time
	ActualCost   float64
	ExpectedCost float64
	Deviation    float64
	Severity     string // "low", "medium", "high"
}

// NewForecaster creates a new forecasting engine
func NewForecaster() *Forecaster {
	return &Forecaster{
		historicalData: make([]DataPoint, 0),
	}
}

// AddDataPoint adds a new cost data point
func (f *Forecaster) AddDataPoint(cost, savings float64) {
	f.historicalData = append(f.historicalData, DataPoint{
		Timestamp: time.Now(),
		Cost:      cost,
		Savings:   savings,
	})

	// Keep only last 90 days
	if len(f.historicalData) > 90 {
		f.historicalData = f.historicalData[len(f.historicalData)-90:]
	}
}

// PredictCost predicts future cost using simple moving average
func (f *Forecaster) PredictCost(daysAhead int) (mean float64, stddev float64) {
	if len(f.historicalData) == 0 {
		return 0, 0
	}

	// Calculate moving average
	sum := 0.0
	for _, dp := range f.historicalData {
		sum += dp.Cost
	}
	mean = sum / float64(len(f.historicalData))

	// Calculate standard deviation
	variance := 0.0
	for _, dp := range f.historicalData {
		variance += math.Pow(dp.Cost-mean, 2)
	}
	stddev = math.Sqrt(variance / float64(len(f.historicalData)))

	// Simple projection (could use more sophisticated models)
	return mean, stddev
}

// DetectAnomalies identifies cost anomalies
func (f *Forecaster) DetectAnomalies() []Anomaly {
	if len(f.historicalData) < 7 {
		return nil // Need at least a week of data
	}

	mean, stddev := f.PredictCost(0)
	var anomalies []Anomaly

	// Check recent data points
	for i := len(f.historicalData) - 7; i < len(f.historicalData); i++ {
		dp := f.historicalData[i]
		deviation := math.Abs(dp.Cost - mean)

		// Anomaly if > 2 standard deviations
		if deviation > 2*stddev {
			severity := "medium"
			if deviation > 3*stddev {
				severity = "high"
			}

			anomalies = append(anomalies, Anomaly{
				Timestamp:    dp.Timestamp,
				ActualCost:   dp.Cost,
				ExpectedCost: mean,
				Deviation:    deviation,
				Severity:     severity,
			})
		}
	}

	return anomalies
}

// GetTrend calculates the cost trend (increasing/decreasing)
func (f *Forecaster) GetTrend() string {
	if len(f.historicalData) < 14 {
		return "insufficient_data"
	}

	// Compare last 7 days vs previous 7 days
	recent := f.historicalData[len(f.historicalData)-7:]
	previous := f.historicalData[len(f.historicalData)-14 : len(f.historicalData)-7]

	recentAvg := 0.0
	for _, dp := range recent {
		recentAvg += dp.Cost
	}
	recentAvg /= float64(len(recent))

	previousAvg := 0.0
	for _, dp := range previous {
		previousAvg += dp.Cost
	}
	previousAvg /= float64(len(previous))

	if recentAvg > previousAvg*1.1 {
		return "increasing"
	} else if recentAvg < previousAvg*0.9 {
		return "decreasing"
	}
	return "stable"
}
