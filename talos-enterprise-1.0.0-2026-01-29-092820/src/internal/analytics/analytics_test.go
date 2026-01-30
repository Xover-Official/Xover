package analytics

import (
	"testing"
)

func TestTokenTracker_RecordUsage(t *testing.T) {
	tracker := NewTokenTracker("")

	// Test recording usage
	tracker.RecordUsage("gemini-2.0-flash-exp", 1000)

	if tracker.TotalTokens != 1000 {
		t.Errorf("Expected 1000 tokens, got %d", tracker.TotalTokens)
	}

	if tracker.TotalCostUSD <= 0 {
		t.Error("Expected positive cost")
	}
}

func TestTokenTracker_ROI(t *testing.T) {
	tracker := NewTokenTracker("")

	// Record usage and savings
	tracker.RecordUsage("gemini-2.0-flash-exp", 1000)
	tracker.RecordSavings(100.0)

	roi := tracker.GetROI()
	if roi <= 0 {
		t.Error("Expected positive ROI")
	}
}

func TestForecaster_PredictCost(t *testing.T) {
	forecaster := NewForecaster()

	// Add sample data
	for i := 0; i < 30; i++ {
		forecaster.AddDataPoint(100.0+float64(i), 50.0)
	}

	mean, stddev := forecaster.PredictCost(7)

	if mean <= 0 {
		t.Error("Expected positive mean")
	}

	if stddev < 0 {
		t.Error("Expected non-negative stddev")
	}
}

func TestForecaster_DetectAnomalies(t *testing.T) {
	forecaster := NewForecaster()

	// Add normal data
	for i := 0; i < 20; i++ {
		forecaster.AddDataPoint(100.0, 50.0)
	}

	// Add anomaly
	forecaster.AddDataPoint(500.0, 50.0)

	anomalies := forecaster.DetectAnomalies()

	if len(anomalies) == 0 {
		t.Error("Expected to detect anomaly")
	}
}

func TestForecaster_GetTrend(t *testing.T) {
	forecaster := NewForecaster()

	// Add increasing trend
	for i := 0; i < 14; i++ {
		forecaster.AddDataPoint(100.0+float64(i*10), 50.0)
	}

	trend := forecaster.GetTrend()

	if trend != "increasing" {
		t.Errorf("Expected increasing trend, got %s", trend)
	}
}
