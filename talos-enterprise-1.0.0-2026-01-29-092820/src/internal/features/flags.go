package features

import (
	"sync"
)

// Flag represents a feature flag
type Flag struct {
	Name        string                 `json:"name"`
	Enabled     bool                   `json:"enabled"`
	Description string                 `json:"description"`
	Rollout     float64                `json:"rollout"` // 0.0 to 1.0 (percentage)
	Constraints map[string]interface{} `json:"constraints,omitempty"`
}

// FlagManager manages feature flags
type FlagManager struct {
	mu    sync.RWMutex
	flags map[string]*Flag
}

// NewFlagManager creates a new feature flag manager
func NewFlagManager() *FlagManager {
	return &FlagManager{
		flags: make(map[string]*Flag),
	}
}

// Register registers a new feature flag
func (f *FlagManager) Register(flag *Flag) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.flags[flag.Name] = flag
}

// IsEnabled checks if a feature is enabled
func (f *FlagManager) IsEnabled(name string) bool {
	f.mu.RLock()
	defer f.mu.RUnlock()

	flag, exists := f.flags[name]
	if !exists {
		return false
	}

	return flag.Enabled
}

// IsEnabledFor checks if feature is enabled for specific user/context
func (f *FlagManager) IsEnabledFor(name string, userID string) bool {
	f.mu.RLock()
	defer f.mu.RUnlock()

	flag, exists := f.flags[name]
	if !exists || !flag.Enabled {
		return false
	}

	// Rollout: use hash of userID to determine if in rollout percentage
	if flag.Rollout < 1.0 {
		hash := simpleHash(userID)
		return float64(hash%100)/100.0 < flag.Rollout
	}

	return true
}

// Enable enables a feature flag
func (f *FlagManager) Enable(name string) {
	f.mu.Lock()
	defer f.mu.Unlock()

	if flag, exists := f.flags[name]; exists {
		flag.Enabled = true
	}
}

// Disable disables a feature flag
func (f *FlagManager) Disable(name string) {
	f.mu.Lock()
	defer f.mu.Unlock()

	if flag, exists := f.flags[name]; exists {
		flag.Enabled = false
	}
}

// SetRollout sets the rollout percentage
func (f *FlagManager) SetRollout(name string, percentage float64) {
	f.mu.Lock()
	defer f.mu.Unlock()

	if flag, exists := f.flags[name]; exists {
		flag.Rollout = clamp(percentage, 0.0, 1.0)
	}
}

// GetAll returns all feature flags
func (f *FlagManager) GetAll() map[string]*Flag {
	f.mu.RLock()
	defer f.mu.RUnlock()

	result := make(map[string]*Flag)
	for k, v := range f.flags {
		flagCopy := *v
		result[k] = &flagCopy
	}

	return result
}

// simpleHash creates a simple hash from string
func simpleHash(s string) uint32 {
	hash := uint32(0)
	for i := 0; i < len(s); i++ {
		hash = hash*31 + uint32(s[i])
	}
	return hash
}

func clamp(val, min, max float64) float64 {
	if val < min {
		return min
	}
	if val > max {
		return max
	}
	return val
}

// Predefined feature flags
const (
	FeatureMultiCloudArbitrage  = "multi_cloud_arbitrage"
	FeatureAIPromptChaining     = "ai_prompt_chaining"
	FeatureAdvancedAnalytics    = "advanced_analytics"
	FeatureAutoScaling          = "auto_scaling"
	FeatureCarbonTracking       = "carbon_tracking"
	FeatureComplianceMonitoring = "compliance_monitoring"
)

// InitializeDefaultFlags sets up default feature flags
func (f *FlagManager) InitializeDefaultFlags() {
	defaults := []*Flag{
		{
			Name:        FeatureMultiCloudArbitrage,
			Enabled:     false,
			Description: "Cross-cloud cost optimization",
			Rollout:     0.0,
		},
		{
			Name:        FeatureAIPromptChaining,
			Enabled:     true,
			Description: "Multi-tier AI prompt chaining",
			Rollout:     1.0,
		},
		{
			Name:        FeatureAdvancedAnalytics,
			Enabled:     true,
			Description: "Advanced cost analytics and forecasting",
			Rollout:     1.0,
		},
		{
			Name:        FeatureAutoScaling,
			Enabled:     false,
			Description: "Automatic resource scaling",
			Rollout:     0.1, // 10% rollout
		},
		{
			Name:        FeatureCarbonTracking,
			Enabled:     true,
			Description: "Carbon footprint tracking",
			Rollout:     1.0,
		},
		{
			Name:        FeatureComplianceMonitoring,
			Enabled:     true,
			Description: "Real-time compliance monitoring",
			Rollout:     1.0,
		},
	}

	for _, flag := range defaults {
		f.Register(flag)
	}
}
