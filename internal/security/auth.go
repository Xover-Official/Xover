package security

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
	"go.uber.org/zap"
	"sync"
)

// SecurityManager handles all security-related operations
type SecurityManager struct {
	jwtSecret     []byte
	tokenExpiry   time.Duration
	refreshExpiry time.Duration
	logger        *zap.Logger
	rateLimiter   *RateLimiter
}

// NewSecurityManager creates a new security manager
func NewSecurityManager(jwtSecret string, tokenExpiry, refreshExpiry time.Duration, logger *zap.Logger) *SecurityManager {
	return &SecurityManager{
		jwtSecret:     []byte(jwtSecret),
		tokenExpiry:   tokenExpiry,
		refreshExpiry: refreshExpiry,
		logger:        logger,
		rateLimiter:   NewRateLimiter(100, time.Hour), // 100 requests per hour
	}
}

// Claims represents JWT claims
type Claims struct {
	UserID   string   `json:"user_id"`
	Username string   `json:"username"`
	Roles    []string `json:"roles"`
	jwt.RegisteredClaims
}

// GenerateTokenPair generates access and refresh tokens
func (sm *SecurityManager) GenerateTokenPair(userID, username string, roles []string) (accessToken, refreshToken string, err error) {
	now := time.Now()
	
	// Generate access token
	accessClaims := &Claims{
		UserID:   userID,
		Username: username,
		Roles:    roles,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(sm.tokenExpiry)),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
			Issuer:    "talos",
			Subject:   userID,
		},
	}

	accessToken, err = jwt.NewWithClaims(jwt.SigningMethodHS256, accessClaims).SignedString(sm.jwtSecret)
	if err != nil {
		return "", "", fmt.Errorf("failed to generate access token: %w", err)
	}

	// Generate refresh token
	refreshClaims := &Claims{
		UserID:   userID,
		Username: username,
		Roles:    roles,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(sm.refreshExpiry)),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
			Issuer:    "talos",
			Subject:   userID,
		},
	}

	refreshToken, err = jwt.NewWithClaims(jwt.SigningMethodHS256, refreshClaims).SignedString(sm.jwtSecret)
	if err != nil {
		return "", "", fmt.Errorf("failed to generate refresh token: %w", err)
	}

	return accessToken, refreshToken, nil
}

// ValidateToken validates a JWT token
func (sm *SecurityManager) ValidateToken(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return sm.jwtSecret, nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to parse token: %w", err)
	}

	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		return claims, nil
	}

	return nil, fmt.Errorf("invalid token")
}

// HashPassword hashes a password using bcrypt
func (sm *SecurityManager) HashPassword(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", fmt.Errorf("failed to hash password: %w", err)
	}
	return string(hash), nil
}

// CheckPassword checks if a password matches the hash
func (sm *SecurityManager) CheckPassword(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

// GenerateAPIKey generates a secure API key
func (sm *SecurityManager) GenerateAPIKey() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", fmt.Errorf("failed to generate random bytes: %w", err)
	}
	return base64.URLEncoding.EncodeToString(bytes), nil
}

// RateLimiter implements thread-safe rate limiting
type RateLimiter struct {
	mu       sync.RWMutex
	requests map[string][]time.Time
	limit    int
	window   time.Duration
}

// NewRateLimiter creates a new rate limiter
func NewRateLimiter(limit int, window time.Duration) *RateLimiter {
	return &RateLimiter{
		requests: make(map[string][]time.Time),
		limit:    limit,
		window:   window,
	}
}

// Allow checks if a request is allowed (thread-safe)
func (rl *RateLimiter) Allow(key string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	
	// Clean old requests
	requests, exists := rl.requests[key]
	if exists {
		var validRequests []time.Time
		for _, req := range requests {
			if now.Sub(req) < rl.window {
				validRequests = append(validRequests, req)
			}
		}
		rl.requests[key] = validRequests
	}
	
	// Check if under limit
	if len(rl.requests[key]) >= rl.limit {
		return false
	}
	
	// Add current request
	rl.requests[key] = append(rl.requests[key], now)
	return true
}

// SecurityMiddleware provides HTTP security middleware
type SecurityMiddleware struct {
	securityManager *SecurityManager
	logger          *zap.Logger
}

// NewSecurityMiddleware creates new security middleware
func NewSecurityMiddleware(sm *SecurityManager, logger *zap.Logger) *SecurityMiddleware {
	return &SecurityMiddleware{
		securityManager: sm,
		logger:          logger,
	}
}

// AuthMiddleware authenticates requests
func (sm *SecurityMiddleware) AuthMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Skip auth for health checks and metrics
		if r.URL.Path == "/health" || r.URL.Path == "/metrics" {
			next(w, r)
			return
		}

		// Check rate limiting
		clientIP := getClientIP(r)
		if !sm.securityManager.rateLimiter.Allow(clientIP) {
			sm.logger.Warn("Rate limit exceeded", zap.String("ip", clientIP))
			http.Error(w, "Rate limit exceeded", http.StatusTooManyRequests)
			return
		}

		// Extract token from Authorization header
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			http.Error(w, "Authorization header required", http.StatusUnauthorized)
			return
		}

		// Check Bearer token format
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			http.Error(w, "Invalid authorization format", http.StatusUnauthorized)
			return
		}

		// Validate token
		claims, err := sm.securityManager.ValidateToken(parts[1])
		if err != nil {
			sm.logger.Warn("Invalid token", zap.String("error", err.Error()))
			http.Error(w, "Invalid token", http.StatusUnauthorized)
			return
		}

		// Add user context to request
		ctx := context.WithValue(r.Context(), "user_claims", claims)
		next(w, r.WithContext(ctx))
	}
}

// CORSMiddleware adds CORS headers
func (sm *SecurityMiddleware) CORSMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		w.Header().Set("Access-Control-Max-Age", "86400")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next(w, r)
	}
}

// SecurityHeadersMiddleware adds security headers
func (sm *SecurityMiddleware) SecurityHeadersMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("X-Frame-Options", "DENY")
		w.Header().Set("X-XSS-Protection", "1; mode=block")
		w.Header().Set("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
		w.Header().Set("Content-Security-Policy", "default-src 'self'")
		
		next(w, r)
	}
}

// LoggingMiddleware logs HTTP requests
func (sm *SecurityMiddleware) LoggingMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		
		// Create response writer wrapper to capture status code
		wrapped := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}
		
		next(wrapped, r)
		
		duration := time.Since(start)
		sm.logger.Info("HTTP request",
			zap.String("method", r.Method),
			zap.String("path", r.URL.Path),
			zap.Int("status", wrapped.statusCode),
			zap.Duration("duration", duration),
			zap.String("ip", getClientIP(r)),
			zap.String("user_agent", r.UserAgent()),
		)
	}
}

// responseWriter wraps http.ResponseWriter to capture status code
type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

// getClientIP extracts the real client IP
func getClientIP(r *http.Request) string {
	// Check X-Forwarded-For header
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		ips := strings.Split(xff, ",")
		return strings.TrimSpace(ips[0])
	}
	
	// Check X-Real-IP header
	if xri := r.Header.Get("X-Real-IP"); xri != "" {
		return xri
	}
	
	// Fall back to RemoteAddr
	return strings.Split(r.RemoteAddr, ":")[0]
}

// InputValidator provides input validation utilities
type InputValidator struct{}

// NewInputValidator creates a new input validator
func NewInputValidator() *InputValidator {
	return &InputValidator{}
}

// ValidateAPIKey validates an API key format
func (iv *InputValidator) ValidateAPIKey(apiKey string) bool {
	if len(apiKey) < 32 {
		return false
	}
	
	// Check if it's valid base64
	_, err := base64.URLEncoding.DecodeString(apiKey)
	return err == nil
}

// ValidateResourceID validates a resource ID format
func (iv *InputValidator) ValidateResourceID(resourceID string) bool {
	if len(resourceID) < 1 || len(resourceID) > 128 {
		return false
	}
	
	// Allow alphanumeric, hyphens, and underscores
	for _, char := range resourceID {
		if !((char >= 'a' && char <= 'z') || 
			 (char >= 'A' && char <= 'Z') || 
			 (char >= '0' && char <= '9') || 
			 char == '-' || char == '_') {
			return false
		}
	}
	
	return true
}

// SanitizeInput sanitizes user input
func (iv *InputValidator) SanitizeInput(input string) string {
	// Remove potentially dangerous characters
	input = strings.ReplaceAll(input, "<", "&lt;")
	input = strings.ReplaceAll(input, ">", "&gt;")
	input = strings.ReplaceAll(input, "&", "&amp;")
	input = strings.ReplaceAll(input, "\"", "&quot;")
	input = strings.ReplaceAll(input, "'", "&#x27;")
	
	return input
}
