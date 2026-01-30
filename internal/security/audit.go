package security

import (
	"encoding/json"
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/google/uuid"
)

// AuditLevel defines the severity of the audit log
type AuditLevel string

const (
	AuditLevelInfo     AuditLevel = "INFO"
	AuditLevelWarning  AuditLevel = "WARNING"
	AuditLevelCritical AuditLevel = "CRITICAL"
)

// AuditEntry represents a single audit log entry
type AuditEntry struct {
	ID        string                 `json:"id"`
	Timestamp time.Time              `json:"timestamp"`
	ActorID   string                 `json:"actor_id"` // User or System Component
	Action    string                 `json:"action"`
	Resource  string                 `json:"resource"`
	Changes   map[string]interface{} `json:"changes,omitempty"`
	Status    string                 `json:"status"` // SUCCESS, FAILURE
	IPAddress string                 `json:"ip_address,omitempty"`
	Level     AuditLevel             `json:"level"`
	Metadata  map[string]string      `json:"metadata,omitempty"`
	Signature string                 `json:"signature"` // HMAC for immutability check
}

// AuditLogger handles audit logging
type AuditLogger struct {
	mu         sync.Mutex
	logFile    *os.File
	signingKey []byte // For HMAC signature
}

// NewAuditLogger creates a new audit logger
func NewAuditLogger(filePath string, signingKey string) (*AuditLogger, error) {
	f, err := os.OpenFile(filePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0600)
	if err != nil {
		return nil, err
	}

	return &AuditLogger{
		logFile:    f,
		signingKey: []byte(signingKey),
	}, nil
}

// Log records an action
func (l *AuditLogger) Log(actorID, action, resource, ip string, level AuditLevel, changes map[string]interface{}) error {
	entry := AuditEntry{
		ID:        uuid.New().String(),
		Timestamp: time.Now().UTC(),
		ActorID:   actorID,
		Action:    action,
		Resource:  resource,
		Changes:   changes,
		Status:    "SUCCESS",
		IPAddress: ip,
		Level:     level,
	}

	// Generate signature (simplified)
	entry.Signature = l.generateSignature(entry)

	// Serialize
	data, err := json.Marshal(entry)
	if err != nil {
		return err
	}

	l.mu.Lock()
	defer l.mu.Unlock()

	// Write to file with newline
	if _, err := l.logFile.Write(append(data, '\n')); err != nil {
		return err
	}

	return nil
}

func (l *AuditLogger) generateSignature(entry AuditEntry) string {
	// In production, use HMAC-SHA256 with fields
	return fmt.Sprintf("sig_%s_%d", entry.ID, entry.Timestamp.Unix())
}

// Close closes the log file
func (l *AuditLogger) Close() error {
	return l.logFile.Close()
}
