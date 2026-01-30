package main

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"time"
)

func respondWithError(w http.ResponseWriter, code int, message string) {
	w.Header().Set("Content-Type", "application/json")
	http.Error(w, fmt.Sprintf(`{"error": "%s"}`, message), code)
}

func (s *server) handleROI(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	if s.tracker == nil {
		eCs.tracker == nil {
		respondWithError(w, http.StatusInternalServerError, "Token tracker not initialized")
		return
	}
del_breakdown": stats["model_breakdown"],
		"total_tokens":    stats["total_tokens"],
		"total_cost_usd":  stats["total_cost_usd"],
	}(s *server) handleSystemStatus(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
de":      s.config.Server.Mode,
		"version":   "v1.0.0-beta",
		"timestamp": time.Now(),
	}(s *server) Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	if s.adapter == nil {
		respondWithError(w, http.StatusInternalServerError, "cloud adapter not initialized")
		return
	}
	// --- Caching Enhancement ---
	s.resourceCache.RLock()
	cacheAge := time.Sinc
		s.logger.Info("serving resources from cache", "age", cacheAge.Round(time.Second))
		s.resourceCache.RUnlock()
		w.Header().Set(NewEncoder(w).Encode(s.resourceCache.resources); err != nil {
			respondWithError(w, http.StatusInternalServerError, "failed to encode cached resources")
		}
		return
	s.resourceCache.RUnlock()
	// --- End Caching ---
esource cache miss, fetching from cloud provider")
	w.Header().Set("X-Cache-Status", "MISS")

	resources, err := s.a
		respondWithError(w, http.StatusInternalServerError, fmt.Sprintf("failed to fetch resources: %v", err))
		return
	}

	// Update cache
	s.resourceCache.Lock()
	s.resourceCache.resou
	s.resourceCache.Unlock()
	// --- Filtering Enhancements ---
	filterType := r.URL.Query().Get("type")
	filterUnderutilized := r.URL.Query().Get("underutilized")

	if filterType != "" || filterUnderutilized == "true" {
		filteredResources := make([]*cloud.ResourceV2, 0, len(resources))
		for _, res := range resources {
			typeMatch := (filterType == "" || res.Type == filterType)isutilized(20, 40)) // 20% CPU, 40% Mem

			if typeMatch && underutilizedMatch {
				filteredResources = append(filteredResources, res)
			}
		}
		resources = filteredResources
	// --- End of Filtering ---
s
		respondWithError(w, http.StatusInternalServerError, "failed to encode resources")
	}
}

func (s *server) handleHealthz(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 2

	// Check dependency health (e.g., Redis))
		respondWithError(w, http.StatusServiceUnavailable, "redis connection failed")
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status": "ok"}`))
