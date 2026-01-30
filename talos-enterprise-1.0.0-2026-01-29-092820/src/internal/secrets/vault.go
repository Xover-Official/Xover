package secrets

import (
	"fmt"

	vault "github.com/hashicorp/vault/api"
)

// VaultClient manages secrets using HashiCorp Vault
type VaultClient struct {
	client *vault.Client
}

// NewVaultClient creates a new Vault client
func NewVaultClient(address, token string) (*VaultClient, error) {
	config := vault.DefaultConfig()
	config.Address = address

	client, err := vault.NewClient(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create Vault client: %w", err)
	}

	client.SetToken(token)

	return &VaultClient{client: client}, nil
}

// GetSecret retrieves a secret from Vault
func (v *VaultClient) GetSecret(path string) (string, error) {
	secret, err := v.client.Logical().Read(path)
	if err != nil {
		return "", fmt.Errorf("failed to read secret: %w", err)
	}

	if secret == nil {
		return "", fmt.Errorf("secret not found at path: %s", path)
	}

	// Extract the value (assuming it's stored under "value" key)
	if val, ok := secret.Data["value"].(string); ok {
		return val, nil
	}

	return "", fmt.Errorf("secret value not found or invalid type")
}

// SetSecret stores a secret in Vault
func (v *VaultClient) SetSecret(path, value string) error {
	data := map[string]interface{}{
		"value": value,
	}

	_, err := v.client.Logical().Write(path, data)
	if err != nil {
		return fmt.Errorf("failed to write secret: %w", err)
	}

	return nil
}

// RotateAPIKey rotates an API key for a provider
func (v *VaultClient) RotateAPIKey(provider string) error {
	// This would integrate with the provider's API to generate a new key
	// Then store it in Vault
	path := fmt.Sprintf("secret/data/talos/%s/api_key", provider)

	// Placeholder - would call provider API to generate new key
	newKey := "rotated-key-placeholder"

	return v.SetSecret(path, newKey)
}

// GetDatabaseCredentials retrieves database credentials
func (v *VaultClient) GetDatabaseCredentials() (string, string, error) {
	userSecret, err := v.GetSecret("secret/data/talos/database/user")
	if err != nil {
		return "", "", err
	}

	passSecret, err := v.GetSecret("secret/data/talos/database/password")
	if err != nil {
		return "", "", err
	}

	return userSecret, passSecret, nil
}
