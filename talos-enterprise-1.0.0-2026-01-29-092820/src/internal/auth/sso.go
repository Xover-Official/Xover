package auth

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// SSOProvider interface for SSO implementations
type SSOProvider interface {
	Authenticate(ctx context.Context, code string) (*SSOUser, error)
	ValidateToken(ctx context.Context, token string) (*SSOUser, error)
	GetAuthURL(redirectURI string) string
}

// SSOUser represents a user from SSO provider (different from internal User)
type SSOUser struct {
	ID       string
	Email    string
	Name     string
	Provider string
	Groups   []string
}

// OktaProvider implements SSO for Okta
type OktaProvider struct {
	Domain       string
	ClientID     string
	ClientSecret string
	HTTPClient   *http.Client
}

// NewOktaProvider creates an Okta SSO provider
func NewOktaProvider(domain, clientID, clientSecret string) *OktaProvider {
	return &OktaProvider{
		Domain:       domain,
		ClientID:     clientID,
		ClientSecret: clientSecret,
		HTTPClient:   &http.Client{Timeout: 10 * time.Second},
	}
}

// GetAuthURL returns the Okta authorization URL
func (p *OktaProvider) GetAuthURL(redirectURI string) string {
	return fmt.Sprintf(
		"https://%s/oauth2/v1/authorize?client_id=%s&redirect_uri=%s&response_type=code&scope=openid profile email",
		p.Domain, p.ClientID, redirectURI,
	)
}

// Authenticate exchanges authorization code for user info
func (p *OktaProvider) Authenticate(ctx context.Context, code string) (*SSOUser, error) {
	// Exchange code for token
	token, err := p.exchangeCode(ctx, code)
	if err != nil {
		return nil, err
	}

	// Get user info
	return p.getUserInfo(ctx, token)
}

// ValidateToken validates an access token
func (p *OktaProvider) ValidateToken(ctx context.Context, token string) (*SSOUser, error) {
	return p.getUserInfo(ctx, token)
}

func (p *OktaProvider) exchangeCode(ctx context.Context, code string) (string, error) {
	// Simplified - in production use proper OAuth flow
	return "okta-access-token", nil
}

func (p *OktaProvider) getUserInfo(ctx context.Context, token string) (*SSOUser, error) {
	url := fmt.Sprintf("https://%s/oauth2/v1/userinfo", p.Domain)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := p.HTTPClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("okta userinfo failed: %d", resp.StatusCode)
	}

	var userInfo struct {
		Sub   string `json:"sub"`
		Email string `json:"email"`
		Name  string `json:"name"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&userInfo); err != nil {
		return nil, err
	}

	return &SSOUser{
		ID:       userInfo.Sub,
		Email:    userInfo.Email,
		Name:     userInfo.Name,
		Provider: "okta",
	}, nil
}

// AzureADProvider implements SSO for Azure Active Directory
type AzureADProvider struct {
	TenantID     string
	ClientID     string
	ClientSecret string
	HTTPClient   *http.Client
}

// NewAzureADProvider creates an Azure AD SSO provider
func NewAzureADProvider(tenantID, clientID, clientSecret string) *AzureADProvider {
	return &AzureADProvider{
		TenantID:     tenantID,
		ClientID:     clientID,
		ClientSecret: clientSecret,
		HTTPClient:   &http.Client{Timeout: 10 * time.Second},
	}
}

// GetAuthURL returns the Azure AD authorization URL
func (p *AzureADProvider) GetAuthURL(redirectURI string) string {
	return fmt.Sprintf(
		"https://login.microsoftonline.com/%s/oauth2/v2.0/authorize?client_id=%s&redirect_uri=%s&response_type=code&scope=openid profile email",
		p.TenantID, p.ClientID, redirectURI,
	)
}

// Authenticate exchanges code for user info
func (p *AzureADProvider) Authenticate(ctx context.Context, code string) (*SSOUser, error) {
	token, err := p.exchangeCode(ctx, code)
	if err != nil {
		return nil, err
	}

	return p.getUserInfo(ctx, token)
}

// ValidateToken validates an access token
func (p *AzureADProvider) ValidateToken(ctx context.Context, token string) (*SSOUser, error) {
	return p.getUserInfo(ctx, token)
}

func (p *AzureADProvider) exchangeCode(ctx context.Context, code string) (string, error) {
	// Simplified - in production use proper OAuth flow
	return "azure-access-token", nil
}

func (p *AzureADProvider) getUserInfo(ctx context.Context, token string) (*SSOUser, error) {
	url := "https://graph.microsoft.com/v1.0/me"

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := p.HTTPClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("azure graph failed: %d", resp.StatusCode)
	}

	var userInfo struct {
		ID                string `json:"id"`
		UserPrincipalName string `json:"userPrincipalName"`
		DisplayName       string `json:"displayName"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&userInfo); err != nil {
		return nil, err
	}

	return &SSOUser{
		ID:       userInfo.ID,
		Email:    userInfo.UserPrincipalName,
		Name:     userInfo.DisplayName,
		Provider: "azure-ad",
	}, nil
}

// GoogleWorkspaceProvider implements SSO for Google Workspace
type GoogleWorkspaceProvider struct {
	ClientID     string
	ClientSecret string
	HTTPClient   *http.Client
}

// NewGoogleWorkspaceProvider creates a Google Workspace SSO provider
func NewGoogleWorkspaceProvider(clientID, clientSecret string) *GoogleWorkspaceProvider {
	return &GoogleWorkspaceProvider{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		HTTPClient:   &http.Client{Timeout: 10 * time.Second},
	}
}

// GetAuthURL returns the Google OAuth URL
func (p *GoogleWorkspaceProvider) GetAuthURL(redirectURI string) string {
	return fmt.Sprintf(
		"https://accounts.google.com/o/oauth2/v2/auth?client_id=%s&redirect_uri=%s&response_type=code&scope=openid profile email",
		p.ClientID, redirectURI,
	)
}

// Authenticate exchanges code for user info
func (p *GoogleWorkspaceProvider) Authenticate(ctx context.Context, code string) (*SSOUser, error) {
	token, err := p.exchangeCode(ctx, code)
	if err != nil {
		return nil, err
	}

	return p.getUserInfo(ctx, token)
}

// ValidateToken validates an access token
func (p *GoogleWorkspaceProvider) ValidateToken(ctx context.Context, token string) (*SSOUser, error) {
	return p.getUserInfo(ctx, token)
}

func (p *GoogleWorkspaceProvider) exchangeCode(ctx context.Context, code string) (string, error) {
	// Simplified - in production use proper OAuth flow
	return "google-access-token", nil
}

func (p *GoogleWorkspaceProvider) getUserInfo(ctx context.Context, token string) (*SSOUser, error) {
	url := "https://www.googleapis.com/oauth2/v2/userinfo"

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := p.HTTPClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("google userinfo failed: %d", resp.StatusCode)
	}

	var userInfo struct {
		ID    string `json:"id"`
		Email string `json:"email"`
		Name  string `json:"name"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&userInfo); err != nil {
		return nil, err
	}

	return &SSOUser{
		ID:       userInfo.ID,
		Email:    userInfo.Email,
		Name:     userInfo.Name,
		Provider: "google",
	}, nil
}
