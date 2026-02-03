package errors

import (
	"context"
	"encoding/json"
	"fmt"
	"runtime"
	"time"

	"sync"

	"github.com/google/uuid"
	"github.com/Xover-Official/Xover/internal/logger"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
)

// ErrorCode represents different types of errors in the system
type ErrorCode string

const (
	// Validation errors
	ErrInvalidInput     ErrorCode = "VALIDATION_INVALID_INPUT"
	ErrMissingParameter ErrorCode = "VALIDATION_MISSING_PARAMETER"
	ErrInvalidFormat    ErrorCode = "VALIDATION_INVALID_FORMAT"

	// Authentication/Authorization errors
	ErrUnauthorized ErrorCode = "AUTH_UNAUTHORIZED"
	ErrForbidden    ErrorCode = "AUTH_FORBIDDEN"
	ErrTokenExpired ErrorCode = "AUTH_TOKEN_EXPIRED"
	ErrInvalidToken ErrorCode = "AUTH_INVALID_TOKEN"

	// Resource errors
	ErrResourceNotFound ErrorCode = "RESOURCE_NOT_FOUND"
	ErrResourceExists   ErrorCode = "RESOURCE_ALREADY_EXISTS"
	ErrResourceLocked   ErrorCode = "RESOURCE_LOCKED"

	// Cloud provider errors
	ErrCloudAPIError      ErrorCode = "CLOUD_API_ERROR"
	ErrCloudTimeout       ErrorCode = "CLOUD_TIMEOUT"
	ErrCloudRateLimit     ErrorCode = "CLOUD_RATE_LIMIT"
	ErrCloudQuotaExceeded ErrorCode = "CLOUD_QUOTA_EXCEEDED"

	// AI/ML errors
	ErrAIServiceUnavailable ErrorCode = "AI_SERVICE_UNAVAILABLE"
	ErrAIModelNotFound      ErrorCode = "AI_MODEL_NOT_FOUND"
	ErrAIRequestFailed      ErrorCode = "AI_REQUEST_FAILED"
	ErrAIInsufficientTokens ErrorCode = "AI_INSUFFICIENT_TOKENS"

	// System errors
	ErrDatabaseError      ErrorCode = "DATABASE_ERROR"
	ErrCacheError         ErrorCode = "CACHE_ERROR"
	ErrNetworkError       ErrorCode = "NETWORK_ERROR"
	ErrInternalError      ErrorCode = "INTERNAL_ERROR"
	ErrServiceUnavailable ErrorCode = "SERVICE_UNAVAILABLE"

	// Business logic errors
	ErrOptimizationFailed ErrorCode = "OPTIMIZATION_FAILED"
	ErrRiskTooHigh        ErrorCode = "RISK_TOO_HIGH"
	ErrInsufficientData   ErrorCode = "INSUFFICIENT_DATA"
)

// ErrorSeverity indicates the severity level of an error
type ErrorSeverity string

const (
	SeverityLow      ErrorSeverity = "low"
	SeverityMedium   ErrorSeverity = "medium"
	SeverityHigh     ErrorSeverity = "high"
	SeverityCritical ErrorSeverity = "critical"
)

// TalosError represents a structured error in the Talos system
type TalosError struct {
	ID          string                 `json:"id"`
	Code        ErrorCode              `json:"code"`
	Message     string                 `json:"message"`
	Description string                 `json:"description,omitempty"`
	Severity    ErrorSeverity          `json:"severity"`
	Timestamp   time.Time              `json:"timestamp"`
	Context     map[string]interface{} `json:"context,omitempty"`
	Cause       error                  `json:"-"`
	StackTrace  []string               `json:"stack_trace,omitempty"`
	Retryable   bool                   `json:"retryable"`
	RetryAfter  *time.Duration         `json:"retry_after,omitempty"`
	TraceID     string                 `json:"trace_id,omitempty"`
}

// Error implements the error interface
func (e *TalosError) Error() string {
	return fmt.Sprintf("[%s] %s: %s", e.Code, e.Message, e.Description)
}

// Unwrap returns the underlying cause
func (e *TalosError) Unwrap() error {
	return e.Cause
}

// WithContext adds context to the error
func (e *TalosError) WithContext(key string, value interface{}) *TalosError {
	if e.Context == nil {
		e.Context = make(map[string]interface{})
	}
	e.Context[key] = value
	return e
}

// WithTrace adds trace information
func (e *TalosError) WithTrace(ctx context.Context) *TalosError {
	if spanCtx := trace.SpanContextFromContext(ctx); spanCtx.IsValid() {
		e.TraceID = spanCtx.TraceID().String()
	}
	return e
}

// WithRetry configures retry behavior
func (e *TalosError) WithRetry(retryable bool, after time.Duration) *TalosError {
	e.Retryable = retryable
	e.RetryAfter = &after
	return e
}

// ToJSON converts the error to JSON
func (e *TalosError) ToJSON() ([]byte, error) {
	return json.Marshal(e)
}

// ErrorBuilder provides a fluent interface for building errors
type ErrorBuilder struct {
	error *TalosError
}

// NewErrorBuilder creates a new error builder
func NewErrorBuilder(code ErrorCode, message string) *ErrorBuilder {
	return &ErrorBuilder{
		error: &TalosError{
			ID:        uuid.New().String(),
			Code:      code,
			Message:   message,
			Timestamp: time.Now(),
			Context:   make(map[string]interface{}),
		},
	}
}

// Description sets the error description
func (b *ErrorBuilder) Description(desc string) *ErrorBuilder {
	b.error.Description = desc
	return b
}

// Severity sets the error severity
func (b *ErrorBuilder) Severity(severity ErrorSeverity) *ErrorBuilder {
	b.error.Severity = severity
	return b
}

// Cause sets the underlying cause
func (b *ErrorBuilder) Cause(cause error) *ErrorBuilder {
	b.error.Cause = cause
	return b
}

// Context adds context information
func (b *ErrorBuilder) Context(key string, value interface{}) *ErrorBuilder {
	b.error.Context[key] = value
	return b
}

// WithRetry configures retry behavior
func (b *ErrorBuilder) WithRetry(retryable bool, after time.Duration) *ErrorBuilder {
	b.error.Retryable = retryable
	b.error.RetryAfter = &after
	return b
}

// Build creates the final error
func (b *ErrorBuilder) Build() *TalosError {
	// Add stack trace for high and critical severity errors
	if b.error.Severity == SeverityHigh || b.error.Severity == SeverityCritical {
		b.error.StackTrace = getStackTrace()
	}

	// Set default severity if not specified
	if b.error.Severity == "" {
		b.error.Severity = SeverityMedium
	}

	return b.error
}

// Predefined error constructors

// NewValidationError creates a validation error
func NewValidationError(message string) *TalosError {
	return NewErrorBuilder(ErrInvalidInput, message).
		Severity(SeverityLow).
		Build()
}

// NewUnauthorizedError creates an unauthorized error
func NewUnauthorizedError(message string) *TalosError {
	return NewErrorBuilder(ErrUnauthorized, message).
		Severity(SeverityMedium).
		Build()
}

// NewResourceNotFoundError creates a resource not found error
func NewResourceNotFoundError(resourceType, resourceID string) *TalosError {
	return NewErrorBuilder(ErrResourceNotFound, fmt.Sprintf("%s not found", resourceType)).
		Description(fmt.Sprintf("The %s with ID %s could not be found", resourceType, resourceID)).
		Severity(SeverityMedium).
		Context("resource_type", resourceType).
		Context("resource_id", resourceID).
		Build()
}

// NewCloudAPIError creates a cloud API error
func NewCloudAPIError(provider, operation string, cause error) *TalosError {
	return NewErrorBuilder(ErrCloudAPIError, fmt.Sprintf("Cloud API error for %s", provider)).
		Description(fmt.Sprintf("Failed to execute %s operation on %s", operation, provider)).
		Severity(SeverityHigh).
		Context("provider", provider).
		Context("operation", operation).
		Cause(cause).
		WithRetry(true, 30*time.Second).
		Build()
}

// NewAIServiceError creates an AI service error
func NewAIServiceError(model, service string, cause error) *TalosError {
	return NewErrorBuilder(ErrAIServiceUnavailable, fmt.Sprintf("AI service unavailable: %s", service)).
		Description(fmt.Sprintf("Failed to get response from %s model %s", service, model)).
		Severity(SeverityHigh).
		Context("model", model).
		Context("service", service).
		Cause(cause).
		WithRetry(true, 10*time.Second).
		Build()
}

// NewOptimizationFailedError creates an optimization failed error
func NewOptimizationFailedError(resourceID, action string, cause error) *TalosError {
	return NewErrorBuilder(ErrOptimizationFailed, fmt.Sprintf("Optimization failed for resource %s", resourceID)).
		Description(fmt.Sprintf("Failed to execute %s action on resource %s", action, resourceID)).
		Severity(SeverityMedium).
		Context("resource_id", resourceID).
		Context("action", action).
		Cause(cause).
		Build()
}

// NewInternalError creates an internal error
func NewInternalError(message string, cause error) *TalosError {
	return NewErrorBuilder(ErrInternalError, message).
		Description("An internal system error occurred").
		Severity(SeverityCritical).
		Cause(cause).
		Build()
}

// ErrorHandler provides centralized error handling using zap
type ErrorHandler struct {
	logger *zap.Logger
}

// NewErrorHandler creates a new error handler with zap
func NewErrorHandler(l *zap.Logger) *ErrorHandler {
	if l == nil {
		l = logger.GetLogger()
	}
	return &ErrorHandler{logger: l}
}

// Handle processes an error and returns a user-friendly response
func (h *ErrorHandler) Handle(ctx context.Context, err error) *TalosError {
	if err == nil {
		return nil
	}

	// If it's already a TalosError, enhance it
	if talosErr, ok := err.(*TalosError); ok {
		talosErr.WithTrace(ctx)
		h.logError(talosErr)
		return talosErr
	}

	// Convert generic error to TalosError
	talosErr := NewInternalError(err.Error(), err)
	talosErr.WithTrace(ctx)

	h.logError(talosErr)
	return talosErr
}

// logError logs the error based on severity using structured zap logging
func (h *ErrorHandler) logError(err *TalosError) {
	fields := []zap.Field{
		zap.String("error_id", err.ID),
		zap.String("code", string(err.Code)),
		zap.String("message", err.Message),
		zap.String("severity", string(err.Severity)),
		zap.Time("timestamp", err.Timestamp),
	}

	if err.Context != nil {
		fields = append(fields, zap.Any("context", err.Context))
	}

	if err.TraceID != "" {
		fields = append(fields, zap.String("trace_id", err.TraceID))
	}

	if err.Cause != nil {
		fields = append(fields, zap.Error(err.Cause))
	}

	if len(err.StackTrace) > 0 {
		fields = append(fields, zap.Strings("stack_trace", err.StackTrace))
	}

	msg := fmt.Sprintf("Error occurred: %s", err.Message)

	switch err.Severity {
	case SeverityLow, SeverityMedium:
		h.logger.Info(msg, fields...)
	case SeverityHigh:
		h.logger.Warn(msg, fields...)
	case SeverityCritical:
		h.logger.Error(msg, fields...)
	default:
		h.logger.Info(msg, fields...)
	}
}

// RecoveryMiddleware provides panic recovery
type RecoveryMiddleware struct {
	handler *ErrorHandler
}

// NewRecoveryMiddleware creates a new recovery middleware
func NewRecoveryMiddleware(handler *ErrorHandler) *RecoveryMiddleware {
	return &RecoveryMiddleware{handler: handler}
}

// Recover recovers from panics and converts them to errors
func (r *RecoveryMiddleware) Recover() error {
	if p := recover(); p != nil {
		var err error
		switch x := p.(type) {
		case string:
			err = fmt.Errorf("panic: %s", x)
		case error:
			err = x
		default:
			err = fmt.Errorf("panic: %v", x)
		}

		return r.handler.Handle(context.Background(),
			NewInternalError("System panic recovered", err))
	}
	return nil
}

// getStackTrace captures the current stack trace
func getStackTrace() []string {
	var stack []string
	for i := 1; i < 10; i++ { // Limit to 10 frames
		pc, file, line, ok := runtime.Caller(i)
		if !ok {
			break
		}

		fn := runtime.FuncForPC(pc)
		if fn == nil {
			continue
		}

		frame := fmt.Sprintf("%s:%d %s", file, line, fn.Name())
		stack = append(stack, frame)
	}

	return stack
}

// ErrorMetrics tracks error statistics (thread-safe)
type ErrorMetrics struct {
	mu             sync.RWMutex
	TotalErrors    int64
	ErrorsByCode   map[ErrorCode]int64
	ErrorsByHour   map[string]int64
	CriticalErrors int64
}

// NewErrorMetrics creates new error metrics
func NewErrorMetrics() *ErrorMetrics {
	return &ErrorMetrics{
		ErrorsByCode: make(map[ErrorCode]int64),
		ErrorsByHour: make(map[string]int64),
	}
}

// Record records an error in metrics (thread-safe)
func (m *ErrorMetrics) Record(err *TalosError) {
	if err == nil {
		return
	}
	m.mu.Lock()
	defer m.mu.Unlock()

	m.TotalErrors++
	m.ErrorsByCode[err.Code]++

	hour := err.Timestamp.Format("2006-01-02T15")
	m.ErrorsByHour[hour]++

	// Cleanup old metrics to prevent memory leak (keep last 24 hours)
	if len(m.ErrorsByHour) > 48 {
		cutoff := time.Now().Add(-24 * time.Hour).Format("2006-01-02T15")
		for k := range m.ErrorsByHour {
			if k < cutoff {
				delete(m.ErrorsByHour, k)
			}
		}
	}

	if err.Severity == SeverityCritical {
		m.CriticalErrors++
	}
}

// GetStats returns current error statistics (thread-safe)
func (m *ErrorMetrics) GetStats() map[string]interface{} {
	m.mu.RLock()
	defer m.mu.RUnlock()

	stats := map[string]interface{}{
		"total_errors":    m.TotalErrors,
		"critical_errors": m.CriticalErrors,
		"top_errors":      m.getTopErrorsUnsafe(),
	}

	// Copy maps to avoid exposure of internal state
	codes := make(map[string]int64)
	for k, v := range m.ErrorsByCode {
		codes[string(k)] = v
	}
	stats["errors_by_code"] = codes

	hours := make(map[string]int64)
	for k, v := range m.ErrorsByHour {
		hours[k] = v
	}
	stats["errors_by_hour"] = hours

	return stats
}

// getTopErrorsUnsafe helper for GetStats
func (m *ErrorMetrics) getTopErrorsUnsafe() []map[string]interface{} {
	type errorCount struct {
		code  ErrorCode
		count int64
	}

	var errors []errorCount
	for code, count := range m.ErrorsByCode {
		errors = append(errors, errorCount{code, count})
	}

	// Simple sort by count (descending)
	for i := 0; i < len(errors); i++ {
		for j := i + 1; j < len(errors); j++ {
			if errors[j].count > errors[i].count {
				errors[i], errors[j] = errors[j], errors[i]
			}
		}
	}

	// Return top 10
	result := make([]map[string]interface{}, 0, 10)
	for i := 0; i < len(errors) && i < 10; i++ {
		result = append(result, map[string]interface{}{
			"code":  errors[i].code,
			"count": errors[i].count,
		})
	}

	return result
}

// getTopErrors returns the most frequent errors
func (m *ErrorMetrics) getTopErrors() []map[string]interface{} {
	type errorCount struct {
		code  ErrorCode
		count int64
	}

	var errors []errorCount
	for code, count := range m.ErrorsByCode {
		errors = append(errors, errorCount{code, count})
	}

	// Sort by count (descending)
	for i := 0; i < len(errors); i++ {
		for j := i + 1; j < len(errors); j++ {
			if errors[j].count > errors[i].count {
				errors[i], errors[j] = errors[j], errors[i]
			}
		}
	}

	// Return top 10
	result := make([]map[string]interface{}, 0, 10)
	for i := 0; i < len(errors) && i < 10; i++ {
		result = append(result, map[string]interface{}{
			"code":  errors[i].code,
			"count": errors[i].count,
		})
	}

	return result
}

// RecoveryStrategy defines how to recover from an error
type RecoveryStrategy struct {
	Type        RecoveryType    `json:"type"`
	Description string          `json:"description"`
	MaxAttempts int             `json:"max_attempts"`
	Delay       time.Duration   `json:"delay"`
	Backoff     BackoffStrategy `json:"backoff"`
}

// RecoveryType represents different recovery strategies
type RecoveryType string

const (
	RecoveryTypeRetry          RecoveryType = "retry"
	RecoveryTypeFallback       RecoveryType = "fallback"
	RecoveryTypeCircuitBreaker RecoveryType = "circuit_breaker"
	RecoveryTypeGraceful       RecoveryType = "graceful_degradation"
	RecoveryTypeManual         RecoveryType = "manual_intervention"
)

// BackoffStrategy defines backoff strategies for retries
type BackoffStrategy string

const (
	BackoffFixed       BackoffStrategy = "fixed"
	BackoffExponential BackoffStrategy = "exponential"
	BackoffLinear      BackoffStrategy = "linear"
	BackoffRandom      BackoffStrategy = "random"
)

// EnhancedTalosError represents an enhanced error with recovery capabilities
type EnhancedTalosError struct {
	*TalosError
	Retryable  bool                   `json:"retryable"`
	RetryCount int                    `json:"retry_count"`
	MaxRetries int                    `json:"max_retries"`
	RetryDelay time.Duration          `json:"retry_delay"`
	Recovery   *RecoveryStrategy      `json:"recovery"`
	Context    map[string]interface{} `json:"context,omitempty"`
	Metadata   map[string]interface{} `json:"metadata,omitempty"`
}

// NewError creates a basic TalosError
func NewError(code ErrorCode, message string, severity ErrorSeverity) *TalosError {
	return &TalosError{
		ID:        uuid.New().String(),
		Code:      code,
		Message:   message,
		Severity:  severity,
		Timestamp: time.Now(),
		Context:   make(map[string]interface{}),
	}
}

// NewEnhancedError creates an enhanced error with recovery capabilities
func NewEnhancedError(code ErrorCode, message string, severity ErrorSeverity) *EnhancedTalosError {
	return &EnhancedTalosError{
		TalosError: NewError(code, message, severity),
		Retryable:  true,
		RetryCount: 0,
		MaxRetries: 3,
		RetryDelay: time.Second * 5,
		Context:    make(map[string]interface{}),
		Metadata:   make(map[string]interface{}),
	}
}

// IsRetryable checks if an error is retryable
func (e *EnhancedTalosError) IsRetryable() bool {
	return e.Retryable && e.RetryCount < e.MaxRetries
}

// ShouldRetry determines if a retry should be attempted
func (e *EnhancedTalosError) ShouldRetry() bool {
	return e.IsRetryable() && e.RetryCount < e.MaxRetries
}

// IncrementRetryCount increments the retry count
func (e *EnhancedTalosError) IncrementRetryCount() {
	e.RetryCount++
}

// GetRetryDelay returns the next retry delay with backoff
func (e *EnhancedTalosError) GetRetryDelay() time.Duration {
	if e.Recovery == nil {
		return e.RetryDelay
	}

	switch e.Recovery.Backoff {
	case BackoffExponential:
		return e.RetryDelay * time.Duration(e.RetryCount+1)
	case BackoffLinear:
		return e.RetryDelay * time.Duration(e.RetryCount+1)
	case BackoffRandom:
		// Simple random backoff implementation
		return e.RetryDelay + time.Duration(e.RetryCount*int(time.Second))
	default:
		return e.RetryDelay
	}
}

// WithRetryable sets whether the error is retryable
func (e *EnhancedTalosError) WithRetryable(retryable bool) *EnhancedTalosError {
	e.Retryable = retryable
	return e
}

// WithMaxRetries sets the maximum number of retries
func (e *EnhancedTalosError) WithMaxRetries(maxRetries int) *EnhancedTalosError {
	e.MaxRetries = maxRetries
	return e
}

// WithRetryDelay sets the delay between retries
func (e *EnhancedTalosError) WithRetryDelay(delay time.Duration) *EnhancedTalosError {
	e.RetryDelay = delay
	return e
}

// WithRecovery sets the recovery strategy
func (e *EnhancedTalosError) WithRecovery(strategy *RecoveryStrategy) *EnhancedTalosError {
	e.Recovery = strategy
	return e
}

// WithContext adds context information
func (e *EnhancedTalosError) WithContext(key string, value interface{}) *EnhancedTalosError {
	if e.Context == nil {
		e.Context = make(map[string]interface{})
	}
	e.Context[key] = value
	return e
}

// WithCause adds cause information
func (e *EnhancedTalosError) WithCause(cause error) *EnhancedTalosError {
	e.Cause = cause
	return e
}

// Recovery helpers
func WithRetry(ctx context.Context, maxRetries int, delay time.Duration, fn func() error) error {
	var err error
	for i := 0; i < maxRetries; i++ {
		if ctx.Err() != nil {
			return ctx.Err()
		}
		err = fn()
		if err == nil {
			break
		}

		// Log retry attempt
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(delay):
		}
	}
	return err
}

func WithFallback(primary, fallback func() error) error {
	err := primary()
	if err == nil {
		return nil
	}

	// Log fallback
	return fallback()
}

func WithCircuitBreaker(threshold int, timeout time.Duration, fn func() error) error {
	// Simple circuit breaker implementation
	// In production, this would track failure rates and open/close the circuit
	return fn()
}

func WithGracefulDegradation(degradedService string, normal, degraded func() error) error {
	err := normal()
	if err == nil {
		return nil
	}

	// Log degradation
	return degraded()
}

// Common enhanced error builders
var (
	// Enhanced validation errors
	EnhancedErrValidation = func(field, message string) *EnhancedTalosError {
		return NewEnhancedError(ErrInvalidInput, fmt.Sprintf("Validation failed for %s: %s", field, message), SeverityMedium).
			WithRetryable(false).
			WithContext("field", field).
			WithContext("validation_message", message)
	}

	// Enhanced authentication errors
	EnhancedErrAuthentication = func(reason string) *EnhancedTalosError {
		return NewEnhancedError(ErrUnauthorized, fmt.Sprintf("Authentication failed: %s", reason), SeverityHigh).
			WithRetryable(false).
			WithContext("auth_reason", reason)
	}

	// Enhanced network errors
	EnhancedErrNetwork = func(operation, endpoint string, cause error) *EnhancedTalosError {
		return NewEnhancedError(ErrNetworkError, fmt.Sprintf("Network error during %s to %s: %v", operation, endpoint, cause), SeverityMedium).
			WithRetryable(true).
			WithMaxRetries(3).
			WithRetryDelay(time.Second*5).
			WithRecovery(&RecoveryStrategy{
				Type:        RecoveryTypeRetry,
				Description: "Retry with exponential backoff",
				MaxAttempts: 3,
				Delay:       time.Second * 5,
				Backoff:     BackoffExponential,
			}).
			WithContext("operation", operation).
			WithContext("endpoint", endpoint).
			WithCause(cause)
	}

	// Enhanced cloud API errors
	EnhancedErrCloudAPI = func(provider, operation, resource string, cause error) *EnhancedTalosError {
		return NewEnhancedError(ErrCloudAPIError, fmt.Sprintf("Cloud API error for %s during %s on %s: %v", provider, operation, resource, cause), SeverityMedium).
			WithRetryable(true).
			WithMaxRetries(3).
			WithRetryDelay(time.Second*10).
			WithRecovery(&RecoveryStrategy{
				Type:        RecoveryTypeRetry,
				Description: fmt.Sprintf("Retry %s API call with exponential backoff", provider),
				MaxAttempts: 3,
				Delay:       time.Second * 10,
				Backoff:     BackoffExponential,
			}).
			WithContext("provider", provider).
			WithContext("operation", operation).
			WithContext("resource", resource).
			WithCause(cause)
	}

	// Enhanced AI service errors
	EnhancedErrAIServiceUnavailable = func(service, model string, cause error) *EnhancedTalosError {
		return NewEnhancedError(ErrAIServiceUnavailable, fmt.Sprintf("AI service unavailable: %s model %s", service, model), SeverityMedium).
			WithRetryable(true).
			WithMaxRetries(2).
			WithRetryDelay(time.Second*15).
			WithRecovery(&RecoveryStrategy{
				Type:        RecoveryTypeFallback,
				Description: "Fallback to alternative AI model",
				MaxAttempts: 2,
				Delay:       time.Second * 15,
				Backoff:     BackoffFixed,
			}).
			WithContext("ai_service", service).
			WithContext("ai_model", model).
			WithCause(cause)
	}

	// Enhanced database errors
	EnhancedErrDatabase = func(operation, query string, cause error) *EnhancedTalosError {
		return NewEnhancedError(ErrDatabaseError, fmt.Sprintf("Database error during %s: %s (query: %s)", operation, cause, query), SeverityHigh).
			WithRetryable(true).
			WithMaxRetries(2).
			WithRetryDelay(time.Second*3).
			WithRecovery(&RecoveryStrategy{
				Type:        RecoveryTypeCircuitBreaker,
				Description: "Database circuit breaker",
				MaxAttempts: 2,
				Delay:       time.Second * 3,
				Backoff:     BackoffExponential,
			}).
			WithContext("operation", operation).
			WithContext("query", query).
			WithCause(cause)
	}
)
