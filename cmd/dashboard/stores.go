package main

import "github.com/Xover-Official/Xover/internal/auth"

// UserStore defines the interface for user persistence.
// This allows for swapping the backend (e.g., Postgres, Mongo) without changing business logic.
type UserStore interface {
	// Upsert creates a new user or updates an existing one based on SSO data.
	Upsert(ssoUser *auth.SSOUser) (*auth.User, error)
}

// InMemoryUserStore is a temporary, non-production-ready implementation of UserStore.
// TODO: Replace with a real database-backed implementation of UserStore.
type InMemoryUserStore struct {
	users map[string]*auth.User
}

// NewInMemoryUserStore creates a new in-memory user store.
func NewInMemoryUserStore() *InMemoryUserStore {
	// Pre-populate with a user for demonstration purposes.
	return &InMemoryUserStore{
		users: map[string]*auth.User{
			"test@example.com": {
				ID:             "user-1",
				Email:          "test@example.com",
				OrganizationID: "org-123",
				Role:           auth.RoleAdmin, // Give admin for testing
			},
		},
	}
}

// Upsert finds a user by email or creates a new one.
func (s *InMemoryUserStore) Upsert(ssoUser *auth.SSOUser) (*auth.User, error) {
	// In a real implementation, this would query the database.
	// If the user exists, update their details. If not, create them.
	// The logic for assigning OrganizationID and Role would be more complex,
	// potentially based on email domain, SSO groups, etc.
	if user, ok := s.users[ssoUser.Email]; ok {
		return user, nil
	}

	// Create new user with default role if not found.
	newUser := &auth.User{
		ID:             ssoUser.ID,
		Email:          ssoUser.Email,
		OrganizationID: "org-123",       // FIXME: Hardcoded organization ID
		Role:           auth.RoleViewer, // FIXME: Default role assignment
	}
	s.users[ssoUser.Email] = newUser
	return newUser, nil
}
