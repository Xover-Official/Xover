package security

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
)

// SecurityAuditEvent represents a security-related audit event
type SecurityAuditEvent struct {
	Timestamp time.Time              `json:"timestamp"`
	EventType string                 `json:"event_type"`
	UserID    string                 `json:"user_id,omitempty"`
	Username  string                 `json:"username,omitempty"`
	IPAddress string                 `json:"ip_address"`
	UserAgent string                 `json:"user_agent"`
	Resource  string                 `json:"resource,omitempty"`
	Action    string                 `json:"action,omitempty"`
	Success   bool                   `json:"success"`
	Reason    string                 `json:"reason,omitempty"`
	RiskScore int                    `json:"risk_score"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
	RequestID string                 `json:"request_id"`
}

// EnhancedSecurityManager handles all security-related operations with enhanced audit logging
type EnhancedSecurityManager struct {
	jwtSecret     []byte
	tokenExpiry   time.Duration
	refreshExpiry time.Duration
	logger        *zap.Logger
	rateLimiter   *RateLimiter
	auditLogger   *zap.Logger
}

// NewEnhancedSecurityManager creates a new security manager with audit logging
func NewEnhancedSecurityManager(jwtSecret string, tokenExpiry, refreshExpiry time.Duration, logger *zap.Logger) *EnhancedSecurityManager {
	// Create dedicated audit logger
	auditLogger := logger.WithOptions(zap.IncreaseLevel(zap.InfoLevel)).Named("security_audit")

	return &EnhancedSecurityManager{
		jwtSecret:     []byte(jwtSecret),
		tokenExpiry:   tokenExpiry,
		refreshExpiry: refreshExpiry,
		logger:        logger,
		rateLimiter:   NewRateLimiter(100, time.Hour), // 100 requests per hour
		auditLogger:   auditLogger,
	}
}

// EnhancedClaims represents JWT claims with enhanced security
type EnhancedClaims struct {
	UserID    string   `json:"user_id"`
	Username  string   `json:"username"`
	Roles     []string `json:"roles"`
	SessionID string   `json:"session_id"`
	LastLogin int64    `json:"last_login"`
	JTI       string   `json:"jti"` // JWT ID for token revocation
	jwt.RegisteredClaims
}

// GenerateTokenPair generates access and refresh tokens with audit logging
func (sm *EnhancedSecurityManager) GenerateTokenPair(userID, username string, roles []string, ipAddress, userAgent string) (accessToken, refreshToken string, err error) {
	requestID := sm.generateRequestID()

	// Log token generation attempt
	sm.logSecurityEvent(SecurityAuditEvent{
		Timestamp: time.Now(),
		EventType: "token_generation_attempt",
		UserID:    userID,
		IPAddress: ipAddress,
		UserAgent: userAgent,
		Resource:  "jwt_token",
		Action:    "generate",
		Success:   false,
		RequestID: requestID,
		RiskScore: sm.calculateRiskScore(ipAddress, userAgent),
	})

	now := time.Now()
	sessionID := sm.generateSessionID()
	jti := sm.generateJTI()

	// Generate access token
	accessClaims := &EnhancedClaims{
		UserID:    userID,
		Username:  username,
		Roles:     roles,
		SessionID: sessionID,
		LastLogin: now.Unix(),
		JTI:       jti,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(sm.tokenExpiry)),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
			Issuer:    "talos-atlas",
			Subject:   userID,
			ID:        jti,
		},
	}

	accessToken, err = jwt.NewWithClaims(jwt.SigningMethodHS256, accessClaims).SignedString(sm.jwtSecret)
	if err != nil {
		sm.logSecurityEvent(SecurityAuditEvent{
			Timestamp: time.Now(),
			EventType: "token_generation_failed",
			UserID:    userID,
			IPAddress: ipAddress,
			UserAgent: userAgent,
			Resource:  "jwt_token",
			Action:    "generate",
			Success:   false,
			Reason:    err.Error(),
			RequestID: requestID,
			RiskScore: 8, // High risk for token generation failure
		})
		return "", "", fmt.Errorf("failed to generate access token: %w", err)
	}

	// Generate refresh token
	refreshClaims := &EnhancedClaims{
		UserID:    userID,
		Username:  username,
		Roles:     roles,
		SessionID: sessionID,
		LastLogin: now.Unix(),
		JTI:       sm.generateJTI(),
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(sm.refreshExpiry)),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
			Issuer:    "talos-atlas",
			Subject:   userID,
		},
	}

	refreshToken, err = jwt.NewWithClaims(jwt.SigningMethodHS256, refreshClaims).SignedString(sm.jwtSecret)
	if err != nil {
		sm.logSecurityEvent(SecurityAuditEvent{
			Timestamp: time.Now(),
			EventType: "token_generation_failed",
			UserID:    userID,
			IPAddress: ipAddress,
			UserAgent: userAgent,
			Resource:  "jwt_refresh_token",
			Action:    "generate",
			Success:   false,
			Reason:    err.Error(),
			RequestID: requestID,
			RiskScore: 8,
		})
		return "", "", fmt.Errorf("failed to generate refresh token: %w", err)
	}

	// Log successful token generation
	sm.logSecurityEvent(SecurityAuditEvent{
		Timestamp: time.Now(),
		EventType: "token_generation_success",
		UserID:    userID,
		IPAddress: ipAddress,
		UserAgent: userAgent,
		Resource:  "jwt_token",
		Action:    "generate",
		Success:   true,
		RequestID: requestID,
		RiskScore: sm.calculateRiskScore(ipAddress, userAgent),
		Metadata: map[string]interface{}{
			"session_id": sessionID,
			"expires_at": now.Add(sm.tokenExpiry),
		},
	})

	return accessToken, refreshToken, nil
}

// ValidateToken validates a JWT token with comprehensive audit logging
func (sm *EnhancedSecurityManager) ValidateToken(tokenString, ipAddress, userAgent string) (*EnhancedClaims, error) {
	requestID := sm.generateRequestID()

	// Log validation attempt
	sm.logSecurityEvent(SecurityAuditEvent{
		Timestamp: time.Now(),
		EventType: "token_validation_attempt",
		IPAddress: ipAddress,
		UserAgent: userAgent,
		Resource:  "jwt_token",
		Action:    "validate",
		Success:   false,
		RequestID: requestID,
		RiskScore: sm.calculateRiskScore(ipAddress, userAgent),
	})

	token, err := jwt.ParseWithClaims(tokenString, &EnhancedClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return sm.jwtSecret, nil
	})

	if err != nil {
		sm.logSecurityEvent(SecurityAuditEvent{
			Timestamp: time.Now(),
			EventType: "token_validation_failed",
			IPAddress: ipAddress,
			UserAgent: userAgent,
			Resource:  "jwt_token",
			Action:    "validate",
			Success:   false,
			Reason:    err.Error(),
			RequestID: requestID,
			RiskScore: 9, // High risk for validation failure
		})
		return nil, fmt.Errorf("failed to parse token: %w", err)
	}

	if claims, ok := token.Claims.(*EnhancedClaims); ok && token.Valid {
		// Additional security checks
		if claims.SessionID == "" {
			sm.logSecurityEvent(SecurityAuditEvent{
				Timestamp: time.Now(),
				EventType: "token_validation_failed",
				UserID:    claims.UserID,
				IPAddress: ipAddress,
				UserAgent: userAgent,
				Resource:  "jwt_token",
				Action:    "validate",
				Success:   false,
				Reason:    "missing session_id",
				RequestID: requestID,
				RiskScore: 8,
			})
			return nil, fmt.Errorf("invalid token: missing session_id")
		}

		// Log successful validation
		sm.logSecurityEvent(SecurityAuditEvent{
			Timestamp: time.Now(),
			EventType: "token_validation_success",
			UserID:    claims.UserID,
			Username:  claims.Username,
			IPAddress: ipAddress,
			UserAgent: userAgent,
			Resource:  "jwt_token",
			Action:    "validate",
			Success:   true,
			RequestID: requestID,
			RiskScore: sm.calculateRiskScore(ipAddress, userAgent),
			Metadata: map[string]interface{}{
				"session_id": claims.SessionID,
				"roles":      claims.Roles,
				"issuer":     claims.Issuer,
			},
		})

		return claims, nil
	} else {
		sm.logSecurityEvent(SecurityAuditEvent{
			Timestamp: time.Now(),
			EventType: "token_validation_failed",
			IPAddress: ipAddress,
			UserAgent: userAgent,
			Resource:  "jwt_token",
			Action:    "validate",
			Success:   false,
			Reason:    "invalid token claims",
			RequestID: requestID,
			RiskScore: 9,
		})
		return nil, fmt.Errorf("invalid token")
	}
}

// HashPassword hashes a password with audit logging
func (sm *EnhancedSecurityManager) HashPassword(password string) (string, error) {
	requestID := sm.generateRequestID()

	sm.logSecurityEvent(SecurityAuditEvent{
		Timestamp: time.Now(),
		EventType: "password_hash_attempt",
		Resource:  "user_password",
		Action:    "hash",
		Success:   false,
		RequestID: requestID,
		RiskScore: 3, // Low risk for password hashing
	})

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		sm.logSecurityEvent(SecurityAuditEvent{
			Timestamp: time.Now(),
			EventType: "password_hash_failed",
			Resource:  "user_password",
			Action:    "hash",
			Success:   false,
			Reason:    err.Error(),
			RequestID: requestID,
			RiskScore: 5,
		})
		return "", fmt.Errorf("failed to hash password: %w", err)
	}

	sm.logSecurityEvent(SecurityAuditEvent{
		Timestamp: time.Now(),
		EventType: "password_hash_success",
		Resource:  "user_password",
		Action:    "hash",
		Success:   true,
		RequestID: requestID,
		RiskScore: 3,
	})

	return string(hash), nil
}

// CheckPassword checks a password against hash with audit logging
func (sm *EnhancedSecurityManager) CheckPassword(hash, password, userID, ipAddress, userAgent string) error {
	requestID := sm.generateRequestID()

	sm.logSecurityEvent(SecurityAuditEvent{
		Timestamp: time.Now(),
		EventType: "password_check_attempt",
		UserID:    userID,
		IPAddress: ipAddress,
		UserAgent: userAgent,
		Resource:  "user_password",
		Action:    "check",
		Success:   false,
		RequestID: requestID,
		RiskScore: sm.calculateRiskScore(ipAddress, userAgent),
	})

	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	if err != nil {
		sm.logSecurityEvent(SecurityAuditEvent{
			Timestamp: time.Now(),
			EventType: "password_check_failed",
			UserID:    userID,
			IPAddress: ipAddress,
			UserAgent: userAgent,
			Resource:  "user_password",
			Action:    "check",
			Success:   false,
			Reason:    "invalid password",
			RequestID: requestID,
			RiskScore: 7, // High risk for failed password check
		})
		return fmt.Errorf("invalid password")
	}

	sm.logSecurityEvent(SecurityAuditEvent{
		Timestamp: time.Now(),
		EventType: "password_check_success",
		UserID:    userID,
		IPAddress: ipAddress,
		UserAgent: userAgent,
		Resource:  "user_password",
		Action:    "check",
		Success:   true,
		RequestID: requestID,
		RiskScore: sm.calculateRiskScore(ipAddress, userAgent),
	})

	return nil
}

// logSecurityEvent logs a security audit event
func (sm *EnhancedSecurityManager) logSecurityEvent(event SecurityAuditEvent) {
	// Convert to JSON for structured logging
	eventJSON, _ := json.Marshal(event)

	// Log with appropriate level based on risk score
	switch {
	case event.RiskScore >= 8:
		sm.auditLogger.Error("high_risk_security_event",
			zap.String("event_json", string(eventJSON)),
			zap.String("event_type", event.EventType),
			zap.String("user_id", event.UserID),
			zap.String("ip_address", event.IPAddress),
			zap.Int("risk_score", event.RiskScore),
		)
	case event.RiskScore >= 5:
		sm.auditLogger.Warn("medium_risk_security_event",
			zap.String("event_json", string(eventJSON)),
			zap.String("event_type", event.EventType),
			zap.String("user_id", event.UserID),
			zap.String("ip_address", event.IPAddress),
			zap.Int("risk_score", event.RiskScore),
		)
	default:
		sm.auditLogger.Info("security_event",
			zap.String("event_json", string(eventJSON)),
			zap.String("event_type", event.EventType),
			zap.String("user_id", event.UserID),
			zap.String("ip_address", event.IPAddress),
			zap.Int("risk_score", event.RiskScore),
		)
	}
}

// calculateRiskScore calculates a risk score based on IP and user agent
func (sm *EnhancedSecurityManager) calculateRiskScore(ipAddress, userAgent string) int {
	risk := 0

	// Check for suspicious IP patterns
	if strings.Contains(ipAddress, "127.0.0.1") || strings.Contains(ipAddress, "::1") {
		risk += 1 // Localhost - low risk
	} else if strings.HasPrefix(ipAddress, "10.") || strings.HasPrefix(ipAddress, "192.168.") {
		risk += 2 // Private IP - low-medium risk
	} else {
		risk += 3 // Public IP - medium risk
	}

	// Check user agent patterns
	if userAgent == "" {
		risk += 3 // No user agent - suspicious
	} else if strings.Contains(strings.ToLower(userAgent), "bot") || strings.Contains(strings.ToLower(userAgent), "crawler") {
		risk += 2 // Bot/crawler - medium risk
	} else {
		risk += 1 // Normal user agent - low risk
	}

	return risk
}

// generateRequestID generates a unique request ID
func (sm *EnhancedSecurityManager) generateRequestID() string {
	b := make([]byte, 16)
	rand.Read(b)
	return base64.URLEncoding.EncodeToString(b)
}

// generateSessionID generates a unique session ID
func (sm *EnhancedSecurityManager) generateSessionID() string {
	b := make([]byte, 32)
	rand.Read(b)
	return base64.URLEncoding.EncodeToString(b)
}

// generateJTI generates a unique JWT ID
func (sm *EnhancedSecurityManager) generateJTI() string {
	b := make([]byte, 16)
	rand.Read(b)
	return base64.URLEncoding.EncodeToString(b)
}

// GetSecurityMiddleware returns HTTP middleware with security audit logging
func (sm *EnhancedSecurityManager) GetSecurityMiddleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Extract IP address
			ipAddress := r.RemoteAddr
			if forwarded := r.Header.Get("X-Forwarded-For"); forwarded != "" {
				ipAddress = strings.Split(forwarded, ",")[0]
			}

			// Log request
			sm.logSecurityEvent(SecurityAuditEvent{
				Timestamp: time.Now(),
				EventType: "http_request",
				IPAddress: ipAddress,
				UserAgent: r.Header.Get("User-Agent"),
				Resource:  r.URL.Path,
				Action:    r.Method,
				Success:   true,
				RequestID: sm.generateRequestID(),
				RiskScore: sm.calculateRiskScore(ipAddress, r.Header.Get("User-Agent")),
				Metadata: map[string]interface{}{
					"method": r.Method,
					"path":   r.URL.Path,
					"query":  r.URL.RawQuery,
				},
			})

			next.ServeHTTP(w, r)
		})
	}
}
