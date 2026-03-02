package main

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"sort"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
)

const (
	serviceName = "anomaly-service"
	moduleName  = "ERP-AIOps"
	basePath    = "/v1/aiops/anomalies"
	dbName      = "erp_aiops"
	tableName   = "aiops_anomalies"
	eventTopic  = "erp.aiops.anomaly"
	cacheTTL    = 30 * time.Second
)

var (
	validSeverities     = map[string]bool{"critical": true, "high": true, "medium": true, "low": true}
	validStatuses       = map[string]bool{"detected": true, "investigating": true, "confirmed": true, "false_positive": true, "resolved": true}
	validClassifications = map[string]bool{"spike": true, "drop": true, "drift": true, "pattern_change": true, "outlier": true, "unknown": true}
)

type anomaly struct {
	ID             string  `json:"id"`
	TenantID       string  `json:"tenant_id"`
	MetricName     string  `json:"metric_name"`
	MetricValue    float64 `json:"metric_value"`
	ExpectedValue  float64 `json:"expected_value"`
	DeviationPct   float64 `json:"deviation_percent"`
	Classification string  `json:"classification"`
	Severity       string  `json:"severity"`
	Status         string  `json:"status"`
	ServiceName    *string `json:"service_name,omitempty"`
	Environment    *string `json:"environment,omitempty"`
	Source         *string `json:"source,omitempty"`
	Description    *string `json:"description,omitempty"`
	DetectedAt     string  `json:"detected_at"`
	ResolvedAt     *string `json:"resolved_at,omitempty"`
	CorrelationID  *string `json:"correlation_id,omitempty"`
	Tags           *string `json:"tags,omitempty"`
	CreatedAt      string  `json:"created_at"`
	UpdatedAt      string  `json:"updated_at"`
}

func newID() string { b := make([]byte, 16); _, _ = rand.Read(b); return hex.EncodeToString(b) }
func writeJSON(w http.ResponseWriter, code int, v any) { w.Header().Set("Content-Type", "application/json"); w.WriteHeader(code); _ = json.NewEncoder(w).Encode(v) }
func readJSON(r *http.Request) (map[string]any, error) {
	body, err := io.ReadAll(io.LimitReader(r.Body, 1<<20)); if err != nil { return nil, err }; defer r.Body.Close()
	if len(body) == 0 { return nil, errors.New("empty body") }
	var m map[string]any; if err := json.Unmarshal(body, &m); err != nil { return nil, err }; return m, nil
}
func strPtr(v any) *string { if v == nil { return nil }; s := fmt.Sprintf("%v", v); return &s }
func strVal(p *string) string { if p == nil { return "" }; return *p }

type cacheEntry struct { data any; expiresAt time.Time }
type ttlCache struct { mu sync.RWMutex; entries map[string]cacheEntry }
func newCache() *ttlCache { c := &ttlCache{entries: make(map[string]cacheEntry)}; go func() { t := time.NewTicker(30 * time.Second); defer t.Stop(); for range t.C { c.evict() } }(); return c }
func (c *ttlCache) get(key string) (any, bool) { c.mu.RLock(); defer c.mu.RUnlock(); e, ok := c.entries[key]; if !ok || time.Now().After(e.expiresAt) { return nil, false }; return e.data, true }
func (c *ttlCache) set(key string, data any) { c.mu.Lock(); defer c.mu.Unlock(); c.entries[key] = cacheEntry{data: data, expiresAt: time.Now().Add(cacheTTL)} }
func (c *ttlCache) invalidate(prefix string) { c.mu.Lock(); defer c.mu.Unlock(); for k := range c.entries { if strings.HasPrefix(k, prefix) { delete(c.entries, k) } } }
func (c *ttlCache) evict() { c.mu.Lock(); defer c.mu.Unlock(); now := time.Now(); for k, e := range c.entries { if now.After(e.expiresAt) { delete(c.entries, k) } } }

var requestCount atomic.Int64

func securityHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Content-Type-Options", "nosniff"); w.Header().Set("X-Frame-Options", "DENY")
		w.Header().Set("X-XSS-Protection", "1; mode=block"); w.Header().Set("Cache-Control", "no-store")
		w.Header().Set("Strict-Transport-Security", "max-age=31536000; includeSubDomains"); next.ServeHTTP(w, r)
	})
}

type store interface {
	List(ctx context.Context, tenantID, cursor string, limit int, filters map[string]string) ([]anomaly, string, error)
	GetByID(ctx context.Context, tenantID, id string) (*anomaly, error)
	Create(ctx context.Context, a *anomaly) error
	Update(ctx context.Context, a *anomaly) error
	Delete(ctx context.Context, tenantID, id string) error
}

type memoryStore struct { mu sync.RWMutex; records map[string]anomaly }
func newMemoryStore() *memoryStore { return &memoryStore{records: make(map[string]anomaly)} }

func (m *memoryStore) List(_ context.Context, tenantID, cursor string, limit int, filters map[string]string) ([]anomaly, string, error) {
	m.mu.RLock(); defer m.mu.RUnlock()
	var all []anomaly
	for _, a := range m.records {
		if a.TenantID != tenantID { continue }
		if v, ok := filters["severity"]; ok && a.Severity != v { continue }
		if v, ok := filters["status"]; ok && a.Status != v { continue }
		if v, ok := filters["classification"]; ok && a.Classification != v { continue }
		if v, ok := filters["metric_name"]; ok && a.MetricName != v { continue }
		all = append(all, a)
	}
	sort.Slice(all, func(i, j int) bool { return all[i].CreatedAt < all[j].CreatedAt })
	start := 0; if cursor != "" { for idx, a := range all { if a.ID == cursor { start = idx + 1; break } } }
	if start >= len(all) { return []anomaly{}, "", nil }
	end := start + limit; if end > len(all) { end = len(all) }
	result := all[start:end]; nextCursor := ""; if end < len(all) { nextCursor = result[len(result)-1].ID }
	return result, nextCursor, nil
}

func (m *memoryStore) GetByID(_ context.Context, tenantID, id string) (*anomaly, error) { m.mu.RLock(); defer m.mu.RUnlock(); a, ok := m.records[id]; if !ok || a.TenantID != tenantID { return nil, errors.New("not found") }; return &a, nil }
func (m *memoryStore) Create(_ context.Context, a *anomaly) error { m.mu.Lock(); defer m.mu.Unlock(); m.records[a.ID] = *a; return nil }
func (m *memoryStore) Update(_ context.Context, a *anomaly) error { m.mu.Lock(); defer m.mu.Unlock(); if _, ok := m.records[a.ID]; !ok { return errors.New("not found") }; m.records[a.ID] = *a; return nil }
func (m *memoryStore) Delete(_ context.Context, tenantID, id string) error { m.mu.Lock(); defer m.mu.Unlock(); a, ok := m.records[id]; if !ok || a.TenantID != tenantID { return errors.New("not found") }; delete(m.records, id); return nil }

type postgresStore struct{ db *sql.DB }
func newPostgresStore(dsn string) (*postgresStore, error) {
	db, err := sql.Open("pgx", dsn); if err != nil { return nil, fmt.Errorf("open db: %w", err) }
	db.SetMaxOpenConns(25); db.SetMaxIdleConns(5); db.SetConnMaxLifetime(5 * time.Minute)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second); defer cancel()
	if err := db.PingContext(ctx); err != nil { return nil, fmt.Errorf("ping db: %w", err) }
	if _, err := db.ExecContext(ctx, createTableSQL); err != nil { return nil, fmt.Errorf("create table: %w", err) }
	return &postgresStore{db: db}, nil
}

const createTableSQL = `CREATE TABLE IF NOT EXISTS ` + tableName + ` (
	id TEXT PRIMARY KEY, tenant_id TEXT NOT NULL, metric_name TEXT NOT NULL,
	metric_value DOUBLE PRECISION DEFAULT 0, expected_value DOUBLE PRECISION DEFAULT 0,
	deviation_percent DOUBLE PRECISION DEFAULT 0,
	classification TEXT CHECK (classification IN ('spike','drop','drift','pattern_change','outlier','unknown')) DEFAULT 'unknown',
	severity TEXT CHECK (severity IN ('critical','high','medium','low')) DEFAULT 'medium',
	status TEXT CHECK (status IN ('detected','investigating','confirmed','false_positive','resolved')) DEFAULT 'detected',
	service_name TEXT, environment TEXT, source TEXT, description TEXT,
	detected_at TIMESTAMPTZ NOT NULL DEFAULT NOW(), resolved_at TIMESTAMPTZ,
	correlation_id TEXT, tags TEXT,
	created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(), updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_` + tableName + `_tenant ON ` + tableName + ` (tenant_id);
CREATE INDEX IF NOT EXISTS idx_` + tableName + `_status ON ` + tableName + ` (tenant_id, status);
CREATE INDEX IF NOT EXISTS idx_` + tableName + `_severity ON ` + tableName + ` (tenant_id, severity);`

func (p *postgresStore) List(ctx context.Context, tenantID, cursor string, limit int, filters map[string]string) ([]anomaly, string, error) {
	query := `SELECT id,tenant_id,metric_name,metric_value,expected_value,deviation_percent,classification,severity,status,service_name,environment,source,description,detected_at,resolved_at,correlation_id,tags,created_at,updated_at FROM ` + tableName + ` WHERE tenant_id=$1`
	args := []any{tenantID}; idx := 2
	for _, f := range []struct{ k, c string }{{"severity", "severity"}, {"status", "status"}, {"classification", "classification"}, {"metric_name", "metric_name"}} {
		if v, ok := filters[f.k]; ok { query += fmt.Sprintf(" AND %s=$%d", f.c, idx); args = append(args, v); idx++ }
	}
	if cursor != "" { query += fmt.Sprintf(" AND created_at>(SELECT created_at FROM "+tableName+" WHERE id=$%d)", idx); args = append(args, cursor); idx++ }
	query += " ORDER BY created_at ASC" + fmt.Sprintf(" LIMIT $%d", idx); args = append(args, limit+1)
	rows, err := p.db.QueryContext(ctx, query, args...); if err != nil { return nil, "", fmt.Errorf("query: %w", err) }; defer rows.Close()
	var results []anomaly
	for rows.Next() {
		var a anomaly; var svcName, env, src, desc, corrID, tags sql.NullString; var detectedAt, resolvedAt, createdAt, updatedAt sql.NullTime
		if err := rows.Scan(&a.ID, &a.TenantID, &a.MetricName, &a.MetricValue, &a.ExpectedValue, &a.DeviationPct, &a.Classification, &a.Severity, &a.Status, &svcName, &env, &src, &desc, &detectedAt, &resolvedAt, &corrID, &tags, &createdAt, &updatedAt); err != nil { return nil, "", fmt.Errorf("scan: %w", err) }
		if svcName.Valid { a.ServiceName = &svcName.String }; if env.Valid { a.Environment = &env.String }
		if src.Valid { a.Source = &src.String }; if desc.Valid { a.Description = &desc.String }
		if detectedAt.Valid { a.DetectedAt = detectedAt.Time.Format(time.RFC3339) }
		if resolvedAt.Valid { s := resolvedAt.Time.Format(time.RFC3339); a.ResolvedAt = &s }
		if corrID.Valid { a.CorrelationID = &corrID.String }; if tags.Valid { a.Tags = &tags.String }
		if createdAt.Valid { a.CreatedAt = createdAt.Time.Format(time.RFC3339) }; if updatedAt.Valid { a.UpdatedAt = updatedAt.Time.Format(time.RFC3339) }
		results = append(results, a)
	}
	nextCursor := ""; if len(results) > limit { nextCursor = results[limit-1].ID; results = results[:limit] }
	return results, nextCursor, nil
}

func (p *postgresStore) GetByID(ctx context.Context, tenantID, id string) (*anomaly, error) {
	query := `SELECT id,tenant_id,metric_name,metric_value,expected_value,deviation_percent,classification,severity,status,service_name,environment,source,description,detected_at,resolved_at,correlation_id,tags,created_at,updated_at FROM ` + tableName + ` WHERE id=$1 AND tenant_id=$2`
	var a anomaly; var svcName, env, src, desc, corrID, tags sql.NullString; var detectedAt, resolvedAt, createdAt, updatedAt sql.NullTime
	err := p.db.QueryRowContext(ctx, query, id, tenantID).Scan(&a.ID, &a.TenantID, &a.MetricName, &a.MetricValue, &a.ExpectedValue, &a.DeviationPct, &a.Classification, &a.Severity, &a.Status, &svcName, &env, &src, &desc, &detectedAt, &resolvedAt, &corrID, &tags, &createdAt, &updatedAt)
	if err != nil { if errors.Is(err, sql.ErrNoRows) { return nil, errors.New("not found") }; return nil, fmt.Errorf("query row: %w", err) }
	if svcName.Valid { a.ServiceName = &svcName.String }; if env.Valid { a.Environment = &env.String }
	if src.Valid { a.Source = &src.String }; if desc.Valid { a.Description = &desc.String }
	if detectedAt.Valid { a.DetectedAt = detectedAt.Time.Format(time.RFC3339) }
	if resolvedAt.Valid { s := resolvedAt.Time.Format(time.RFC3339); a.ResolvedAt = &s }
	if corrID.Valid { a.CorrelationID = &corrID.String }; if tags.Valid { a.Tags = &tags.String }
	if createdAt.Valid { a.CreatedAt = createdAt.Time.Format(time.RFC3339) }; if updatedAt.Valid { a.UpdatedAt = updatedAt.Time.Format(time.RFC3339) }
	return &a, nil
}

func (p *postgresStore) Create(ctx context.Context, a *anomaly) error {
	query := `INSERT INTO ` + tableName + ` (id,tenant_id,metric_name,metric_value,expected_value,deviation_percent,classification,severity,status,service_name,environment,source,description,detected_at,resolved_at,correlation_id,tags,created_at,updated_at) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17,$18,$19)`
	_, err := p.db.ExecContext(ctx, query, a.ID, a.TenantID, a.MetricName, a.MetricValue, a.ExpectedValue, a.DeviationPct, a.Classification, a.Severity, a.Status, a.ServiceName, a.Environment, a.Source, a.Description, parseTime(a.DetectedAt), parseTimePtr(a.ResolvedAt), a.CorrelationID, a.Tags, parseTime(a.CreatedAt), parseTime(a.UpdatedAt))
	return err
}

func (p *postgresStore) Update(ctx context.Context, a *anomaly) error {
	query := `UPDATE ` + tableName + ` SET metric_name=$1,metric_value=$2,expected_value=$3,deviation_percent=$4,classification=$5,severity=$6,status=$7,service_name=$8,environment=$9,source=$10,description=$11,resolved_at=$12,correlation_id=$13,tags=$14,updated_at=$15 WHERE id=$16 AND tenant_id=$17`
	res, err := p.db.ExecContext(ctx, query, a.MetricName, a.MetricValue, a.ExpectedValue, a.DeviationPct, a.Classification, a.Severity, a.Status, a.ServiceName, a.Environment, a.Source, a.Description, parseTimePtr(a.ResolvedAt), a.CorrelationID, a.Tags, parseTime(a.UpdatedAt), a.ID, a.TenantID)
	if err != nil { return err }; n, _ := res.RowsAffected(); if n == 0 { return errors.New("not found") }; return nil
}

func (p *postgresStore) Delete(ctx context.Context, tenantID, id string) error {
	res, err := p.db.ExecContext(ctx, `DELETE FROM `+tableName+` WHERE id=$1 AND tenant_id=$2`, id, tenantID)
	if err != nil { return err }; n, _ := res.RowsAffected(); if n == 0 { return errors.New("not found") }; return nil
}

func parseTime(s string) time.Time { t, err := time.Parse(time.RFC3339, s); if err != nil { return time.Now() }; return t }
func parseTimePtr(s *string) *time.Time { if s == nil { return nil }; t, err := time.Parse(time.RFC3339, *s); if err != nil { return nil }; return &t }

type server struct { store store; cache *ttlCache }

func (sv *server) handleList(w http.ResponseWriter, r *http.Request) {
	tenantID := r.Header.Get("X-Tenant-ID"); cursor := r.URL.Query().Get("cursor")
	limit := 20; if v, err := strconv.Atoi(r.URL.Query().Get("limit")); err == nil && v > 0 && v <= 100 { limit = v }
	filters := make(map[string]string); for _, key := range []string{"severity", "status", "classification", "metric_name"} { if v := r.URL.Query().Get(key); v != "" { filters[key] = v } }
	cacheKey := fmt.Sprintf("list:%s:%s:%d:%v", tenantID, cursor, limit, filters)
	if cached, ok := sv.cache.get(cacheKey); ok { writeJSON(w, http.StatusOK, cached); return }
	items, nextCursor, err := sv.store.List(r.Context(), tenantID, cursor, limit, filters)
	if err != nil { writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()}); return }
	resp := map[string]any{"items": items, "next_cursor": nextCursor, "limit": limit, "count": len(items), "event_topic": eventTopic + ".listed"}
	sv.cache.set(cacheKey, resp); writeJSON(w, http.StatusOK, resp)
}

func (sv *server) handleGet(w http.ResponseWriter, r *http.Request, id string) {
	tenantID := r.Header.Get("X-Tenant-ID"); cacheKey := fmt.Sprintf("get:%s:%s", tenantID, id)
	if cached, ok := sv.cache.get(cacheKey); ok { writeJSON(w, http.StatusOK, cached); return }
	item, err := sv.store.GetByID(r.Context(), tenantID, id)
	if err != nil { if err.Error() == "not found" { writeJSON(w, http.StatusNotFound, map[string]string{"error": "not found"}); return }; writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()}); return }
	resp := map[string]any{"item": item, "event_topic": eventTopic + ".read"}; sv.cache.set(cacheKey, resp); writeJSON(w, http.StatusOK, resp)
}

func (sv *server) handleCreate(w http.ResponseWriter, r *http.Request) {
	tenantID := r.Header.Get("X-Tenant-ID"); body, err := readJSON(r)
	if err != nil { writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid JSON: " + err.Error()}); return }
	metricName, _ := body["metric_name"].(string)
	if metricName == "" { writeJSON(w, http.StatusBadRequest, map[string]string{"error": "metric_name is required"}); return }
	now := time.Now().UTC().Format(time.RFC3339)
	a := &anomaly{ID: newID(), TenantID: tenantID, MetricName: metricName, MetricValue: 0, ExpectedValue: 0, DeviationPct: 0, Classification: "unknown", Severity: "medium", Status: "detected", ServiceName: strPtr(body["service_name"]), Environment: strPtr(body["environment"]), Source: strPtr(body["source"]), Description: strPtr(body["description"]), DetectedAt: now, CorrelationID: strPtr(body["correlation_id"]), Tags: strPtr(body["tags"]), CreatedAt: now, UpdatedAt: now}
	if v, ok := body["metric_value"].(float64); ok { a.MetricValue = v }
	if v, ok := body["expected_value"].(float64); ok { a.ExpectedValue = v }
	if a.ExpectedValue != 0 { a.DeviationPct = ((a.MetricValue - a.ExpectedValue) / a.ExpectedValue) * 100 }
	if v, ok := body["classification"].(string); ok && validClassifications[v] { a.Classification = v }
	if v, ok := body["severity"].(string); ok && validSeverities[v] { a.Severity = v }
	if err := sv.store.Create(r.Context(), a); err != nil { writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()}); return }
	sv.cache.invalidate("list:" + tenantID); writeJSON(w, http.StatusCreated, map[string]any{"item": a, "event_topic": eventTopic + ".created"})
}

func (sv *server) handleUpdate(w http.ResponseWriter, r *http.Request, id string) {
	tenantID := r.Header.Get("X-Tenant-ID")
	existing, err := sv.store.GetByID(r.Context(), tenantID, id)
	if err != nil { if err.Error() == "not found" { writeJSON(w, http.StatusNotFound, map[string]string{"error": "not found"}); return }; writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()}); return }
	body, err := readJSON(r)
	if err != nil { writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid JSON: " + err.Error()}); return }
	if v, ok := body["metric_name"].(string); ok && v != "" { existing.MetricName = v }
	if v, ok := body["metric_value"].(float64); ok { existing.MetricValue = v }
	if v, ok := body["expected_value"].(float64); ok { existing.ExpectedValue = v }
	if v, ok := body["classification"].(string); ok && validClassifications[v] { existing.Classification = v }
	if v, ok := body["severity"].(string); ok && validSeverities[v] { existing.Severity = v }
	if v, ok := body["status"].(string); ok && validStatuses[v] { existing.Status = v }
	if v, exists := body["service_name"]; exists { existing.ServiceName = strPtr(v) }
	if v, exists := body["environment"]; exists { existing.Environment = strPtr(v) }
	if v, exists := body["source"]; exists { existing.Source = strPtr(v) }
	if v, exists := body["description"]; exists { existing.Description = strPtr(v) }
	if v, exists := body["correlation_id"]; exists { existing.CorrelationID = strPtr(v) }
	if v, exists := body["tags"]; exists { existing.Tags = strPtr(v) }
	if existing.ExpectedValue != 0 { existing.DeviationPct = ((existing.MetricValue - existing.ExpectedValue) / existing.ExpectedValue) * 100 }
	now := time.Now().UTC().Format(time.RFC3339)
	if existing.Status == "resolved" && existing.ResolvedAt == nil { existing.ResolvedAt = &now }
	existing.UpdatedAt = now
	if err := sv.store.Update(r.Context(), existing); err != nil { writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()}); return }
	sv.cache.invalidate("list:" + tenantID); sv.cache.invalidate("get:" + tenantID + ":" + id)
	writeJSON(w, http.StatusOK, map[string]any{"item": existing, "event_topic": eventTopic + ".updated"})
}

func (sv *server) handleDelete(w http.ResponseWriter, r *http.Request, id string) {
	tenantID := r.Header.Get("X-Tenant-ID")
	if err := sv.store.Delete(r.Context(), tenantID, id); err != nil { if err.Error() == "not found" { writeJSON(w, http.StatusNotFound, map[string]string{"error": "not found"}); return }; writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()}); return }
	sv.cache.invalidate("list:" + tenantID); sv.cache.invalidate("get:" + tenantID + ":" + id)
	writeJSON(w, http.StatusOK, map[string]any{"deleted": true, "id": id, "event_topic": eventTopic + ".deleted"})
}

func handleExplain(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]any{"service": serviceName, "module": moduleName, "base_path": basePath, "database": dbName, "table": tableName, "event_topic": eventTopic, "entity": "anomaly",
		"fields": []string{"id", "tenant_id", "metric_name", "metric_value", "expected_value", "deviation_percent", "classification", "severity", "status", "service_name", "environment", "source", "description", "detected_at", "resolved_at", "correlation_id", "tags", "created_at", "updated_at"},
		"filters": []string{"severity", "status", "classification", "metric_name"},
	})
}

func main() {
	port := os.Getenv("PORT"); if port == "" { port = "8080" }
	var st store; dsn := os.Getenv("DATABASE_URL")
	if dsn != "" { pg, err := newPostgresStore(dsn); if err != nil { log.Fatalf("postgres: %v", err) }; st = pg; log.Println("Using PostgreSQL store") } else { st = newMemoryStore(); log.Println("Using in-memory store (set DATABASE_URL for PostgreSQL)") }
	cache := newCache(); srv := &server{store: st, cache: cache}; mux := http.NewServeMux()
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) { writeJSON(w, http.StatusOK, map[string]string{"status": "healthy", "module": moduleName, "service": serviceName}) })
	mux.HandleFunc("/readyz", func(w http.ResponseWriter, r *http.Request) { writeJSON(w, http.StatusOK, map[string]string{"status": "ready"}) })
	mux.HandleFunc("/metrics", func(w http.ResponseWriter, r *http.Request) { writeJSON(w, http.StatusOK, map[string]any{"requests_total": requestCount.Load(), "service": serviceName}) })
	mux.HandleFunc(basePath+"/_explain", handleExplain)
	mux.HandleFunc(basePath, func(w http.ResponseWriter, r *http.Request) {
		requestCount.Add(1); tenantID := r.Header.Get("X-Tenant-ID")
		if tenantID == "" { writeJSON(w, http.StatusBadRequest, map[string]string{"error": "missing X-Tenant-ID header"}); return }
		switch r.Method { case http.MethodGet: srv.handleList(w, r); case http.MethodPost: srv.handleCreate(w, r); default: writeJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "method not allowed"}) }
	})
	mux.HandleFunc(basePath+"/", func(w http.ResponseWriter, r *http.Request) {
		requestCount.Add(1); tenantID := r.Header.Get("X-Tenant-ID")
		if tenantID == "" { writeJSON(w, http.StatusBadRequest, map[string]string{"error": "missing X-Tenant-ID header"}); return }
		id := strings.TrimPrefix(r.URL.Path, basePath+"/"); if id == "" { writeJSON(w, http.StatusBadRequest, map[string]string{"error": "missing id"}); return }
		switch r.Method { case http.MethodGet: srv.handleGet(w, r, id); case http.MethodPut, http.MethodPatch: srv.handleUpdate(w, r, id); case http.MethodDelete: srv.handleDelete(w, r, id); default: writeJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "method not allowed"}) }
	})
	handler := securityHeaders(mux); log.Printf("%s listening on :%s", serviceName, port); log.Fatal(http.ListenAndServe(":"+port, handler))
}
