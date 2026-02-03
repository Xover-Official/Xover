package main

import "net/http"

// routes sets up all the HTTP handlers for the dashboard application.
func (s *server) routes() http.Handler {
	router := http.NewServeMux()

	// Publicly accessible static files for the frontend.
	fs := http.FileServer(http.Dir("./web"))
	router.Handle("/", fs)

	// Publicly accessible health check.
	router.HandleFunc("/healthz", s.handleHealthz)

	// Public auth endpoints for the SSO login/logout/callback flow.
	router.HandleFunc("/auth/login/", s.handleLogin)
	router.HandleFunc("/auth/callback/", s.handleCallback)
	router.HandleFunc("/auth/logout", s.handleLogout)

	// API endpoints are grouped together and protected by the authentication middleware.
	api := http.NewServeMux()
	api.HandleFunc("/roi", s.handleROI)
	api.HandleFunc("/token-breakdown", s.handleTokenBreakdown)
	api.HandleFunc("/system/status", s.handleSystemStatus)
	api.HandleFunc("/resources", s.handleResources)
	api.HandleFunc("/token-stats", s.handleTokenStats)
	api.HandleFunc("/resource-metrics", s.handleResourceMetrics)
	api.HandleFunc("/optimization-suggestions", s.handleOptimizationSuggestions)
	api.HandleFunc("/dashboard/stats", s.handleDashboardStats)
	api.HandleFunc("/dashboard/opportunities", s.handleOpportunities)
	api.HandleFunc("/dashboard/anomalies", s.handleAnomalies)
	api.HandleFunc("/feedback", s.handleSubmitFeedback)

	// Mount the protected API endpoints under the /api/ path.
	// http.StripPrefix is used to remove the "/api" prefix before the request reaches the 'api' mux,
	// so that handlers can be registered with paths like "/roi" instead of "/api/roi".
	router.Handle("/api/", http.StripPrefix("/api", s.authMiddleware(api)))

	return router
}
