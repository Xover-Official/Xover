package main

import (
	"errors"
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/Xover-Official/Xover/internal/auth"
	"github.com/Xover-Official/Xover/internal/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockSSOProvider is a mock implementation of the SSOProvider interface
type MockSSOProvider struct {
	mock.Mock
}

func (m *MockSSOProvider) Authenticate(ctx context.Context, code string) (*auth.SSOUser, error) {
	args := m.Called(ctx, code)
	var user *auth.SSOUser
	if args.Get(0) != nil {
		user = args.Get(0).(*auth.SSOUser)
	}
	return user, args.Error(1)
}

func (m *MockSSOProvider) ValidateToken(ctx context.Context, token string) (*auth.SSOUser, error) {
	args := m.Called(ctx, token)
	return args.Get(0).(*auth.SSOUser), args.Error(1)
}

func (m *MockSSOProvider) GetAuthURL(redirectURI string) string {
	args := m.Called(redirectURI)
	return args.String(0)
}

// MockUserStore is a mock implementation of the UserStore interface
type MockUserStore struct {
	mock.Mock
}

func (m *MockUserStore) Upsert(ssoUser *auth.SSOUser) (*auth.User, error) {
	args := m.Called(ssoUser)
	var user *auth.User
	if args.Get(0) != nil {
		user = args.Get(0).(*auth.User)
	}
	return user, args.Error(1)
}

func TestAuthMiddleware(t *testing.T) {
	jwtMgr := auth.NewJWTManager("test-secret", time.Hour)
	srv := &server{jwtManager: jwtMgr}

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

func TestHandleCallback(t *testing.T) {
	jwtMgr := auth.NewJWTManager("test-secret", time.Hour)
	mockSSOProvider := new(MockSSOProvider)
	mockUserStore := new(MockUserStore)

	srv := &server{
		jwtManager: jwtMgr,
		userStore:  mockUserStore,
		logger:     zap.NewNop(),
		config: &config.Config{
			SSO: config.SSOConfig{
				Google: config.SSOProviderConfig{ClientID: "test-client-id", ClientSecret: "test-client-secret"},
			},
		},
	}

	// Override the getSSOProvider to return our mock
	srv.getSSOProvider = func(name string) (auth.SSOProvider, error) {
		if name == "google" {
			return mockSSOProvider, nil
		}
		return nil, errors.New("unknown provider")
	}

	t.Run("Successful callback", func(t *testing.T) {
		ssoUser := &auth.SSOUser{ID: "sso-123", Email: "test@example.com"}
		dbUser := &auth.User{ID: "user-1", Email: "test@example.com", Role: auth.RoleAdmin}

		mockSSOProvider.On("Authenticate", mock.Anything, "good-code").Return(ssoUser, nil).Once()
		mockUserStore.On("Upsert", ssoUser).Return(dbUser, nil).Once()

		req := httptest.NewRequest("GET", "/auth/callback/google?code=good-code", nil)
		rr := httptest.NewRecorder()

		srv.handleCallback(rr, req)

		assert.Equal(t, http.StatusTemporaryRedirect, rr.Code)
		assert.Equal(t, "/", rr.Header().Get("Location"))

		cookie := rr.Result().Cookies()[0]
		assert.Equal(t, "atlas_token", cookie.Name)
		assert.NotEmpty(t, cookie.Value)

		claims, err := jwtMgr.Verify(cookie.Value)
		assert.NoError(t, err)
		assert.Equal(t, dbUser.ID, claims.ID)
		assert.Equal(t, dbUser.Email, claims.Email)

		mockSSOProvider.AssertExpectations(t)
		mockUserStore.AssertExpectations(t)
	})

	t.Run("SSO authentication fails", func(t *testing.T) {
		mockSSOProvider.On("Authenticate", mock.Anything, "bad-code").Return(nil, errors.New("invalid code")).Once()

		req := httptest.NewRequest("GET", "/auth/callback/google?code=bad-code", nil)
		rr := httptest.NewRecorder()

		srv.handleCallback(rr, req)

		assert.Equal(t, http.StatusInternalServerError, rr.Code)
		assert.Contains(t, rr.Body.String(), "sso authentication failed")
		mockSSOProvider.AssertExpectations(t)
	})

	t.Run("User resolution fails", func(t *testing.T) {
		ssoUser := &auth.SSOUser{ID: "sso-123", Email: "test@example.com"}
		mockSSOProvider.On("Authenticate", mock.Anything, "good-code-bad-user").Return(ssoUser, nil).Once()
		mockUserStore.On("Upsert", ssoUser).Return(nil, errors.New("db error")).Once()

		req := httptest.NewRequest("GET", "/auth/callback/google?code=good-code-bad-user", nil)
		rr := httptest.NewRecorder()
		srv.handleCallback(rr, req)

		assert.Equal(t, http.StatusInternalServerError, rr.Code)
		assert.Contains(t, rr.Body.String(), "failed to process user login")
	})
}
