package config

import (
	"sync"
)

// FeatureManager handles dynamic feature flags
type FeatureManager struct {
	flags map[string]bool
	mu    sync.RWMutex
}

var instance *FeatureManager
var once sync.Once

func GetFeatureManager() *FeatureManager {
	once.Do(func() {
		instance = &FeatureManager{
			flags: make(map[string]bool),
		}
		// Default flags
		instance.flags["new_ui"] = true
		instance.flags["autonomous_mode"] = false
		instance.flags["beta_features"] = false
	})
	return instance
}

func (fm *FeatureManager) IsEnabled(feature string) bool {
	fm.mu.RLock()
	defer fm.mu.RUnlock()
	return fm.flags[feature]
}

func (fm *FeatureManager) SetFlag(feature string, enabled bool) {
	fm.mu.Lock()
	defer fm.mu.Unlock()
	fm.flags[feature] = enabled
}
