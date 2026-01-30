package risk

import "time"

type CloudMetrics struct {
	CPUUsage      float64
	MemoryUsage   float64
	MeasuredAt    time.Time
	AnalysisEpoch time.Duration
}
