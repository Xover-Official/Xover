package performance

import (
	"context"
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"go.uber.org/zap"
)

// MonitoringService handles metrics and observability
type MonitoringService struct {
	activeWorkers       prometheus.Gauge
	httpRequests        *prometheus.CounterVec
	cloudOperations     *prometheus.CounterVec
	aiRequests          *prometheus.CounterVec
	costSavings         *prometheus.CounterVec
	optimizationActions *prometheus.CounterVec
}

// NewMonitoringService creates a new monitoring service
func NewMonitoringService() (*MonitoringService, error) {
	s := &MonitoringService{
		activeWorkers: prometheus.NewGauge(prometheus.GaugeOpts{
			Name: "talos_active_workers",
			Help: "Current number of active workers",
		}),
		httpRequests: prometheus.NewCounterVec(prometheus.CounterOpts{
			Name: "talos_http_requests_total",
			Help: "Total number of HTTP requests",
		}, []string{"method", "path", "status"}),
		cloudOperations: prometheus.NewCounterVec(prometheus.CounterOpts{
			Name: "talos_cloud_operations_total",
			Help: "Total number of cloud operations",
		}, []string{"provider", "operation", "status"}),
		aiRequests: prometheus.NewCounterVec(prometheus.CounterOpts{
			Name: "talos_ai_requests_total",
			Help: "Total number of AI requests",
		}, []string{"provider", "model", "status"}),
		costSavings: prometheus.NewCounterVec(prometheus.CounterOpts{
			Name: "talos_cost_savings_total",
			Help: "Total cost savings estimated",
		}, []string{"provider", "category"}),
		optimizationActions: prometheus.NewCounterVec(prometheus.CounterOpts{
			Name: "talos_optimization_actions_total",
			Help: "Total optimization actions taken",
		}, []string{"provider", "action", "status"}),
	}

	// Register metrics
	prometheus.MustRegister(s.activeWorkers)
	prometheus.MustRegister(s.httpRequests)
	prometheus.MustRegister(s.cloudOperations)
	prometheus.MustRegister(s.aiRequests)
	prometheus.MustRegister(s.costSavings)
	prometheus.MustRegister(s.optimizationActions)

	return s, nil
}

// GetMetrics returns the underlying metrics registry
func (s *MonitoringService) GetMetrics() interface{} {
	return prometheus.DefaultRegisterer
}

// SetActiveWorkers updates the active worker count metric
func (s *MonitoringService) SetActiveWorkers(count int) {
	s.activeWorkers.Set(float64(count))
}

// RecordHTTPRequest records HTTP request metrics
func (s *MonitoringService) RecordHTTPRequest(method, path, status string, duration time.Duration) {
	s.httpRequests.WithLabelValues(method, path, status).Inc()
}

// RecordCloudOperation records cloud API operation metrics
func (s *MonitoringService) RecordCloudOperation(provider, operation, status string, duration time.Duration) {
	s.cloudOperations.WithLabelValues(provider, operation, status).Inc()
}

// RecordAIRequest records AI service request metrics
func (s *MonitoringService) RecordAIRequest(provider, model, status string, duration time.Duration, tokens int) {
	s.aiRequests.WithLabelValues(provider, model, status).Inc()
}

// RecordCostSavings records estimated cost savings
func (s *MonitoringService) RecordCostSavings(provider, category string, amount float64) {
	s.costSavings.WithLabelValues(provider, category).Add(amount)
}

// RecordOptimizationAction records optimization actions taken
func (s *MonitoringService) RecordOptimizationAction(provider, action, status string) {
	s.optimizationActions.WithLabelValues(provider, action, status).Inc()
}

// HealthCheck handles system health checks
type HealthCheck struct {
	logger *zap.Logger
	checks map[string]func(context.Context) error
	mu     sync.RWMutex
}

func NewHealthCheck(logger *zap.Logger) *HealthCheck {
	return &HealthCheck{
		logger: logger,
		checks: make(map[string]func(context.Context) error),
	}
}

func (hc *HealthCheck) AddCheck(name string, check func(context.Context) error) {
	hc.mu.Lock()
	defer hc.mu.Unlock()
	hc.checks[name] = check
}

// RunChecks executes all registered health checks
func (hc *HealthCheck) RunChecks(ctx context.Context) map[string]error {
	hc.mu.RLock()
	defer hc.mu.RUnlock()

	results := make(map[string]error)
	for name, check := range hc.checks {
		results[name] = check(ctx)
	}
	return results
}
