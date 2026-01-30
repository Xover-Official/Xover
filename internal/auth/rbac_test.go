package auth

import (
	"testing"
	"time"
)

func TestRole_HasPermission(t *testing.T) {
	tests := []struct {
		role       Role
		permission Permission
		expected   bool
	}{
		{RoleAdmin, Permission{Resource: "resources", Action: "delete"}, true},
		{RoleOperator, Permission{Resource: "resources", Action: "write"}, true},
		{RoleOperator, Permission{Resource: "settings", Action: "write"}, false},
		{RoleViewer, Permission{Resource: "resources", Action: "read"}, true},
		{RoleViewer, Permission{Resource: "resources", Action: "write"}, false},
	}

	for _, tt := range tests {
		result := tt.role.HasPermission(tt.permission)
		if result != tt.expected {
			t.Errorf("Role %s, Permission %+v: expected %v, got %v",
				tt.role, tt.permission, tt.expected, result)
		}
	}
}

func TestJWTManager_GenerateAndVerify(t *testing.T) {
	manager := NewJWTManager("test-secret-key", 24*time.Hour)

	user := User{
		ID:             "user-123",
		Email:          "test@example.com",
		OrganizationID: "org-456",
		Role:           RoleOperator,
	}

	// Generate token
	token, err := manager.Generate(user)
	if err != nil {
		t.Fatalf("Failed to generate token: %v", err)
	}

	// Verify token
	claims, err := manager.Verify(token)
	if err != nil {
		t.Fatalf("Failed to verify token: %v", err)
	}

	if claims.UserID != user.ID {
		t.Errorf("Expected UserID %s, got %s", user.ID, claims.UserID)
	}

	if claims.Role != user.Role {
		t.Errorf("Expected Role %s, got %s", user.Role, claims.Role)
	}
}

func TestJWTManager_VerifyInvalidToken(t *testing.T) {
	manager := NewJWTManager("test-secret-key", 24*time.Hour)

	_, err := manager.Verify("invalid-token")
	if err == nil {
		t.Error("Expected error for invalid token")
	}
}
