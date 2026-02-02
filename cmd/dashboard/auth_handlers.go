package main

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/project-atlas/atlas/internal/auth"
	"go.uber.org/zap"
)

func (s *server) handleLogin(w http.ResponseWriter, r *http.Request) {
	providerName := strings.TrimPrefix(r.URL.Path, "/auth/login/")
	var provider auth.SSOProvider
	switch providerName {
	case "google":
		provider = auth.NewGoogleWorkspaceProvider(s.config.SSO.Google.ClientID, s.config.SSO.Google.ClientSecret)
	case "okta":
		provider = auth.NewOktaProvider(s.config.SSO.Okta.Domain, s.config.SSO.Okta.ClientID, s.config.SSO.Okta.ClientSecret)
	case "azure":
		provider = auth.NewAzureADProvider(s.config.SSO.Azure.TenantID, s.config.SSO.Azure.ClientID, s.config.SSO.Azure.ClientSecret)
	default:
		respondWithError(w, http.StatusBadRequest, "unknown provider")
		return
	}

	redirectURI := fmt.Sprintf("http://%s/auth/callback/%s", r.Host, providerName)
	authURL := provider.GetAuthURL(redirectURI)
	http.Redirect(w, r, authURL, http.StatusTemporaryRedirect)
}

func (s *server) handleCallback(w http.ResponseWriter, r *http.Request) {
	providerName := strings.TrimPrefix(r.URL.Path, "/auth/callback/")
	var provider auth.SSOProvider
	switch providerName {
	case "google":
		provider = auth.NewGoogleWorkspaceProvider(s.config.SSO.Google.ClientID, s.config.SSO.Google.ClientSecret)
	case "okta":
		provider = auth.NewOktaProvider(s.config.SSO.Okta.Domain, s.config.SSO.Okta.ClientID, s.config.SSO.Okta.ClientSecret)
	case "azure":
		provider = auth.NewAzureADProvider(s.config.SSO.Azure.TenantID, s.config.SSO.Azure.ClientID, s.config.SSO.Azure.ClientSecret)
	default:
		respondWithError(w, http.StatusBadRequest, "unknown provider")
		return
	}

	code := r.URL.Query().Get("code")
	ssoUser, err := provider.Authenticate(r.Context(), code)
	if err != nil {
		s.logger.Error("sso authentication failed", zap.Error(err))
		respondWithError(w, http.StatusInternalServerError, "sso authentication failed")
		return
	}

	// In a real application, you would look up the user in your database
	// and assign roles based on the ssoUser.Groups or other attributes.
	// For this example, we'll create a new user with a default role.
	user := auth.User{
		ID:             ssoUser.ID,
		Email:          ssoUser.Email,
		OrganizationID: "org-123",       // Example org
		Role:           auth.RoleViewer, // Default role
	}

	token, err := s.jwtManager.Generate(user)
	if err != nil {
		s.logger.Error("failed to generate jwt", zap.Error(err))
		respondWithError(w, http.StatusInternalServerError, "failed to generate token")
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "atlas_token",
		Value:    token,
		Path:     "/",
		Expires:  time.Now().Add(24 * time.Hour), // Use 24 hours as default
		HttpOnly: true,
		Secure:   r.TLS != nil,
	})

	http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
}

func (s *server) handleLogout(w http.ResponseWriter, r *http.Request) {
	http.SetCookie(w, &http.Cookie{
		Name:     "atlas_token",
		Value:    "",
		Path:     "/",
		Expires:  time.Unix(0, 0),
		HttpOnly: true,
		Secure:   r.TLS != nil,
	})
	http.Redirect(w, r, "/login", http.StatusTemporaryRedirect)
}

func (s *server) authMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie("atlas_token")
		if err != nil {
			if r.URL.Path == "/login" || strings.HasPrefix(r.URL.Path, "/auth/") {
				next.ServeHTTP(w, r)
				return
			}
			http.Redirect(w, r, "/login", http.StatusTemporaryRedirect)
			return
		}

		claims, err := s.jwtManager.Verify(cookie.Value)
		if err != nil {
			http.Redirect(w, r, "/login", http.StatusTemporaryRedirect)
			return
		}

		ctx := context.WithValue(r.Context(), "user", claims)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func (s *server) requirePermission(permission auth.Permission, next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userClaims, ok := r.Context().Value("user").(*auth.Claims)
		if !ok {
			respondWithError(w, http.StatusUnauthorized, "no user in context")
			return
		}

		if !userClaims.Role.HasPermission(permission) {
			respondWithError(w, http.StatusForbidden, "insufficient permissions")
			return
		}

		next.ServeHTTP(w, r)
	}
}
