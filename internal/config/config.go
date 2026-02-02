package config

import (
	"fmt"
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

type ServerConfig struct {
	Port         string        `yaml:"port"`
	Mode         string        `yaml:"mode"` // e.g., "development", "production"
	ReadTimeout  time.Duration `yaml:"read_timeout"`
	WriteTimeout time.Duration `yaml:"write_timeout"`
	IdleTimeout  time.Duration `yaml:"idle_timeout"`
}

type AIConfig struct {
	OpenRouterKey        string        `yaml:"openrouter_key"`
	GeminiAPIKey         string        `yaml:"gemini_api_key"`
	ClaudeAPIKey         string        `yaml:"claude_api_key"`
	GPT5MiniAPIKey       string        `yaml:"gpt5_mini_api_key"`
	DevinKey             string        `yaml:"devin_key"`
	DevinsAPIKey         string        `yaml:"devins_api_key"`
	CacheEnabled         bool          `yaml:"cache_enabled"`
	MaxTokensPerRequest  int           `yaml:"max_tokens_per_request"`
	MaxRequestsPerMinute int           `yaml:"max_requests_per_minute"`
	Timeout              time.Duration `yaml:"timeout"`
}

type AITiersConfig struct {
	Sentinel   string `yaml:"sentinel"`
	Strategist string `yaml:"strategist"`
	Arbiter    string `yaml:"arbiter"`
	Oracle     string `yaml:"oracle"`
}

type RedisConfig struct {
	Address      string        `yaml:"address"`
	Password     string        `yaml:"password"`
	DB           int           `yaml:"db"`
	CacheTTL     time.Duration `yaml:"cache_ttl"`
	MaxRetries   int           `yaml:"max_retries"`
	PoolSize     int           `yaml:"pool_size"`
	MinIdleConns int           `yaml:"min_idle_conns"`
	DialTimeout  time.Duration `yaml:"dial_timeout"`
	ReadTimeout  time.Duration `yaml:"read_timeout"`
	WriteTimeout time.Duration `yaml:"write_timeout"`
}

type DatabaseConfig struct {
	DSN string `yaml:"dsn"`
}

type CloudConfig struct {
	Provider             string        `yaml:"provider"`
	Region               string        `yaml:"region"`
	DryRun               bool          `yaml:"dry_run"`
	MaxAPICallsPerMinute int           `yaml:"max_api_calls_per_minute"`
	RetryAttempts        int           `yaml:"retry_attempts"`
	RetryDelay           time.Duration `yaml:"retry_delay"`
	ResourceTypes        []string      `yaml:"resource_types"`
}

type JWTConfig struct {
	SecretKey     string        `yaml:"secret_key"`
	TokenDuration time.Duration `yaml:"token_duration"`
}

type SSOProviderConfig struct {
	ClientID     string `yaml:"client_id"`
	ClientSecret string `yaml:"client_secret"`
	Domain       string `yaml:"domain,omitempty"`    // For Okta
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
	AITiers   AITiersConfig   `yaml:"ai_tiers"`
	Redis     RedisConfig     `yaml:"redis"`
	Database  DatabaseConfig  `yaml:"database"`
	Cloud     CloudConfig     `yaml:"cloud"`
	Analytics AnalyticsConfig `yaml:"analytics"`
	JWT       JWTConfig       `yaml:"jwt"`
	SSO       SSOConfig       `yaml:"sso"`
}

type AnalyticsConfig struct {
	PersistPath string `yaml:"persist_path"`
}

// Validate checks the configuration for required fields and valid values
func (c *Config) Validate() error {
	if c.Server.Port == "" {
		return fmt.Errorf("server port is required")
	}

	if c.Server.Mode != "development" && c.Server.Mode != "production" {
		return fmt.Errorf("server mode must be 'development' or 'production'")
	}

	if c.AI.OpenRouterKey == "" {
		return fmt.Errorf("OpenRouter API key is required")
	}

	if c.JWT.SecretKey == "" {
		return fmt.Errorf("JWT secret key is required")
	}

	if len(c.JWT.SecretKey) < 32 {
		return fmt.Errorf("JWT secret key must be at least 32 characters")
	}

	if c.Cloud.Provider == "" {
		return fmt.Errorf("cloud provider is required")
	}

	if c.Cloud.Region == "" {
		return fmt.Errorf("cloud region is required")
	}

	return nil
}

// Load reads configuration from a YAML file and overrides with environment variables.
func Load(path string) (*Config, error) {
	cfg := &Config{
		// Set production-safe defaults
		Server: ServerConfig{
			Port:         "8080",
			Mode:         "production",
			ReadTimeout:  30 * time.Second,
			WriteTimeout: 30 * time.Second,
			IdleTimeout:  120 * time.Second,
		},
		Cloud: CloudConfig{
			Provider:             "aws",
			Region:               "us-east-1",
			DryRun:               true,
			MaxAPICallsPerMinute: 100,
			RetryAttempts:        3,
			RetryDelay:           1 * time.Second,
			ResourceTypes:        []string{"ec2", "rds", "lambda", "ebs"},
		},
		Redis: RedisConfig{
			Address:      "localhost:6379",
			CacheTTL:     5 * time.Minute,
			MaxRetries:   3,
			PoolSize:     10,
			MinIdleConns: 5,
			DialTimeout:  5 * time.Second,
			ReadTimeout:  3 * time.Second,
			WriteTimeout: 3 * time.Second,
		},
		Database:  DatabaseConfig{DSN: "host=localhost user=atlas dbname=atlas sslmode=disable"},
		Analytics: AnalyticsConfig{PersistPath: "./talos_tracker_state.json"},
		AI: AIConfig{
			CacheEnabled:         true,
			MaxTokensPerRequest:  4000,
			MaxRequestsPerMinute: 60,
			Timeout:              30 * time.Second,
		},
		AITiers: AITiersConfig{
			Sentinel:   "gemini-1.5-flash",
			Strategist: "gemini-1.5-pro",
			Arbiter:    "claude-3-5-sonnet-20240620",
			Oracle:     "gpt-4o",
		},
	}

	f, err := os.Open(path)
	if err != nil {
		return cfg, err
	}
	defer f.Close()

	decoder := yaml.NewDecoder(f)
	if err := decoder.Decode(cfg); err != nil {
		return cfg, err
	}

	// Override with environment variables for container-friendly deployment
	if port := os.Getenv("PORT"); port != "" {
		cfg.Server.Port = port
	}
	if mode := os.Getenv("MODE"); mode != "" {
		cfg.Server.Mode = mode
	}
	if region := os.Getenv("AWS_REGION"); region != "" {
		cfg.Cloud.Region = region
	}
	if openRouterKey := os.Getenv("OPENROUTER_API_KEY"); openRouterKey != "" {
		cfg.AI.OpenRouterKey = openRouterKey
	}
	if devinKey := os.Getenv("DEVIN_API_KEY"); devinKey != "" {
		cfg.AI.DevinKey = devinKey
	}
	if redisAddr := os.Getenv("REDIS_ADDR"); redisAddr != "" {
		cfg.Redis.Address = redisAddr
	}
	if redisPassword := os.Getenv("REDIS_PASSWORD"); redisPassword != "" {
		cfg.Redis.Password = redisPassword
	}
	if dbDsn := os.Getenv("DATABASE_DSN"); dbDsn != "" {
		cfg.Database.DSN = dbDsn
	}
	if jwtSecret := os.Getenv("JWT_SECRET_KEY"); jwtSecret != "" {
		cfg.JWT.SecretKey = jwtSecret
	}

	// Validate configuration after loading
	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("configuration validation failed: %w", err)
	}

	return cfg, nil
}
