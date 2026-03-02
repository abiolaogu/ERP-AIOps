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
	serviceName = "runbook-service"
	moduleName  = "ERP-AIOps"
	basePath    = "/v1/aiops/runbooks"
	dbName      = "erp_aiops"
	tableName   = "aiops_runbooks"
	eventTopic  = "erp.aiops.runbook"
	cacheTTL    = 30 * time.Second
)

var validStatuses = map[string]bool{"draft": true, "active": true, "running": true, "completed": true, "failed": true, "cancelled": true}

type runbook struct {
	ID            string  `json:"id"`
	TenantID      string  `json:"tenant_id"`
	Name          string  `json:"name"`
	Description   *string `json:"description,omitempty"`
	Status        string  `json:"status"`
	Steps         *string `json:"steps,omitempty"`
	TotalSteps    int     `json:"total_steps"`
	CurrentStep   int     `json:"current_step"`
	TriggerType   string  `json:"trigger_type"`
	TriggerConfig *string `json:"trigger_config,omitempty"`
	ServiceName   *string `json:"service_name,omitempty"`
	Environment   *string `json:"environment,omitempty"`
	LastRunAt     *string `json:"last_run_at,omitempty"`
	LastRunResult *string `json:"last_run_result,omitempty"`
	LastRunLog    *string `json:"last_run_log,omitempty"`
	OwnerID       *string `json:"owner_id,omitempty"`
	Tags          *string `json:"tags,omitempty"`
	CreatedAt     string  `json:"created_at"`
	UpdatedAt     string  `json:"updated_at"`
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
	List(ctx context.Context, tenantID, cursor string, limit int, filters map[string]string) ([]runbook, string, error)
	GetByID(ctx context.Context, tenantID, id string) (*runbook, error)
	Create(ctx context.Context, rb *runbook) error
	Update(ctx context.Context, rb *runbook) error
	Delete(ctx context.Context, tenantID, id string) error
}

type memoryStore struct { mu sync.RWMutex; records map[string]runbook }
func newMemoryStore() *memoryStore { return &memoryStore{records: make(map[string]runbook)} }

func (m *memoryStore) List(_ context.Context, tenantID, cursor string, limit int, filters map[string]string) ([]runbook, string, error) {
	m.mu.RLock(); defer m.mu.RUnlock()
	var all []runbook
	for _, rb := range m.records {
		if rb.TenantID != tenantID { continue }
		if v, ok := filters["status"]; ok && rb.Status != v { continue }
		if v, ok := filters["trigger_type"]; ok && rb.TriggerType != v { continue }
		if v, ok := filters["owner_id"]; ok && strVal(rb.OwnerID) != v { continue }
		all = append(all, rb)
	}
	sort.Slice(all, func(i, j int) bool { return all[i].CreatedAt < all[j].CreatedAt })
	start := 0; if cursor != "" { for idx, rb := range all { if rb.ID == cursor { start = idx + 1; break } } }
	if start >= len(all) { return []runbook{}, "", nil }
	end := start + limit; if end > len(all) { end = len(all) }
	result := all[start:end]; nextCursor := ""; if end < len(all) { nextCursor = result[len(result)-1].ID }
	return result, nextCursor, nil
}

func (m *memoryStore) GetByID(_ context.Context, tenantID, id string) (*runbook, error) {
	m.mu.RLock(); defer m.mu.RUnlock(); rb, ok := m.records[id]; if !ok || rb.TenantID != tenantID { return nil, errors.New("not found") }; return &rb, nil
}
func (m *memoryStore) Create(_ context.Context, rb *runbook) error { m.mu.Lock(); defer m.mu.Unlock(); m.records[rb.ID] = *rb; return nil }
func (m *memoryStore) Update(_ context.Context, rb *runbook) error { m.mu.Lock(); defer m.mu.Unlock(); if _, ok := m.records[rb.ID]; !ok { return errors.New("not found") }; m.records[rb.ID] = *rb; return nil }
func (m *memoryStore) Delete(_ context.Context, tenantID, id string) error { m.mu.Lock(); defer m.mu.Unlock(); rb, ok := m.records[id]; if !ok || rb.TenantID != tenantID { return errors.New("not found") }; delete(m.records, id); return nil }

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
	status TEXT CHECK (status IN ('draft','active','running','completed','failed','cancelled')) DEFAULT 'draft',
	steps TEXT, total_steps INT DEFAULT 0, current_step INT DEFAULT 0,
	trigger_type TEXT DEFAULT 'manual', trigger_config TEXT, service_name TEXT, environment TEXT,
	last_run_at TIMESTAMPTZ, last_run_result TEXT, last_run_log TEXT, owner_id TEXT, tags TEXT,
	created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(), updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_` + tableName + `_tenant ON ` + tableName + ` (tenant_id);
CREATE INDEX IF NOT EXISTS idx_` + tableName + `_status ON ` + tableName + ` (tenant_id, status);`

func (p *postgresStore) List(ctx context.Context, tenantID, cursor string, limit int, filters map[string]string) ([]runbook, string, error) {
	query := `SELECT id,tenant_id,name,description,status,steps,total_steps,current_step,trigger_type,trigger_config,service_name,environment,last_run_at,last_run_result,last_run_log,owner_id,tags,created_at,updated_at FROM ` + tableName + ` WHERE tenant_id=$1`
	args := []any{tenantID}; idx := 2
	for _, f := range []struct{ k, c string }{{"status", "status"}, {"trigger_type", "trigger_type"}, {"owner_id", "owner_id"}} {
		if v, ok := filters[f.k]; ok { query += fmt.Sprintf(" AND %s=$%d", f.c, idx); args = append(args, v); idx++ }
	}
	if cursor != "" { query += fmt.Sprintf(" AND created_at>(SELECT created_at FROM "+tableName+" WHERE id=$%d)", idx); args = append(args, cursor); idx++ }
	query += " ORDER BY created_at ASC" + fmt.Sprintf(" LIMIT $%d", idx); args = append(args, limit+1)
	rows, err := p.db.QueryContext(ctx, query, args...); if err != nil { return nil, "", fmt.Errorf("query: %w", err) }; defer rows.Close()
	var results []runbook
	for rows.Next() {
		var rb runbook; var desc, steps, trigCfg, svcName, env, lastResult, lastLog, ownerID, tags sql.NullString; var lastRunAt, createdAt, updatedAt sql.NullTime
		if err := rows.Scan(&rb.ID, &rb.TenantID, &rb.Name, &desc, &rb.Status, &steps, &rb.TotalSteps, &rb.CurrentStep, &rb.TriggerType, &trigCfg, &svcName, &env, &lastRunAt, &lastResult, &lastLog, &ownerID, &tags, &createdAt, &updatedAt); err != nil { return nil, "", fmt.Errorf("scan: %w", err) }
		if desc.Valid { rb.Description = &desc.String }; if steps.Valid { rb.Steps = &steps.String }
		if trigCfg.Valid { rb.TriggerConfig = &trigCfg.String }; if svcName.Valid { rb.ServiceName = &svcName.String }
		if env.Valid { rb.Environment = &env.String }; if lastRunAt.Valid { s := lastRunAt.Time.Format(time.RFC3339); rb.LastRunAt = &s }
		if lastResult.Valid { rb.LastRunResult = &lastResult.String }; if lastLog.Valid { rb.LastRunLog = &lastLog.String }
		if ownerID.Valid { rb.OwnerID = &ownerID.String }; if tags.Valid { rb.Tags = &tags.String }
		if createdAt.Valid { rb.CreatedAt = createdAt.Time.Format(time.RFC3339) }; if updatedAt.Valid { rb.UpdatedAt = updatedAt.Time.Format(time.RFC3339) }
		results = append(results, rb)
	}
	nextCursor := ""; if len(results) > limit { nextCursor = results[limit-1].ID; results = results[:limit] }
	return results, nextCursor, nil
}

func (p *postgresStore) GetByID(ctx context.Context, tenantID, id string) (*runbook, error) {
	query := `SELECT id,tenant_id,name,description,status,steps,total_steps,current_step,trigger_type,trigger_config,service_name,environment,last_run_at,last_run_result,last_run_log,owner_id,tags,created_at,updated_at FROM ` + tableName + ` WHERE id=$1 AND tenant_id=$2`
	var rb runbook; var desc, steps, trigCfg, svcName, env, lastResult, lastLog, ownerID, tags sql.NullString; var lastRunAt, createdAt, updatedAt sql.NullTime
	err := p.db.QueryRowContext(ctx, query, id, tenantID).Scan(&rb.ID, &rb.TenantID, &rb.Name, &desc, &rb.Status, &steps, &rb.TotalSteps, &rb.CurrentStep, &rb.TriggerType, &trigCfg, &svcName, &env, &lastRunAt, &lastResult, &lastLog, &ownerID, &tags, &createdAt, &updatedAt)
	if err != nil { if errors.Is(err, sql.ErrNoRows) { return nil, errors.New("not found") }; return nil, fmt.Errorf("query row: %w", err) }
	if desc.Valid { rb.Description = &desc.String }; if steps.Valid { rb.Steps = &steps.String }
	if trigCfg.Valid { rb.TriggerConfig = &trigCfg.String }; if svcName.Valid { rb.ServiceName = &svcName.String }
	if env.Valid { rb.Environment = &env.String }; if lastRunAt.Valid { s := lastRunAt.Time.Format(time.RFC3339); rb.LastRunAt = &s }
	if lastResult.Valid { rb.LastRunResult = &lastResult.String }; if lastLog.Valid { rb.LastRunLog = &lastLog.String }
	if ownerID.Valid { rb.OwnerID = &ownerID.String }; if tags.Valid { rb.Tags = &tags.String }
	if createdAt.Valid { rb.CreatedAt = createdAt.Time.Format(time.RFC3339) }; if updatedAt.Valid { rb.UpdatedAt = updatedAt.Time.Format(time.RFC3339) }
	return &rb, nil
}

func (p *postgresStore) Create(ctx context.Context, rb *runbook) error {
	query := `INSERT INTO ` + tableName + ` (id,tenant_id,name,description,status,steps,total_steps,current_step,trigger_type,trigger_config,service_name,environment,last_run_at,last_run_result,last_run_log,owner_id,tags,created_at,updated_at) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17,$18,$19)`
	_, err := p.db.ExecContext(ctx, query, rb.ID, rb.TenantID, rb.Name, rb.Description, rb.Status, rb.Steps, rb.TotalSteps, rb.CurrentStep, rb.TriggerType, rb.TriggerConfig, rb.ServiceName, rb.Environment, parseTimePtr(rb.LastRunAt), rb.LastRunResult, rb.LastRunLog, rb.OwnerID, rb.Tags, parseTime(rb.CreatedAt), parseTime(rb.UpdatedAt))
	return err
}

func (p *postgresStore) Update(ctx context.Context, rb *runbook) error {
	query := `UPDATE ` + tableName + ` SET name=$1,description=$2,status=$3,steps=$4,total_steps=$5,current_step=$6,trigger_type=$7,trigger_config=$8,service_name=$9,environment=$10,last_run_at=$11,last_run_result=$12,last_run_log=$13,owner_id=$14,tags=$15,updated_at=$16 WHERE id=$17 AND tenant_id=$18`
	res, err := p.db.ExecContext(ctx, query, rb.Name, rb.Description, rb.Status, rb.Steps, rb.TotalSteps, rb.CurrentStep, rb.TriggerType, rb.TriggerConfig, rb.ServiceName, rb.Environment, parseTimePtr(rb.LastRunAt), rb.LastRunResult, rb.LastRunLog, rb.OwnerID, rb.Tags, parseTime(rb.UpdatedAt), rb.ID, rb.TenantID)
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
	filters := make(map[string]string); for _, key := range []string{"status", "trigger_type", "owner_id"} { if v := r.URL.Query().Get(key); v != "" { filters[key] = v } }
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
	name, _ := body["name"].(string); if name == "" { writeJSON(w, http.StatusBadRequest, map[string]string{"error": "name is required"}); return }
	now := time.Now().UTC().Format(time.RFC3339)
	rb := &runbook{ID: newID(), TenantID: tenantID, Name: name, Description: strPtr(body["description"]), Status: "draft", Steps: strPtr(body["steps"]), TotalSteps: 0, CurrentStep: 0, TriggerType: "manual", TriggerConfig: strPtr(body["trigger_config"]), ServiceName: strPtr(body["service_name"]), Environment: strPtr(body["environment"]), OwnerID: strPtr(body["owner_id"]), Tags: strPtr(body["tags"]), CreatedAt: now, UpdatedAt: now}
	if v, ok := body["status"].(string); ok && validStatuses[v] { rb.Status = v }
	if v, ok := body["total_steps"].(float64); ok { rb.TotalSteps = int(v) }
	if v, ok := body["trigger_type"].(string); ok && v != "" { rb.TriggerType = v }
	if err := s.store.Create(r.Context(), rb); err != nil { writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()}); return }
	s.cache.invalidate("list:" + tenantID); writeJSON(w, http.StatusCreated, map[string]any{"item": rb, "event_topic": eventTopic + ".created"})
}

func (s *server) handleUpdate(w http.ResponseWriter, r *http.Request, id string) {
	tenantID := r.Header.Get("X-Tenant-ID")
	existing, err := s.store.GetByID(r.Context(), tenantID, id)
	if err != nil { if err.Error() == "not found" { writeJSON(w, http.StatusNotFound, map[string]string{"error": "not found"}); return }; writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()}); return }
	body, err := readJSON(r)
	if err != nil { writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid JSON: " + err.Error()}); return }
	if v, ok := body["name"].(string); ok && v != "" { existing.Name = v }
	if v, exists := body["description"]; exists { existing.Description = strPtr(v) }
	if v, ok := body["status"].(string); ok && validStatuses[v] { existing.Status = v }
	if v, exists := body["steps"]; exists { existing.Steps = strPtr(v) }
	if v, ok := body["total_steps"].(float64); ok { existing.TotalSteps = int(v) }
	if v, ok := body["current_step"].(float64); ok { existing.CurrentStep = int(v) }
	if v, ok := body["trigger_type"].(string); ok && v != "" { existing.TriggerType = v }
	if v, exists := body["trigger_config"]; exists { existing.TriggerConfig = strPtr(v) }
	if v, exists := body["service_name"]; exists { existing.ServiceName = strPtr(v) }
	if v, exists := body["environment"]; exists { existing.Environment = strPtr(v) }
	if v, exists := body["owner_id"]; exists { existing.OwnerID = strPtr(v) }
	if v, exists := body["tags"]; exists { existing.Tags = strPtr(v) }
	if v, exists := body["last_run_result"]; exists { existing.LastRunResult = strPtr(v) }
	if v, exists := body["last_run_log"]; exists { existing.LastRunLog = strPtr(v) }
	existing.UpdatedAt = time.Now().UTC().Format(time.RFC3339)
	if err := s.store.Update(r.Context(), existing); err != nil { writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()}); return }
	s.cache.invalidate("list:" + tenantID); s.cache.invalidate("get:" + tenantID + ":" + id)
	writeJSON(w, http.StatusOK, map[string]any{"item": existing, "event_topic": eventTopic + ".updated"})
}

func (s *server) handleExecute(w http.ResponseWriter, r *http.Request, id string) {
	tenantID := r.Header.Get("X-Tenant-ID")
	existing, err := s.store.GetByID(r.Context(), tenantID, id)
	if err != nil { if err.Error() == "not found" { writeJSON(w, http.StatusNotFound, map[string]string{"error": "not found"}); return }; writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()}); return }
	now := time.Now().UTC().Format(time.RFC3339)
	existing.Status = "running"; existing.CurrentStep = 0; existing.LastRunAt = &now; existing.UpdatedAt = now
	if err := s.store.Update(r.Context(), existing); err != nil { writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()}); return }
	s.cache.invalidate("list:" + tenantID); s.cache.invalidate("get:" + tenantID + ":" + id)
	writeJSON(w, http.StatusOK, map[string]any{"item": existing, "event_topic": eventTopic + ".executed"})
}

func (s *server) handleDelete(w http.ResponseWriter, r *http.Request, id string) {
	tenantID := r.Header.Get("X-Tenant-ID")
	if err := s.store.Delete(r.Context(), tenantID, id); err != nil { if err.Error() == "not found" { writeJSON(w, http.StatusNotFound, map[string]string{"error": "not found"}); return }; writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()}); return }
	s.cache.invalidate("list:" + tenantID); s.cache.invalidate("get:" + tenantID + ":" + id)
	writeJSON(w, http.StatusOK, map[string]any{"deleted": true, "id": id, "event_topic": eventTopic + ".deleted"})
}

func handleExplain(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]any{"service": serviceName, "module": moduleName, "base_path": basePath, "database": dbName, "table": tableName, "event_topic": eventTopic, "entity": "runbook",
		"fields": []string{"id", "tenant_id", "name", "description", "status", "steps", "total_steps", "current_step", "trigger_type", "trigger_config", "service_name", "environment", "last_run_at", "last_run_result", "last_run_log", "owner_id", "tags", "created_at", "updated_at"},
		"filters": []string{"status", "trigger_type", "owner_id"},
		"endpoints": map[string]string{"list": "GET " + basePath, "get": "GET " + basePath + "/{id}", "create": "POST " + basePath, "update": "PUT/PATCH " + basePath + "/{id}", "execute": "POST " + basePath + "/{id}/execute", "delete": "DELETE " + basePath + "/{id}"},
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
		sub := strings.TrimPrefix(r.URL.Path, basePath+"/"); parts := strings.SplitN(sub, "/", 2); id := parts[0]
		if id == "" { writeJSON(w, http.StatusBadRequest, map[string]string{"error": "missing id"}); return }
		if len(parts) == 2 && parts[1] == "execute" && r.Method == http.MethodPost { srv.handleExecute(w, r, id); return }
		switch r.Method { case http.MethodGet: srv.handleGet(w, r, id); case http.MethodPut, http.MethodPatch: srv.handleUpdate(w, r, id); case http.MethodDelete: srv.handleDelete(w, r, id); default: writeJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "method not allowed"}) }
	})
	handler := securityHeaders(mux); log.Printf("%s listening on :%s", serviceName, port); log.Fatal(http.ListenAndServe(":"+port, handler))
}
