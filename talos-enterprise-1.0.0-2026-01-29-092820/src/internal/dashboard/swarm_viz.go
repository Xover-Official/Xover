package dashboard

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"
	
	"github.com/redis/go-redis/v9"
)

// SwarmVisualization represents real-time AI swarm state
type SwarmVisualization struct {
	Timestamp     time.Time    `json:"timestamp"`
	ActiveTier    int          `json:"active_tier"`
	TierStatus    []TierStatus `json:"tier_status"`
	CurrentAction string       `json:"current_action"`
	QueueDepth    int          `json:"queue_depth"`
}

// TierStatus represents the status of one AI tier
type TierStatus struct {
	Tier          int     `json:"tier"`
	Name          string  `json:"name"`
	Model         string  `json:"model"`
	Active        bool    `json:"active"`
	RequestsToday int     `json:"requests_today"`
	AvgLatency    float64 `json:"avg_latency_ms"`
	SuccessRate   float64 `json:"success_rate"`
	Status        string  `json:"status"` // "healthy", "degraded", "offline"
}

// QueueProvider defines the interface for fetching distributed queue metrics
type QueueProvider interface {
	GetQueueDepth() int
}

// MetricsProvider defines interface for fetching swarm metrics
type MetricsProvider interface {
	GetTierMetrics(model string) TierMetrics
}

type TierMetrics struct {
	LastUsed    time.Time
	TotalCalls  int
	AvgLatency  float64
	SuccessRate float64
}

// RedisMetricsProvider implements MetricsProvider
type RedisMetricsProvider struct {
	client *redis.Client
}

func NewRedisMetricsProvider(client *redis.Client) *RedisMetricsProvider {
	return &RedisMetricsProvider{client: client}
}

func (r *RedisMetricsProvider) GetTierMetrics(model string) TierMetrics {
	key := fmt.Sprintf("talos:metrics:%s", model)
	val, err := r.client.HGetAll(context.Background(), key).Result()
	if err != nil || len(val) == 0 {
		return TierMetrics{}
	}

	lastUsed, _ := time.Parse(time.RFC3339, val["last_used"])
	totalCalls, _ := strconv.Atoi(val["total_calls"])
	avgLatency, _ := strconv.ParseFloat(val["avg_latency"], 64)
	successRate, _ := strconv.ParseFloat(val["success_rate"], 64)

	return TierMetrics{
		LastUsed:    lastUsed,
		TotalCalls:  totalCalls,
		AvgLatency:  avgLatency,
		SuccessRate: successRate,
	}
}

// LiveSwarmHandler serves real-time swarm visualization data
type LiveSwarmHandler struct {
	metrics       MetricsProvider
	queueProvider QueueProvider
}

func NewLiveSwarmHandler(metrics MetricsProvider, qp QueueProvider) *LiveSwarmHandler {
	return &LiveSwarmHandler{
		metrics:       metrics,
		queueProvider: qp,
	}
}

func (h *LiveSwarmHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Map real metrics to dashboard tiers
	tiers := []TierStatus{
		h.buildTierStatus(1, "Sentinel", "gemini-2.0-flash-exp"),
		h.buildTierStatus(2, "Strategist", "gemini-pro"),
		h.buildTierStatus(3, "Arbiter", "claude-3.5-sonnet"),
		h.buildTierStatus(4, "Reasoning", "gpt-4o-mini"),
		h.buildTierStatus(5, "Oracle", "devin"),
	}

	// Determine active tier (highest tier with recent activity)
	activeTier := 1
	for _, t := range tiers {
		if t.Active && t.Tier > activeTier {
			activeTier = t.Tier
		}
	}

	// Determine current action based on most recently active model
	currentAction := "Monitoring cloud estate..."
	var lastActiveTime time.Time
	var lastActiveModel string

	for _, t := range tiers {
		// Check if active based on the TierStatus we just built
		// We need to parse the LastUsed from the metrics provider again or rely on the Active flag
		// For simplicity, we trust the Active flag which is based on 5 min window
		if t.Active {
			// We don't have the exact timestamp here easily without refetching, 
			// but we can assume the highest tier is the "current" action
			lastActiveModel = t.Name
		}
	}

	if lastActiveModel != "" {
		currentAction = fmt.Sprintf("%s analyzing opportunities...", lastActiveModel)
	}

	// Fetch real queue depth if provider is available
	queueDepth := 0
	if h.queueProvider != nil {
		queueDepth = h.queueProvider.GetQueueDepth()
	}

	viz := SwarmVisualization{
		Timestamp:     time.Now(),
		ActiveTier:    activeTier,
		CurrentAction: currentAction,
		QueueDepth:    queueDepth,
		TierStatus:    tiers,
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	json.NewEncoder(w).Encode(viz)
}

func (h *LiveSwarmHandler) buildTierStatus(tier int, name, model string) TierStatus {
	m := h.metrics.GetTierMetrics(model)

	// Consider active if used in the last 5 minutes
	isActive := !m.LastUsed.IsZero() && time.Since(m.LastUsed) < 5*time.Minute

	return TierStatus{
		Tier:          tier,
		Name:          name,
		Model:         model,
		Active:        isActive,
		RequestsToday: m.TotalCalls,
		AvgLatency:    m.AvgLatency,
		SuccessRate:   m.SuccessRate,
		Status:        "healthy",
	}
}
