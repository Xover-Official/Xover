package metrics

import (
	"fmt"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	// AI Metrics
	AIRequestsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "talos_ai_requests_total",
			Help: "Total number of AI requests by model and status",
		},
		[]string{"model", "tier", "status"},
	)

	AIRequestDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "talos_ai_request_duration_seconds",
			Help:    "AI request duration in seconds",
			Buckets: prometheus.ExponentialBuckets(0.1, 2, 10),
		},
		[]string{"model", "tier"},
	)

	AITokensConsumed = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "talos_ai_tokens_consumed_total",
			Help: "Total AI tokens consumed by model",
		},
		[]string{"model"},
	)

	AICostUSD = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "talos_ai_cost_usd_total",
			Help: "Total AI cost in USD",
		},
		[]string{"model"},
	)

	// Cloud Resource Metrics
	ResourcesDiscovered = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "talos_resources_discovered_total",
			Help: "Total cloud resources discovered",
		},
		[]string{"provider", "region", "type"},
	)

	ResourcesOptimized = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "talos_resources_optimized_total",
			Help: "Total resources optimized",
		},
		[]string{"provider", "type", "action"},
	)

	ResourceCostPerHour = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "talos_resource_cost_per_hour_usd",
			Help: "Resource cost per hour in USD",
		},
		[]string{"provider", "region", "type", "resource_id"},
	)

	// Optimization Metrics
	OptimizationSavingsUSD = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "talos_optimization_savings_usd_total",
			Help: "Total optimization savings in USD",
		},
		[]string{"provider", "type"},
	)

	OptimizationROI = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "talos_optimization_roi_ratio",
			Help: "Optimization ROI ratio (savings/cost)",
		},
	)

	// OODA Loop Metrics
	OODALoopDuration = promauto.NewHistogram(
		prometheus.HistogramOpts{
			Name:    "talos_ooda_loop_duration_seconds",
			Help:    "OODA loop execution duration",
			Buckets: prometheus.LinearBuckets(60, 30, 10),
		},
	)

	OODALoopPhaseErrors = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "talos_ooda_phase_errors_total",
			Help: "OODA loop phase errors",
		},
		[]string{"phase"}, // observe, orient, decide, act
	)

	// System Health Metrics
	HealthCheckStatus = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "talos_health_check_status",
			Help: "Health check status (1=healthy, 0.5=degraded, 0=unhealthy)",
		},
		[]string{"component"},
	)

	HealthCheckDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "talos_health_check_duration_seconds",
			Help:    "Health check duration",
			Buckets: prometheus.ExponentialBuckets(0.001, 2, 10),
		},
		[]string{"component"},
	)

	// Database Metrics
	DatabaseQueryDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "talos_database_query_duration_seconds",
			Help:    "Database query duration",
			Buckets: prometheus.ExponentialBuckets(0.001, 2, 10),
		},
		[]string{"operation"},
	)

	DatabaseConnectionsActive = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "talos_database_connections_active",
			Help: "Active database connections",
		},
	)

	// Cache Metrics
	CacheHitRatio = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "talos_cache_hit_ratio",
			Help: "Cache hit ratio (0-1)",
		},
	)

	CacheOperations = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "talos_cache_operations_total",
			Help: "Cache operations by type and result",
		},
		[]string{"operation", "result"}, // get/set, hit/miss/error
	)

	// API Metrics
	HTTPRequestsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "talos_http_requests_total",
			Help: "Total HTTP requests",
		},
		[]string{"method", "path", "status"},
	)

	HTTPRequestDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "talos_http_request_duration_seconds",
			Help:    "HTTP request duration",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"method", "path"},
	)
)

// RecordAIRequest records an AI request metric
func RecordAIRequest(model string, tier int, status string, duration float64, tokens int, cost float64) {
	tierStr := fmt.Sprintf("%d", tier)
	AIRequestsTotal.WithLabelValues(model, tierStr, status).Inc()
	AIRequestDuration.WithLabelValues(model, tierStr).Observe(duration)
	AITokensConsumed.WithLabelValues(model).Add(float64(tokens))
	AICostUSD.WithLabelValues(model).Add(cost)
}

// RecordOptimization records an optimization metric
func RecordOptimization(provider, resourceType, action string, savingsUSD float64) {
	ResourcesOptimized.WithLabelValues(provider, resourceType, action).Inc()
	OptimizationSavingsUSD.WithLabelValues(provider, resourceType).Add(savingsUSD)
}

// UpdateHealthStatus updates health check status
func UpdateHealthStatus(component string, healthy bool) {
	status := 0.0
	if healthy {
		status = 1.0
	}
	HealthCheckStatus.WithLabelValues(component).Set(status)
}

func RecordCacheOperation(operation, result string) {
	CacheOperations.WithLabelValues(operation, result).Inc()
}

func UpdateCacheHitRatio(ratio float64) {
	CacheHitRatio.Set(ratio)
}
