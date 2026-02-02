package telemetry

import (
	"context"
	"fmt"
	"log"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/baggage"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/exporters/stdout/stdouttrace"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/trace"
)

// TelemetryConfig holds configuration for telemetry
type TelemetryConfig struct {
	ServiceName    string  `yaml:"service_name"`
	ServiceVersion string  `yaml:"service_version"`
	Environment    string  `yaml:"environment"`
	Enabled        bool    `yaml:"enabled"`
	JaegerEndpoint string  `yaml:"jaeger_endpoint"`
	OTLPEndpoint   string  `yaml:"otlp_endpoint"`
	SampleRate     float64 `yaml:"sample_rate"`
}

// TelemetryManager manages distributed tracing
type TelemetryManager struct {
	config       TelemetryConfig
	tracer       trace.Tracer
	shutdownFunc func(context.Context) error
	enabled      bool
}

// NewTelemetryManager creates a new telemetry manager
func NewTelemetryManager(config TelemetryConfig) (*TelemetryManager, error) {
	if !config.Enabled {
		return &TelemetryManager{
			config:  config,
			enabled: false,
		}, nil
	}

	// Create resource
	res, err := resource.New(
		context.Background(),
		resource.WithAttributes(
			attribute.String("service.name", config.ServiceName),
			attribute.String("service.version", config.ServiceVersion),
			attribute.String("environment", config.Environment),
		),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create resource: %w", err)
	}

	// Create trace provider
	var exporter sdktrace.SpanExporter

	// Try OTLP exporter first
	if config.OTLPEndpoint != "" {
		exporter, err = otlptracegrpc.New(
			context.Background(),
			otlptracegrpc.WithEndpoint(config.OTLPEndpoint),
			otlptracegrpc.WithInsecure(),
		)
		if err != nil {
			log.Printf("Failed to create OTLP exporter: %v", err)
		}
	}

	// Fallback to stdout
	if exporter == nil {
		exporter, err = stdouttrace.New(
			stdouttrace.WithPrettyPrint(),
		)
		if err != nil {
			return nil, fmt.Errorf("failed to create stdout exporter: %w", err)
		}
	}

	// Create trace provider
	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exporter),
		sdktrace.WithResource(res),
		sdktrace.WithSampler(sdktrace.TraceIDRatioBased(config.SampleRate)),
	)

	// Register globally
	otel.SetTracerProvider(tp)
	otel.SetTextMapPropagator(propagation.TraceContext{})

	return &TelemetryManager{
		config:       config,
		tracer:       tp.Tracer(config.ServiceName),
		shutdownFunc: tp.Shutdown,
		enabled:      true,
	}, nil
}

// Shutdown shuts down the telemetry manager
func (tm *TelemetryManager) Shutdown(ctx context.Context) error {
	if tm.shutdownFunc != nil {
		return tm.shutdownFunc(ctx)
	}
	return nil
}

// StartSpan starts a new span
func (tm *TelemetryManager) StartSpan(ctx context.Context, name string, opts ...trace.SpanStartOption) (context.Context, trace.Span) {
	if !tm.enabled {
		return ctx, trace.SpanFromContext(ctx)
	}
	return tm.tracer.Start(ctx, name, opts...)
}

// SpanFromContext extracts a span from context
func (tm *TelemetryManager) SpanFromContext(ctx context.Context) trace.Span {
	return trace.SpanFromContext(ctx)
}

// AddSpanAttributes adds attributes to the current span
func (tm *TelemetryManager) AddSpanAttributes(ctx context.Context, attrs ...attribute.KeyValue) {
	if !tm.enabled {
		return
	}

	span := trace.SpanFromContext(ctx)
	if span.IsRecording() {
		span.SetAttributes(attrs...)
	}
}

// AddSpanEvent adds an event to the current span
func (tm *TelemetryManager) AddSpanEvent(ctx context.Context, name string, attrs ...attribute.KeyValue) {
	if !tm.enabled {
		return
	}

	span := trace.SpanFromContext(ctx)
	if span.IsRecording() {
		span.AddEvent(name, trace.WithAttributes(attrs...))
	}
}

// RecordError records an error in the current span
func (tm *TelemetryManager) RecordError(ctx context.Context, err error, attrs ...attribute.KeyValue) {
	if !tm.enabled {
		return
	}

	span := trace.SpanFromContext(ctx)
	if span.IsRecording() {
		span.RecordError(err, trace.WithAttributes(attrs...))
	}
}

// SetSpanStatus sets the status of the current span
func (tm *TelemetryManager) SetSpanStatus(ctx context.Context, code codes.Code, message string) {
	if !tm.enabled {
		return
	}

	span := trace.SpanFromContext(ctx)
	if span.IsRecording() {
		span.SetStatus(code, message)
	}
}

// WithSpan creates a new span and executes the function
func (tm *TelemetryManager) WithSpan(ctx context.Context, name string, fn func(context.Context) error, opts ...trace.SpanStartOption) error {
	if !tm.enabled {
		return fn(ctx)
	}

	ctx, span := tm.StartSpan(ctx, name, opts...)
	defer span.End()

	err := fn(ctx)
	if err != nil {
		tm.RecordError(ctx, err)
		tm.SetSpanStatus(ctx, codes.Error, err.Error())
	} else {
		tm.SetSpanStatus(ctx, codes.Ok, "Operation completed successfully")
	}

	return err
}

// TraceFunction wraps a function with tracing
func (tm *TelemetryManager) TraceFunction(ctx context.Context, functionName string, fn func(context.Context) error) error {
	return tm.WithSpan(ctx, fmt.Sprintf("function.%s", functionName), fn)
}

// TraceHTTPRequest traces an HTTP request
func (tm *TelemetryManager) TraceHTTPRequest(ctx context.Context, method, url string, fn func(context.Context) error) error {
	attrs := []attribute.KeyValue{
		attribute.String("http.method", method),
		attribute.String("http.url", url),
	}

	return tm.WithSpan(ctx, fmt.Sprintf("http.%s", method), func(ctx context.Context) error {
		tm.AddSpanAttributes(ctx, attrs...)
		return fn(ctx)
	})
}

// TraceDatabaseOperation traces a database operation
func (tm *TelemetryManager) TraceDatabaseOperation(ctx context.Context, operation, table string, fn func(context.Context) error) error {
	attrs := []attribute.KeyValue{
		attribute.String("db.operation", operation),
		attribute.String("db.table", table),
	}

	return tm.WithSpan(ctx, fmt.Sprintf("db.%s", operation), func(ctx context.Context) error {
		tm.AddSpanAttributes(ctx, attrs...)
		return fn(ctx)
	})
}

// TraceCloudOperation traces a cloud operation
func (tm *TelemetryManager) TraceCloudOperation(ctx context.Context, provider, operation, resourceType string, fn func(context.Context) error) error {
	attrs := []attribute.KeyValue{
		attribute.String("cloud.provider", provider),
		attribute.String("cloud.operation", operation),
		attribute.String("cloud.resource_type", resourceType),
	}

	return tm.WithSpan(ctx, fmt.Sprintf("cloud.%s.%s", provider, operation), func(ctx context.Context) error {
		tm.AddSpanAttributes(ctx, attrs...)
		return fn(ctx)
	})
}

// TraceAIOperation traces an AI operation
func (tm *TelemetryManager) TraceAIOperation(ctx context.Context, model, operation string, fn func(context.Context) error) error {
	attrs := []attribute.KeyValue{
		attribute.String("ai.model", model),
		attribute.String("ai.operation", operation),
	}

	return tm.WithSpan(ctx, fmt.Sprintf("ai.%s.%s", model, operation), func(ctx context.Context) error {
		tm.AddSpanAttributes(ctx, attrs...)
		return fn(ctx)
	})
}

// GetTraceID returns the trace ID from context
func (tm *TelemetryManager) GetTraceID(ctx context.Context) string {
	if !tm.enabled {
		return ""
	}

	spanCtx := trace.SpanFromContext(ctx).SpanContext()
	return spanCtx.TraceID().String()
}

// GetSpanID returns the span ID from context
func (tm *TelemetryManager) GetSpanID(ctx context.Context) string {
	if !tm.enabled {
		return ""
	}

	spanCtx := trace.SpanFromContext(ctx).SpanContext()
	return spanCtx.SpanID().String()
}

// CreateSpanBaggage creates baggage items for context propagation
func (tm *TelemetryManager) CreateSpanBaggage(ctx context.Context, items map[string]string) context.Context {
	if !tm.enabled {
		return ctx
	}

	bag := baggage.FromContext(ctx)
	for key, value := range items {
		member, err := baggage.NewMember(key, value)
		if err == nil {
			bag, err = bag.SetMember(member)
			if err != nil {
				log.Printf("Failed to set baggage member: %v", err)
				continue
			}
		}
	}

	return baggage.ContextWithBaggage(ctx, bag)
}

// GetBaggageItem retrieves a baggage item from context
func (tm *TelemetryManager) GetBaggageItem(ctx context.Context, key string) string {
	if !tm.enabled {
		return ""
	}

	bag := baggage.FromContext(ctx)
	member := bag.Member(key)
	if member.Value() != "" {
		return member.Value()
	}

	return ""
}

// Metrics represents application metrics
type Metrics struct {
	// HTTP metrics
	HTTPRequestsTotal   int64         `json:"http_requests_total"`
	HTTPRequestDuration time.Duration `json:"http_request_duration"`

	// Cloud metrics
	CloudOperationsTotal   int64         `json:"cloud_operations_total"`
	CloudOperationDuration time.Duration `json:"cloud_operation_duration"`

	// AI metrics
	AIRequestsTotal   int64         `json:"ai_requests_total"`
	AIRequestDuration time.Duration `json:"ai_request_duration"`

	// Error metrics
	ErrorCount int64 `json:"error_count"`
}

// RecordMetrics records application metrics
func (tm *TelemetryManager) RecordMetrics(ctx context.Context, metrics Metrics) {
	if !tm.enabled {
		return
	}

	span := trace.SpanFromContext(ctx)
	if span.IsRecording() {
		span.SetAttributes(
			attribute.Int64("metrics.http_requests_total", metrics.HTTPRequestsTotal),
			attribute.Int64("metrics.http_request_duration_ms", metrics.HTTPRequestDuration.Milliseconds()),
			attribute.Int64("metrics.cloud_operations_total", metrics.CloudOperationsTotal),
			attribute.Int64("metrics.cloud_operation_duration_ms", metrics.CloudOperationDuration.Milliseconds()),
			attribute.Int64("metrics.ai_requests_total", metrics.AIRequestsTotal),
			attribute.Int64("metrics.ai_request_duration_ms", metrics.AIRequestDuration.Milliseconds()),
			attribute.Int64("metrics.error_count", metrics.ErrorCount),
		)
	}
}

// DefaultTelemetryConfig returns a default configuration
func DefaultTelemetryConfig() TelemetryConfig {
	return TelemetryConfig{
		ServiceName:    "talos-atlas",
		ServiceVersion: "1.0.0",
		Environment:    "development",
		Enabled:        true,
		SampleRate:     1.0, // Sample all traces in development
	}
}

// ProductionTelemetryConfig returns a production configuration
func ProductionTelemetryConfig() TelemetryConfig {
	return TelemetryConfig{
		ServiceName:    "talos-atlas",
		ServiceVersion: "1.0.0",
		Environment:    "production",
		Enabled:        true,
		SampleRate:     0.1, // Sample 10% of traces in production
	}
}
