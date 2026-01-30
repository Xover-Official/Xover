package main

import (
	"context" // Fix 1: Added missing context import
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/project-atlas/atlas/internal/auth"
	"github.com/project-atlas/atlas/internal/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// Fix 2: Defined the server struct so the tests can reference it
type server struct {
	jwtManager *auth.JWTManager
	config     *config.Config
}

// MockSSOProvider is a mock implementation of the SSOProvider interface
type MockSSOProvider struct {
	mock.Mock
}

func (m *MockSSOProvider) Authenticate(ctx context.Context, code string) (*auth.SSOUser, error) {
	args := m.Called(ctx, code)
	return args.Get(0).(*auth.SSOUser), args.Error(1)
}

func (m *MockSSOProvider) ValidateToken(ctx context.Context, token string) (*auth.SSOUser, error) {
	args := m.Called(ctx, token)
	return args.Get(0).(*auth.SSOUser), args.Error(1)
}

func (m *MockSSOProvider) GetAuthURL(redirectURI string) string {
	args := m.Called(redirectURI)
	return args.String(0)
}

// NOTE: These methods (authMiddleware and handleLogin) need to be defined on the server struct 
// in your main application code for these tests to pass.

func TestAuthMiddleware(t *testing.T) {
	jwtMgr := auth.NewJWTManager("test-secret", time.Hour)
	srv := &server{jwtManager: jwtMgr}

	// Assuming your middleware implementation exists
	handler := srv.authMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	t.Run("No token redirects to login", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/", nil)
		rr := httptest.NewRecorder()
		handler.ServeHTTP(rr, req)
		assert.Equal(t, http.StatusTemporaryRedirect, rr.Code)
		assert.Equal(t, "/login", rr.Header().Get("Location"))
	})

	t.Run("Invalid token redirects to login", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/", nil)
		req.AddCookie(&http.Cookie{Name: "atlas_token", Value: "invalid-token"})
		rr := httptest.NewRecorder()
		handler.ServeHTTP(rr, req)
		assert.Equal(t, http.StatusTemporaryRedirect, rr.Code)
		assert.Equal(t, "/login", rr.Header().Get("Location"))
	})

	t.Run("Valid token allows access", func(t *testing.T) {
		user := auth.User{ID: "user-1", Email: "test@example.com", Role: auth.RoleViewer}
		token, _ := jwtMgr.Generate(user)

		req := httptest.NewRequest("GET", "/", nil)
		req.AddCookie(&http.Cookie{Name: "atlas_token", Value: token})
		rr := httptest.NewRecorder()
		handler.ServeHTTP(rr, req)
		assert.Equal(t, http.StatusOK, rr.Code)
	})
}

func TestHandleLogin(t *testing.T) {
	srv := &server{
		config: &config.Config{
			SSO: config.SSOConfig{
				Google: config.SSOProviderConfig{ClientID: "test-client-id", ClientSecret: "test-client-secret"},
			},
		},
	}
	req := httptest.NewRequest("GET", "/auth/login/google", nil)
	rr := httptest.NewRecorder()
	srv.handleLogin(rr, req)
	assert.Equal(t, http.StatusTemporaryRedirect, rr.Code)
	assert.Contains(t, rr.Header().Get("Location"), "accounts.google.com")
}