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
	serviceName = "slo-service"
	moduleName  = "ERP-AIOps"
	basePath    = "/v1/aiops/slos"
	dbName      = "erp_aiops"
	tableName   = "aiops_slos"
	eventTopic  = "erp.aiops.slo"
	cacheTTL    = 30 * time.Second
)

var validStatuses = map[string]bool{"healthy": true, "at_risk": true, "breached": true, "inactive": true}

type slo struct {
	ID               string  `json:"id"`
	TenantID         string  `json:"tenant_id"`
	Name             string  `json:"name"`
	Description      *string `json:"description,omitempty"`
	ServiceName      string  `json:"service_name"`
	SLIMetric        string  `json:"sli_metric"`
	TargetPercent    float64 `json:"target_percent"`
	CurrentPercent   float64 `json:"current_percent"`
	ErrorBudgetTotal float64 `json:"error_budget_total"`
	ErrorBudgetUsed  float64 `json:"error_budget_used"`
	ErrorBudgetRemaining float64 `json:"error_budget_remaining"`
	BurnRate         float64 `json:"burn_rate"`
	WindowDays       int     `json:"window_days"`
	Status           string  `json:"status"`
	Environment      *string `json:"environment,omitempty"`
	OwnerID          *string `json:"owner_id,omitempty"`
	AlertChannel     *string `json:"alert_channel,omitempty"`
	Tags             *string `json:"tags,omitempty"`
	CreatedAt        string  `json:"created_at"`
	UpdatedAt        string  `json:"updated_at"`
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
	List(ctx context.Context, tenantID, cursor string, limit int, filters map[string]string) ([]slo, string, error)
	GetByID(ctx context.Context, tenantID, id string) (*slo, error)
	Create(ctx context.Context, s *slo) error
	Update(ctx context.Context, s *slo) error
	Delete(ctx context.Context, tenantID, id string) error
}

type memoryStore struct { mu sync.RWMutex; records map[string]slo }
func newMemoryStore() *memoryStore { return &memoryStore{records: make(map[string]slo)} }

func (m *memoryStore) List(_ context.Context, tenantID, cursor string, limit int, filters map[string]string) ([]slo, string, error) {
	m.mu.RLock(); defer m.mu.RUnlock()
	var all []slo
	for _, s := range m.records {
		if s.TenantID != tenantID { continue }
		if v, ok := filters["status"]; ok && s.Status != v { continue }
		if v, ok := filters["service_name"]; ok && s.ServiceName != v { continue }
		if v, ok := filters["owner_id"]; ok && strVal(s.OwnerID) != v { continue }
		all = append(all, s)
	}
	sort.Slice(all, func(i, j int) bool { return all[i].CreatedAt < all[j].CreatedAt })
	start := 0; if cursor != "" { for idx, s := range all { if s.ID == cursor { start = idx + 1; break } } }
	if start >= len(all) { return []slo{}, "", nil }
	end := start + limit; if end > len(all) { end = len(all) }
	result := all[start:end]; nextCursor := ""; if end < len(all) { nextCursor = result[len(result)-1].ID }
	return result, nextCursor, nil
}

func (m *memoryStore) GetByID(_ context.Context, tenantID, id string) (*slo, error) { m.mu.RLock(); defer m.mu.RUnlock(); s, ok := m.records[id]; if !ok || s.TenantID != tenantID { return nil, errors.New("not found") }; return &s, nil }
func (m *memoryStore) Create(_ context.Context, s *slo) error { m.mu.Lock(); defer m.mu.Unlock(); m.records[s.ID] = *s; return nil }
func (m *memoryStore) Update(_ context.Context, s *slo) error { m.mu.Lock(); defer m.mu.Unlock(); if _, ok := m.records[s.ID]; !ok { return errors.New("not found") }; m.records[s.ID] = *s; return nil }
func (m *memoryStore) Delete(_ context.Context, tenantID, id string) error { m.mu.Lock(); defer m.mu.Unlock(); s, ok := m.records[id]; if !ok || s.TenantID != tenantID { return errors.New("not found") }; delete(m.records, id); return nil }

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
	id TEXT PRIMARY KEY, tenant_id TEXT NOT NULL, name TEXT NOT NULL, description TEXT,
	service_name TEXT NOT NULL, sli_metric TEXT NOT NULL,
	target_percent DOUBLE PRECISION DEFAULT 99.9, current_percent DOUBLE PRECISION DEFAULT 100,
	error_budget_total DOUBLE PRECISION DEFAULT 0, error_budget_used DOUBLE PRECISION DEFAULT 0,
	error_budget_remaining DOUBLE PRECISION DEFAULT 0, burn_rate DOUBLE PRECISION DEFAULT 0,
	window_days INT DEFAULT 30,
	status TEXT CHECK (status IN ('healthy','at_risk','breached','inactive')) DEFAULT 'healthy',
	environment TEXT, owner_id TEXT, alert_channel TEXT, tags TEXT,
	created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(), updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_` + tableName + `_tenant ON ` + tableName + ` (tenant_id);
CREATE INDEX IF NOT EXISTS idx_` + tableName + `_status ON ` + tableName + ` (tenant_id, status);`

func (p *postgresStore) List(ctx context.Context, tenantID, cursor string, limit int, filters map[string]string) ([]slo, string, error) {
	query := `SELECT id,tenant_id,name,description,service_name,sli_metric,target_percent,current_percent,error_budget_total,error_budget_used,error_budget_remaining,burn_rate,window_days,status,environment,owner_id,alert_channel,tags,created_at,updated_at FROM ` + tableName + ` WHERE tenant_id=$1`
	args := []any{tenantID}; idx := 2
	for _, f := range []struct{ k, c string }{{"status", "status"}, {"service_name", "service_name"}, {"owner_id", "owner_id"}} {
		if v, ok := filters[f.k]; ok { query += fmt.Sprintf(" AND %s=$%d", f.c, idx); args = append(args, v); idx++ }
	}
	if cursor != "" { query += fmt.Sprintf(" AND created_at>(SELECT created_at FROM "+tableName+" WHERE id=$%d)", idx); args = append(args, cursor); idx++ }
	query += " ORDER BY created_at ASC" + fmt.Sprintf(" LIMIT $%d", idx); args = append(args, limit+1)
	rows, err := p.db.QueryContext(ctx, query, args...); if err != nil { return nil, "", fmt.Errorf("query: %w", err) }; defer rows.Close()
	var results []slo
	for rows.Next() {
		var s slo; var desc, env, ownerID, alertCh, tags sql.NullString; var createdAt, updatedAt sql.NullTime
		if err := rows.Scan(&s.ID, &s.TenantID, &s.Name, &desc, &s.ServiceName, &s.SLIMetric, &s.TargetPercent, &s.CurrentPercent, &s.ErrorBudgetTotal, &s.ErrorBudgetUsed, &s.ErrorBudgetRemaining, &s.BurnRate, &s.WindowDays, &s.Status, &env, &ownerID, &alertCh, &tags, &createdAt, &updatedAt); err != nil { return nil, "", fmt.Errorf("scan: %w", err) }
		if desc.Valid { s.Description = &desc.String }; if env.Valid { s.Environment = &env.String }
		if ownerID.Valid { s.OwnerID = &ownerID.String }; if alertCh.Valid { s.AlertChannel = &alertCh.String }
		if tags.Valid { s.Tags = &tags.String }
		if createdAt.Valid { s.CreatedAt = createdAt.Time.Format(time.RFC3339) }; if updatedAt.Valid { s.UpdatedAt = updatedAt.Time.Format(time.RFC3339) }
		results = append(results, s)
	}
	nextCursor := ""; if len(results) > limit { nextCursor = results[limit-1].ID; results = results[:limit] }
	return results, nextCursor, nil
}

func (p *postgresStore) GetByID(ctx context.Context, tenantID, id string) (*slo, error) {
	query := `SELECT id,tenant_id,name,description,service_name,sli_metric,target_percent,current_percent,error_budget_total,error_budget_used,error_budget_remaining,burn_rate,window_days,status,environment,owner_id,alert_channel,tags,created_at,updated_at FROM ` + tableName + ` WHERE id=$1 AND tenant_id=$2`
	var s slo; var desc, env, ownerID, alertCh, tags sql.NullString; var createdAt, updatedAt sql.NullTime
	err := p.db.QueryRowContext(ctx, query, id, tenantID).Scan(&s.ID, &s.TenantID, &s.Name, &desc, &s.ServiceName, &s.SLIMetric, &s.TargetPercent, &s.CurrentPercent, &s.ErrorBudgetTotal, &s.ErrorBudgetUsed, &s.ErrorBudgetRemaining, &s.BurnRate, &s.WindowDays, &s.Status, &env, &ownerID, &alertCh, &tags, &createdAt, &updatedAt)
	if err != nil { if errors.Is(err, sql.ErrNoRows) { return nil, errors.New("not found") }; return nil, fmt.Errorf("query row: %w", err) }
	if desc.Valid { s.Description = &desc.String }; if env.Valid { s.Environment = &env.String }
	if ownerID.Valid { s.OwnerID = &ownerID.String }; if alertCh.Valid { s.AlertChannel = &alertCh.String }
	if tags.Valid { s.Tags = &tags.String }
	if createdAt.Valid { s.CreatedAt = createdAt.Time.Format(time.RFC3339) }; if updatedAt.Valid { s.UpdatedAt = updatedAt.Time.Format(time.RFC3339) }
	return &s, nil
}

func (p *postgresStore) Create(ctx context.Context, s *slo) error {
	query := `INSERT INTO ` + tableName + ` (id,tenant_id,name,description,service_name,sli_metric,target_percent,current_percent,error_budget_total,error_budget_used,error_budget_remaining,burn_rate,window_days,status,environment,owner_id,alert_channel,tags,created_at,updated_at) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17,$18,$19,$20)`
	_, err := p.db.ExecContext(ctx, query, s.ID, s.TenantID, s.Name, s.Description, s.ServiceName, s.SLIMetric, s.TargetPercent, s.CurrentPercent, s.ErrorBudgetTotal, s.ErrorBudgetUsed, s.ErrorBudgetRemaining, s.BurnRate, s.WindowDays, s.Status, s.Environment, s.OwnerID, s.AlertChannel, s.Tags, parseTime(s.CreatedAt), parseTime(s.UpdatedAt))
	return err
}

func (p *postgresStore) Update(ctx context.Context, s *slo) error {
	query := `UPDATE ` + tableName + ` SET name=$1,description=$2,service_name=$3,sli_metric=$4,target_percent=$5,current_percent=$6,error_budget_total=$7,error_budget_used=$8,error_budget_remaining=$9,burn_rate=$10,window_days=$11,status=$12,environment=$13,owner_id=$14,alert_channel=$15,tags=$16,updated_at=$17 WHERE id=$18 AND tenant_id=$19`
	res, err := p.db.ExecContext(ctx, query, s.Name, s.Description, s.ServiceName, s.SLIMetric, s.TargetPercent, s.CurrentPercent, s.ErrorBudgetTotal, s.ErrorBudgetUsed, s.ErrorBudgetRemaining, s.BurnRate, s.WindowDays, s.Status, s.Environment, s.OwnerID, s.AlertChannel, s.Tags, parseTime(s.UpdatedAt), s.ID, s.TenantID)
	if err != nil { return err }; n, _ := res.RowsAffected(); if n == 0 { return errors.New("not found") }; return nil
}

func (p *postgresStore) Delete(ctx context.Context, tenantID, id string) error {
	res, err := p.db.ExecContext(ctx, `DELETE FROM `+tableName+` WHERE id=$1 AND tenant_id=$2`, id, tenantID)
	if err != nil { return err }; n, _ := res.RowsAffected(); if n == 0 { return errors.New("not found") }; return nil
}

func parseTime(s string) time.Time { t, err := time.Parse(time.RFC3339, s); if err != nil { return time.Now() }; return t }

type server struct { store store; cache *ttlCache }

func (sv *server) handleList(w http.ResponseWriter, r *http.Request) {
	tenantID := r.Header.Get("X-Tenant-ID"); cursor := r.URL.Query().Get("cursor")
	limit := 20; if v, err := strconv.Atoi(r.URL.Query().Get("limit")); err == nil && v > 0 && v <= 100 { limit = v }
	filters := make(map[string]string); for _, key := range []string{"status", "service_name", "owner_id"} { if v := r.URL.Query().Get(key); v != "" { filters[key] = v } }
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
	name, _ := body["name"].(string); svcName, _ := body["service_name"].(string); sliMetric, _ := body["sli_metric"].(string)
	if name == "" || svcName == "" || sliMetric == "" { writeJSON(w, http.StatusBadRequest, map[string]string{"error": "name, service_name, and sli_metric are required"}); return }
	now := time.Now().UTC().Format(time.RFC3339)
	s := &slo{ID: newID(), TenantID: tenantID, Name: name, Description: strPtr(body["description"]), ServiceName: svcName, SLIMetric: sliMetric, TargetPercent: 99.9, CurrentPercent: 100, ErrorBudgetTotal: 0.1, ErrorBudgetUsed: 0, ErrorBudgetRemaining: 0.1, BurnRate: 0, WindowDays: 30, Status: "healthy", Environment: strPtr(body["environment"]), OwnerID: strPtr(body["owner_id"]), AlertChannel: strPtr(body["alert_channel"]), Tags: strPtr(body["tags"]), CreatedAt: now, UpdatedAt: now}
	if v, ok := body["target_percent"].(float64); ok { s.TargetPercent = v; s.ErrorBudgetTotal = 100 - v; s.ErrorBudgetRemaining = s.ErrorBudgetTotal }
	if v, ok := body["window_days"].(float64); ok { s.WindowDays = int(v) }
	if err := sv.store.Create(r.Context(), s); err != nil { writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()}); return }
	sv.cache.invalidate("list:" + tenantID); writeJSON(w, http.StatusCreated, map[string]any{"item": s, "event_topic": eventTopic + ".created"})
}

func (sv *server) handleUpdate(w http.ResponseWriter, r *http.Request, id string) {
	tenantID := r.Header.Get("X-Tenant-ID")
	existing, err := sv.store.GetByID(r.Context(), tenantID, id)
	if err != nil { if err.Error() == "not found" { writeJSON(w, http.StatusNotFound, map[string]string{"error": "not found"}); return }; writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()}); return }
	body, err := readJSON(r)
	if err != nil { writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid JSON: " + err.Error()}); return }
	if v, ok := body["name"].(string); ok && v != "" { existing.Name = v }
	if v, exists := body["description"]; exists { existing.Description = strPtr(v) }
	if v, ok := body["service_name"].(string); ok && v != "" { existing.ServiceName = v }
	if v, ok := body["sli_metric"].(string); ok && v != "" { existing.SLIMetric = v }
	if v, ok := body["target_percent"].(float64); ok { existing.TargetPercent = v }
	if v, ok := body["current_percent"].(float64); ok { existing.CurrentPercent = v }
	if v, ok := body["error_budget_used"].(float64); ok { existing.ErrorBudgetUsed = v }
	if v, ok := body["burn_rate"].(float64); ok { existing.BurnRate = v }
	if v, ok := body["window_days"].(float64); ok { existing.WindowDays = int(v) }
	if v, exists := body["environment"]; exists { existing.Environment = strPtr(v) }
	if v, exists := body["owner_id"]; exists { existing.OwnerID = strPtr(v) }
	if v, exists := body["alert_channel"]; exists { existing.AlertChannel = strPtr(v) }
	if v, exists := body["tags"]; exists { existing.Tags = strPtr(v) }
	existing.ErrorBudgetTotal = 100 - existing.TargetPercent
	existing.ErrorBudgetRemaining = existing.ErrorBudgetTotal - existing.ErrorBudgetUsed
	if existing.ErrorBudgetRemaining < 0 { existing.ErrorBudgetRemaining = 0 }
	if existing.CurrentPercent < existing.TargetPercent { existing.Status = "breached" } else if existing.BurnRate > 1 { existing.Status = "at_risk" } else { existing.Status = "healthy" }
	if v, ok := body["status"].(string); ok && validStatuses[v] { existing.Status = v }
	existing.UpdatedAt = time.Now().UTC().Format(time.RFC3339)
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
	writeJSON(w, http.StatusOK, map[string]any{"service": serviceName, "module": moduleName, "base_path": basePath, "database": dbName, "table": tableName, "event_topic": eventTopic, "entity": "slo",
		"fields": []string{"id", "tenant_id", "name", "description", "service_name", "sli_metric", "target_percent", "current_percent", "error_budget_total", "error_budget_used", "error_budget_remaining", "burn_rate", "window_days", "status", "environment", "owner_id", "alert_channel", "tags", "created_at", "updated_at"},
		"filters": []string{"status", "service_name", "owner_id"},
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
