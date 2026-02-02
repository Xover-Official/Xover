package monitoring

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.uber.org/zap"
)

// Metrics holds all application metrics
type Metrics struct {
	// HTTP metrics
	httpRequestsTotal   *prometheus.CounterVec
	httpRequestDuration *prometheus.HistogramVec

	// Cloud operations
	cloudOperationsTotal     *prometheus.CounterVec
	cloudOperationDuration   *prometheus.HistogramVec
	cloudResourcesDiscovered *prometheus.GaugeVec

	// AI operations
	aiRequestsTotal   *prometheus.CounterVec
	aiRequestDuration *prometheus.HistogramVec
	aiTokensUsed      *prometheus.CounterVec

	// Cost optimization
	costSavingsTotal    *prometheus.CounterVec
	optimizationActions *prometheus.CounterVec

	// System metrics
	systemErrors  *prometheus.CounterVec
	systemUptime  prometheus.Gauge
	activeWorkers prometheus.Gauge
}

// NewMetrics creates a new metrics instance
func NewMetrics() *Metrics {
	return &Metrics{
		httpRequestsTotal: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "http_requests_total",
				Help: "Total number of HTTP requests",
			},
			[]string{"method", "endpoint", "status"},
		),
		httpRequestDuration: prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "http_request_duration_seconds",
				Help:    "HTTP request duration in seconds",
				Buckets: prometheus.DefBuckets,
			},
			[]string{"method", "endpoint"},
		),
		cloudOperationsTotal: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "cloud_operations_total",
				Help: "Total number of cloud operations",
			},
			[]string{"provider", "operation", "status"},
		),
		cloudOperationDuration: prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "cloud_operation_duration_seconds",
				Help:    "Cloud operation duration in seconds",
				Buckets: prometheus.DefBuckets,
			},
			[]string{"provider", "operation"},
		),
		cloudResourcesDiscovered: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "cloud_resources_discovered",
				Help: "Number of cloud resources discovered",
			},
			[]string{"provider", "type", "region"},
		),
		aiRequestsTotal: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "ai_requests_total",
				Help: "Total number of AI requests",
			},
			[]string{"service", "model", "status"},
		),
		aiRequestDuration: prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "ai_request_duration_seconds",
				Help:    "AI request duration in seconds",
				Buckets: prometheus.DefBuckets,
			},
			[]string{"service", "model"},
		),
		aiTokensUsed: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "ai_tokens_used_total",
				Help: "Total number of AI tokens used",
			},
			[]string{"service", "model"},
		),
		costSavingsTotal: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "cost_savings_total",
				Help: "Total cost savings in USD",
			},
			[]string{"provider", "optimization_type"},
		),
		optimizationActions: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "optimization_actions_total",
				Help: "Total optimization actions taken",
			},
			[]string{"provider", "action", "status"},
		),
		systemErrors: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "system_errors_total",
				Help: "Total number of system errors",
			},
			[]string{"component", "error_type"},
		),
		systemUptime: prometheus.NewGauge(
			prometheus.GaugeOpts{
				Name: "system_uptime_seconds",
				Help: "System uptime in seconds",
			},
		),
		activeWorkers: prometheus.NewGauge(
			prometheus.GaugeOpts{
				Name: "active_workers",
				Help: "Number of active workers",
			},
		),
	}
}

// Register registers all metrics with Prometheus
func (m *Metrics) Register() error {
	metrics := []prometheus.Collector{
		m.httpRequestsTotal,
		m.httpRequestDuration,
		m.cloudOperationsTotal,
		m.cloudOperationDuration,
		m.cloudResourcesDiscovered,
		m.aiRequestsTotal,
		m.aiRequestDuration,
		m.aiTokensUsed,
		m.costSavingsTotal,
		m.optimizationActions,
		m.systemErrors,
		m.systemUptime,
		m.activeWorkers,
	}

	for _, metric := range metrics {
		if err := prometheus.Register(metric); err != nil {
			return fmt.Errorf("failed to register metric: %w", err)
		}
	}

	return nil
}

// MonitoringService provides monitoring capabilities
type MonitoringService struct {
	metrics   *Metrics
	logger    *zap.Logger
	startTime time.Time
}

// NewMonitoringService creates a new monitoring service
func NewMonitoringService() (*MonitoringService, error) {
	// Initialize structured logger
	logger, err := zap.NewProduction()
	if err != nil {
		return nil, fmt.Errorf("failed to create logger: %w", err)
	}

	metrics := NewMetrics()
	if err := metrics.Register(); err != nil {
		return nil, fmt.Errorf("failed to register metrics: %w", err)
	}

	service := &MonitoringService{
		metrics:   metrics,
		logger:    logger,
		startTime: time.Now(),
	}

	// Start uptime monitoring
	go service.updateUptime()

	return service, nil
}

// updateUptime updates the system uptime metric
func (ms *MonitoringService) updateUptime() {
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	for range ticker.C {
		ms.metrics.systemUptime.Set(time.Since(ms.startTime).Seconds())
	}
}

// RecordHTTPRequest records an HTTP request
func (ms *MonitoringService) RecordHTTPRequest(method, endpoint, status string, duration time.Duration) {
	ms.metrics.httpRequestsTotal.WithLabelValues(method, endpoint, status).Inc()
	ms.metrics.httpRequestDuration.WithLabelValues(method, endpoint).Observe(duration.Seconds())
}

// RecordCloudOperation records a cloud operation
func (ms *MonitoringService) RecordCloudOperation(provider, operation, status string, duration time.Duration) {
	ms.metrics.cloudOperationsTotal.WithLabelValues(provider, operation, status).Inc()
	ms.metrics.cloudOperationDuration.WithLabelValues(provider, operation).Observe(duration.Seconds())
}

// RecordCloudResources records discovered cloud resources
func (ms *MonitoringService) RecordCloudResources(provider, resourceType, region string, count float64) {
	ms.metrics.cloudResourcesDiscovered.WithLabelValues(provider, resourceType, region).Set(count)
}

// RecordAIRequest records an AI request
func (ms *MonitoringService) RecordAIRequest(service, model, status string, duration time.Duration, tokens int) {
	ms.metrics.aiRequestsTotal.WithLabelValues(service, model, status).Inc()
	ms.metrics.aiRequestDuration.WithLabelValues(service, model).Observe(duration.Seconds())
	ms.metrics.aiTokensUsed.WithLabelValues(service, model).Add(float64(tokens))
}

// RecordCostSavings records cost savings
func (ms *MonitoringService) RecordCostSavings(provider, optimizationType string, savings float64) {
	ms.metrics.costSavingsTotal.WithLabelValues(provider, optimizationType).Add(savings)
}

// RecordOptimizationAction records an optimization action
func (ms *MonitoringService) RecordOptimizationAction(provider, action, status string) {
	ms.metrics.optimizationActions.WithLabelValues(provider, action, status).Inc()
}

// RecordError records a system error
func (ms *MonitoringService) RecordError(component, errorType string) {
	ms.metrics.systemErrors.WithLabelValues(component, errorType).Inc()
	ms.logger.Error("System error", zap.String("component", component), zap.String("error_type", errorType))
}

// SetActiveWorkers sets the number of active workers
func (ms *MonitoringService) SetActiveWorkers(count int) {
	ms.metrics.activeWorkers.Set(float64(count))
}

// GetMetricsHandler returns the Prometheus metrics handler
func (ms *MonitoringService) GetMetricsHandler() http.Handler {
	return promhttp.Handler()
}

// GetMetrics returns current metrics as a map
func (ms *MonitoringService) GetMetrics() map[string]interface{} {
	return map[string]interface{}{
		"uptime_seconds": time.Since(ms.startTime).Seconds(),
		"system_errors":  ms.metrics.systemErrors,
		"cost_savings":   ms.metrics.costSavingsTotal,
	}
}

// HealthCheck provides health check functionality
type HealthCheck struct {
	checks map[string]HealthCheckFunc
	logger *zap.Logger
}

// HealthCheckFunc is a function that performs a health check
type HealthCheckFunc func(ctx context.Context) error

// NewHealthCheck creates a new health check instance
func NewHealthCheck(logger *zap.Logger) *HealthCheck {
	return &HealthCheck{
		checks: make(map[string]HealthCheckFunc),
		logger: logger,
	}
}

// AddCheck adds a health check
func (hc *HealthCheck) AddCheck(name string, check HealthCheckFunc) {
	hc.checks[name] = check
}

// CheckHealth performs all health checks
func (hc *HealthCheck) CheckHealth(ctx context.Context) map[string]string {
	results := make(map[string]string)

	for name, check := range hc.checks {
		if err := check(ctx); err != nil {
			results[name] = fmt.Sprintf("unhealthy: %v", err)
		} else {
			results[name] = "healthy"
		}
	}

	return results
}

// RunChecks performs all health checks (alias for CheckHealth)
func (hc *HealthCheck) RunChecks(ctx context.Context) map[string]string {
	return hc.CheckHealth(ctx)
}

// GetHealthHandler returns an HTTP handler for health checks
func (hc *HealthCheck) GetHealthHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
		defer cancel()

		results := hc.CheckHealth(ctx)

		// Check if all checks are healthy
		allHealthy := true
		for _, status := range results {
			if status != "healthy" {
				allHealthy = false
				break
			}
		}

		w.Header().Set("Content-Type", "application/json")
		if allHealthy {
			w.WriteHeader(http.StatusOK)
		} else {
			w.WriteHeader(http.StatusServiceUnavailable)
		}

		// Write JSON response
		w.Write([]byte(`{"status":"` + map[bool]string{true: "healthy", false: "unhealthy"}[allHealthy] + `","checks":`))
		// Simple JSON encoding for health checks
		w.Write([]byte("{"))
		first := true
		for name, status := range results {
			if !first {
				w.Write([]byte(","))
			}
			w.Write([]byte(`"` + name + `":"` + status + `"`))
			first = false
		}
		w.Write([]byte("}}"))
	}
}
