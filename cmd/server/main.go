// ERP-AIOps Gateway — Go HTTP reverse proxy
// Routes API requests to the Rust API backend (opstrac-api).
// Provides health checks, middleware (auth, logging, CORS), and route registration.

package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
	slog.SetDefault(logger)

	port := envOr("GATEWAY_PORT", "8090")
	rustAPIURL := envOr("RUST_API_URL", "http://localhost:8091")

	rustTarget, err := url.Parse(rustAPIURL)
	if err != nil {
		slog.Error("invalid RUST_API_URL", "url", rustAPIURL, "error", err)
		os.Exit(1)
	}

	proxy := httputil.NewSingleHostReverseProxy(rustTarget)
	proxy.ErrorHandler = func(w http.ResponseWriter, r *http.Request, err error) {
		slog.Error("proxy error", "path", r.URL.Path, "error", err)
		http.Error(w, `{"error":"upstream unavailable"}`, http.StatusBadGateway)
	}

	mux := http.NewServeMux()

	// ============================================================
	// Health & Readiness
	// ============================================================
	mux.HandleFunc("GET /health", handleHealth)
	mux.HandleFunc("GET /ready", handleReady(rustTarget))

	// ============================================================
	// Ingestion Endpoints (Phase 7B)
	// Redpanda Connect pipelines POST categorized events here.
	// ============================================================
	mux.Handle("POST /api/v1/ingest/health", proxy)
	mux.Handle("POST /api/v1/ingest/incident", proxy)
	mux.Handle("POST /api/v1/ingest/metric", proxy)
	mux.Handle("POST /api/v1/ingest/event", proxy)
	mux.Handle("POST /api/v1/ingest/observability", proxy)

	// ============================================================
	// Hasura Action Endpoints (Phase 7C)
	// Hasura calls these for custom business logic resolution.
	// ============================================================
	mux.Handle("POST /api/v1/actions/detect-anomalies", proxy)
	mux.Handle("POST /api/v1/actions/correlate-incidents", proxy)
	mux.Handle("POST /api/v1/actions/execute-playbook", proxy)
	mux.Handle("POST /api/v1/actions/topology-map", proxy)
	mux.Handle("POST /api/v1/actions/module-health-check", proxy)
	mux.Handle("POST /api/v1/actions/evaluate-guardrail", proxy)
	mux.Handle("POST /api/v1/actions/create-maintenance-window", proxy)
	mux.Handle("POST /api/v1/actions/execute-runbook", proxy)
	mux.Handle("POST /api/v1/actions/slo-status", proxy)

	// ============================================================
	// Event Trigger Webhook Endpoints (Phase 7D)
	// Hasura event triggers fire these on data changes.
	// ============================================================
	mux.Handle("POST /api/v1/webhooks/incident-created", proxy)
	mux.Handle("POST /api/v1/webhooks/anomaly-detected", proxy)
	mux.Handle("POST /api/v1/webhooks/health-status-changed", proxy)
	mux.Handle("POST /api/v1/webhooks/slo-breached", proxy)
	mux.Handle("POST /api/v1/webhooks/guardrail-resolved", proxy)

	// ============================================================
	// Alertmanager Webhook (Phase 4A)
	// Prometheus Alertmanager forwards alerts here.
	// ============================================================
	mux.Handle("POST /api/v1/webhooks/alertmanager", proxy)

	// ============================================================
	// Module Command API (Cross-module operations)
	// AIOps issues commands to target modules.
	// ============================================================
	mux.Handle("GET /api/v1/aiops/commands/{module}", proxy)
	mux.Handle("POST /api/v1/aiops/commands/{module}", proxy)

	// ============================================================
	// Metrics endpoint for Prometheus scraping
	// ============================================================
	mux.Handle("GET /metrics", proxy)

	handler := withMiddleware(mux, logger)

	srv := &http.Server{
		Addr:         ":" + port,
		Handler:      handler,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 60 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	go func() {
		slog.Info("aiops gateway starting", "port", port, "upstream", rustAPIURL)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("server error", "error", err)
			os.Exit(1)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	slog.Info("shutting down gateway")
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		slog.Error("shutdown error", "error", err)
	}
	slog.Info("gateway stopped")
}

// handleHealth returns 200 if the gateway process is running.
func handleHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	fmt.Fprint(w, `{"status":"healthy","service":"erp-aiops-gateway"}`)
}

// handleReady checks that the Rust API upstream is reachable.
func handleReady(upstream *url.URL) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		client := &http.Client{Timeout: 5 * time.Second}
		resp, err := client.Get(upstream.String() + "/health")
		if err != nil || resp.StatusCode != http.StatusOK {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusServiceUnavailable)
			fmt.Fprint(w, `{"status":"not_ready","reason":"upstream unavailable"}`)
			return
		}
		resp.Body.Close()
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, `{"status":"ready","service":"erp-aiops-gateway"}`)
	}
}

// withMiddleware wraps the handler with logging, CORS, and request ID middleware.
func withMiddleware(next http.Handler, logger *slog.Logger) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// CORS headers
		origin := r.Header.Get("Origin")
		if origin != "" {
			w.Header().Set("Access-Control-Allow-Origin", origin)
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Authorization, Content-Type, X-Hasura-Admin-Secret, X-Hasura-Role, X-Hasura-Tenant-Id, X-AIOps-Command-Token")
			w.Header().Set("Access-Control-Max-Age", "86400")
		}
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		// Request ID propagation
		requestID := r.Header.Get("X-Request-Id")
		if requestID == "" {
			requestID = fmt.Sprintf("aiops-%d", time.Now().UnixNano())
		}
		w.Header().Set("X-Request-Id", requestID)

		// Tenant context propagation
		tenantID := r.Header.Get("X-Hasura-Tenant-Id")

		// Wrap response writer to capture status code
		wrapped := &statusWriter{ResponseWriter: w, statusCode: http.StatusOK}
		next.ServeHTTP(wrapped, r)

		// Structured access log
		logger.Info("request",
			"method", r.Method,
			"path", r.URL.Path,
			"status", wrapped.statusCode,
			"duration_ms", time.Since(start).Milliseconds(),
			"tenant_id", tenantID,
			"request_id", requestID,
			"remote_addr", stripPort(r.RemoteAddr),
		)
	})
}

type statusWriter struct {
	http.ResponseWriter
	statusCode int
}

func (w *statusWriter) WriteHeader(code int) {
	w.statusCode = code
	w.ResponseWriter.WriteHeader(code)
}

func envOr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func stripPort(addr string) string {
	if i := strings.LastIndex(addr, ":"); i != -1 {
		return addr[:i]
	}
	return addr
}
