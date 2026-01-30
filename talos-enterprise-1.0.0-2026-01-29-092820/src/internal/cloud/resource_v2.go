package cloud

import (
	"time"
)

// ResourceV2 represents unified cloud resource schema (v2)
type ResourceV2 struct {
	// Core Identity
	ID       string            `json:"id"`
	Type     string            `json:"type"`
	Provider string            `json:"provider"`
	Region   string            `json:"region"`
	Account  string            `json:"account"`
	Tags     map[string]string `json:"tags"`

	// State
	State         string    `json:"state"`
	CreatedAt     time.Time `json:"created_at"`
	ModifiedAt    time.Time `json:"modified_at"`
	LastScannedAt time.Time `json:"last_scanned_at"`

	// Performance Metrics
	CPUUsage    float64 `json:"cpu_usage"`
	MemoryUsage float64 `json:"memory_usage"`
	NetworkIn   float64 `json:"network_in"`
	NetworkOut  float64 `json:"network_out"`
	DiskIO      float64 `json:"disk_io"`

	// Cost & Billing
	CostPerHour  float64 `json:"cost_per_hour"`
	CostPerMonth float64 `json:"cost_per_month"`
	CostYTD      float64 `json:"cost_ytd"`
	Currency     string  `json:"currency"`

	// Optimization
	RightSizeRecommendation string  `json:"rightsize_recommendation,omitempty"`
	EstimatedSavings        float64 `json:"estimated_savings"`
	OptimizationScore       float64 `json:"optimization_score"` // 0-100

	// Compliance & Security
	ComplianceTags     []string `json:"compliance_tags"`
	EncryptionEnabled  bool     `json:"encryption_enabled"`
	PubliclyAccessible bool     `json:"publicly_accessible"`
	BackupEnabled      bool     `json:"backup_enabled"`

	// SLA/SLO
	AvailabilityTarget float64 `json:"availability_target"` // e.g., 99.9
	ActualAvailability float64 `json:"actual_availability"`
	LatencyTarget      int     `json:"latency_target_ms"`
	ActualLatency      int     `json:"actual_latency_ms"`

	// Dependencies
	DependsOn  []string `json:"depends_on,omitempty"`
	DependedBy []string `json:"depended_by,omitempty"`

	// Carbon Footprint
	EnergyUsageKWh    float64 `json:"energy_usage_kwh"`
	CarbonFootprintKg float64 `json:"carbon_footprint_kg"`

	// Metadata
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}

// GetEfficiencyScore calculates overall efficiency (0-100)
func (r *ResourceV2) GetEfficiencyScore() float64 {
	// Weighted score based on utilization and cost
	cpuWeight := 0.4
	memWeight := 0.3
	costWeight := 0.3

	cpuScore := min(r.CPUUsage, 100.0)
	memScore := min(r.MemoryUsage, 100.0)

	// Cost score: inverse of wasted capacity
	costScore := 100.0
	if r.CPUUsage < 50 {
		costScore = r.CPUUsage * 2
	}

	return (cpuScore * cpuWeight) + (memScore * memWeight) + (costScore * costWeight)
}

// IsUnderutilized determines if resource is underutilized
func (r *ResourceV2) IsUnderutilized(cpuThreshold, memThreshold float64) bool {
	return r.CPUUsage < cpuThreshold && r.MemoryUsage < memThreshold
}

// IsOverprovisioned checks if resource can be downsized
func (r *ResourceV2) IsOverprovisioned() bool {
	return r.CPUUsage < 30 && r.MemoryUsage < 40
}

// GetMonthlyWaste calculates monthly wasted cost
func (r *ResourceV2) GetMonthlyWaste() float64 {
	utilizationFactor := max(r.CPUUsage, r.MemoryUsage) / 100.0
	return r.CostPerMonth * (1 - utilizationFactor)
}

// IsCompliant checks if resource meets compliance requirements
func (r *ResourceV2) IsCompliant(requiredTags []string) bool {
	for _, required := range requiredTags {
		found := false
		for _, tag := range r.ComplianceTags {
			if tag == required {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}
	return r.EncryptionEnabled && r.BackupEnabled
}

// MeetsS LA checks if resource meets SLA targets
func (r *ResourceV2) MeetsSLA() bool {
	availabilityOK := r.ActualAvailability >= r.AvailabilityTarget
	latencyOK := r.ActualLatency <= r.LatencyTarget
	return availabilityOK && latencyOK
}

func min(a, b float64) float64 {
	if a < b {
		return a
	}
	return b
}

func max(a, b float64) float64 {
	if a > b {
		return a
	}
	return b
}
