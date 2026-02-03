package main

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/Xover-Official/Xover/internal/auth"
	"go.uber.org/zap"
)

type contextKey string

// userContextKey is a type-safe key for storing user claims in the request context.
const userContextKey = contextKey("user")

func (s *server) handleLogin(w http.ResponseWriter, r *http.Request) {
	providerName := strings.TrimPrefix(r.URL.Path, "/auth/login/")
	provider, err := s.getSSOProvider(providerName)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, err.Error())
		return
	}

	// Make redirect URI scheme-aware for production environments (e.g., behind HTTPS proxy)
	scheme := "http"
	if r.TLS != nil || r.Header.Get("X-Forwarded-Proto") == "https" {
		scheme = "https"
	}
	redirectURI := fmt.Sprintf("%s://%s/auth/callback/%s", scheme, r.Host, providerName)
	authURL := provider.GetAuthURL(redirectURI)
	http.Redirect(w, r, authURL, http.StatusTemporaryRedirect)
}

func (s *server) handleCallback(w http.ResponseWriter, r *http.Request) {
	providerName := strings.TrimPrefix(r.URL.Path, "/auth/callback/")
	provider, err := s.getSSOProvider(providerName)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, err.Error())
		return
	}

	code := r.URL.Query().Get("code")
	ssoUser, err := provider.Authenticate(r.Context(), code)
	if err != nil {
		s.logger.Error("sso authentication failed", zap.Error(err))
		respondWithError(w, http.StatusInternalServerError, "sso authentication failed")
		return
	}

	user, err := s.resolveUserFromSSO(ssoUser)
	if err != nil {
		s.logger.Error("failed to resolve user from sso", zap.Error(err))
		respondWithError(w, http.StatusInternalServerError, "failed to process user login")
		return
	}

	token, err := s.jwtManager.Generate(*user)
	if err != nil {
		s.logger.Error("failed to generate jwt", zap.Error(err))
		respondWithError(w, http.StatusInternalServerError, "failed to generate token")
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "atlas_token",
		Value:    token,
		Path:     "/",
		Expires:  time.Now().Add(s.config.JWT.TokenDuration), // Use configured duration
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
		if err != nil { // If cookie is not set, redirect to login.
			http.Redirect(w, r, "/login", http.StatusTemporaryRedirect)
			return
		}

		claims, err := s.jwtManager.Verify(cookie.Value)
		if err != nil {
			http.Redirect(w, r, "/login", http.StatusTemporaryRedirect)
			return
		}

		ctx := context.WithValue(r.Context(), userContextKey, claims)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func (s *server) requirePermission(permission auth.Permission, next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userClaims, ok := r.Context().Value(userContextKey).(*auth.Claims)
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

// getSSOProvider is a helper to instantiate an SSO provider by name.
func (s *server) getSSOProvider(name string) (auth.SSOProvider, error) {
	switch name {
	case "google":
		return auth.NewGoogleWorkspaceProvider(s.config.SSO.Google.ClientID, s.config.SSO.Google.ClientSecret), nil
	case "okta":
		return auth.NewOktaProvider(s.config.SSO.Okta.Domain, s.config.SSO.Okta.ClientID, s.config.SSO.Okta.ClientSecret), nil
	case "azure":
		return auth.NewAzureADProvider(s.config.SSO.Azure.TenantID, s.config.SSO.Azure.ClientID, s.config.SSO.Azure.ClientSecret), nil
	default:
		return nil, fmt.Errorf("unknown provider: %s", name)
	}
}

// resolveUserFromSSO handles the logic of finding or creating a user from SSO data.
func (s *server) resolveUserFromSSO(ssoUser *auth.SSOUser) (*auth.User, error) {
	// Use the UserStore to find or create the user.
	// This decouples the handler from the database implementation.
	user, err := s.userStore.Upsert(ssoUser)
	if err != nil {
		s.logger.Error("failed to upsert user from sso",
			zap.String("user_email", ssoUser.Email),
			zap.Error(err),
		)
		return nil, fmt.Errorf("failed to process user login")
	}

	s.logger.Info("resolved user from sso",
		zap.String("user_email", user.Email),
		zap.String("user_id", user.ID),
	)
	return user, nil
}
