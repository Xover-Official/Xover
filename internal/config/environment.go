package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

// EnvironmentConfig loads configuration from environment variables
type EnvironmentConfig struct {
	*Config
}

// NewEnvironmentConfig creates a new environment-based configuration
func NewEnvironmentConfig() (*EnvironmentConfig, error) {
	cfg := &Config{
		Server: ServerConfig{
			Port: getEnvOrDefault("PORT", "8080"),
			Mode: getEnvOrDefault("MODE", "production"),
		},
		AI: AIConfig{
			OpenRouterKey:  getEnvOrDefault("OPENROUTER_API_KEY", ""),
			GeminiAPIKey:   getEnvOrDefault("GEMINI_API_KEY", ""),
			ClaudeAPIKey:   getEnvOrDefault("CLAUDE_API_KEY", ""),
			GPT5MiniAPIKey: getEnvOrDefault("GPT5MINI_API_KEY", ""),
			DevinKey:       getEnvOrDefault("DEVIN_API_KEY", ""),
			DevinsAPIKey:   getEnvOrDefault("DEVINS_API_KEY", ""),
			CacheEnabled:   getEnvBoolOrDefault("AI_CACHE_ENABLED", true),
		},
		AITiers: AITiersConfig{
			Sentinel:   getEnvOrDefault("AI_TIER_SENTINEL", "gemini-1.5-flash"),
			Strategist: getEnvOrDefault("AI_TIER_STRATEGIST", "gemini-1.5-pro"),
			Arbiter:    getEnvOrDefault("AI_TIER_ARBITER", "claude-3-5-sonnet-20240620"),
			Oracle:     getEnvOrDefault("AI_TIER_ORACLE", "gpt-4o"),
		},
		Redis: RedisConfig{
			Address:  getEnvOrDefault("REDIS_ADDRESS", "localhost:6379"),
			Password: getEnvOrDefault("REDIS_PASSWORD", ""),
			DB:       getEnvIntOrDefault("REDIS_DB", 0),
			CacheTTL: getEnvDurationOrDefault("REDIS_CACHE_TTL", 5*time.Minute),
		},
		Database: DatabaseConfig{
			DSN: getEnvOrDefault("DATABASE_DSN", fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
				getEnvOrDefault("DB_HOST", "localhost"),
				getEnvOrDefault("DB_PORT", "5432"),
				getEnvOrDefault("DB_USER", "postgres"),
				getEnvOrDefault("DB_PASSWORD", ""),
				getEnvOrDefault("DB_NAME", "talos"))),
		},
		Cloud: CloudConfig{
			Provider: getEnvOrDefault("CLOUD_PROVIDER", "aws"),
			Region:   getEnvOrDefault("CLOUD_REGION", "us-east-1"),
			DryRun:   getEnvBoolOrDefault("CLOUD_DRY_RUN", false),
		},
		Analytics: AnalyticsConfig{
			PersistPath: getEnvOrDefault("ANALYTICS_PATH", "./data/analytics"),
		},
		JWT: JWTConfig{
			SecretKey:     getEnvOrDefault("JWT_SECRET", "your-secret-key"),
			TokenDuration: getEnvDurationOrDefault("JWT_EXPIRATION", 24*time.Hour),
		},
		SSO: SSOConfig{
			Google: SSOProviderConfig{
				ClientID:     getEnvOrDefault("GOOGLE_CLIENT_ID", ""),
				ClientSecret: getEnvOrDefault("GOOGLE_CLIENT_SECRET", ""),
			},
			Okta: SSOProviderConfig{
				ClientID:     getEnvOrDefault("OKTA_CLIENT_ID", ""),
				ClientSecret: getEnvOrDefault("OKTA_CLIENT_SECRET", ""),
				Domain:       getEnvOrDefault("OKTA_DOMAIN", ""),
			},
			Azure: SSOProviderConfig{
				ClientID:     getEnvOrDefault("AZURE_CLIENT_ID", ""),
				ClientSecret: getEnvOrDefault("AZURE_CLIENT_SECRET", ""),
				TenantID:     getEnvOrDefault("AZURE_TENANT_ID", ""),
			},
		},
	}

	return &EnvironmentConfig{Config: cfg}, nil
}

// Validate validates the configuration
func (ec *EnvironmentConfig) Validate() error {
	// Validate AI API keys
	if ec.AI.OpenRouterKey == "" && ec.AI.GeminiAPIKey == "" && ec.AI.ClaudeAPIKey == "" && ec.AI.GPT5MiniAPIKey == "" {
		return fmt.Errorf("at least one AI API key must be provided")
	}

	// In production mode, require valid API keys
	if ec.Server.Mode == "production" {
		if ec.AI.GeminiAPIKey == "" || len(ec.AI.GeminiAPIKey) < 20 {
			return fmt.Errorf("valid Gemini API key required in production mode")
		}
		if ec.AI.ClaudeAPIKey == "" || len(ec.AI.ClaudeAPIKey) < 20 {
			return fmt.Errorf("valid Claude API key required in production mode")
		}
	}

	if ec.JWT.SecretKey == "your-secret-key" || ec.JWT.SecretKey == "demo-jwt-secret-change-in-production" {
		return fmt.Errorf("JWT secret must be changed in production")
	}

	if len(ec.JWT.SecretKey) < 32 {
		return fmt.Errorf("JWT secret must be at least 32 characters long")
	}

	if ec.Cloud.Provider != "aws" && ec.Cloud.Provider != "azure" && ec.Cloud.Provider != "gcp" {
		return fmt.Errorf("unsupported cloud provider: %s", ec.Cloud.Provider)
	}

	return nil
}

// GetDatabaseDSN returns the database connection string
func (ec *EnvironmentConfig) GetDatabaseDSN() string {
	return ec.Database.DSN
}

// IsProduction returns true if running in production mode
func (ec *EnvironmentConfig) IsProduction() bool {
	return ec.Server.Mode == "production"
}

// IsDevelopment returns true if running in development mode
func (ec *EnvironmentConfig) IsDevelopment() bool {
	return ec.Server.Mode == "development"
}

// Helper functions

func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvBoolOrDefault(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		if parsed, err := strconv.ParseBool(value); err == nil {
			return parsed
		}
	}
	return defaultValue
}

func getEnvIntOrDefault(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if parsed, err := strconv.Atoi(value); err == nil {
			return parsed
		}
	}
	return defaultValue
}

func getEnvDurationOrDefault(key string, defaultValue time.Duration) time.Duration {
	if value := os.Getenv(key); value != "" {
		if parsed, err := time.ParseDuration(value); err == nil {
			return parsed
		}
	}
	return defaultValue
}

// GetCloudProviderConfig returns provider-specific configuration
func (ec *EnvironmentConfig) GetCloudProviderConfig() (interface{}, error) {
	switch strings.ToLower(ec.Cloud.Provider) {
	case "aws":
		return map[string]string{
			"region":            ec.Cloud.Region,
			"access_key_id":     getEnvOrDefault("AWS_ACCESS_KEY_ID", ""),
			"secret_access_key": getEnvOrDefault("AWS_SECRET_ACCESS_KEY", ""),
			"session_token":     getEnvOrDefault("AWS_SESSION_TOKEN", ""),
		}, nil
	case "azure":
		return map[string]string{
			"subscription_id": getEnvOrDefault("AZURE_SUBSCRIPTION_ID", ""),
			"tenant_id":       getEnvOrDefault("AZURE_TENANT_ID", ""),
			"client_id":       getEnvOrDefault("AZURE_CLIENT_ID", ""),
			"client_secret":   getEnvOrDefault("AZURE_CLIENT_SECRET", ""),
		}, nil
	case "gcp":
		return map[string]string{
			"project_id":       getEnvOrDefault("GCP_PROJECT_ID", ""),
			"credentials_file": getEnvOrDefault("GCP_CREDENTIALS_FILE", ""),
		}, nil
	default:
		return nil, fmt.Errorf("unsupported cloud provider: %s", ec.Cloud.Provider)
	}
}

// GetAIServiceConfig returns AI service configuration
func (ec *EnvironmentConfig) GetAIServiceConfig() map[string]interface{} {
	config := make(map[string]interface{})

	if ec.AI.OpenRouterKey != "" {
		config["openrouter"] = map[string]string{
			"api_key":  ec.AI.OpenRouterKey,
			"base_url": getEnvOrDefault("OPENROUTER_BASE_URL", "https://openrouter.ai/api/v1"),
		}
	}

	if ec.AI.GeminiAPIKey != "" {
		config["gemini"] = map[string]string{
			"api_key": ec.AI.GeminiAPIKey,
		}
	}

	if ec.AI.ClaudeAPIKey != "" {
		config["claude"] = map[string]string{
			"api_key": ec.AI.ClaudeAPIKey,
		}
	}

	return config
}

// GetMonitoringConfig returns monitoring and observability configuration
func (ec *EnvironmentConfig) GetMonitoringConfig() map[string]interface{} {
	return map[string]interface{}{
		"prometheus": map[string]interface{}{
			"enabled": getEnvBoolOrDefault("PROMETHEUS_ENABLED", true),
			"port":    getEnvOrDefault("PROMETHEUS_PORT", "9090"),
		},
		"jaeger": map[string]interface{}{
			"enabled":      getEnvBoolOrDefault("JAEGER_ENABLED", false),
			"endpoint":     getEnvOrDefault("JAEGER_ENDPOINT", "http://localhost:14268/api/traces"),
			"service_name": getEnvOrDefault("JAEGER_SERVICE_NAME", "talos"),
		},
		"health_check": map[string]interface{}{
			"enabled":  getEnvBoolOrDefault("HEALTH_CHECK_ENABLED", true),
			"interval": getEnvDurationOrDefault("HEALTH_CHECK_INTERVAL", 30*time.Second),
			"timeout":  getEnvDurationOrDefault("HEALTH_CHECK_TIMEOUT", 5*time.Second),
		},
	}
}
