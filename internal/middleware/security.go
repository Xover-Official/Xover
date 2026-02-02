package middleware

import (
	"fmt"
	"net"
	"net/http"
	"strings"
	"sync"
)

// SecurityConfig holds security configuration
type SecurityConfig struct {
	IPWhitelist []string
	GeoBanned   []string // ISO country codes
	TrustProxy  bool
}

// SecurityManager handles request filtering
type SecurityManager struct {
	config SecurityConfig
	mu     sync.RWMutex
}

// NewSecurityManager creates a new security manager
func NewSecurityManager(config SecurityConfig) *SecurityManager {
	return &SecurityManager{
		config: config,
	}
}

// IPWhitelistingMiddleware filters requests based on IP
func (m *SecurityManager) IPWhitelistingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		clientIP := m.getClientIP(r)

		if !m.isWhitelisted(clientIP) {
			http.Error(w, "Forbidden: IP not authorized", http.StatusForbidden)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// GeoFencingMiddleware filters requests based on location
func (m *SecurityManager) GeoFencingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		clientIP := m.getClientIP(r)
		country := m.getCountryFromIP(clientIP)

		if m.isBannedCountry(country) {
			http.Error(w, fmt.Sprintf("Forbidden: Access denied from %s", country), http.StatusForbidden)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func (m *SecurityManager) getClientIP(r *http.Request) string {
	if m.config.TrustProxy {
		// specific headers like X-Forwarded-For
		forwarded := r.Header.Get("X-Forwarded-For")
		if forwarded != "" {
			return strings.Split(forwarded, ",")[0]
		}
	}

	ip, _, _ := net.SplitHostPort(r.RemoteAddr)
	return ip
}

func (m *SecurityManager) isWhitelisted(ip string) bool {
	m.mu.RLock()
	defer m.mu.RUnlock()

	// Empty whitelist means allow all (unless restricted mode is on - safer default depends on policy)
	if len(m.config.IPWhitelist) == 0 {
		return true // Or false for strict mode
	}

	for _, allowed := range m.config.IPWhitelist {
		if allowed == ip || allowed == "*" {
			return true
		}
		// In production, add CIDR support
	}
	return false
}

func (m *SecurityManager) getCountryFromIP(ip string) string {
	// Mock implementation
	// In production, use GeoIP database (MaxMind, etc.)
	if ip == "1.2.3.4" {
		return "KP" // North Korea example
	}
	return "US"
}

func (m *SecurityManager) isBannedCountry(country string) bool {
	m.mu.RLock()
	defer m.mu.RUnlock()

	for _, banned := range m.config.GeoBanned {
		if banned == country {
			return true
		}
	}
	return false
}
