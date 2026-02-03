package monitoring

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/Xover-Official/Xover/internal/errors"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

// AlertSeverity represents the severity level of an alert
type AlertSeverity string

const (
	SeverityInfo     AlertSeverity = "info"
	SeverityWarning  AlertSeverity = "warning"
	SeverityError    AlertSeverity = "error"
	SeverityCritical AlertSeverity = "critical"
)

// AlertStatus represents the current status of an alert
type AlertStatus string

const (
	StatusActive   AlertStatus = "active"
	StatusResolved AlertStatus = "resolved"
	StatusSilenced AlertStatus = "silenced"
)

// AlertType represents different types of alerts
type AlertType string

const (
	AlertTypePerformance  AlertType = "performance"
	AlertTypeAvailability AlertType = "availability"
	AlertTypeSecurity     AlertType = "security"
	AlertTypeCost         AlertType = "cost"
	AlertTypeCapacity     AlertType = "capacity"
	AlertTypeOptimization AlertType = "optimization"
	AlertTypeSystem       AlertType = "system"
)

// Alert represents a monitoring alert
type Alert struct {
	ID            string                 `json:"id"`
	Type          AlertType              `json:"type"`
	Severity      AlertSeverity          `json:"severity"`
	Status        AlertStatus            `json:"status"`
	Title         string                 `json:"title"`
	Description   string                 `json:"description"`
	EntityID      string                 `json:"entity_id"`
	EntityType    string                 `json:"entity_type"`
	Timestamp     time.Time              `json:"timestamp"`
	Labels        map[string]string      `json:"labels"`
	Annotations   map[string]interface{} `json:"annotations"`
	Threshold     *Threshold             `json:"threshold,omitempty"`
	Current       float64                `json:"current_value"`
	ResolvedAt    *time.Time             `json:"resolved_at,omitempty"`
	SilencedUntil *time.Time             `json:"silenced_until,omitempty"`
}

// Threshold defines alerting thresholds
type Threshold struct {
	Metric   string  `json:"metric"`
	Operator string  `json:"operator"` // >, <, >=, <=, ==, !=
	Value    float64 `json:"value"`
	Duration string  `json:"duration"` // e.g., "5m", "1h"
}

// AlertRule defines when to trigger alerts
type AlertRule struct {
	ID        string            `json:"id"`
	Name      string            `json:"name"`
	Type      AlertType         `json:"type"`
	Severity  AlertSeverity     `json:"severity"`
	Threshold Threshold         `json:"threshold"`
	Query     string            `json:"query"`
	Labels    map[string]string `json:"labels"`
	Enabled   bool              `json:"enabled"`
	Interval  time.Duration     `json:"interval"`
	LastEval  time.Time         `json:"last_eval"`
}

// NotificationChannel defines how alerts are sent
type NotificationChannel struct {
	ID       string                 `json:"id"`
	Name     string                 `json:"name"`
	Type     string                 `json:"type"` // email, slack, webhook, pagerduty
	Config   map[string]interface{} `json:"config"`
	Enabled  bool                   `json:"enabled"`
	LastSent time.Time              `json:"last_sent"`
}

// AlertManager manages alerts and notifications
type AlertManager struct {
	alerts   map[string]*Alert
	rules    map[string]*AlertRule
	channels map[string]*NotificationChannel
	mu       sync.RWMutex
	logger   *log.Logger
	metrics  *AlertMetrics
	notifier *Notifier
}

// AlertMetrics tracks alert-related metrics
type AlertMetrics struct {
	AlertsTotal      prometheus.Counter
	AlertsActive     prometheus.Gauge
	AlertsResolved   prometheus.Counter
	AlertsByType     *prometheus.CounterVec
	AlertsBySeverity *prometheus.CounterVec
}

// NewAlertMetrics creates new alert metrics
func NewAlertMetrics() *AlertMetrics {
	return &AlertMetrics{
		AlertsTotal: promauto.NewCounter(prometheus.CounterOpts{
			Name: "talos_alerts_total",
			Help: "Total number of alerts generated",
		}),
		AlertsActive: promauto.NewGauge(prometheus.GaugeOpts{
			Name: "talos_alerts_active",
			Help: "Number of currently active alerts",
		}),
		AlertsResolved: promauto.NewCounter(prometheus.CounterOpts{
			Name: "talos_alerts_resolved_total",
			Help: "Total number of alerts resolved",
		}),
		AlertsByType: promauto.NewCounterVec(prometheus.CounterOpts{
			Name: "talos_alerts_by_type_total",
			Help: "Total number of alerts by type",
		}, []string{"type"}),
		AlertsBySeverity: promauto.NewCounterVec(prometheus.CounterOpts{
			Name: "talos_alerts_by_severity_total",
			Help: "Total number of alerts by severity",
		}, []string{"severity"}),
	}
}

// NewAlertManager creates a new alert manager
func NewAlertManager(logger *log.Logger) *AlertManager {
	if logger == nil {
		logger = log.Default()
	}

	return &AlertManager{
		alerts:   make(map[string]*Alert),
		rules:    make(map[string]*AlertRule),
		channels: make(map[string]*NotificationChannel),
		logger:   logger,
		metrics:  NewAlertMetrics(),
		notifier: NewNotifier(logger),
	}
}

// AddRule adds a new alert rule
func (am *AlertManager) AddRule(rule *AlertRule) {
	am.mu.Lock()
	defer am.mu.Unlock()

	am.rules[rule.ID] = rule
	am.logger.Printf("Added alert rule: %s", rule.Name)
}

// RemoveRule removes an alert rule
func (am *AlertManager) RemoveRule(ruleID string) {
	am.mu.Lock()
	defer am.mu.Unlock()

	delete(am.rules, ruleID)
	am.logger.Printf("Removed alert rule: %s", ruleID)
}

// AddChannel adds a new notification channel
func (am *AlertManager) AddChannel(channel *NotificationChannel) {
	am.mu.Lock()
	defer am.mu.Unlock()

	am.channels[channel.ID] = channel
	am.logger.Printf("Added notification channel: %s", channel.Name)
}

// EvaluateRules evaluates all alert rules
func (am *AlertManager) EvaluateRules(ctx context.Context) error {
	am.mu.RLock()
	rules := make([]*AlertRule, 0, len(am.rules))
	for _, rule := range am.rules {
		if rule.Enabled {
			rules = append(rules, rule)
		}
	}
	am.mu.RUnlock()

	for _, rule := range rules {
		if err := am.evaluateRule(ctx, rule); err != nil {
			am.logger.Printf("Error evaluating rule %s: %v", rule.Name, err)
		}
	}

	return nil
}

// evaluateRule evaluates a single alert rule
func (am *AlertManager) evaluateRule(ctx context.Context, rule *AlertRule) error {
	// Check if it's time to evaluate
	if time.Since(rule.LastEval) < rule.Interval {
		return nil
	}

	// Execute the query (this would integrate with your metrics system)
	currentValue, err := am.executeQuery(ctx, rule.Query)
	if err != nil {
		return fmt.Errorf("failed to execute query: %w", err)
	}

	// Check if threshold is breached
	breached := am.checkThreshold(currentValue, rule.Threshold)

	rule.LastEval = time.Now()

	alertID := fmt.Sprintf("%s-%s", rule.ID, rule.Type)

	am.mu.Lock()
	defer am.mu.Unlock()

	existingAlert, exists := am.alerts[alertID]

	if breached && (!exists || existingAlert.Status == StatusResolved) {
		// Create new alert
		alert := &Alert{
			ID:          alertID,
			Type:        rule.Type,
			Severity:    rule.Severity,
			Status:      StatusActive,
			Title:       fmt.Sprintf("%s alert", rule.Name),
			Description: fmt.Sprintf("%s: %.2f %s %.2f", rule.Name, currentValue, rule.Threshold.Operator, rule.Threshold.Value),
			Timestamp:   time.Now(),
			Labels:      rule.Labels,
			Threshold:   &rule.Threshold,
			Current:     currentValue,
		}

		am.alerts[alertID] = alert
		am.metrics.AlertsTotal.Inc()
		am.metrics.AlertsActive.Inc()
		am.metrics.AlertsByType.WithLabelValues(string(rule.Type)).Inc()
		am.metrics.AlertsBySeverity.WithLabelValues(string(rule.Severity)).Inc()

		// Send notifications
		go am.notifier.SendNotifications(ctx, alert, am.channels)

		am.logger.Printf("Alert triggered: %s", alert.Title)

	} else if !breached && exists && existingAlert.Status == StatusActive {
		// Resolve alert
		resolvedAt := time.Now()
		existingAlert.Status = StatusResolved
		existingAlert.ResolvedAt = &resolvedAt

		am.metrics.AlertsActive.Dec()
		am.metrics.AlertsResolved.Inc()

		// Send resolution notifications
		go am.notifier.SendResolutionNotifications(ctx, existingAlert, am.channels)

		am.logger.Printf("Alert resolved: %s", existingAlert.Title)
	}

	return nil
}

// checkThreshold checks if a value breaches the threshold
func (am *AlertManager) checkThreshold(value float64, threshold Threshold) bool {
	switch threshold.Operator {
	case ">":
		return value > threshold.Value
	case "<":
		return value < threshold.Value
	case ">=":
		return value >= threshold.Value
	case "<=":
		return value <= threshold.Value
	case "==":
		return value == threshold.Value
	case "!=":
		return value != threshold.Value
	default:
		return false
	}
}

// executeQuery executes a metrics query (placeholder implementation)
func (am *AlertManager) executeQuery(ctx context.Context, query string) (float64, error) {
	// This would integrate with Prometheus, InfluxDB, or your metrics system
	// For now, return a mock value
	return 75.0, nil
}

// GetAlerts returns all alerts
func (am *AlertManager) GetAlerts() []*Alert {
	am.mu.RLock()
	defer am.mu.RUnlock()

	alerts := make([]*Alert, 0, len(am.alerts))
	for _, alert := range am.alerts {
		alerts = append(alerts, alert)
	}

	return alerts
}

// GetActiveAlerts returns only active alerts
func (am *AlertManager) GetActiveAlerts() []*Alert {
	am.mu.RLock()
	defer am.mu.RUnlock()

	var activeAlerts []*Alert
	for _, alert := range am.alerts {
		if alert.Status == StatusActive {
			activeAlerts = append(activeAlerts, alert)
		}
	}

	return activeAlerts
}

// SilenceAlert silences an alert
func (am *AlertManager) SilenceAlert(alertID string, duration time.Duration) error {
	am.mu.Lock()
	defer am.mu.Unlock()

	alert, exists := am.alerts[alertID]
	if !exists {
		return errors.NewResourceNotFoundError("alert", alertID)
	}

	silencedUntil := time.Now().Add(duration)
	alert.SilencedUntil = &silencedUntil
	alert.Status = StatusSilenced

	am.logger.Printf("Alert silenced: %s until %v", alert.Title, silencedUntil)
	return nil
}

// Notifier handles sending alert notifications
type Notifier struct {
	logger *log.Logger
}

// NewNotifier creates a new notifier
func NewNotifier(logger *log.Logger) *Notifier {
	if logger == nil {
		logger = log.Default()
	}
	return &Notifier{logger: logger}
}

// SendNotifications sends alert notifications through all channels
func (n *Notifier) SendNotifications(ctx context.Context, alert *Alert, channels map[string]*NotificationChannel) {
	for _, channel := range channels {
		if !channel.Enabled {
			continue
		}

		// Rate limiting: don't spam notifications
		if time.Since(channel.LastSent) < 5*time.Minute {
			continue
		}

		if err := n.sendNotification(ctx, alert, channel); err != nil {
			n.logger.Printf("Failed to send notification via %s: %v", channel.Name, err)
		} else {
			channel.LastSent = time.Now()
		}
	}
}

// SendResolutionNotifications sends resolution notifications
func (n *Notifier) SendResolutionNotifications(ctx context.Context, alert *Alert, channels map[string]*NotificationChannel) {
	for _, channel := range channels {
		if !channel.Enabled {
			continue
		}

		if err := n.sendResolutionNotification(ctx, alert, channel); err != nil {
			n.logger.Printf("Failed to send resolution notification via %s: %v", channel.Name, err)
		}
	}
}

// sendNotification sends a single notification
func (n *Notifier) sendNotification(ctx context.Context, alert *Alert, channel *NotificationChannel) error {
	switch channel.Type {
	case "email":
		return n.sendEmailNotification(ctx, alert, channel)
	case "slack":
		return n.sendSlackNotification(ctx, alert, channel)
	case "webhook":
		return n.sendWebhookNotification(ctx, alert, channel)
	case "pagerduty":
		return n.sendPagerDutyNotification(ctx, alert, channel)
	default:
		return fmt.Errorf("unsupported notification channel type: %s", channel.Type)
	}
}

// sendResolutionNotification sends a resolution notification
func (n *Notifier) sendResolutionNotification(ctx context.Context, alert *Alert, channel *NotificationChannel) error {
	// Similar to sendNotification but with resolution message
	return n.sendNotification(ctx, alert, channel)
}

// Placeholder implementations for notification methods
func (n *Notifier) sendEmailNotification(ctx context.Context, alert *Alert, channel *NotificationChannel) error {
	n.logger.Printf("Email notification sent for alert: %s", alert.Title)
	return nil
}

func (n *Notifier) sendSlackNotification(ctx context.Context, alert *Alert, channel *NotificationChannel) error {
	n.logger.Printf("Slack notification sent for alert: %s", alert.Title)
	return nil
}

func (n *Notifier) sendWebhookNotification(ctx context.Context, alert *Alert, channel *NotificationChannel) error {
	n.logger.Printf("Webhook notification sent for alert: %s", alert.Title)
	return nil
}

func (n *Notifier) sendPagerDutyNotification(ctx context.Context, alert *Alert, channel *NotificationChannel) error {
	n.logger.Printf("PagerDuty notification sent for alert: %s", alert.Title)
	return nil
}

// DefaultAlertRules returns a set of default alert rules
func DefaultAlertRules() []*AlertRule {
	return []*AlertRule{
		{
			ID:       "high-cpu-usage",
			Name:     "High CPU Usage",
			Type:     AlertTypePerformance,
			Severity: SeverityWarning,
			Threshold: Threshold{
				Metric:   "cpu_usage",
				Operator: ">",
				Value:    80.0,
				Duration: "5m",
			},
			Query:    "avg(cpu_usage)",
			Labels:   map[string]string{"team": "infrastructure"},
			Enabled:  true,
			Interval: 1 * time.Minute,
		},
		{
			ID:       "high-memory-usage",
			Name:     "High Memory Usage",
			Type:     AlertTypePerformance,
			Severity: SeverityWarning,
			Threshold: Threshold{
				Metric:   "memory_usage",
				Operator: ">",
				Value:    85.0,
				Duration: "5m",
			},
			Query:    "avg(memory_usage)",
			Labels:   map[string]string{"team": "infrastructure"},
			Enabled:  true,
			Interval: 1 * time.Minute,
		},
		{
			ID:       "service-unavailable",
			Name:     "Service Unavailable",
			Type:     AlertTypeAvailability,
			Severity: SeverityCritical,
			Threshold: Threshold{
				Metric:   "up",
				Operator: "==",
				Value:    0.0,
				Duration: "1m",
			},
			Query:    "up",
			Labels:   map[string]string{"team": "sre"},
			Enabled:  true,
			Interval: 30 * time.Second,
		},
		{
			ID:       "high-cost-anomaly",
			Name:     "High Cost Anomaly",
			Type:     AlertTypeCost,
			Severity: SeverityError,
			Threshold: Threshold{
				Metric:   "daily_cost",
				Operator: ">",
				Value:    1000.0,
				Duration: "1h",
			},
			Query:    "sum(daily_cost)",
			Labels:   map[string]string{"team": "finance"},
			Enabled:  true,
			Interval: 5 * time.Minute,
		},
		{
			ID:       "optimization-failed",
			Name:     "Optimization Failed",
			Type:     AlertTypeOptimization,
			Severity: SeverityError,
			Threshold: Threshold{
				Metric:   "optimization_failures",
				Operator: ">",
				Value:    5.0,
				Duration: "10m",
			},
			Query:    "rate(optimization_failures[10m])",
			Labels:   map[string]string{"team": "automation"},
			Enabled:  true,
			Interval: 2 * time.Minute,
		},
	}
}

// DefaultNotificationChannels returns default notification channels
func DefaultNotificationChannels() []*NotificationChannel {
	return []*NotificationChannel{
		{
			ID:   "email-admin",
			Name: "Email Admin",
			Type: "email",
			Config: map[string]interface{}{
				"to":      "admin@talos.io",
				"subject": "Talos Alert",
			},
			Enabled: true,
		},
		{
			ID:   "slack-alerts",
			Name: "Slack Alerts",
			Type: "slack",
			Config: map[string]interface{}{
				"webhook_url": "https://hooks.slack.com/services/...",
				"channel":     "#alerts",
			},
			Enabled: true,
		},
		{
			ID:   "pagerduty-critical",
			Name: "PagerDuty Critical",
			Type: "pagerduty",
			Config: map[string]interface{}{
				"service_key": "your-pagerduty-service-key",
			},
			Enabled: true,
		},
	}
}
