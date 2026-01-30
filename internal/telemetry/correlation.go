package telemetry

import (
	"context"

	"github.com/google/uuid"
)

type contextKey string

const (
	correlationIDKey contextKey = "correlation_id"
	requestIDKey     contextKey = "request_id"
	sessionIDKey     contextKey = "session_id"
)

// WithCorrelationID adds a correlation ID to the context
func WithCorrelationID(ctx context.Context, id string) context.Context {
	if id == "" {
		id = uuid.New().String()
	}
	return context.WithValue(ctx, correlationIDKey, id)
}

// GetCorrelationID retrieves the correlation ID from context
func GetCorrelationID(ctx context.Context) string {
	if id, ok := ctx.Value(correlationIDKey).(string); ok {
		return id
	}
	return ""
}

// WithRequestID adds a request ID to the context
func WithRequestID(ctx context.Context, id string) context.Context {
	if id == "" {
		id = uuid.New().String()
	}
	return context.WithValue(ctx, requestIDKey, id)
}

// GetRequestID retrieves the request ID from context
func GetRequestID(ctx context.Context) string {
	if id, ok := ctx.Value(requestIDKey).(string); ok {
		return id
	}
	return ""
}

// WithSessionID adds a session ID to the context
func WithSessionID(ctx context.Context, id string) context.Context {
	if id == "" {
		id = uuid.New().String()
	}
	return context.WithValue(ctx, sessionIDKey, id)
}

// GetSessionID retrieves the session ID from context
func GetSessionID(ctx context.Context) string {
	if id, ok := ctx.Value(sessionIDKey).(string); ok {
		return id
	}
	return ""
}

// NewContext creates a new context with all tracing IDs
func NewContext() context.Context {
	ctx := context.Background()
	ctx = WithCorrelationID(ctx, "")
	ctx = WithRequestID(ctx, "")
	ctx = WithSessionID(ctx, "")
	return ctx
}
