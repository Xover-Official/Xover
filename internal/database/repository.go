package database

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
)

// Repository provides database operations for entities
type Repository struct {
	db     *DatabaseManager
	logger *zap.Logger
	tracer trace.Tracer
}

// NewRepository creates a new repository
func NewRepository(db *DatabaseManager, logger *zap.Logger, tracer trace.Tracer) *Repository {
	return &Repository{
		db:     db,
		logger: logger,
		tracer: tracer,
	}
}

// Action represents an action in the system
type Action struct {
	ID               string     `json:"id" db:"id"`
	ResourceID       string     `json:"resource_id" db:"resource_id"`
	ActionType       string     `json:"action_type" db:"action_type"`
	Status           string     `json:"status" db:"status"`
	Checksum         string     `json:"checksum" db:"checksum"`
	Payload          string     `json:"payload" db:"payload"`
	RiskScore        float64    `json:"risk_score" db:"risk_score"`
	EstimatedSavings float64    `json:"estimated_savings" db:"estimated_savings"`
	CreatedAt        time.Time  `json:"created_at" db:"created_at"`
	StartedAt        *time.Time `json:"started_at" db:"started_at"`
	CompletedAt      *time.Time `json:"completed_at" db:"completed_at"`
	ErrorMessage     *string    `json:"error_message" db:"error_message"`
}

// AIDecision represents an AI decision
type AIDecision struct {
	ID         string    `json:"id" db:"id"`
	ResourceID string    `json:"resource_id" db:"resource_id"`
	Model      string    `json:"model" db:"model"`
	Decision   string    `json:"decision" db:"decision"`
	Reasoning  *string   `json:"reasoning" db:"reasoning"`
	Confidence *float64  `json:"confidence" db:"confidence"`
	TokensUsed *int      `json:"tokens_used" db:"tokens_used"`
	LatencyMs  *int      `json:"latency_ms" db:"latency_ms"`
	CreatedAt  time.Time `json:"created_at" db:"created_at"`
}

// TokenUsage represents token usage tracking
type TokenUsage struct {
	ID          string    `json:"id" db:"id"`
	Model       string    `json:"model" db:"model"`
	Tokens      int       `json:"tokens" db:"tokens"`
	CostUSD     float64   `json:"cost_usd" db:"cost_usd"`
	RequestType *string   `json:"request_type" db:"request_type"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
}

// SavingsEvent represents a savings event
type SavingsEvent struct {
	ID               string    `json:"id" db:"id"`
	ActionID         *string   `json:"action_id" db:"action_id"`
	ResourceID       string    `json:"resource_id" db:"resource_id"`
	OptimizationType *string   `json:"optimization_type" db:"optimization_type"`
	EstimatedSavings *float64  `json:"estimated_savings" db:"estimated_savings"`
	ActualSavings    *float64  `json:"actual_savings" db:"actual_savings"`
	CreatedAt        time.Time `json:"created_at" db:"created_at"`
}

// Organization represents an organization
type Organization struct {
	ID        string                 `json:"id" db:"id"`
	Name      string                 `json:"name" db:"name"`
	CreatedAt time.Time              `json:"created_at" db:"created_at"`
	Settings  map[string]interface{} `json:"settings" db:"settings"`
}

// User represents a user
type User struct {
	ID             string     `json:"id" db:"id"`
	Email          string     `json:"email" db:"email"`
	OrganizationID *string    `json:"organization_id" db:"organization_id"`
	Role           string     `json:"role" db:"role"`
	CreatedAt      time.Time  `json:"created_at" db:"created_at"`
	LastLogin      *time.Time `json:"last_login" db:"last_login"`
}

// Resource represents a cloud resource
type Resource struct {
	ID              string                 `json:"id" db:"id"`
	OrganizationID  *string                `json:"organization_id" db:"organization_id"`
	CloudResourceID string                 `json:"cloud_resource_id" db:"cloud_resource_id"`
	CloudProvider   string                 `json:"cloud_provider" db:"cloud_provider"`
	ResourceType    *string                `json:"resource_type" db:"resource_type"`
	Region          *string                `json:"region" db:"region"`
	Tags            map[string]interface{} `json:"tags" db:"tags"`
	Metadata        map[string]interface{} `json:"metadata" db:"metadata"`
	MonthlyCost     *float64               `json:"monthly_cost" db:"monthly_cost"`
	LastScanned     *time.Time             `json:"last_scanned" db:"last_scanned"`
	CreatedAt       time.Time              `json:"created_at" db:"created_at"`
}

// AuditLog represents an audit log entry
type AuditLog struct {
	ID           string                 `json:"id" db:"id"`
	UserID       *string                `json:"user_id" db:"user_id"`
	Action       string                 `json:"action" db:"action"`
	ResourceType *string                `json:"resource_type" db:"resource_type"`
	ResourceID   *string                `json:"resource_id" db:"resource_id"`
	Details      map[string]interface{} `json:"details" db:"details"`
	IPAddress    *string                `json:"ip_address" db:"ip_address"`
	CreatedAt    time.Time              `json:"created_at" db:"created_at"`
}

// CreateAction creates a new action
func (r *Repository) CreateAction(ctx context.Context, action *Action) error {
	ctx, span := r.tracer.Start(ctx, "repository.create_action")
	defer span.End()

	query := `
		INSERT INTO actions (id, resource_id, action_type, status, checksum, payload, risk_score, estimated_savings)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`

	_, err := r.db.Exec(ctx, query,
		action.ID, action.ResourceID, action.ActionType, action.Status,
		action.Checksum, action.Payload, action.RiskScore, action.EstimatedSavings,
	)
	if err != nil {
		span.RecordError(err)
		return fmt.Errorf("failed to create action: %w", err)
	}

	return nil
}

// GetActionByID retrieves an action by ID
func (r *Repository) GetActionByID(ctx context.Context, id string) (*Action, error) {
	ctx, span := r.tracer.Start(ctx, "repository.get_action_by_id")
	defer span.End()

	query := `
		SELECT id, resource_id, action_type, status, checksum, payload, risk_score, estimated_savings,
			   created_at, started_at, completed_at, error_message
		FROM actions WHERE id = $1
	`

	var action Action
	err := r.db.QueryRow(ctx, query, id).Scan(
		&action.ID, &action.ResourceID, &action.ActionType, &action.Status,
		&action.Checksum, &action.Payload, &action.RiskScore, &action.EstimatedSavings,
		&action.CreatedAt, &action.StartedAt, &action.CompletedAt, &action.ErrorMessage,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("action not found: %s", id)
		}
		span.RecordError(err)
		return nil, fmt.Errorf("failed to get action: %w", err)
	}

	return &action, nil
}

// UpdateActionStatus updates an action's status
func (r *Repository) UpdateActionStatus(ctx context.Context, id, status string, startedAt, completedAt *time.Time, errorMessage *string) error {
	ctx, span := r.tracer.Start(ctx, "repository.update_action_status")
	defer span.End()

	query := `
		UPDATE actions 
		SET status = $2, started_at = $3, completed_at = $4, error_message = $5
		WHERE id = $1
	`

	_, err := r.db.Exec(ctx, query, id, status, startedAt, completedAt, errorMessage)
	if err != nil {
		span.RecordError(err)
		return fmt.Errorf("failed to update action status: %w", err)
	}

	return nil
}

// GetPendingActions retrieves all pending actions
func (r *Repository) GetPendingActions(ctx context.Context) ([]*Action, error) {
	ctx, span := r.tracer.Start(ctx, "repository.get_pending_actions")
	defer span.End()

	query := `
		SELECT id, resource_id, action_type, status, checksum, payload, risk_score, estimated_savings,
			   created_at, started_at, completed_at, error_message
		FROM actions WHERE status = 'PENDING'
		ORDER BY created_at ASC
		LIMIT 100
	`

	rows, err := r.db.Query(ctx, query)
	if err != nil {
		span.RecordError(err)
		return nil, fmt.Errorf("failed to get pending actions: %w", err)
	}
	defer rows.Close()

	var actions []*Action
	for rows.Next() {
		var action Action
		err := rows.Scan(
			&action.ID, &action.ResourceID, &action.ActionType, &action.Status,
			&action.Checksum, &action.Payload, &action.RiskScore, &action.EstimatedSavings,
			&action.CreatedAt, &action.StartedAt, &action.CompletedAt, &action.ErrorMessage,
		)
		if err != nil {
			span.RecordError(err)
			return nil, fmt.Errorf("failed to scan action: %w", err)
		}
		actions = append(actions, &action)
	}

	return actions, nil
}

// CreateAIDecision creates a new AI decision
func (r *Repository) CreateAIDecision(ctx context.Context, decision *AIDecision) error {
	ctx, span := r.tracer.Start(ctx, "repository.create_ai_decision")
	defer span.End()

	query := `
		INSERT INTO ai_decisions (id, resource_id, model, decision, reasoning, confidence, tokens_used, latency_ms)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`

	_, err := r.db.Exec(ctx, query,
		decision.ID, decision.ResourceID, decision.Model, decision.Decision,
		decision.Reasoning, decision.Confidence, decision.TokensUsed, decision.LatencyMs,
	)
	if err != nil {
		span.RecordError(err)
		return fmt.Errorf("failed to create AI decision: %w", err)
	}

	return nil
}

// RecordTokenUsage records token usage
func (r *Repository) RecordTokenUsage(ctx context.Context, usage *TokenUsage) error {
	ctx, span := r.tracer.Start(ctx, "repository.record_token_usage")
	defer span.End()

	query := `
		INSERT INTO token_usage (id, model, tokens, cost_usd, request_type)
		VALUES ($1, $2, $3, $4, $5)
	`

	_, err := r.db.Exec(ctx, query,
		usage.ID, usage.Model, usage.Tokens, usage.CostUSD, usage.RequestType,
	)
	if err != nil {
		span.RecordError(err)
		return fmt.Errorf("failed to record token usage: %w", err)
	}

	return nil
}

// CreateSavingsEvent creates a new savings event
func (r *Repository) CreateSavingsEvent(ctx context.Context, event *SavingsEvent) error {
	ctx, span := r.tracer.Start(ctx, "repository.create_savings_event")
	defer span.End()

	query := `
		INSERT INTO savings_events (id, action_id, resource_id, optimization_type, estimated_savings, actual_savings)
		VALUES ($1, $2, $3, $4, $5, $6)
	`

	_, err := r.db.Exec(ctx, query,
		event.ID, event.ActionID, event.ResourceID, event.OptimizationType,
		event.EstimatedSavings, event.ActualSavings,
	)
	if err != nil {
		span.RecordError(err)
		return fmt.Errorf("failed to create savings event: %w", err)
	}

	return nil
}

// GetTokenUsageStats retrieves token usage statistics
func (r *Repository) GetTokenUsageStats(ctx context.Context, timeRange time.Duration) (map[string]interface{}, error) {
	ctx, span := r.tracer.Start(ctx, "repository.get_token_usage_stats")
	defer span.End()

	query := `
		SELECT 
			model,
			SUM(tokens) as total_tokens,
			SUM(cost_usd) as total_cost,
			COUNT(*) as request_count
		FROM token_usage 
		WHERE created_at >= NOW() - INTERVAL '1 hour' * $1
		GROUP BY model
		ORDER BY total_cost DESC
	`

	hours := int(timeRange.Hours())
	rows, err := r.db.Query(ctx, query, hours)
	if err != nil {
		span.RecordError(err)
		return nil, fmt.Errorf("failed to get token usage stats: %w", err)
	}
	defer rows.Close()

	var stats []map[string]interface{}
	for rows.Next() {
		var model string
		var totalTokens int
		var totalCost float64
		var requestCount int

		err := rows.Scan(&model, &totalTokens, &totalCost, &requestCount)
		if err != nil {
			span.RecordError(err)
			return nil, fmt.Errorf("failed to scan token usage stats: %w", err)
		}

		stats = append(stats, map[string]interface{}{
			"model":         model,
			"total_tokens":  totalTokens,
			"total_cost":    totalCost,
			"request_count": requestCount,
		})
	}

	return map[string]interface{}{
		"by_model":         stats,
		"time_range_hours": hours,
	}, nil
}

// CreateAuditLog creates a new audit log entry
func (r *Repository) CreateAuditLog(ctx context.Context, log *AuditLog) error {
	ctx, span := r.tracer.Start(ctx, "repository.create_audit_log")
	defer span.End()

	query := `
		INSERT INTO audit_log (id, user_id, action, resource_type, resource_id, details, ip_address)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`

	_, err := r.db.Exec(ctx, query,
		log.ID, log.UserID, log.Action, log.ResourceType, log.ResourceID, log.Details, log.IPAddress,
	)
	if err != nil {
		span.RecordError(err)
		return fmt.Errorf("failed to create audit log: %w", err)
	}

	return nil
}
