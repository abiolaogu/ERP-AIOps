package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"strings"
	"sync/atomic"
	"time"
)

type capabilityDoc struct {
	Module          string   `json:"module"`
	Version         string   `json:"version,omitempty"`
	Capabilities    []string `json:"capabilities"`
	IntegrationMode string   `json:"integration_mode,omitempty"`
	AIDDGovernance  string   `json:"aidd_governance,omitempty"`
}

var reqCounter uint64

func loadCapabilities() capabilityDoc {
	b, err := os.ReadFile("configs/capabilities.json")
	if err != nil {
		return capabilityDoc{Module: "ERP-AIOps", Capabilities: []string{"unconfigured"}}
	}
	var d capabilityDoc
	if err := json.Unmarshal(b, &d); err != nil {
		return capabilityDoc{Module: "ERP-AIOps", Capabilities: []string{"invalid_config"}}
	}
	if d.Module == "" {
		d.Module = "ERP-AIOps"
	}
	return d
}

func writeJSON(w http.ResponseWriter, code int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(payload)
}

func nextRequestID(r *http.Request) string {
	if id := r.Header.Get("X-Request-ID"); id != "" {
		return id
	}
	n := atomic.AddUint64(&reqCounter, 1)
	return fmt.Sprintf("req-%d-%d", time.Now().UnixNano(), n)
}

type statusRecorder struct {
	http.ResponseWriter
	status int
}

func (s *statusRecorder) WriteHeader(status int) {
	s.status = status
	s.ResponseWriter.WriteHeader(status)
}

// envOrDefault returns the value of the environment variable named by key,
// or the provided fallback if the variable is unset or empty.
func envOrDefault(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

// corsOrigins returns the set of allowed CORS origins read from the
// CORS_ORIGINS environment variable (comma-separated). If the variable
// is not set, it defaults to allowing all origins ("*").
func corsOrigins() []string {
	raw := os.Getenv("CORS_ORIGINS")
	if raw == "" {
		return []string{"*"}
	}
	parts := strings.Split(raw, ",")
	origins := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			origins = append(origins, p)
		}
	}
	if len(origins) == 0 {
		return []string{"*"}
	}
	return origins
}

// withCORS wraps a handler with CORS header handling. Allowed origins are
// read once at startup from the CORS_ORIGINS env var.
func withCORS(origins []string, next http.Handler) http.Handler {
	allowAll := len(origins) == 1 && origins[0] == "*"
	originSet := make(map[string]struct{}, len(origins))
	for _, o := range origins {
		originSet[o] = struct{}{}
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")

		if allowAll {
			w.Header().Set("Access-Control-Allow-Origin", "*")
		} else if origin != "" {
			if _, ok := originSet[origin]; ok {
				w.Header().Set("Access-Control-Allow-Origin", origin)
				w.Header().Set("Vary", "Origin")
			}
		}

		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Request-ID, X-Tenant-ID")
		w.Header().Set("Access-Control-Max-Age", "86400")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func withServerDefaults(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		requestID := nextRequestID(r)

		w.Header().Set("X-Request-ID", requestID)
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("X-Frame-Options", "DENY")
		w.Header().Set("Referrer-Policy", "no-referrer")
		w.Header().Set("Cache-Control", "no-store")

		rec := &statusRecorder{ResponseWriter: w, status: http.StatusOK}
		next.ServeHTTP(rec, r)

		tenant := r.Header.Get("X-Tenant-ID")
		if tenant == "" {
			tenant = "system"
		}
		log.Printf("method=%s path=%s status=%d duration_ms=%d tenant=%s request_id=%s remote=%s",
			r.Method,
			r.URL.Path,
			rec.status,
			time.Since(start).Milliseconds(),
			tenant,
			requestID,
			r.RemoteAddr,
		)
	})
}

// proxyRoute registers a reverse proxy for the given path prefix. Requests
// to both "/prefix" and "/prefix/..." are forwarded to the backend URL.
func proxyRoute(mux *http.ServeMux, pathPrefix, backendURL string) {
	target, err := url.Parse(backendURL)
	if err != nil {
		log.Fatalf("invalid backend URL %q for prefix %s: %v", backendURL, pathPrefix, err)
	}
	proxy := httputil.NewSingleHostReverseProxy(target)

	proxy.ErrorHandler = func(w http.ResponseWriter, r *http.Request, err error) {
		log.Printf("proxy error prefix=%s target=%s err=%v", pathPrefix, backendURL, err)
		writeJSON(w, http.StatusBadGateway, map[string]string{
			"error":  "bad_gateway",
			"detail": fmt.Sprintf("upstream %s unreachable", pathPrefix),
		})
	}

	mux.HandleFunc(pathPrefix+"/", func(w http.ResponseWriter, r *http.Request) {
		proxy.ServeHTTP(w, r)
	})
	mux.HandleFunc(pathPrefix, func(w http.ResponseWriter, r *http.Request) {
		proxy.ServeHTTP(w, r)
	})

	log.Printf("route %s -> %s", pathPrefix, backendURL)
}

func main() {
	doc := loadCapabilities()
	origins := corsOrigins()
	mux := http.NewServeMux()

	// --- Health & capabilities (gateway-level) ---

	mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			writeJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "method not allowed"})
			return
		}
		writeJSON(w, http.StatusOK, map[string]string{"status": "healthy", "module": doc.Module})
	})

	mux.HandleFunc("/v1/capabilities", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			writeJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "method not allowed"})
			return
		}
		writeJSON(w, http.StatusOK, doc)
	})

	mux.HandleFunc("/v1/aiops/health", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			writeJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "method not allowed"})
			return
		}
		writeJSON(w, http.StatusOK, map[string]string{
			"status": "healthy",
			"module": doc.Module,
			"time":   time.Now().UTC().Format(time.RFC3339),
		})
	})

	// --- Reverse-proxy routes to Rust API backend ---

	rustAPI := envOrDefault("RUST_API_URL", "http://rust-api:8080")

	proxyRoute(mux, "/v1/aiops/incidents", rustAPI)
	proxyRoute(mux, "/v1/aiops/anomalies", rustAPI)
	proxyRoute(mux, "/v1/aiops/rules", rustAPI)
	proxyRoute(mux, "/v1/aiops/topology", rustAPI)
	proxyRoute(mux, "/v1/aiops/remediation", rustAPI)
	proxyRoute(mux, "/v1/aiops/cost", rustAPI)
	proxyRoute(mux, "/v1/aiops/security", rustAPI)

	// --- Ingestion routes ---

	proxyRoute(mux, "/api/v1/ingest/health", rustAPI)
	proxyRoute(mux, "/api/v1/ingest/incident", rustAPI)
	proxyRoute(mux, "/api/v1/ingest/metric", rustAPI)
	proxyRoute(mux, "/api/v1/ingest/event", rustAPI)
	proxyRoute(mux, "/api/v1/ingest/observability", rustAPI)

	// --- Hasura Action routes ---

	proxyRoute(mux, "/api/v1/actions/module-health-check", rustAPI)
	proxyRoute(mux, "/api/v1/actions/evaluate-guardrail", rustAPI)
	proxyRoute(mux, "/api/v1/actions/create-maintenance-window", rustAPI)
	proxyRoute(mux, "/api/v1/actions/execute-runbook", rustAPI)
	proxyRoute(mux, "/api/v1/actions/slo-status", rustAPI)

	// --- Event Trigger Webhook routes ---

	proxyRoute(mux, "/api/v1/webhooks/incident-created", rustAPI)
	proxyRoute(mux, "/api/v1/webhooks/anomaly-detected", rustAPI)
	proxyRoute(mux, "/api/v1/webhooks/health-status-changed", rustAPI)
	proxyRoute(mux, "/api/v1/webhooks/slo-breached", rustAPI)
	proxyRoute(mux, "/api/v1/webhooks/guardrail-resolved", rustAPI)

	// --- Alertmanager Webhook ---

	proxyRoute(mux, "/api/v1/webhooks/alertmanager", rustAPI)

	// --- Module Command API ---

	proxyRoute(mux, "/api/v1/aiops/commands", rustAPI)

	// --- Reverse-proxy routes to AI Brain backend ---

	aiBrain := envOrDefault("AI_BRAIN_URL", "http://ai-brain:8001")

	proxyRoute(mux, "/v1/aiops/analyze", aiBrain)
	proxyRoute(mux, "/v1/aiops/forecast", aiBrain)

	// --- Server ---

	port := os.Getenv("PORT")
	if port == "" {
		port = "8090"
	}
	addr := ":" + port

	handler := withServerDefaults(withCORS(origins, mux))

	srv := &http.Server{
		Addr:              addr,
		Handler:           handler,
		ReadHeaderTimeout: 2 * time.Second,
		ReadTimeout:       15 * time.Second,
		WriteTimeout:      30 * time.Second,
		IdleTimeout:       120 * time.Second,
		MaxHeaderBytes:    1 << 20,
	}

	log.Printf("%s gateway listening on %s (CORS origins: %v)", doc.Module, addr, origins)
	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatal(err)
	}
}
