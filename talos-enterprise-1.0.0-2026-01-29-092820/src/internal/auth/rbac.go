package auth

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// Role represents a user role in the system
type Role string

const (
	RoleAdmin    Role = "admin"
	RoleOperator Role = "operator"
	RoleViewer   Role = "viewer"
)

// Permission represents an action on a resource
type Permission struct {
	Resource string // e.g., "resources", "actions", "settings"
	Action   string // e.g., "read", "write", "delete"
}

// User represents an authenticated user
type User struct {
	ID             string
	Email          string
	OrganizationID string
	Role           Role
}

// HasPermission checks if a role has a specific permission
func (r Role) HasPermission(p Permission) bool {
	permissions := map[Role][]Permission{
		RoleAdmin: {
			{Resource: "*", Action: "*"}, // Admin has all permissions
		},
		RoleOperator: {
			{Resource: "resources", Action: "read"},
			{Resource: "resources", Action: "write"},
			{Resource: "actions", Action: "read"},
			{Resource: "actions", Action: "write"},
			{Resource: "settings", Action: "read"},
		},
		RoleViewer: {
			{Resource: "resources", Action: "read"},
			{Resource: "actions", Action: "read"},
			{Resource: "settings", Action: "read"},
		},
	}

	rolePerms, ok := permissions[r]
	if !ok {
		return false
	}

	for _, perm := range rolePerms {
		// Wildcard match
		if perm.Resource == "*" && perm.Action == "*" {
			return true
		}
		// Exact match
		if perm.Resource == p.Resource && perm.Action == p.Action {
			return true
		}
		// Resource wildcard
		if perm.Resource == p.Resource && perm.Action == "*" {
			return true
		}
	}

	return false
}

// JWTManager manages JWT tokens
type JWTManager struct {
	secretKey     string
	tokenDuration time.Duration
}

// NewJWTManager creates a new JWT manager
func NewJWTManager(secretKey string, tokenDuration time.Duration) *JWTManager {
	return &JWTManager{
		secretKey:     secretKey,
		tokenDuration: tokenDuration,
	}
}

// Claims represents JWT claims
type Claims struct {
	UserID         string `json:"user_id"`
	Email          string `json:"email"`
	OrganizationID string `json:"org_id"`
	Role           Role   `json:"role"`
	jwt.RegisteredClaims
}

// Generate creates a new JWT token
func (m *JWTManager) Generate(user User) (string, error) {
	claims := Claims{
		UserID:         user.ID,
		Email:          user.Email,
		OrganizationID: user.OrganizationID,
		Role:           user.Role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(m.tokenDuration)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(m.secretKey))
}

// Verify validates a JWT token and returns the claims
func (m *JWTManager) Verify(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(m.secretKey), nil
	})

	if err != nil {
		return nil, err
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, errors.New("invalid token")
	}

	return claims, nil
}
