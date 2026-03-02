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
	serviceName = "cost-service"
	moduleName  = "ERP-AIOps"
	basePath    = "/v1/aiops/costs"
	dbName      = "erp_aiops"
	tableName   = "aiops_cost_items"
	eventTopic  = "erp.aiops.cost"
	cacheTTL    = 30 * time.Second
)

var validStatuses = map[string]bool{"active": true, "optimized": true, "waste": true, "archived": true}

type costItem struct {
	ID                 string  `json:"id"`
	TenantID           string  `json:"tenant_id"`
	ResourceName       string  `json:"resource_name"`
	ResourceType       string  `json:"resource_type"`
	Provider           string  `json:"provider"`
	Region             *string `json:"region,omitempty"`
	Environment        *string `json:"environment,omitempty"`
	ServiceName        *string `json:"service_name,omitempty"`
	MonthlyCost        float64 `json:"monthly_cost"`
	DailyCost          float64 `json:"daily_cost"`
	Currency           string  `json:"currency"`
	UtilizationPercent float64 `json:"utilization_percent"`
	WastePercent       float64 `json:"waste_percent"`
	SavingsPotential   float64 `json:"savings_potential"`
	Recommendation     *string `json:"recommendation,omitempty"`
	Status             string  `json:"status"`
	Tags               *string `json:"tags,omitempty"`
	LastAnalyzedAt     *string `json:"last_analyzed_at,omitempty"`
	CreatedAt          string  `json:"created_at"`
	UpdatedAt          string  `json:"updated_at"`
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
	List(ctx context.Context, tenantID, cursor string, limit int, filters map[string]string) ([]costItem, string, error)
	GetByID(ctx context.Context, tenantID, id string) (*costItem, error)
	Create(ctx context.Context, ci *costItem) error
	Update(ctx context.Context, ci *costItem) error
	Delete(ctx context.Context, tenantID, id string) error
}

type memoryStore struct { mu sync.RWMutex; records map[string]costItem }
func newMemoryStore() *memoryStore { return &memoryStore{records: make(map[string]costItem)} }

func (m *memoryStore) List(_ context.Context, tenantID, cursor string, limit int, filters map[string]string) ([]costItem, string, error) {
	m.mu.RLock(); defer m.mu.RUnlock()
	var all []costItem
	for _, ci := range m.records {
		if ci.TenantID != tenantID { continue }
		if v, ok := filters["status"]; ok && ci.Status != v { continue }
		if v, ok := filters["provider"]; ok && ci.Provider != v { continue }
		if v, ok := filters["resource_type"]; ok && ci.ResourceType != v { continue }
		all = append(all, ci)
	}
	sort.Slice(all, func(i, j int) bool { return all[i].CreatedAt < all[j].CreatedAt })
	start := 0; if cursor != "" { for idx, ci := range all { if ci.ID == cursor { start = idx + 1; break } } }
	if start >= len(all) { return []costItem{}, "", nil }
	end := start + limit; if end > len(all) { end = len(all) }
	result := all[start:end]; nextCursor := ""; if end < len(all) { nextCursor = result[len(result)-1].ID }
	return result, nextCursor, nil
}

func (m *memoryStore) GetByID(_ context.Context, tenantID, id string) (*costItem, error) { m.mu.RLock(); defer m.mu.RUnlock(); ci, ok := m.records[id]; if !ok || ci.TenantID != tenantID { return nil, errors.New("not found") }; return &ci, nil }
func (m *memoryStore) Create(_ context.Context, ci *costItem) error { m.mu.Lock(); defer m.mu.Unlock(); m.records[ci.ID] = *ci; return nil }
func (m *memoryStore) Update(_ context.Context, ci *costItem) error { m.mu.Lock(); defer m.mu.Unlock(); if _, ok := m.records[ci.ID]; !ok { return errors.New("not found") }; m.records[ci.ID] = *ci; return nil }
func (m *memoryStore) Delete(_ context.Context, tenantID, id string) error { m.mu.Lock(); defer m.mu.Unlock(); ci, ok := m.records[id]; if !ok || ci.TenantID != tenantID { return errors.New("not found") }; delete(m.records, id); return nil }

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
	id TEXT PRIMARY KEY, tenant_id TEXT NOT NULL, resource_name TEXT NOT NULL,
	resource_type TEXT NOT NULL, provider TEXT NOT NULL DEFAULT 'aws',
	region TEXT, environment TEXT, service_name TEXT,
	monthly_cost DOUBLE PRECISION DEFAULT 0, daily_cost DOUBLE PRECISION DEFAULT 0,
	currency TEXT DEFAULT 'USD', utilization_percent DOUBLE PRECISION DEFAULT 0,
	waste_percent DOUBLE PRECISION DEFAULT 0, savings_potential DOUBLE PRECISION DEFAULT 0,
	recommendation TEXT,
	status TEXT CHECK (status IN ('active','optimized','waste','archived')) DEFAULT 'active',
	tags TEXT, last_analyzed_at TIMESTAMPTZ,
	created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(), updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_` + tableName + `_tenant ON ` + tableName + ` (tenant_id);
CREATE INDEX IF NOT EXISTS idx_` + tableName + `_status ON ` + tableName + ` (tenant_id, status);`

func (p *postgresStore) List(ctx context.Context, tenantID, cursor string, limit int, filters map[string]string) ([]costItem, string, error) {
	query := `SELECT id,tenant_id,resource_name,resource_type,provider,region,environment,service_name,monthly_cost,daily_cost,currency,utilization_percent,waste_percent,savings_potential,recommendation,status,tags,last_analyzed_at,created_at,updated_at FROM ` + tableName + ` WHERE tenant_id=$1`
	args := []any{tenantID}; idx := 2
	for _, f := range []struct{ k, c string }{{"status", "status"}, {"provider", "provider"}, {"resource_type", "resource_type"}} {
		if v, ok := filters[f.k]; ok { query += fmt.Sprintf(" AND %s=$%d", f.c, idx); args = append(args, v); idx++ }
	}
	if cursor != "" { query += fmt.Sprintf(" AND created_at>(SELECT created_at FROM "+tableName+" WHERE id=$%d)", idx); args = append(args, cursor); idx++ }
	query += " ORDER BY created_at ASC" + fmt.Sprintf(" LIMIT $%d", idx); args = append(args, limit+1)
	rows, err := p.db.QueryContext(ctx, query, args...); if err != nil { return nil, "", fmt.Errorf("query: %w", err) }; defer rows.Close()
	var results []costItem
	for rows.Next() {
		var ci costItem; var region, env, svcName, rec, tags sql.NullString; var lastAnalyzed, createdAt, updatedAt sql.NullTime
		if err := rows.Scan(&ci.ID, &ci.TenantID, &ci.ResourceName, &ci.ResourceType, &ci.Provider, &region, &env, &svcName, &ci.MonthlyCost, &ci.DailyCost, &ci.Currency, &ci.UtilizationPercent, &ci.WastePercent, &ci.SavingsPotential, &rec, &ci.Status, &tags, &lastAnalyzed, &createdAt, &updatedAt); err != nil { return nil, "", fmt.Errorf("scan: %w", err) }
		if region.Valid { ci.Region = &region.String }; if env.Valid { ci.Environment = &env.String }
		if svcName.Valid { ci.ServiceName = &svcName.String }; if rec.Valid { ci.Recommendation = &rec.String }
		if tags.Valid { ci.Tags = &tags.String }
		if lastAnalyzed.Valid { s := lastAnalyzed.Time.Format(time.RFC3339); ci.LastAnalyzedAt = &s }
		if createdAt.Valid { ci.CreatedAt = createdAt.Time.Format(time.RFC3339) }; if updatedAt.Valid { ci.UpdatedAt = updatedAt.Time.Format(time.RFC3339) }
		results = append(results, ci)
	}
	nextCursor := ""; if len(results) > limit { nextCursor = results[limit-1].ID; results = results[:limit] }
	return results, nextCursor, nil
}

func (p *postgresStore) GetByID(ctx context.Context, tenantID, id string) (*costItem, error) {
	query := `SELECT id,tenant_id,resource_name,resource_type,provider,region,environment,service_name,monthly_cost,daily_cost,currency,utilization_percent,waste_percent,savings_potential,recommendation,status,tags,last_analyzed_at,created_at,updated_at FROM ` + tableName + ` WHERE id=$1 AND tenant_id=$2`
	var ci costItem; var region, env, svcName, rec, tags sql.NullString; var lastAnalyzed, createdAt, updatedAt sql.NullTime
	err := p.db.QueryRowContext(ctx, query, id, tenantID).Scan(&ci.ID, &ci.TenantID, &ci.ResourceName, &ci.ResourceType, &ci.Provider, &region, &env, &svcName, &ci.MonthlyCost, &ci.DailyCost, &ci.Currency, &ci.UtilizationPercent, &ci.WastePercent, &ci.SavingsPotential, &rec, &ci.Status, &tags, &lastAnalyzed, &createdAt, &updatedAt)
	if err != nil { if errors.Is(err, sql.ErrNoRows) { return nil, errors.New("not found") }; return nil, fmt.Errorf("query row: %w", err) }
	if region.Valid { ci.Region = &region.String }; if env.Valid { ci.Environment = &env.String }
	if svcName.Valid { ci.ServiceName = &svcName.String }; if rec.Valid { ci.Recommendation = &rec.String }
	if tags.Valid { ci.Tags = &tags.String }
	if lastAnalyzed.Valid { s := lastAnalyzed.Time.Format(time.RFC3339); ci.LastAnalyzedAt = &s }
	if createdAt.Valid { ci.CreatedAt = createdAt.Time.Format(time.RFC3339) }; if updatedAt.Valid { ci.UpdatedAt = updatedAt.Time.Format(time.RFC3339) }
	return &ci, nil
}

func (p *postgresStore) Create(ctx context.Context, ci *costItem) error {
	query := `INSERT INTO ` + tableName + ` (id,tenant_id,resource_name,resource_type,provider,region,environment,service_name,monthly_cost,daily_cost,currency,utilization_percent,waste_percent,savings_potential,recommendation,status,tags,last_analyzed_at,created_at,updated_at) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17,$18,$19,$20)`
	_, err := p.db.ExecContext(ctx, query, ci.ID, ci.TenantID, ci.ResourceName, ci.ResourceType, ci.Provider, ci.Region, ci.Environment, ci.ServiceName, ci.MonthlyCost, ci.DailyCost, ci.Currency, ci.UtilizationPercent, ci.WastePercent, ci.SavingsPotential, ci.Recommendation, ci.Status, ci.Tags, parseTimePtr(ci.LastAnalyzedAt), parseTime(ci.CreatedAt), parseTime(ci.UpdatedAt))
	return err
}

func (p *postgresStore) Update(ctx context.Context, ci *costItem) error {
	query := `UPDATE ` + tableName + ` SET resource_name=$1,resource_type=$2,provider=$3,region=$4,environment=$5,service_name=$6,monthly_cost=$7,daily_cost=$8,currency=$9,utilization_percent=$10,waste_percent=$11,savings_potential=$12,recommendation=$13,status=$14,tags=$15,last_analyzed_at=$16,updated_at=$17 WHERE id=$18 AND tenant_id=$19`
	res, err := p.db.ExecContext(ctx, query, ci.ResourceName, ci.ResourceType, ci.Provider, ci.Region, ci.Environment, ci.ServiceName, ci.MonthlyCost, ci.DailyCost, ci.Currency, ci.UtilizationPercent, ci.WastePercent, ci.SavingsPotential, ci.Recommendation, ci.Status, ci.Tags, parseTimePtr(ci.LastAnalyzedAt), parseTime(ci.UpdatedAt), ci.ID, ci.TenantID)
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
	filters := make(map[string]string); for _, key := range []string{"status", "provider", "resource_type"} { if v := r.URL.Query().Get(key); v != "" { filters[key] = v } }
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
	resName, _ := body["resource_name"].(string); resType, _ := body["resource_type"].(string); provider, _ := body["provider"].(string)
	if resName == "" || resType == "" { writeJSON(w, http.StatusBadRequest, map[string]string{"error": "resource_name and resource_type are required"}); return }
	if provider == "" { provider = "aws" }
	now := time.Now().UTC().Format(time.RFC3339)
	ci := &costItem{ID: newID(), TenantID: tenantID, ResourceName: resName, ResourceType: resType, Provider: provider, Region: strPtr(body["region"]), Environment: strPtr(body["environment"]), ServiceName: strPtr(body["service_name"]), MonthlyCost: 0, DailyCost: 0, Currency: "USD", UtilizationPercent: 0, WastePercent: 0, SavingsPotential: 0, Recommendation: strPtr(body["recommendation"]), Status: "active", Tags: strPtr(body["tags"]), CreatedAt: now, UpdatedAt: now}
	if v, ok := body["monthly_cost"].(float64); ok { ci.MonthlyCost = v; ci.DailyCost = v / 30 }
	if v, ok := body["daily_cost"].(float64); ok { ci.DailyCost = v }
	if v, ok := body["currency"].(string); ok && v != "" { ci.Currency = v }
	if v, ok := body["utilization_percent"].(float64); ok { ci.UtilizationPercent = v; ci.WastePercent = 100 - v; ci.SavingsPotential = ci.MonthlyCost * (ci.WastePercent / 100) }
	if ci.WastePercent > 50 { ci.Status = "waste" }
	if err := sv.store.Create(r.Context(), ci); err != nil { writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()}); return }
	sv.cache.invalidate("list:" + tenantID); writeJSON(w, http.StatusCreated, map[string]any{"item": ci, "event_topic": eventTopic + ".created"})
}

func (sv *server) handleUpdate(w http.ResponseWriter, r *http.Request, id string) {
	tenantID := r.Header.Get("X-Tenant-ID")
	existing, err := sv.store.GetByID(r.Context(), tenantID, id)
	if err != nil { if err.Error() == "not found" { writeJSON(w, http.StatusNotFound, map[string]string{"error": "not found"}); return }; writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()}); return }
	body, err := readJSON(r)
	if err != nil { writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid JSON: " + err.Error()}); return }
	if v, ok := body["resource_name"].(string); ok && v != "" { existing.ResourceName = v }
	if v, ok := body["resource_type"].(string); ok && v != "" { existing.ResourceType = v }
	if v, ok := body["provider"].(string); ok && v != "" { existing.Provider = v }
	if v, exists := body["region"]; exists { existing.Region = strPtr(v) }
	if v, exists := body["environment"]; exists { existing.Environment = strPtr(v) }
	if v, exists := body["service_name"]; exists { existing.ServiceName = strPtr(v) }
	if v, ok := body["monthly_cost"].(float64); ok { existing.MonthlyCost = v; existing.DailyCost = v / 30 }
	if v, ok := body["daily_cost"].(float64); ok { existing.DailyCost = v }
	if v, ok := body["currency"].(string); ok && v != "" { existing.Currency = v }
	if v, ok := body["utilization_percent"].(float64); ok { existing.UtilizationPercent = v; existing.WastePercent = 100 - v; existing.SavingsPotential = existing.MonthlyCost * (existing.WastePercent / 100) }
	if v, exists := body["recommendation"]; exists { existing.Recommendation = strPtr(v) }
	if v, ok := body["status"].(string); ok && validStatuses[v] { existing.Status = v }
	if v, exists := body["tags"]; exists { existing.Tags = strPtr(v) }
	now := time.Now().UTC().Format(time.RFC3339); existing.LastAnalyzedAt = &now; existing.UpdatedAt = now
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
	writeJSON(w, http.StatusOK, map[string]any{"service": serviceName, "module": moduleName, "base_path": basePath, "database": dbName, "table": tableName, "event_topic": eventTopic, "entity": "costItem",
		"fields": []string{"id", "tenant_id", "resource_name", "resource_type", "provider", "region", "environment", "service_name", "monthly_cost", "daily_cost", "currency", "utilization_percent", "waste_percent", "savings_potential", "recommendation", "status", "tags", "last_analyzed_at", "created_at", "updated_at"},
		"filters": []string{"status", "provider", "resource_type"},
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
