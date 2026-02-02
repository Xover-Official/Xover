package secrets

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"os"
	"strings"
	"time"
)

// SecretManager handles secure secret management
type SecretManager struct {
	secrets map[string]string
	logger  Logger
}

// Logger interface for logging
type Logger interface {
	Info(msg string)
	Warn(msg string)
	Error(msg string)
}

// NewSecretManager creates a new secret manager
func NewSecretManager(logger Logger) *SecretManager {
	return &SecretManager{
		secrets: make(map[string]string),
		logger:  logger,
	}
}

// LoadSecrets loads secrets from environment variables
func (sm *SecretManager) LoadSecrets() error {
	// Required secrets
	requiredSecrets := map[string]string{
		"GEMINI_API_KEY": "Gemini API key for AI services",
		"CLAUDE_API_KEY": "Claude API key for AI services",
		"JWT_SECRET":     "JWT secret for authentication",
		"REDIS_PASSWORD": "Redis password for caching",
	}

	// Optional secrets
	optionalSecrets := map[string]string{
		"OPENROUTER_API_KEY":  "OpenRouter API key for AI services",
		"GPT5MINI_API_KEY":    "GPT-5 Mini API key for AI services",
		"DEVIN_API_KEY":       "Devin API key for AI services",
		"AWS_ACCESS_KEY":      "AWS access key for cloud services",
		"AWS_SECRET_KEY":      "AWS secret key for cloud services",
		"AZURE_CLIENT_ID":     "Azure client ID for cloud services",
		"AZURE_CLIENT_SECRET": "Azure client secret for cloud services",
		"GCP_PROJECT_ID":      "GCP project ID for cloud services",
		"GCP_KEY_FILE":        "GCP key file path for cloud services",
		"DATABASE_DSN":        "Database connection string",
		"SLACK_WEBHOOK_URL":   "Slack webhook URL for notifications",
		"TEAMS_WEBHOOK_URL":   "Teams webhook URL for notifications",
	}

	// Load required secrets
	for key, description := range requiredSecrets {
		value := os.Getenv(key)
		if value == "" {
			return fmt.Errorf("required secret %s is not set: %s", key, description)
		}

		// Validate secret strength
		if err := sm.validateSecret(key, value); err != nil {
			return fmt.Errorf("invalid secret %s: %w", key, err)
		}

		sm.secrets[key] = value
		sm.logger.Info(fmt.Sprintf("Loaded required secret: %s", key))
	}

	// Load optional secrets
	for key := range optionalSecrets {
		value := os.Getenv(key)
		if value != "" {
			if err := sm.validateSecret(key, value); err != nil {
				sm.logger.Warn(fmt.Sprintf("Invalid optional secret %s, skipping: %v", key, err))
				continue
			}
			sm.secrets[key] = value
			sm.logger.Info(fmt.Sprintf("Loaded optional secret: %s", key))
		}
	}

	return nil
}

// validateSecret validates a secret's strength
func (sm *SecretManager) validateSecret(key, value string) error {
	// Check for common weak patterns
	weakPatterns := []string{
		"test", "demo", "example", "sample", "fake", "mock",
		"123456", "password", "secret", "key", "admin",
		"change-me", "default", "placeholder",
	}

	lowerValue := strings.ToLower(value)
	for _, pattern := range weakPatterns {
		if strings.Contains(lowerValue, pattern) {
			return fmt.Errorf("secret contains weak pattern: %s", pattern)
		}
	}

	// Validate specific secrets
	switch key {
	case "JWT_SECRET":
		if len(value) < 32 {
			return fmt.Errorf("JWT secret must be at least 32 characters long")
		}
		if !sm.hasGoodEntropy(value) {
			return fmt.Errorf("JWT secret has insufficient entropy")
		}

	case "GEMINI_API_KEY", "CLAUDE_API_KEY", "GPT5MINI_API_KEY", "OPENROUTER_API_KEY", "DEVIN_API_KEY":
		if len(value) < 20 {
			return fmt.Errorf("API key must be at least 20 characters long")
		}
		if !strings.HasPrefix(value, "AIza") && !strings.HasPrefix(value, "sk-") && !strings.HasPrefix(value, "sk-proj-") {
			return fmt.Errorf("API key has invalid format")
		}

	case "AWS_ACCESS_KEY", "AWS_SECRET_KEY":
		if len(value) < 16 {
			return fmt.Errorf("AWS credentials must be at least 16 characters long")
		}
		if strings.Contains(value, "AKIA") && len(value) != 20 {
			return fmt.Errorf("AWS access key must be exactly 20 characters long")
		}

	case "DATABASE_DSN":
		if !strings.Contains(value, "://") {
			return fmt.Errorf("Database DSN must be a valid connection string")
		}
	}

	return nil
}

// hasGoodEntropy checks if a string has good entropy
func (sm *SecretManager) hasGoodEntropy(value string) bool {
	// Simple entropy check: variety of characters
	hasUpper := false
	hasLower := false
	hasNumber := false
	hasSpecial := false

	for _, char := range value {
		switch {
		case char >= 'A' && char <= 'Z':
			hasUpper = true
		case char >= 'a' && char <= 'z':
			hasLower = true
		case char >= '0' && char <= '9':
			hasNumber = true
		default:
			hasSpecial = true
		}
	}

	// Require at least 3 of 4 character types
	score := 0
	if hasUpper {
		score++
	}
	if hasLower {
		score++
	}
	if hasNumber {
		score++
	}
	if hasSpecial {
		score++
	}

	return score >= 3
}

// GetSecret retrieves a secret
func (sm *SecretManager) GetSecret(key string) (string, error) {
	value, exists := sm.secrets[key]
	if !exists {
		return "", fmt.Errorf("secret not found: %s", key)
	}
	return value, nil
}

// GetSecretWithDefault retrieves a secret with a default value
func (sm *SecretManager) GetSecretWithDefault(key, defaultValue string) string {
	value, err := sm.GetSecret(key)
	if err != nil {
		sm.logger.Warn(fmt.Sprintf("Secret not found, using default: %s", key))
		return defaultValue
	}
	return value
}

// ValidateSecrets validates all loaded secrets
func (sm *SecretManager) ValidateSecrets() error {
	for key, value := range sm.secrets {
		if err := sm.validateSecret(key, value); err != nil {
			return fmt.Errorf("invalid secret %s: %w", key, err)
		}
	}
	return nil
}

// RotateSecret rotates a secret (placeholder for actual rotation logic)
func (sm *SecretManager) RotateSecret(key string) error {
	// In a real implementation, this would:
	// 1. Generate a new secure secret
	// 2. Update the secret in the secret store
	// 3. Update any dependent services
	// 4. Log the rotation

	newSecret := sm.generateSecureSecret()
	sm.secrets[key] = newSecret

	sm.logger.Info(fmt.Sprintf("Rotated secret: %s", key))
	return nil
}

// generateSecureSecret generates a secure random secret
func (sm *SecretManager) generateSecureSecret() string {
	// Generate 32 random bytes
	randomBytes := make([]byte, 32)
	if _, err := rand.Read(randomBytes); err != nil {
		// Fallback to less secure method
		sm.logger.Error("Failed to read from OS random, using fallback method")
		return sm.generateFallbackSecret()
	}

	// Encode to base64
	return base64.StdEncoding.EncodeToString(randomBytes)
}

// generateFallbackSecret generates a fallback secret
func (sm *SecretManager) generateFallbackSecret() string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789!@#$%^&*()"
	secret := make([]byte, 32)

	for i := range secret {
		secret[i] = charset[i%len(charset)]
	}

	return string(secret)
}

// MaskSecret masks a secret for logging
func (sm *SecretManager) MaskSecret(secret string) string {
	if len(secret) <= 8 {
		return strings.Repeat("*", len(secret))
	}

	return secret[:4] + strings.Repeat("*", len(secret)-8) + secret[len(secret)-4:]
}

// GetSecretStatus returns the status of all secrets
func (sm *SecretManager) GetSecretStatus() map[string]SecretStatus {
	status := make(map[string]SecretStatus)

	for key, value := range sm.secrets {
		status[key] = SecretStatus{
			Loaded:    true,
			Masked:    sm.MaskSecret(value),
			Length:    len(value),
			LastCheck: time.Now(),
		}
	}

	return status
}

// SecretStatus represents the status of a secret
type SecretStatus struct {
	Loaded    bool      `json:"loaded"`
	Masked    string    `json:"masked"`
	Length    int       `json:"length"`
	LastCheck time.Time `json:"last_check"`
}

// EnvironmentValidator validates environment configuration
type EnvironmentValidator struct {
	logger Logger
}

// NewEnvironmentValidator creates a new environment validator
func NewEnvironmentValidator(logger Logger) *EnvironmentValidator {
	return &EnvironmentValidator{
		logger: logger,
	}
}

// ValidateEnvironment validates the entire environment
func (ev *EnvironmentValidator) ValidateEnvironment() error {
	// Check if we're in production
	isProduction := os.Getenv("MODE") == "production"

	if isProduction {
		ev.logger.Info("Validating production environment")

		// Additional production validations
		if err := ev.validateProductionEnvironment(); err != nil {
			return fmt.Errorf("production environment validation failed: %w", err)
		}
	} else {
		ev.logger.Info("Validating development environment")
	}

	// Validate required environment variables
	if err := ev.validateRequiredEnvironment(); err != nil {
		return fmt.Errorf("required environment validation failed: %w", err)
	}

	// Validate optional environment variables
	ev.validateOptionalEnvironment()

	return nil
}

// validateProductionEnvironment validates production-specific requirements
func (ev *EnvironmentValidator) validateProductionEnvironment() error {
	// Check for development defaults
	devDefaults := []string{
		"your-secret-key",
		"demo-jwt-secret-change-in-production",
		"localhost",
		"test-key",
		"example-key",
		"change-me",
		"password",
		"admin",
	}

	for _, key := range devDefaults {
		if value := os.Getenv(key); value != "" && strings.Contains(strings.ToLower(value), strings.ToLower(key)) {
			return fmt.Errorf("production environment contains development default for %s", key)
		}
	}

	// Check for insecure configurations
	if os.Getenv("TLS_DISABLED") == "true" {
		return fmt.Errorf("TLS cannot be disabled in production")
	}

	if os.Getenv("DEBUG") == "true" {
		return fmt.Errorf("debug mode cannot be enabled in production")
	}

	if os.Getenv("LOG_LEVEL") == "debug" {
		return fmt.Errorf("debug logging not recommended in production")
	}

	return nil
}

// validateRequiredEnvironment validates required environment variables
func (ev *EnvironmentValidator) validateRequiredEnvironment() error {
	required := map[string]string{
		"PORT":          "Server port",
		"MODE":          "Application mode",
		"JWT_SECRET":    "JWT secret for authentication",
		"REDIS_ADDRESS": "Redis server address",
	}

	for key, description := range required {
		value := os.Getenv(key)
		if value == "" {
			return fmt.Errorf("required environment variable %s is not set: %s", key, description)
		}
		ev.logger.Info(fmt.Sprintf("Environment variable set: %s", key))
	}

	return nil
}

// validateOptionalEnvironment validates optional environment variables
func (ev *EnvironmentValidator) validateOptionalEnvironment() {
	optional := map[string]string{
		"DATABASE_DSN":       "Database connection string",
		"PROMETHEUS_ENABLED": "Prometheus metrics enabled",
		"JAEGER_ENABLED":     "Jaeger tracing enabled",
		"LOG_LEVEL":          "Logging level",
		"TLS_CERT_FILE":      "TLS certificate file",
		"TLS_KEY_FILE":       "TLS private key file",
	}

	for key, description := range optional {
		value := os.Getenv(key)
		if value != "" {
			ev.logger.Info(fmt.Sprintf("Optional environment variable set: %s (%s = %s)", key, description, value))
		}
	}
}

// GetEnvironmentSummary returns a summary of the environment
func (ev *EnvironmentValidator) GetEnvironmentSummary() map[string]interface{} {
	summary := make(map[string]interface{})

	// Basic info
	summary["mode"] = os.Getenv("MODE")
	summary["port"] = os.Getenv("PORT")
	summary["environment"] = "production"
	if os.Getenv("MODE") != "production" {
		summary["environment"] = "development"
	}

	// Feature flags
	summary["features"] = map[string]bool{
		"prometheus": os.Getenv("PROMETHEUS_ENABLED") == "true",
		"jaeger":     os.Getenv("JAEGER_ENABLED") == "true",
		"tls":        os.Getenv("TLS_ENABLED") == "true",
		"debug":      os.Getenv("DEBUG") == "true",
	}

	// Logging
	summary["logging"] = map[string]string{
		"level":  os.Getenv("LOG_LEVEL"),
		"format": os.Getenv("LOG_FORMAT"),
	}

	// External services
	summary["external_services"] = map[string]bool{
		"database":   os.Getenv("DATABASE_DSN") != "",
		"redis":      os.Getenv("REDIS_ADDRESS") != "",
		"prometheus": os.Getenv("PROMETHEUS_ENABLED") == "true",
		"jaeger":     os.Getenv("JAEGER_ENABLED") == "true",
	}

	return summary
}
