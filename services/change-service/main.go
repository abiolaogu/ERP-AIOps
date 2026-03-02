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
	serviceName = "change-service"
	moduleName  = "ERP-AIOps"
	basePath    = "/v1/aiops/changes"
	dbName      = "erp_aiops"
	tableName   = "aiops_changes"
	eventTopic  = "erp.aiops.change"
	cacheTTL    = 30 * time.Second
)

var (
	validStatuses  = map[string]bool{"draft": true, "pending_approval": true, "approved": true, "scheduled": true, "in_progress": true, "completed": true, "failed": true, "rolled_back": true, "cancelled": true}
	validRiskLevels = map[string]bool{"low": true, "medium": true, "high": true, "critical": true}
	validTypes      = map[string]bool{"standard": true, "normal": true, "emergency": true}
)

type changeRequest struct {
	ID            string  `json:"id"`
	TenantID      string  `json:"tenant_id"`
	Title         string  `json:"title"`
	Description   *string `json:"description,omitempty"`
	ChangeType    string  `json:"change_type"`
	Status        string  `json:"status"`
	RiskLevel     string  `json:"risk_level"`
	Impact        *string `json:"impact,omitempty"`
	ServiceName   *string `json:"service_name,omitempty"`
	Environment   *string `json:"environment,omitempty"`
	RequestedBy   string  `json:"requested_by"`
	ApprovedBy    *string `json:"approved_by,omitempty"`
	ScheduledAt   *string `json:"scheduled_at,omitempty"`
	StartedAt     *string `json:"started_at,omitempty"`
	CompletedAt   *string `json:"completed_at,omitempty"`
	RollbackPlan  *string `json:"rollback_plan,omitempty"`
	RolledBackAt  *string `json:"rolled_back_at,omitempty"`
	Notes         *string `json:"notes,omitempty"`
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
	List(ctx context.Context, tenantID, cursor string, limit int, filters map[string]string) ([]changeRequest, string, error)
	GetByID(ctx context.Context, tenantID, id string) (*changeRequest, error)
	Create(ctx context.Context, cr *changeRequest) error
	Update(ctx context.Context, cr *changeRequest) error
	Delete(ctx context.Context, tenantID, id string) error
}

type memoryStore struct { mu sync.RWMutex; records map[string]changeRequest }
func newMemoryStore() *memoryStore { return &memoryStore{records: make(map[string]changeRequest)} }

func (m *memoryStore) List(_ context.Context, tenantID, cursor string, limit int, filters map[string]string) ([]changeRequest, string, error) {
	m.mu.RLock(); defer m.mu.RUnlock()
	var all []changeRequest
	for _, cr := range m.records {
		if cr.TenantID != tenantID { continue }
		if v, ok := filters["status"]; ok && cr.Status != v { continue }
		if v, ok := filters["change_type"]; ok && cr.ChangeType != v { continue }
		if v, ok := filters["risk_level"]; ok && cr.RiskLevel != v { continue }
		if v, ok := filters["requested_by"]; ok && cr.RequestedBy != v { continue }
		all = append(all, cr)
	}
	sort.Slice(all, func(i, j int) bool { return all[i].CreatedAt < all[j].CreatedAt })
	start := 0; if cursor != "" { for idx, cr := range all { if cr.ID == cursor { start = idx + 1; break } } }
	if start >= len(all) { return []changeRequest{}, "", nil }
	end := start + limit; if end > len(all) { end = len(all) }
	result := all[start:end]; nextCursor := ""; if end < len(all) { nextCursor = result[len(result)-1].ID }
	return result, nextCursor, nil
}

func (m *memoryStore) GetByID(_ context.Context, tenantID, id string) (*changeRequest, error) { m.mu.RLock(); defer m.mu.RUnlock(); cr, ok := m.records[id]; if !ok || cr.TenantID != tenantID { return nil, errors.New("not found") }; return &cr, nil }
func (m *memoryStore) Create(_ context.Context, cr *changeRequest) error { m.mu.Lock(); defer m.mu.Unlock(); m.records[cr.ID] = *cr; return nil }
func (m *memoryStore) Update(_ context.Context, cr *changeRequest) error { m.mu.Lock(); defer m.mu.Unlock(); if _, ok := m.records[cr.ID]; !ok { return errors.New("not found") }; m.records[cr.ID] = *cr; return nil }
func (m *memoryStore) Delete(_ context.Context, tenantID, id string) error { m.mu.Lock(); defer m.mu.Unlock(); cr, ok := m.records[id]; if !ok || cr.TenantID != tenantID { return errors.New("not found") }; delete(m.records, id); return nil }

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
	id TEXT PRIMARY KEY, tenant_id TEXT NOT NULL, title TEXT NOT NULL, description TEXT,
	change_type TEXT CHECK (change_type IN ('standard','normal','emergency')) DEFAULT 'normal',
	status TEXT CHECK (status IN ('draft','pending_approval','approved','scheduled','in_progress','completed','failed','rolled_back','cancelled')) DEFAULT 'draft',
	risk_level TEXT CHECK (risk_level IN ('low','medium','high','critical')) DEFAULT 'medium',
	impact TEXT, service_name TEXT, environment TEXT,
	requested_by TEXT NOT NULL, approved_by TEXT,
	scheduled_at TIMESTAMPTZ, started_at TIMESTAMPTZ, completed_at TIMESTAMPTZ,
	rollback_plan TEXT, rolled_back_at TIMESTAMPTZ,
	notes TEXT, tags TEXT,
	created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(), updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_` + tableName + `_tenant ON ` + tableName + ` (tenant_id);
CREATE INDEX IF NOT EXISTS idx_` + tableName + `_status ON ` + tableName + ` (tenant_id, status);`

func (p *postgresStore) List(ctx context.Context, tenantID, cursor string, limit int, filters map[string]string) ([]changeRequest, string, error) {
	query := `SELECT id,tenant_id,title,description,change_type,status,risk_level,impact,service_name,environment,requested_by,approved_by,scheduled_at,started_at,completed_at,rollback_plan,rolled_back_at,notes,tags,created_at,updated_at FROM ` + tableName + ` WHERE tenant_id=$1`
	args := []any{tenantID}; idx := 2
	for _, f := range []struct{ k, c string }{{"status", "status"}, {"change_type", "change_type"}, {"risk_level", "risk_level"}, {"requested_by", "requested_by"}} {
		if v, ok := filters[f.k]; ok { query += fmt.Sprintf(" AND %s=$%d", f.c, idx); args = append(args, v); idx++ }
	}
	if cursor != "" { query += fmt.Sprintf(" AND created_at>(SELECT created_at FROM "+tableName+" WHERE id=$%d)", idx); args = append(args, cursor); idx++ }
	query += " ORDER BY created_at ASC" + fmt.Sprintf(" LIMIT $%d", idx); args = append(args, limit+1)
	rows, err := p.db.QueryContext(ctx, query, args...); if err != nil { return nil, "", fmt.Errorf("query: %w", err) }; defer rows.Close()
	var results []changeRequest
	for rows.Next() {
		var cr changeRequest
		var desc, impact, svcName, env, approvedBy, rollbackPlan, notes, tags sql.NullString
		var scheduledAt, startedAt, completedAt, rolledBackAt, createdAt, updatedAt sql.NullTime
		if err := rows.Scan(&cr.ID, &cr.TenantID, &cr.Title, &desc, &cr.ChangeType, &cr.Status, &cr.RiskLevel, &impact, &svcName, &env, &cr.RequestedBy, &approvedBy, &scheduledAt, &startedAt, &completedAt, &rollbackPlan, &rolledBackAt, &notes, &tags, &createdAt, &updatedAt); err != nil { return nil, "", fmt.Errorf("scan: %w", err) }
		if desc.Valid { cr.Description = &desc.String }; if impact.Valid { cr.Impact = &impact.String }
		if svcName.Valid { cr.ServiceName = &svcName.String }; if env.Valid { cr.Environment = &env.String }
		if approvedBy.Valid { cr.ApprovedBy = &approvedBy.String }; if rollbackPlan.Valid { cr.RollbackPlan = &rollbackPlan.String }
		if notes.Valid { cr.Notes = &notes.String }; if tags.Valid { cr.Tags = &tags.String }
		if scheduledAt.Valid { s := scheduledAt.Time.Format(time.RFC3339); cr.ScheduledAt = &s }
		if startedAt.Valid { s := startedAt.Time.Format(time.RFC3339); cr.StartedAt = &s }
		if completedAt.Valid { s := completedAt.Time.Format(time.RFC3339); cr.CompletedAt = &s }
		if rolledBackAt.Valid { s := rolledBackAt.Time.Format(time.RFC3339); cr.RolledBackAt = &s }
		if createdAt.Valid { cr.CreatedAt = createdAt.Time.Format(time.RFC3339) }; if updatedAt.Valid { cr.UpdatedAt = updatedAt.Time.Format(time.RFC3339) }
		results = append(results, cr)
	}
	nextCursor := ""; if len(results) > limit { nextCursor = results[limit-1].ID; results = results[:limit] }
	return results, nextCursor, nil
}

func (p *postgresStore) GetByID(ctx context.Context, tenantID, id string) (*changeRequest, error) {
	query := `SELECT id,tenant_id,title,description,change_type,status,risk_level,impact,service_name,environment,requested_by,approved_by,scheduled_at,started_at,completed_at,rollback_plan,rolled_back_at,notes,tags,created_at,updated_at FROM ` + tableName + ` WHERE id=$1 AND tenant_id=$2`
	var cr changeRequest
	var desc, impact, svcName, env, approvedBy, rollbackPlan, notes, tags sql.NullString
	var scheduledAt, startedAt, completedAt, rolledBackAt, createdAt, updatedAt sql.NullTime
	err := p.db.QueryRowContext(ctx, query, id, tenantID).Scan(&cr.ID, &cr.TenantID, &cr.Title, &desc, &cr.ChangeType, &cr.Status, &cr.RiskLevel, &impact, &svcName, &env, &cr.RequestedBy, &approvedBy, &scheduledAt, &startedAt, &completedAt, &rollbackPlan, &rolledBackAt, &notes, &tags, &createdAt, &updatedAt)
	if err != nil { if errors.Is(err, sql.ErrNoRows) { return nil, errors.New("not found") }; return nil, fmt.Errorf("query row: %w", err) }
	if desc.Valid { cr.Description = &desc.String }; if impact.Valid { cr.Impact = &impact.String }
	if svcName.Valid { cr.ServiceName = &svcName.String }; if env.Valid { cr.Environment = &env.String }
	if approvedBy.Valid { cr.ApprovedBy = &approvedBy.String }; if rollbackPlan.Valid { cr.RollbackPlan = &rollbackPlan.String }
	if notes.Valid { cr.Notes = &notes.String }; if tags.Valid { cr.Tags = &tags.String }
	if scheduledAt.Valid { s := scheduledAt.Time.Format(time.RFC3339); cr.ScheduledAt = &s }
	if startedAt.Valid { s := startedAt.Time.Format(time.RFC3339); cr.StartedAt = &s }
	if completedAt.Valid { s := completedAt.Time.Format(time.RFC3339); cr.CompletedAt = &s }
	if rolledBackAt.Valid { s := rolledBackAt.Time.Format(time.RFC3339); cr.RolledBackAt = &s }
	if createdAt.Valid { cr.CreatedAt = createdAt.Time.Format(time.RFC3339) }; if updatedAt.Valid { cr.UpdatedAt = updatedAt.Time.Format(time.RFC3339) }
	return &cr, nil
}

func (p *postgresStore) Create(ctx context.Context, cr *changeRequest) error {
	query := `INSERT INTO ` + tableName + ` (id,tenant_id,title,description,change_type,status,risk_level,impact,service_name,environment,requested_by,approved_by,scheduled_at,started_at,completed_at,rollback_plan,rolled_back_at,notes,tags,created_at,updated_at) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17,$18,$19,$20,$21)`
	_, err := p.db.ExecContext(ctx, query, cr.ID, cr.TenantID, cr.Title, cr.Description, cr.ChangeType, cr.Status, cr.RiskLevel, cr.Impact, cr.ServiceName, cr.Environment, cr.RequestedBy, cr.ApprovedBy, parseTimePtr(cr.ScheduledAt), parseTimePtr(cr.StartedAt), parseTimePtr(cr.CompletedAt), cr.RollbackPlan, parseTimePtr(cr.RolledBackAt), cr.Notes, cr.Tags, parseTime(cr.CreatedAt), parseTime(cr.UpdatedAt))
	return err
}

func (p *postgresStore) Update(ctx context.Context, cr *changeRequest) error {
	query := `UPDATE ` + tableName + ` SET title=$1,description=$2,change_type=$3,status=$4,risk_level=$5,impact=$6,service_name=$7,environment=$8,requested_by=$9,approved_by=$10,scheduled_at=$11,started_at=$12,completed_at=$13,rollback_plan=$14,rolled_back_at=$15,notes=$16,tags=$17,updated_at=$18 WHERE id=$19 AND tenant_id=$20`
	res, err := p.db.ExecContext(ctx, query, cr.Title, cr.Description, cr.ChangeType, cr.Status, cr.RiskLevel, cr.Impact, cr.ServiceName, cr.Environment, cr.RequestedBy, cr.ApprovedBy, parseTimePtr(cr.ScheduledAt), parseTimePtr(cr.StartedAt), parseTimePtr(cr.CompletedAt), cr.RollbackPlan, parseTimePtr(cr.RolledBackAt), cr.Notes, cr.Tags, parseTime(cr.UpdatedAt), cr.ID, cr.TenantID)
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
	filters := make(map[string]string); for _, key := range []string{"status", "change_type", "risk_level", "requested_by"} { if v := r.URL.Query().Get(key); v != "" { filters[key] = v } }
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
	title, _ := body["title"].(string); requestedBy, _ := body["requested_by"].(string)
	if title == "" || requestedBy == "" { writeJSON(w, http.StatusBadRequest, map[string]string{"error": "title and requested_by are required"}); return }
	now := time.Now().UTC().Format(time.RFC3339)
	cr := &changeRequest{ID: newID(), TenantID: tenantID, Title: title, Description: strPtr(body["description"]), ChangeType: "normal", Status: "draft", RiskLevel: "medium", Impact: strPtr(body["impact"]), ServiceName: strPtr(body["service_name"]), Environment: strPtr(body["environment"]), RequestedBy: requestedBy, RollbackPlan: strPtr(body["rollback_plan"]), Notes: strPtr(body["notes"]), Tags: strPtr(body["tags"]), CreatedAt: now, UpdatedAt: now}
	if v, ok := body["change_type"].(string); ok && validTypes[v] { cr.ChangeType = v }
	if v, ok := body["risk_level"].(string); ok && validRiskLevels[v] { cr.RiskLevel = v }
	if v, ok := body["scheduled_at"].(string); ok { cr.ScheduledAt = &v }
	if err := s.store.Create(r.Context(), cr); err != nil { writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()}); return }
	s.cache.invalidate("list:" + tenantID); writeJSON(w, http.StatusCreated, map[string]any{"item": cr, "event_topic": eventTopic + ".created"})
}

func (s *server) handleUpdate(w http.ResponseWriter, r *http.Request, id string) {
	tenantID := r.Header.Get("X-Tenant-ID")
	existing, err := s.store.GetByID(r.Context(), tenantID, id)
	if err != nil { if err.Error() == "not found" { writeJSON(w, http.StatusNotFound, map[string]string{"error": "not found"}); return }; writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()}); return }
	body, err := readJSON(r)
	if err != nil { writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid JSON: " + err.Error()}); return }
	if v, ok := body["title"].(string); ok && v != "" { existing.Title = v }
	if v, exists := body["description"]; exists { existing.Description = strPtr(v) }
	if v, ok := body["change_type"].(string); ok && validTypes[v] { existing.ChangeType = v }
	if v, ok := body["status"].(string); ok && validStatuses[v] { existing.Status = v }
	if v, ok := body["risk_level"].(string); ok && validRiskLevels[v] { existing.RiskLevel = v }
	if v, exists := body["impact"]; exists { existing.Impact = strPtr(v) }
	if v, exists := body["service_name"]; exists { existing.ServiceName = strPtr(v) }
	if v, exists := body["environment"]; exists { existing.Environment = strPtr(v) }
	if v, exists := body["approved_by"]; exists { existing.ApprovedBy = strPtr(v) }
	if v, ok := body["scheduled_at"].(string); ok { existing.ScheduledAt = &v }
	if v, exists := body["rollback_plan"]; exists { existing.RollbackPlan = strPtr(v) }
	if v, exists := body["notes"]; exists { existing.Notes = strPtr(v) }
	if v, exists := body["tags"]; exists { existing.Tags = strPtr(v) }
	now := time.Now().UTC().Format(time.RFC3339)
	if existing.Status == "approved" && existing.ApprovedBy != nil { /* already set */ }
	if existing.Status == "in_progress" && existing.StartedAt == nil { existing.StartedAt = &now }
	if existing.Status == "completed" && existing.CompletedAt == nil { existing.CompletedAt = &now }
	if existing.Status == "rolled_back" && existing.RolledBackAt == nil { existing.RolledBackAt = &now }
	existing.UpdatedAt = now
	if err := s.store.Update(r.Context(), existing); err != nil { writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()}); return }
	s.cache.invalidate("list:" + tenantID); s.cache.invalidate("get:" + tenantID + ":" + id)
	writeJSON(w, http.StatusOK, map[string]any{"item": existing, "event_topic": eventTopic + ".updated"})
}

func (s *server) handleApprove(w http.ResponseWriter, r *http.Request, id string) {
	tenantID := r.Header.Get("X-Tenant-ID")
	existing, err := s.store.GetByID(r.Context(), tenantID, id)
	if err != nil { if err.Error() == "not found" { writeJSON(w, http.StatusNotFound, map[string]string{"error": "not found"}); return }; writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()}); return }
	body, _ := readJSON(r); approver := ""; if body != nil { approver, _ = body["approved_by"].(string) }
	now := time.Now().UTC().Format(time.RFC3339)
	existing.Status = "approved"; if approver != "" { existing.ApprovedBy = &approver }; existing.UpdatedAt = now
	if err := s.store.Update(r.Context(), existing); err != nil { writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()}); return }
	s.cache.invalidate("list:" + tenantID); s.cache.invalidate("get:" + tenantID + ":" + id)
	writeJSON(w, http.StatusOK, map[string]any{"item": existing, "event_topic": eventTopic + ".approved"})
}

func (s *server) handleRollback(w http.ResponseWriter, r *http.Request, id string) {
	tenantID := r.Header.Get("X-Tenant-ID")
	existing, err := s.store.GetByID(r.Context(), tenantID, id)
	if err != nil { if err.Error() == "not found" { writeJSON(w, http.StatusNotFound, map[string]string{"error": "not found"}); return }; writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()}); return }
	now := time.Now().UTC().Format(time.RFC3339)
	existing.Status = "rolled_back"; existing.RolledBackAt = &now; existing.UpdatedAt = now
	if err := s.store.Update(r.Context(), existing); err != nil { writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()}); return }
	s.cache.invalidate("list:" + tenantID); s.cache.invalidate("get:" + tenantID + ":" + id)
	writeJSON(w, http.StatusOK, map[string]any{"item": existing, "event_topic": eventTopic + ".rolled_back"})
}

func (s *server) handleDelete(w http.ResponseWriter, r *http.Request, id string) {
	tenantID := r.Header.Get("X-Tenant-ID")
	if err := s.store.Delete(r.Context(), tenantID, id); err != nil { if err.Error() == "not found" { writeJSON(w, http.StatusNotFound, map[string]string{"error": "not found"}); return }; writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()}); return }
	s.cache.invalidate("list:" + tenantID); s.cache.invalidate("get:" + tenantID + ":" + id)
	writeJSON(w, http.StatusOK, map[string]any{"deleted": true, "id": id, "event_topic": eventTopic + ".deleted"})
}

func handleExplain(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]any{"service": serviceName, "module": moduleName, "base_path": basePath, "database": dbName, "table": tableName, "event_topic": eventTopic, "entity": "changeRequest",
		"fields":  []string{"id", "tenant_id", "title", "description", "change_type", "status", "risk_level", "impact", "service_name", "environment", "requested_by", "approved_by", "scheduled_at", "started_at", "completed_at", "rollback_plan", "rolled_back_at", "notes", "tags", "created_at", "updated_at"},
		"filters": []string{"status", "change_type", "risk_level", "requested_by"},
		"endpoints": map[string]string{"list": "GET " + basePath, "get": "GET " + basePath + "/{id}", "create": "POST " + basePath, "update": "PUT/PATCH " + basePath + "/{id}", "approve": "POST " + basePath + "/{id}/approve", "rollback": "POST " + basePath + "/{id}/rollback", "delete": "DELETE " + basePath + "/{id}"},
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
		if len(parts) == 2 && r.Method == http.MethodPost {
			switch parts[1] { case "approve": srv.handleApprove(w, r, id); return; case "rollback": srv.handleRollback(w, r, id); return }
		}
		switch r.Method { case http.MethodGet: srv.handleGet(w, r, id); case http.MethodPut, http.MethodPatch: srv.handleUpdate(w, r, id); case http.MethodDelete: srv.handleDelete(w, r, id); default: writeJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "method not allowed"}) }
	})
	handler := securityHeaders(mux); log.Printf("%s listening on :%s", serviceName, port); log.Fatal(http.ListenAndServe(":"+port, handler))
}
