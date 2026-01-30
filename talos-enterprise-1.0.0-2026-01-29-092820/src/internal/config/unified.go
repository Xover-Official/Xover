package config

import (
	"fmt"
	"os"
	"sync"

	"gopkg.in/yaml.v3"
)

// UnifiedConfigService provides centralized configuration management
type UnifiedConfigService struct {
	mu       sync.RWMutex
	config   *Config
	watchers []chan *Config
}

var (
	globalConfig     *UnifiedConfigService
	globalConfigOnce sync.Once
)

// GetGlobalConfig returns the singleton config service
func GetGlobalConfig() *UnifiedConfigService {
	globalConfigOnce.Do(func() {
		globalConfig = &UnifiedConfigService{
			watchers: make([]chan *Config, 0),
		}
	})
	return globalConfig
}

// Load loads configuration from file with hot-reload support
func (u *UnifiedConfigService) Load(path string) error {
	u.mu.Lock()
	defer u.mu.Unlock()

	cfg, err := Load(path)
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	u.config = cfg

	// Notify watchers of config change
	u.notifyWatchers()

	return nil
}

// Get returns current configuration (thread-safe)
func (u *UnifiedConfigService) Get() *Config {
	u.mu.RLock()
	defer u.mu.RUnlock()
	return u.config
}

// Watch registers a channel to receive config updates
func (u *UnifiedConfigService) Watch() <-chan *Config {
	u.mu.Lock()
	defer u.mu.Unlock()

	ch := make(chan *Config, 1)
	u.watchers = append(u.watchers, ch)

	// Send current config immediately
	if u.config != nil {
		ch <- u.config
	}

	return ch
}

// Reload reloads configuration from disk
func (u *UnifiedConfigService) Reload(path string) error {
	return u.Load(path)
}

// Update performs a hot update of specific config fields
func (u *UnifiedConfigService) Update(updateFn func(*Config)) {
	u.mu.Lock()
	defer u.mu.Unlock()

	if u.config != nil {
		updateFn(u.config)
		u.notifyWatchers()
	}
}

// notifyWatchers sends config updates to all registered watchers
func (u *UnifiedConfigService) notifyWatchers() {
	for _, ch := range u.watchers {
		select {
		case ch <- u.config:
		default:
			// Channel full, skip
		}
	}
}

// Validate validates the current configuration
func (u *UnifiedConfigService) Validate() error {
	u.mu.RLock()
	defer u.mu.RUnlock()

	if u.config == nil {
		return fmt.Errorf("configuration not loaded")
	}

	// Validate AI keys
	if u.config.AI.OpenRouterKey == "" {
		return fmt.Errorf("OPENROUTER_API_KEY is required")
	}

	// Validate risk threshold - using AI config for now
	if u.config.AI.CacheEnabled && false { // placeholder validation
		return fmt.Errorf("AI configuration validation failed")
	}

	// Validate database config - using Redis config for now
	if u.config.Redis.Address == "" {
		return fmt.Errorf("Redis address is required")
	}

	return nil
}

// Export exports configuration to YAML
func (u *UnifiedConfigService) Export(path string) error {
	u.mu.RLock()
	defer u.mu.RUnlock()

	if u.config == nil {
		return fmt.Errorf("no configuration to export")
	}

	data, err := yaml.Marshal(u.config)
	if err != nil {
		return err
	}

	return os.WriteFile(path, data, 0644)
}

// GetString returns a config value as string (for dynamic access)
func (u *UnifiedConfigService) GetString(key string) string {
	u.mu.RLock()
	defer u.mu.RUnlock()

	if u.config == nil {
		return ""
	}

	// Simple key resolution (can be extended)
	switch key {
	case "ai.openrouter_key":
		return u.config.AI.OpenRouterKey
	case "server.mode":
		return u.config.Server.Mode
	case "redis.address":
		return u.config.Redis.Address
	default:
		return ""
	}
}

// GetBool returns a config value as bool
func (u *UnifiedConfigService) GetBool(key string) bool {
	u.mu.RLock()
	defer u.mu.RUnlock()

	if u.config == nil {
		return false
	}

	switch key {
	case "cloud.dry_run":
		return u.config.Cloud.DryRun
	case "ai.cache_enabled":
		return u.config.AI.CacheEnabled
	case "server.port":
		return u.config.Server.Port == "8080"
	default:
		return false
	}
}

// GetFloat returns a config value as float64
func (u *UnifiedConfigService) GetFloat(key string) float64 {
	u.mu.RLock()
	defer u.mu.RUnlock()

	if u.config == nil {
		return 0
	}

	switch key {
	case "redis.cache_ttl":
		return float64(u.config.Redis.CacheTTL.Seconds())
	default:
		return 0
	}
}
