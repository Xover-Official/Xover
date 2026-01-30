package main

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/project-atlas/atlas/internal/health"
)

// HealthHandler serves health check endpoints
type HealthHandler struct {
	manager *health.HealthManager
}

// NewHealthHandler creates a new health handler
func NewHealthHandler(manager *health.HealthManager) *HealthHandler {
	return &HealthHandler{manager: manager}
}

// HandleHealthz serves the /healthz endpoint (simple liveness check)
func (h *HealthHandler) HandleHealthz(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":    "ok",
		"timestamp": time.Now().Unix(),
	})
}

// HandleReady serves the /ready endpoint (full readiness check)
func (h *HealthHandler) HandleReady(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	status := h.manager.GetOverallStatus(ctx)

	response := map[string]interface{}{
		"status":    status,
		"timestamp": time.Now().Unix(),
	}

	w.Header().Set("Content-Type", "application/json")

	switch status {
	case health.StatusHealthy:
		w.WriteHeader(http.StatusOK)
	case health.StatusDegraded:
		w.WriteHeader(http.StatusOK) // Still operational
		response["warning"] = "Some components degraded"
	case health.StatusUnhealthy:
		w.WriteHeader(http.StatusServiceUnavailable)
		response["error"] = "System unhealthy"
	}

	json.NewEncoder(w).Encode(response)
}

// HandleHealth serves the /health endpoint (detailed health info)
func (h *HealthHandler) HandleHealth(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	results := h.manager.CheckAll(ctx)
	overallStatus := h.manager.GetOverallStatus(ctx)

	response := map[string]interface{}{
		"status":    overallStatus,
		"checks":    results,
		"timestamp": time.Now().Unix(),
	}

	w.Header().Set("Content-Type", "application/json")

	if overallStatus == health.StatusUnhealthy {
		w.WriteHeader(http.StatusServiceUnavailable)
	} else {
		w.WriteHeader(http.StatusOK)
	}

	json.NewEncoder(w).Encode(response)
}
