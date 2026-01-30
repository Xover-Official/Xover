package errors

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"runtime"
	"time"

	"github.com/google/uuid"
	"go.opentelemetry.io/otel/trace"
)

// ErrorCode represents different types of errors in the system
type ErrorCode string

const (
	// Validation errors
	ErrInvalidInput     ErrorCode = "VALIDATION_INVALID_INPUT"
	ErrMissingParameter ErrorCode = "VALIDATION_MISSING_PARAMETER"
	ErrInvalidFormat    ErrorCode = "VALIDATION_INVALID_FORMAT"

	// Authentication/Authorization errors
	ErrUnauthorized     ErrorCode = "AUTH_UNAUTHORIZED"
	ErrForbidden        ErrorCode = "AUTH_FORBIDDEN"
	ErrTokenExpired     ErrorCode = "AUTH_TOKEN_EXPIRED"
	ErrInvalidToken     ErrorCode = "AUTH_INVALID_TOKEN"

	// Resource errors
	ErrResourceNotFound ErrorCode = "RESOURCE_NOT_FOUND"
	ErrResourceExists   ErrorCode = "RESOURCE_ALREADY_EXISTS"
	ErrResourceLocked   ErrorCode = "RESOURCE_LOCKED"

	// Cloud provider errors
	ErrCloudAPIError    ErrorCode = "CLOUD_API_ERROR"
	ErrCloudTimeout     ErrorCode = "CLOUD_TIMEOUT"
	ErrCloudRateLimit   ErrorCode = "CLOUD_RATE_LIMIT"
	ErrCloudQuotaExceeded ErrorCode = "CLOUD_QUOTA_EXCEEDED"

	// AI/ML errors
	ErrAIServiceUnavailable ErrorCode = "AI_SERVICE_UNAVAILABLE"
	ErrAIModelNotFound     ErrorCode = "AI_MODEL_NOT_FOUND"
	ErrAIRequestFailed     ErrorCode = "AI_REQUEST_FAILED"
	ErrAIInsufficientTokens ErrorCode = "AI_INSUFFICIENT_TOKENS"

	// System errors
	ErrDatabaseError    ErrorCode = "DATABASE_ERROR"
	ErrCacheError       ErrorCode = "CACHE_ERROR"
	ErrNetworkError     ErrorCode = "NETWORK_ERROR"
	ErrInternalError    ErrorCode = "INTERNAL_ERROR"
	ErrServiceUnavailable ErrorCode = "SERVICE_UNAVAILABLE"

	// Business logic errors
	ErrOptimizationFailed ErrorCode = "OPTIMIZATION_FAILED"
	ErrRiskTooHigh       ErrorCode = "RISK_TOO_HIGH"
	ErrInsufficientData  ErrorCode = "INSUFFICIENT_DATA"
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

// ErrorHandler provides centralized error handling
type ErrorHandler struct {
	logger *log.Logger
}

// NewErrorHandler creates a new error handler
func NewErrorHandler(logger *log.Logger) *ErrorHandler {
	if logger == nil {
		logger = log.Default()
	}
	return &ErrorHandler{logger: logger}
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
	talosErr := NewInternalError(err.Error(), err).
		WithTrace(ctx).
		Build()
	
	h.logError(talosErr)
	return talosErr
}

// logError logs the error based on severity
func (h *ErrorHandler) logError(err *TalosError) {
	logMsg := fmt.Sprintf("Error [%s]: %s", err.ID, err.Error())
	
	if err.Context != nil && len(err.Context) > 0 {
		if contextJSON, ctxErr := json.Marshal(err.Context); ctxErr == nil {
			logMsg += fmt.Sprintf(" | Context: %s", string(contextJSON))
		}
	}
	
	if err.TraceID != "" {
		logMsg += fmt.Sprintf(" | TraceID: %s", err.TraceID)
	}
	
	switch err.Severity {
	case SeverityLow:
		h.logger.Print(logMsg)
	case SeverityMedium:
		h.logger.Print(logMsg)
	case SeverityHigh:
		h.logger.Printf("WARNING: %s", logMsg)
	case SeverityCritical:
		h.logger.Printf("CRITICAL: %s", logMsg)
		
		// For critical errors, also log stack trace
		if len(err.StackTrace) > 0 {
			h.logger.Printf("Stack Trace: %+v", err.StackTrace)
		}
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

// ErrorMetrics tracks error statistics
type ErrorMetrics struct {
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

// Record records an error in metrics
func (m *ErrorMetrics) Record(err *TalosError) {
	m.TotalErrors++
	m.ErrorsByCode[err.Code]++
	
	hour := err.Timestamp.Format("2006-01-02T15")
	m.ErrorsByHour[hour]++
	
	if err.Severity == SeverityCritical {
		m.CriticalErrors++
	}
}

// GetStats returns current error statistics
func (m *ErrorMetrics) GetStats() map[string]interface{} {
	return map[string]interface{}{
		"total_errors":     m.TotalErrors,
		"critical_errors":  m.CriticalErrors,
		"errors_by_code":   m.ErrorsByCode,
		"errors_by_hour":   m.ErrorsByHour,
		"top_errors":       m.getTopErrors(),
	}
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
