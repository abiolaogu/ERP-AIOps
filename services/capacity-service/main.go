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
	serviceName = "capacity-service"
	moduleName  = "ERP-AIOps"
	basePath    = "/v1/aiops/capacity"
	dbName      = "erp_aiops"
	tableName   = "aiops_capacity_plans"
	eventTopic  = "erp.aiops.capacity"
	cacheTTL    = 30 * time.Second
)

var validStatuses = map[string]bool{"active": true, "warning": true, "critical": true, "archived": true}

type capacityPlan struct {
	ID               string  `json:"id"`
	TenantID         string  `json:"tenant_id"`
	ResourceName     string  `json:"resource_name"`
	ResourceType     string  `json:"resource_type"`
	CurrentUsage     float64 `json:"current_usage"`
	MaxCapacity      float64 `json:"max_capacity"`
	UsagePercent     float64 `json:"usage_percent"`
	WarningThreshold float64 `json:"warning_threshold"`
	CriticalThreshold float64 `json:"critical_threshold"`
	ForecastDays     int     `json:"forecast_days"`
	ForecastUsage    float64 `json:"forecast_usage"`
	Unit             string  `json:"unit"`
	Region           *string `json:"region,omitempty"`
	Environment      *string `json:"environment,omitempty"`
	ServiceName      *string `json:"service_name,omitempty"`
	Status           string  `json:"status"`
	Notes            *string `json:"notes,omitempty"`
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
	List(ctx context.Context, tenantID, cursor string, limit int, filters map[string]string) ([]capacityPlan, string, error)
	GetByID(ctx context.Context, tenantID, id string) (*capacityPlan, error)
	Create(ctx context.Context, cp *capacityPlan) error
	Update(ctx context.Context, cp *capacityPlan) error
	Delete(ctx context.Context, tenantID, id string) error
}

type memoryStore struct { mu sync.RWMutex; records map[string]capacityPlan }
func newMemoryStore() *memoryStore { return &memoryStore{records: make(map[string]capacityPlan)} }

func (m *memoryStore) List(_ context.Context, tenantID, cursor string, limit int, filters map[string]string) ([]capacityPlan, string, error) {
	m.mu.RLock(); defer m.mu.RUnlock()
	var all []capacityPlan
	for _, cp := range m.records {
		if cp.TenantID != tenantID { continue }
		if v, ok := filters["status"]; ok && cp.Status != v { continue }
		if v, ok := filters["resource_type"]; ok && cp.ResourceType != v { continue }
		if v, ok := filters["environment"]; ok && strVal(cp.Environment) != v { continue }
		all = append(all, cp)
	}
	sort.Slice(all, func(i, j int) bool { return all[i].CreatedAt < all[j].CreatedAt })
	start := 0; if cursor != "" { for idx, cp := range all { if cp.ID == cursor { start = idx + 1; break } } }
	if start >= len(all) { return []capacityPlan{}, "", nil }
	end := start + limit; if end > len(all) { end = len(all) }
	result := all[start:end]; nextCursor := ""; if end < len(all) { nextCursor = result[len(result)-1].ID }
	return result, nextCursor, nil
}

func (m *memoryStore) GetByID(_ context.Context, tenantID, id string) (*capacityPlan, error) { m.mu.RLock(); defer m.mu.RUnlock(); cp, ok := m.records[id]; if !ok || cp.TenantID != tenantID { return nil, errors.New("not found") }; return &cp, nil }
func (m *memoryStore) Create(_ context.Context, cp *capacityPlan) error { m.mu.Lock(); defer m.mu.Unlock(); m.records[cp.ID] = *cp; return nil }
func (m *memoryStore) Update(_ context.Context, cp *capacityPlan) error { m.mu.Lock(); defer m.mu.Unlock(); if _, ok := m.records[cp.ID]; !ok { return errors.New("not found") }; m.records[cp.ID] = *cp; return nil }
func (m *memoryStore) Delete(_ context.Context, tenantID, id string) error { m.mu.Lock(); defer m.mu.Unlock(); cp, ok := m.records[id]; if !ok || cp.TenantID != tenantID { return errors.New("not found") }; delete(m.records, id); return nil }

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
	id TEXT PRIMARY KEY, tenant_id TEXT NOT NULL, resource_name TEXT NOT NULL, resource_type TEXT NOT NULL,
	current_usage DOUBLE PRECISION DEFAULT 0, max_capacity DOUBLE PRECISION DEFAULT 100,
	usage_percent DOUBLE PRECISION DEFAULT 0, warning_threshold DOUBLE PRECISION DEFAULT 75,
	critical_threshold DOUBLE PRECISION DEFAULT 90, forecast_days INT DEFAULT 30,
	forecast_usage DOUBLE PRECISION DEFAULT 0, unit TEXT DEFAULT 'percent',
	region TEXT, environment TEXT, service_name TEXT,
	status TEXT CHECK (status IN ('active','warning','critical','archived')) DEFAULT 'active',
	notes TEXT, tags TEXT,
	created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(), updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_` + tableName + `_tenant ON ` + tableName + ` (tenant_id);
CREATE INDEX IF NOT EXISTS idx_` + tableName + `_status ON ` + tableName + ` (tenant_id, status);`

func (p *postgresStore) List(ctx context.Context, tenantID, cursor string, limit int, filters map[string]string) ([]capacityPlan, string, error) {
	query := `SELECT id,tenant_id,resource_name,resource_type,current_usage,max_capacity,usage_percent,warning_threshold,critical_threshold,forecast_days,forecast_usage,unit,region,environment,service_name,status,notes,tags,created_at,updated_at FROM ` + tableName + ` WHERE tenant_id=$1`
	args := []any{tenantID}; idx := 2
	for _, f := range []struct{ k, c string }{{"status", "status"}, {"resource_type", "resource_type"}, {"environment", "environment"}} {
		if v, ok := filters[f.k]; ok { query += fmt.Sprintf(" AND %s=$%d", f.c, idx); args = append(args, v); idx++ }
	}
	if cursor != "" { query += fmt.Sprintf(" AND created_at>(SELECT created_at FROM "+tableName+" WHERE id=$%d)", idx); args = append(args, cursor); idx++ }
	query += " ORDER BY created_at ASC" + fmt.Sprintf(" LIMIT $%d", idx); args = append(args, limit+1)
	rows, err := p.db.QueryContext(ctx, query, args...); if err != nil { return nil, "", fmt.Errorf("query: %w", err) }; defer rows.Close()
	var results []capacityPlan
	for rows.Next() {
		var cp capacityPlan; var region, env, svcName, notes, tags sql.NullString; var createdAt, updatedAt sql.NullTime
		if err := rows.Scan(&cp.ID, &cp.TenantID, &cp.ResourceName, &cp.ResourceType, &cp.CurrentUsage, &cp.MaxCapacity, &cp.UsagePercent, &cp.WarningThreshold, &cp.CriticalThreshold, &cp.ForecastDays, &cp.ForecastUsage, &cp.Unit, &region, &env, &svcName, &cp.Status, &notes, &tags, &createdAt, &updatedAt); err != nil { return nil, "", fmt.Errorf("scan: %w", err) }
		if region.Valid { cp.Region = &region.String }; if env.Valid { cp.Environment = &env.String }
		if svcName.Valid { cp.ServiceName = &svcName.String }; if notes.Valid { cp.Notes = &notes.String }
		if tags.Valid { cp.Tags = &tags.String }
		if createdAt.Valid { cp.CreatedAt = createdAt.Time.Format(time.RFC3339) }; if updatedAt.Valid { cp.UpdatedAt = updatedAt.Time.Format(time.RFC3339) }
		results = append(results, cp)
	}
	nextCursor := ""; if len(results) > limit { nextCursor = results[limit-1].ID; results = results[:limit] }
	return results, nextCursor, nil
}

func (p *postgresStore) GetByID(ctx context.Context, tenantID, id string) (*capacityPlan, error) {
	query := `SELECT id,tenant_id,resource_name,resource_type,current_usage,max_capacity,usage_percent,warning_threshold,critical_threshold,forecast_days,forecast_usage,unit,region,environment,service_name,status,notes,tags,created_at,updated_at FROM ` + tableName + ` WHERE id=$1 AND tenant_id=$2`
	var cp capacityPlan; var region, env, svcName, notes, tags sql.NullString; var createdAt, updatedAt sql.NullTime
	err := p.db.QueryRowContext(ctx, query, id, tenantID).Scan(&cp.ID, &cp.TenantID, &cp.ResourceName, &cp.ResourceType, &cp.CurrentUsage, &cp.MaxCapacity, &cp.UsagePercent, &cp.WarningThreshold, &cp.CriticalThreshold, &cp.ForecastDays, &cp.ForecastUsage, &cp.Unit, &region, &env, &svcName, &cp.Status, &notes, &tags, &createdAt, &updatedAt)
	if err != nil { if errors.Is(err, sql.ErrNoRows) { return nil, errors.New("not found") }; return nil, fmt.Errorf("query row: %w", err) }
	if region.Valid { cp.Region = &region.String }; if env.Valid { cp.Environment = &env.String }
	if svcName.Valid { cp.ServiceName = &svcName.String }; if notes.Valid { cp.Notes = &notes.String }
	if tags.Valid { cp.Tags = &tags.String }
	if createdAt.Valid { cp.CreatedAt = createdAt.Time.Format(time.RFC3339) }; if updatedAt.Valid { cp.UpdatedAt = updatedAt.Time.Format(time.RFC3339) }
	return &cp, nil
}

func (p *postgresStore) Create(ctx context.Context, cp *capacityPlan) error {
	query := `INSERT INTO ` + tableName + ` (id,tenant_id,resource_name,resource_type,current_usage,max_capacity,usage_percent,warning_threshold,critical_threshold,forecast_days,forecast_usage,unit,region,environment,service_name,status,notes,tags,created_at,updated_at) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17,$18,$19,$20)`
	_, err := p.db.ExecContext(ctx, query, cp.ID, cp.TenantID, cp.ResourceName, cp.ResourceType, cp.CurrentUsage, cp.MaxCapacity, cp.UsagePercent, cp.WarningThreshold, cp.CriticalThreshold, cp.ForecastDays, cp.ForecastUsage, cp.Unit, cp.Region, cp.Environment, cp.ServiceName, cp.Status, cp.Notes, cp.Tags, parseTime(cp.CreatedAt), parseTime(cp.UpdatedAt))
	return err
}

func (p *postgresStore) Update(ctx context.Context, cp *capacityPlan) error {
	query := `UPDATE ` + tableName + ` SET resource_name=$1,resource_type=$2,current_usage=$3,max_capacity=$4,usage_percent=$5,warning_threshold=$6,critical_threshold=$7,forecast_days=$8,forecast_usage=$9,unit=$10,region=$11,environment=$12,service_name=$13,status=$14,notes=$15,tags=$16,updated_at=$17 WHERE id=$18 AND tenant_id=$19`
	res, err := p.db.ExecContext(ctx, query, cp.ResourceName, cp.ResourceType, cp.CurrentUsage, cp.MaxCapacity, cp.UsagePercent, cp.WarningThreshold, cp.CriticalThreshold, cp.ForecastDays, cp.ForecastUsage, cp.Unit, cp.Region, cp.Environment, cp.ServiceName, cp.Status, cp.Notes, cp.Tags, parseTime(cp.UpdatedAt), cp.ID, cp.TenantID)
	if err != nil { return err }; n, _ := res.RowsAffected(); if n == 0 { return errors.New("not found") }; return nil
}

func (p *postgresStore) Delete(ctx context.Context, tenantID, id string) error {
	res, err := p.db.ExecContext(ctx, `DELETE FROM `+tableName+` WHERE id=$1 AND tenant_id=$2`, id, tenantID)
	if err != nil { return err }; n, _ := res.RowsAffected(); if n == 0 { return errors.New("not found") }; return nil
}

func parseTime(s string) time.Time { t, err := time.Parse(time.RFC3339, s); if err != nil { return time.Now() }; return t }
func parseTimePtr(s *string) *time.Time { if s == nil { return nil }; t, err := time.Parse(time.RFC3339, *s); if err != nil { return nil }; return &t }

type server struct { store store; cache *ttlCache }

func (s *server) handleList(w http.ResponseWriter, r *http.Request) {
	tenantID := r.Header.Get("X-Tenant-ID"); cursor := r.URL.Query().Get("cursor")
	limit := 20; if v, err := strconv.Atoi(r.URL.Query().Get("limit")); err == nil && v > 0 && v <= 100 { limit = v }
	filters := make(map[string]string); for _, key := range []string{"status", "resource_type", "environment"} { if v := r.URL.Query().Get(key); v != "" { filters[key] = v } }
	cacheKey := fmt.Sprintf("list:%s:%s:%d:%v", tenantID, cursor, limit, filters)
	if cached, ok := s.cache.get(cacheKey); ok { writeJSON(w, http.StatusOK, cached); return }
	items, nextCursor, err := s.store.List(r.Context(), tenantID, cursor, limit, filters)
	if err != nil { writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()}); return }
	resp := map[string]any{"items": items, "next_cursor": nextCursor, "limit": limit, "count": len(items), "event_topic": eventTopic + ".listed"}
	s.cache.set(cacheKey, resp); writeJSON(w, http.StatusOK, resp)
}

func (s *server) handleGet(w http.ResponseWriter, r *http.Request, id string) {
	tenantID := r.Header.Get("X-Tenant-ID"); cacheKey := fmt.Sprintf("get:%s:%s", tenantID, id)
	if cached, ok := s.cache.get(cacheKey); ok { writeJSON(w, http.StatusOK, cached); return }
	item, err := s.store.GetByID(r.Context(), tenantID, id)
	if err != nil { if err.Error() == "not found" { writeJSON(w, http.StatusNotFound, map[string]string{"error": "not found"}); return }; writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()}); return }
	resp := map[string]any{"item": item, "event_topic": eventTopic + ".read"}; s.cache.set(cacheKey, resp); writeJSON(w, http.StatusOK, resp)
}

func (s *server) handleCreate(w http.ResponseWriter, r *http.Request) {
	tenantID := r.Header.Get("X-Tenant-ID"); body, err := readJSON(r)
	if err != nil { writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid JSON: " + err.Error()}); return }
	resName, _ := body["resource_name"].(string); resType, _ := body["resource_type"].(string)
	if resName == "" || resType == "" { writeJSON(w, http.StatusBadRequest, map[string]string{"error": "resource_name and resource_type are required"}); return }
	now := time.Now().UTC().Format(time.RFC3339)
	cp := &capacityPlan{ID: newID(), TenantID: tenantID, ResourceName: resName, ResourceType: resType, CurrentUsage: 0, MaxCapacity: 100, UsagePercent: 0, WarningThreshold: 75, CriticalThreshold: 90, ForecastDays: 30, ForecastUsage: 0, Unit: "percent", Region: strPtr(body["region"]), Environment: strPtr(body["environment"]), ServiceName: strPtr(body["service_name"]), Status: "active", Notes: strPtr(body["notes"]), Tags: strPtr(body["tags"]), CreatedAt: now, UpdatedAt: now}
	if v, ok := body["current_usage"].(float64); ok { cp.CurrentUsage = v }
	if v, ok := body["max_capacity"].(float64); ok { cp.MaxCapacity = v }
	if v, ok := body["warning_threshold"].(float64); ok { cp.WarningThreshold = v }
	if v, ok := body["critical_threshold"].(float64); ok { cp.CriticalThreshold = v }
	if v, ok := body["forecast_days"].(float64); ok { cp.ForecastDays = int(v) }
	if v, ok := body["unit"].(string); ok && v != "" { cp.Unit = v }
	if cp.MaxCapacity > 0 { cp.UsagePercent = (cp.CurrentUsage / cp.MaxCapacity) * 100 }
	if cp.UsagePercent >= cp.CriticalThreshold { cp.Status = "critical" } else if cp.UsagePercent >= cp.WarningThreshold { cp.Status = "warning" }
	if err := s.store.Create(r.Context(), cp); err != nil { writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()}); return }
	s.cache.invalidate("list:" + tenantID); writeJSON(w, http.StatusCreated, map[string]any{"item": cp, "event_topic": eventTopic + ".created"})
}

func (s *server) handleUpdate(w http.ResponseWriter, r *http.Request, id string) {
	tenantID := r.Header.Get("X-Tenant-ID")
	existing, err := s.store.GetByID(r.Context(), tenantID, id)
	if err != nil { if err.Error() == "not found" { writeJSON(w, http.StatusNotFound, map[string]string{"error": "not found"}); return }; writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()}); return }
	body, err := readJSON(r)
	if err != nil { writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid JSON: " + err.Error()}); return }
	if v, ok := body["resource_name"].(string); ok && v != "" { existing.ResourceName = v }
	if v, ok := body["resource_type"].(string); ok && v != "" { existing.ResourceType = v }
	if v, ok := body["current_usage"].(float64); ok { existing.CurrentUsage = v }
	if v, ok := body["max_capacity"].(float64); ok { existing.MaxCapacity = v }
	if v, ok := body["warning_threshold"].(float64); ok { existing.WarningThreshold = v }
	if v, ok := body["critical_threshold"].(float64); ok { existing.CriticalThreshold = v }
	if v, ok := body["forecast_days"].(float64); ok { existing.ForecastDays = int(v) }
	if v, ok := body["forecast_usage"].(float64); ok { existing.ForecastUsage = v }
	if v, ok := body["unit"].(string); ok && v != "" { existing.Unit = v }
	if v, exists := body["region"]; exists { existing.Region = strPtr(v) }
	if v, exists := body["environment"]; exists { existing.Environment = strPtr(v) }
	if v, exists := body["service_name"]; exists { existing.ServiceName = strPtr(v) }
	if v, exists := body["notes"]; exists { existing.Notes = strPtr(v) }
	if v, exists := body["tags"]; exists { existing.Tags = strPtr(v) }
	if existing.MaxCapacity > 0 { existing.UsagePercent = (existing.CurrentUsage / existing.MaxCapacity) * 100 }
	if existing.UsagePercent >= existing.CriticalThreshold { existing.Status = "critical" } else if existing.UsagePercent >= existing.WarningThreshold { existing.Status = "warning" } else { existing.Status = "active" }
	existing.UpdatedAt = time.Now().UTC().Format(time.RFC3339)
	if err := s.store.Update(r.Context(), existing); err != nil { writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()}); return }
	s.cache.invalidate("list:" + tenantID); s.cache.invalidate("get:" + tenantID + ":" + id)
	writeJSON(w, http.StatusOK, map[string]any{"item": existing, "event_topic": eventTopic + ".updated"})
}

func (s *server) handleDelete(w http.ResponseWriter, r *http.Request, id string) {
	tenantID := r.Header.Get("X-Tenant-ID")
	if err := s.store.Delete(r.Context(), tenantID, id); err != nil { if err.Error() == "not found" { writeJSON(w, http.StatusNotFound, map[string]string{"error": "not found"}); return }; writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()}); return }
	s.cache.invalidate("list:" + tenantID); s.cache.invalidate("get:" + tenantID + ":" + id)
	writeJSON(w, http.StatusOK, map[string]any{"deleted": true, "id": id, "event_topic": eventTopic + ".deleted"})
}

func handleExplain(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]any{"service": serviceName, "module": moduleName, "base_path": basePath, "database": dbName, "table": tableName, "event_topic": eventTopic, "entity": "capacityPlan",
		"fields": []string{"id", "tenant_id", "resource_name", "resource_type", "current_usage", "max_capacity", "usage_percent", "warning_threshold", "critical_threshold", "forecast_days", "forecast_usage", "unit", "region", "environment", "service_name", "status", "notes", "tags", "created_at", "updated_at"},
		"filters": []string{"status", "resource_type", "environment"},
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
