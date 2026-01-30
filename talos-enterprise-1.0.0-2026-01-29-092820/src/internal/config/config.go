package config

import (
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

type ServerConfig struct {
	Port string `yaml:"port"`
	Mode string `yaml:"mode"` // e.g., "development", "production"
}

type AIConfig struct {
	OpenRouterKey string `yaml:"openrouter_key"`
	DevinKey      string `yaml:"devin_key"`
	CacheEnabled  bool   `yaml:"cache_enabled"`
}

type RedisConfig struct {
	Address  string        `yaml:"address"`
	Password string        `yaml:"password"`
	DB       int           `yaml:"db"`
	CacheTTL time.Duration `yaml:"cache_ttl"`
}

type CloudConfig struct {
	Provider string `yaml:"provider"`
	Region   string `yaml:"region"`
	DryRun   bool   `yaml:"dry_run"`
}

type JWTConfig struct {
	SecretKey     string        `yaml:"secret_key"`
	TokenDuration time.Duration `yaml:"token_duration"`
}

type SSOProviderConfig struct {
	ClientID     string `yaml:"client_id"`
	ClientSecret string `yaml:"client_secret"`
	Domain       string `yaml:"domain,omitempty"`  // For Okta
	TenantID     string `yaml:"tenant_id,omitempty"` // For Azure
}

type SSOConfig struct {
	Google SSOProviderConfig `yaml:"google"`
	Okta   SSOProviderConfig `yaml:"okta"`
	Azure  SSOProviderConfig `yaml:"azure"`
}

type Config struct {
	Server    ServerConfig    `yaml:"server"`
	AI        AIConfig        `yaml:"ai"`
	Redis     RedisConfig     `yaml:"redis"`
	Cloud     CloudConfig     `yaml:"cloud"`
	Analytics AnalyticsConfig `yaml:"analytics"`
	JWT       JWTConfig       `yaml:"jwt"`
	SSO       SSOConfig       `yaml:"sso"`
}

type AnalyticsConfig struct {
	PersistPath string `yaml:"persist_path"`
}

// Load reads configuration from a YAML file and overrides with environment variables.
func Load(path string) (*Config, error) {
	cfg := &Config{
		// Set production-safe defaults
		Server:    ServerConfig{Port: "8080", Mode: "production"},
		Cloud:     CloudConfig{Provider: "aws", Region: "us-east-1", DryRun: true},
		Redis:     RedisConfig{Address: "localhost:6379", CacheTTL: 5 * time.Minute},
		Analytics: AnalyticsConfig{PersistPath: "./talos_tracker_state.json"},
		AI:        AIConfig{CacheEnabled: true},
	}

	// Override with environment variables for container-friendly deployment
	if port := os.Getenv("PORT"); port != "" { cfg.Server.Port = port }
	if region := os.Getenv("AWS_REGION"); region != "" { cfg.Cloud.Region = region }
	if openRouterKey := os.Getenv("OPENROUTER_API_KEY"); openRouterKey != "" { cfg.AI.OpenRouterKey = openRouterKey }
	if devinKey := os.Getenv("DEVIN_API_KEY"); devinKey != "" { cfg.AI.DevinKey = devinKey }
	if redisAddr := os.Getenv("REDIS_ADDR"); redisAddr != "" { cfg.Redis.Address = redisAddr }

	return cfg, nil
}