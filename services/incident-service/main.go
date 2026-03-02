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

// ---------------------------------------------------------------------------
// Constants
// ---------------------------------------------------------------------------

const (
	serviceName = "incident-service"
	moduleName  = "ERP-AIOps"
	basePath    = "/v1/aiops/incidents"
	dbName      = "erp_aiops"
	tableName   = "aiops_incidents"
	eventTopic  = "erp.aiops.incident"
	cacheTTL    = 30 * time.Second
)

var (
	validSeverities = map[string]bool{"critical": true, "high": true, "medium": true, "low": true, "info": true}
	validStatuses   = map[string]bool{"open": true, "acknowledged": true, "investigating": true, "resolved": true, "closed": true}
	validPriorities = map[string]bool{"P1": true, "P2": true, "P3": true, "P4": true, "P5": true}
)

// ---------------------------------------------------------------------------
// Entity
// ---------------------------------------------------------------------------

type incident struct {
	ID           string  `json:"id"`
	TenantID     string  `json:"tenant_id"`
	Title        string  `json:"title"`
	Description  *string `json:"description,omitempty"`
	Severity     string  `json:"severity"`
	Priority     string  `json:"priority"`
	Status       string  `json:"status"`
	Source       *string `json:"source,omitempty"`
	AssigneeID   *string `json:"assignee_id,omitempty"`
	TeamID       *string `json:"team_id,omitempty"`
	ServiceName  *string `json:"service_name,omitempty"`
	Environment  *string `json:"environment,omitempty"`
	AlertID      *string `json:"alert_id,omitempty"`
	RootCause    *string `json:"root_cause,omitempty"`
	Resolution   *string `json:"resolution,omitempty"`
	EscalatedAt  *string `json:"escalated_at,omitempty"`
	ResolvedAt   *string `json:"resolved_at,omitempty"`
	AcknowledgedAt *string `json:"acknowledged_at,omitempty"`
	Tags         *string `json:"tags,omitempty"`
	CreatedAt    string  `json:"created_at"`
	UpdatedAt    string  `json:"updated_at"`
}

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

func newID() string {
	b := make([]byte, 16)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}

func writeJSON(w http.ResponseWriter, code int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(v)
}

func readJSON(r *http.Request) (map[string]any, error) {
	body, err := io.ReadAll(io.LimitReader(r.Body, 1<<20))
	if err != nil {
		return nil, err
	}
	defer r.Body.Close()
	if len(body) == 0 {
		return nil, errors.New("empty body")
	}
	var m map[string]any
	if err := json.Unmarshal(body, &m); err != nil {
		return nil, err
	}
	return m, nil
}

func strPtr(v any) *string {
	if v == nil {
		return nil
	}
	s := fmt.Sprintf("%v", v)
	return &s
}

func strVal(p *string) string {
	if p == nil {
		return ""
	}
	return *p
}

// ---------------------------------------------------------------------------
// Cache
// ---------------------------------------------------------------------------

type cacheEntry struct {
	data      any
	expiresAt time.Time
}

type ttlCache struct {
	mu      sync.RWMutex
	entries map[string]cacheEntry
}

func newCache() *ttlCache {
	c := &ttlCache{entries: make(map[string]cacheEntry)}
	go func() {
		ticker := time.NewTicker(30 * time.Second)
		defer ticker.Stop()
		for range ticker.C {
			c.evict()
		}
	}()
	return c
}

func (c *ttlCache) get(key string) (any, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	e, ok := c.entries[key]
	if !ok || time.Now().After(e.expiresAt) {
		return nil, false
	}
	return e.data, true
}

func (c *ttlCache) set(key string, data any) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.entries[key] = cacheEntry{data: data, expiresAt: time.Now().Add(cacheTTL)}
}

func (c *ttlCache) invalidate(prefix string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	for k := range c.entries {
		if strings.HasPrefix(k, prefix) {
			delete(c.entries, k)
		}
	}
}

func (c *ttlCache) evict() {
	c.mu.Lock()
	defer c.mu.Unlock()
	now := time.Now()
	for k, e := range c.entries {
		if now.After(e.expiresAt) {
			delete(c.entries, k)
		}
	}
}

var requestCount atomic.Int64

func securityHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("X-Frame-Options", "DENY")
		w.Header().Set("X-XSS-Protection", "1; mode=block")
		w.Header().Set("Cache-Control", "no-store")
		w.Header().Set("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
		next.ServeHTTP(w, r)
	})
}

// ---------------------------------------------------------------------------
// Store
// ---------------------------------------------------------------------------

type store interface {
	List(ctx context.Context, tenantID, cursor string, limit int, filters map[string]string) ([]incident, string, error)
	GetByID(ctx context.Context, tenantID, id string) (*incident, error)
	Create(ctx context.Context, i *incident) error
	Update(ctx context.Context, i *incident) error
	Delete(ctx context.Context, tenantID, id string) error
}

type memoryStore struct {
	mu      sync.RWMutex
	records map[string]incident
}

func newMemoryStore() *memoryStore {
	return &memoryStore{records: make(map[string]incident)}
}

func (m *memoryStore) List(_ context.Context, tenantID, cursor string, limit int, filters map[string]string) ([]incident, string, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	var all []incident
	for _, i := range m.records {
		if i.TenantID != tenantID {
			continue
		}
		if v, ok := filters["severity"]; ok && i.Severity != v {
			continue
		}
		if v, ok := filters["status"]; ok && i.Status != v {
			continue
		}
		if v, ok := filters["priority"]; ok && i.Priority != v {
			continue
		}
		if v, ok := filters["assignee_id"]; ok && strVal(i.AssigneeID) != v {
			continue
		}
		all = append(all, i)
	}
	sort.Slice(all, func(a, b int) bool { return all[a].CreatedAt < all[b].CreatedAt })
	start := 0
	if cursor != "" {
		for idx, i := range all {
			if i.ID == cursor {
				start = idx + 1
				break
			}
		}
	}
	if start >= len(all) {
		return []incident{}, "", nil
	}
	end := start + limit
	if end > len(all) {
		end = len(all)
	}
	result := all[start:end]
	nextCursor := ""
	if end < len(all) {
		nextCursor = result[len(result)-1].ID
	}
	return result, nextCursor, nil
}

func (m *memoryStore) GetByID(_ context.Context, tenantID, id string) (*incident, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	i, ok := m.records[id]
	if !ok || i.TenantID != tenantID {
		return nil, errors.New("not found")
	}
	return &i, nil
}

func (m *memoryStore) Create(_ context.Context, i *incident) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.records[i.ID] = *i
	return nil
}

func (m *memoryStore) Update(_ context.Context, i *incident) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if _, ok := m.records[i.ID]; !ok {
		return errors.New("not found")
	}
	m.records[i.ID] = *i
	return nil
}

func (m *memoryStore) Delete(_ context.Context, tenantID, id string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	i, ok := m.records[id]
	if !ok || i.TenantID != tenantID {
		return errors.New("not found")
	}
	delete(m.records, id)
	return nil
}

// ---------------------------------------------------------------------------
// Postgres store
// ---------------------------------------------------------------------------

type postgresStore struct{ db *sql.DB }

func newPostgresStore(dsn string) (*postgresStore, error) {
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		return nil, fmt.Errorf("open db: %w", err)
	}
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(5 * time.Minute)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := db.PingContext(ctx); err != nil {
		return nil, fmt.Errorf("ping db: %w", err)
	}
	if _, err := db.ExecContext(ctx, createTableSQL); err != nil {
		return nil, fmt.Errorf("create table: %w", err)
	}
	return &postgresStore{db: db}, nil
}

const createTableSQL = `CREATE TABLE IF NOT EXISTS ` + tableName + ` (
	id TEXT PRIMARY KEY,
	tenant_id TEXT NOT NULL,
	title TEXT NOT NULL,
	description TEXT,
	severity TEXT CHECK (severity IN ('critical','high','medium','low','info')) DEFAULT 'medium',
	priority TEXT CHECK (priority IN ('P1','P2','P3','P4','P5')) DEFAULT 'P3',
	status TEXT CHECK (status IN ('open','acknowledged','investigating','resolved','closed')) DEFAULT 'open',
	source TEXT,
	assignee_id TEXT,
	team_id TEXT,
	service_name TEXT,
	environment TEXT,
	alert_id TEXT,
	root_cause TEXT,
	resolution TEXT,
	escalated_at TIMESTAMPTZ,
	resolved_at TIMESTAMPTZ,
	acknowledged_at TIMESTAMPTZ,
	tags TEXT,
	created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
	updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_` + tableName + `_tenant ON ` + tableName + ` (tenant_id);
CREATE INDEX IF NOT EXISTS idx_` + tableName + `_status ON ` + tableName + ` (tenant_id, status);
CREATE INDEX IF NOT EXISTS idx_` + tableName + `_severity ON ` + tableName + ` (tenant_id, severity);`

func (p *postgresStore) List(ctx context.Context, tenantID, cursor string, limit int, filters map[string]string) ([]incident, string, error) {
	query := `SELECT id,tenant_id,title,description,severity,priority,status,source,assignee_id,team_id,service_name,environment,alert_id,root_cause,resolution,escalated_at,resolved_at,acknowledged_at,tags,created_at,updated_at FROM ` + tableName + ` WHERE tenant_id=$1`
	args := []any{tenantID}
	idx := 2
	for _, f := range []struct{ key, col string }{{"severity", "severity"}, {"status", "status"}, {"priority", "priority"}, {"assignee_id", "assignee_id"}} {
		if v, ok := filters[f.key]; ok {
			query += fmt.Sprintf(" AND %s=$%d", f.col, idx)
			args = append(args, v)
			idx++
		}
	}
	if cursor != "" {
		query += fmt.Sprintf(" AND created_at>(SELECT created_at FROM "+tableName+" WHERE id=$%d)", idx)
		args = append(args, cursor)
		idx++
	}
	query += " ORDER BY created_at ASC"
	query += fmt.Sprintf(" LIMIT $%d", idx)
	args = append(args, limit+1)
	rows, err := p.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, "", fmt.Errorf("query: %w", err)
	}
	defer rows.Close()
	var results []incident
	for rows.Next() {
		var i incident
		var desc, src, assignee, team, svcName, env, alertID, rootCause, resolution, tags sql.NullString
		var escalatedAt, resolvedAt, ackedAt, createdAt, updatedAt sql.NullTime
		if err := rows.Scan(&i.ID, &i.TenantID, &i.Title, &desc, &i.Severity, &i.Priority, &i.Status, &src, &assignee, &team, &svcName, &env, &alertID, &rootCause, &resolution, &escalatedAt, &resolvedAt, &ackedAt, &tags, &createdAt, &updatedAt); err != nil {
			return nil, "", fmt.Errorf("scan: %w", err)
		}
		if desc.Valid { i.Description = &desc.String }
		if src.Valid { i.Source = &src.String }
		if assignee.Valid { i.AssigneeID = &assignee.String }
		if team.Valid { i.TeamID = &team.String }
		if svcName.Valid { i.ServiceName = &svcName.String }
		if env.Valid { i.Environment = &env.String }
		if alertID.Valid { i.AlertID = &alertID.String }
		if rootCause.Valid { i.RootCause = &rootCause.String }
		if resolution.Valid { i.Resolution = &resolution.String }
		if escalatedAt.Valid { s := escalatedAt.Time.Format(time.RFC3339); i.EscalatedAt = &s }
		if resolvedAt.Valid { s := resolvedAt.Time.Format(time.RFC3339); i.ResolvedAt = &s }
		if ackedAt.Valid { s := ackedAt.Time.Format(time.RFC3339); i.AcknowledgedAt = &s }
		if tags.Valid { i.Tags = &tags.String }
		if createdAt.Valid { i.CreatedAt = createdAt.Time.Format(time.RFC3339) }
		if updatedAt.Valid { i.UpdatedAt = updatedAt.Time.Format(time.RFC3339) }
		results = append(results, i)
	}
	nextCursor := ""
	if len(results) > limit {
		nextCursor = results[limit-1].ID
		results = results[:limit]
	}
	return results, nextCursor, nil
}

func (p *postgresStore) GetByID(ctx context.Context, tenantID, id string) (*incident, error) {
	query := `SELECT id,tenant_id,title,description,severity,priority,status,source,assignee_id,team_id,service_name,environment,alert_id,root_cause,resolution,escalated_at,resolved_at,acknowledged_at,tags,created_at,updated_at FROM ` + tableName + ` WHERE id=$1 AND tenant_id=$2`
	var i incident
	var desc, src, assignee, team, svcName, env, alertID, rootCause, resolution, tags sql.NullString
	var escalatedAt, resolvedAt, ackedAt, createdAt, updatedAt sql.NullTime
	err := p.db.QueryRowContext(ctx, query, id, tenantID).Scan(&i.ID, &i.TenantID, &i.Title, &desc, &i.Severity, &i.Priority, &i.Status, &src, &assignee, &team, &svcName, &env, &alertID, &rootCause, &resolution, &escalatedAt, &resolvedAt, &ackedAt, &tags, &createdAt, &updatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) { return nil, errors.New("not found") }
		return nil, fmt.Errorf("query row: %w", err)
	}
	if desc.Valid { i.Description = &desc.String }
	if src.Valid { i.Source = &src.String }
	if assignee.Valid { i.AssigneeID = &assignee.String }
	if team.Valid { i.TeamID = &team.String }
	if svcName.Valid { i.ServiceName = &svcName.String }
	if env.Valid { i.Environment = &env.String }
	if alertID.Valid { i.AlertID = &alertID.String }
	if rootCause.Valid { i.RootCause = &rootCause.String }
	if resolution.Valid { i.Resolution = &resolution.String }
	if escalatedAt.Valid { s := escalatedAt.Time.Format(time.RFC3339); i.EscalatedAt = &s }
	if resolvedAt.Valid { s := resolvedAt.Time.Format(time.RFC3339); i.ResolvedAt = &s }
	if ackedAt.Valid { s := ackedAt.Time.Format(time.RFC3339); i.AcknowledgedAt = &s }
	if tags.Valid { i.Tags = &tags.String }
	if createdAt.Valid { i.CreatedAt = createdAt.Time.Format(time.RFC3339) }
	if updatedAt.Valid { i.UpdatedAt = updatedAt.Time.Format(time.RFC3339) }
	return &i, nil
}

func (p *postgresStore) Create(ctx context.Context, i *incident) error {
	query := `INSERT INTO ` + tableName + ` (id,tenant_id,title,description,severity,priority,status,source,assignee_id,team_id,service_name,environment,alert_id,root_cause,resolution,escalated_at,resolved_at,acknowledged_at,tags,created_at,updated_at) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17,$18,$19,$20,$21)`
	_, err := p.db.ExecContext(ctx, query, i.ID, i.TenantID, i.Title, i.Description, i.Severity, i.Priority, i.Status, i.Source, i.AssigneeID, i.TeamID, i.ServiceName, i.Environment, i.AlertID, i.RootCause, i.Resolution, parseTimePtr(i.EscalatedAt), parseTimePtr(i.ResolvedAt), parseTimePtr(i.AcknowledgedAt), i.Tags, parseTime(i.CreatedAt), parseTime(i.UpdatedAt))
	return err
}

func (p *postgresStore) Update(ctx context.Context, i *incident) error {
	query := `UPDATE ` + tableName + ` SET title=$1,description=$2,severity=$3,priority=$4,status=$5,source=$6,assignee_id=$7,team_id=$8,service_name=$9,environment=$10,alert_id=$11,root_cause=$12,resolution=$13,escalated_at=$14,resolved_at=$15,acknowledged_at=$16,tags=$17,updated_at=$18 WHERE id=$19 AND tenant_id=$20`
	res, err := p.db.ExecContext(ctx, query, i.Title, i.Description, i.Severity, i.Priority, i.Status, i.Source, i.AssigneeID, i.TeamID, i.ServiceName, i.Environment, i.AlertID, i.RootCause, i.Resolution, parseTimePtr(i.EscalatedAt), parseTimePtr(i.ResolvedAt), parseTimePtr(i.AcknowledgedAt), i.Tags, parseTime(i.UpdatedAt), i.ID, i.TenantID)
	if err != nil { return err }
	n, _ := res.RowsAffected()
	if n == 0 { return errors.New("not found") }
	return nil
}

func (p *postgresStore) Delete(ctx context.Context, tenantID, id string) error {
	res, err := p.db.ExecContext(ctx, `DELETE FROM `+tableName+` WHERE id=$1 AND tenant_id=$2`, id, tenantID)
	if err != nil { return err }
	n, _ := res.RowsAffected()
	if n == 0 { return errors.New("not found") }
	return nil
}

func parseTime(s string) time.Time {
	t, err := time.Parse(time.RFC3339, s)
	if err != nil { return time.Now() }
	return t
}

func parseTimePtr(s *string) *time.Time {
	if s == nil { return nil }
	t, err := time.Parse(time.RFC3339, *s)
	if err != nil { return nil }
	return &t
}

// ---------------------------------------------------------------------------
// Handlers
// ---------------------------------------------------------------------------

type server struct {
	store store
	cache *ttlCache
}

func (s *server) handleList(w http.ResponseWriter, r *http.Request) {
	tenantID := r.Header.Get("X-Tenant-ID")
	cursor := r.URL.Query().Get("cursor")
	limit := 20
	if v, err := strconv.Atoi(r.URL.Query().Get("limit")); err == nil && v > 0 && v <= 100 {
		limit = v
	}
	filters := make(map[string]string)
	for _, key := range []string{"severity", "status", "priority", "assignee_id"} {
		if v := r.URL.Query().Get(key); v != "" {
			filters[key] = v
		}
	}
	cacheKey := fmt.Sprintf("list:%s:%s:%d:%v", tenantID, cursor, limit, filters)
	if cached, ok := s.cache.get(cacheKey); ok {
		writeJSON(w, http.StatusOK, cached)
		return
	}
	items, nextCursor, err := s.store.List(r.Context(), tenantID, cursor, limit, filters)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	resp := map[string]any{"items": items, "next_cursor": nextCursor, "limit": limit, "count": len(items), "event_topic": eventTopic + ".listed"}
	s.cache.set(cacheKey, resp)
	writeJSON(w, http.StatusOK, resp)
}

func (s *server) handleGet(w http.ResponseWriter, r *http.Request, id string) {
	tenantID := r.Header.Get("X-Tenant-ID")
	cacheKey := fmt.Sprintf("get:%s:%s", tenantID, id)
	if cached, ok := s.cache.get(cacheKey); ok {
		writeJSON(w, http.StatusOK, cached)
		return
	}
	item, err := s.store.GetByID(r.Context(), tenantID, id)
	if err != nil {
		if err.Error() == "not found" { writeJSON(w, http.StatusNotFound, map[string]string{"error": "not found"}); return }
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	resp := map[string]any{"item": item, "event_topic": eventTopic + ".read"}
	s.cache.set(cacheKey, resp)
	writeJSON(w, http.StatusOK, resp)
}

func (s *server) handleCreate(w http.ResponseWriter, r *http.Request) {
	tenantID := r.Header.Get("X-Tenant-ID")
	body, err := readJSON(r)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid JSON: " + err.Error()})
		return
	}
	title, _ := body["title"].(string)
	if title == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "title is required"})
		return
	}
	now := time.Now().UTC().Format(time.RFC3339)
	i := &incident{
		ID: newID(), TenantID: tenantID, Title: title, Description: strPtr(body["description"]),
		Severity: "medium", Priority: "P3", Status: "open", Source: strPtr(body["source"]),
		AssigneeID: strPtr(body["assignee_id"]), TeamID: strPtr(body["team_id"]),
		ServiceName: strPtr(body["service_name"]), Environment: strPtr(body["environment"]),
		AlertID: strPtr(body["alert_id"]), Tags: strPtr(body["tags"]),
		CreatedAt: now, UpdatedAt: now,
	}
	if v, ok := body["severity"].(string); ok && validSeverities[v] { i.Severity = v }
	if v, ok := body["priority"].(string); ok && validPriorities[v] { i.Priority = v }
	if v, ok := body["status"].(string); ok && validStatuses[v] { i.Status = v }
	if err := s.store.Create(r.Context(), i); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	s.cache.invalidate("list:" + tenantID)
	writeJSON(w, http.StatusCreated, map[string]any{"item": i, "event_topic": eventTopic + ".created"})
}

func (s *server) handleUpdate(w http.ResponseWriter, r *http.Request, id string) {
	tenantID := r.Header.Get("X-Tenant-ID")
	existing, err := s.store.GetByID(r.Context(), tenantID, id)
	if err != nil {
		if err.Error() == "not found" { writeJSON(w, http.StatusNotFound, map[string]string{"error": "not found"}); return }
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	body, err := readJSON(r)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid JSON: " + err.Error()})
		return
	}
	if v, ok := body["title"].(string); ok && v != "" { existing.Title = v }
	if v, exists := body["description"]; exists { existing.Description = strPtr(v) }
	if v, ok := body["severity"].(string); ok && validSeverities[v] { existing.Severity = v }
	if v, ok := body["priority"].(string); ok && validPriorities[v] { existing.Priority = v }
	if v, ok := body["status"].(string); ok && validStatuses[v] { existing.Status = v }
	if v, exists := body["source"]; exists { existing.Source = strPtr(v) }
	if v, exists := body["assignee_id"]; exists { existing.AssigneeID = strPtr(v) }
	if v, exists := body["team_id"]; exists { existing.TeamID = strPtr(v) }
	if v, exists := body["service_name"]; exists { existing.ServiceName = strPtr(v) }
	if v, exists := body["environment"]; exists { existing.Environment = strPtr(v) }
	if v, exists := body["alert_id"]; exists { existing.AlertID = strPtr(v) }
	if v, exists := body["root_cause"]; exists { existing.RootCause = strPtr(v) }
	if v, exists := body["resolution"]; exists { existing.Resolution = strPtr(v) }
	if v, exists := body["tags"]; exists { existing.Tags = strPtr(v) }
	now := time.Now().UTC().Format(time.RFC3339)
	if existing.Status == "acknowledged" && existing.AcknowledgedAt == nil { existing.AcknowledgedAt = &now }
	if existing.Status == "resolved" && existing.ResolvedAt == nil { existing.ResolvedAt = &now }
	existing.UpdatedAt = now
	if err := s.store.Update(r.Context(), existing); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	s.cache.invalidate("list:" + tenantID)
	s.cache.invalidate("get:" + tenantID + ":" + id)
	writeJSON(w, http.StatusOK, map[string]any{"item": existing, "event_topic": eventTopic + ".updated"})
}

func (s *server) handleEscalate(w http.ResponseWriter, r *http.Request, id string) {
	tenantID := r.Header.Get("X-Tenant-ID")
	existing, err := s.store.GetByID(r.Context(), tenantID, id)
	if err != nil {
		if err.Error() == "not found" { writeJSON(w, http.StatusNotFound, map[string]string{"error": "not found"}); return }
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	now := time.Now().UTC().Format(time.RFC3339)
	existing.EscalatedAt = &now
	if existing.Priority == "P5" { existing.Priority = "P4" } else if existing.Priority == "P4" { existing.Priority = "P3" } else if existing.Priority == "P3" { existing.Priority = "P2" } else if existing.Priority == "P2" { existing.Priority = "P1" }
	existing.UpdatedAt = now
	if err := s.store.Update(r.Context(), existing); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	s.cache.invalidate("list:" + tenantID)
	s.cache.invalidate("get:" + tenantID + ":" + id)
	writeJSON(w, http.StatusOK, map[string]any{"item": existing, "event_topic": eventTopic + ".escalated"})
}

func (s *server) handleDelete(w http.ResponseWriter, r *http.Request, id string) {
	tenantID := r.Header.Get("X-Tenant-ID")
	if err := s.store.Delete(r.Context(), tenantID, id); err != nil {
		if err.Error() == "not found" { writeJSON(w, http.StatusNotFound, map[string]string{"error": "not found"}); return }
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	s.cache.invalidate("list:" + tenantID)
	s.cache.invalidate("get:" + tenantID + ":" + id)
	writeJSON(w, http.StatusOK, map[string]any{"deleted": true, "id": id, "event_topic": eventTopic + ".deleted"})
}

func handleExplain(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]any{
		"service": serviceName, "module": moduleName, "base_path": basePath,
		"database": dbName, "table": tableName, "event_topic": eventTopic, "entity": "incident",
		"fields":  []string{"id", "tenant_id", "title", "description", "severity", "priority", "status", "source", "assignee_id", "team_id", "service_name", "environment", "alert_id", "root_cause", "resolution", "escalated_at", "resolved_at", "acknowledged_at", "tags", "created_at", "updated_at"},
		"filters": []string{"severity", "status", "priority", "assignee_id"},
		"endpoints": map[string]string{
			"list": "GET " + basePath, "get": "GET " + basePath + "/{id}",
			"create": "POST " + basePath, "update": "PUT/PATCH " + basePath + "/{id}",
			"escalate": "POST " + basePath + "/{id}/escalate", "delete": "DELETE " + basePath + "/{id}",
		},
	})
}

// ---------------------------------------------------------------------------
// Main
// ---------------------------------------------------------------------------

func main() {
	port := os.Getenv("PORT")
	if port == "" { port = "8080" }
	var st store
	dsn := os.Getenv("DATABASE_URL")
	if dsn != "" {
		pg, err := newPostgresStore(dsn)
		if err != nil { log.Fatalf("postgres: %v", err) }
		st = pg
		log.Println("Using PostgreSQL store")
	} else {
		st = newMemoryStore()
		log.Println("Using in-memory store (set DATABASE_URL for PostgreSQL)")
	}
	cache := newCache()
	srv := &server{store: st, cache: cache}
	mux := http.NewServeMux()

	mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, map[string]string{"status": "healthy", "module": moduleName, "service": serviceName})
	})
	mux.HandleFunc("/readyz", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, map[string]string{"status": "ready"})
	})
	mux.HandleFunc("/metrics", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, map[string]any{"requests_total": requestCount.Load(), "service": serviceName})
	})
	mux.HandleFunc(basePath+"/_explain", handleExplain)

	mux.HandleFunc(basePath, func(w http.ResponseWriter, r *http.Request) {
		requestCount.Add(1)
		tenantID := r.Header.Get("X-Tenant-ID")
		if tenantID == "" { writeJSON(w, http.StatusBadRequest, map[string]string{"error": "missing X-Tenant-ID header"}); return }
		switch r.Method {
		case http.MethodGet: srv.handleList(w, r)
		case http.MethodPost: srv.handleCreate(w, r)
		default: writeJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "method not allowed"})
		}
	})
	mux.HandleFunc(basePath+"/", func(w http.ResponseWriter, r *http.Request) {
		requestCount.Add(1)
		tenantID := r.Header.Get("X-Tenant-ID")
		if tenantID == "" { writeJSON(w, http.StatusBadRequest, map[string]string{"error": "missing X-Tenant-ID header"}); return }
		sub := strings.TrimPrefix(r.URL.Path, basePath+"/")
		parts := strings.SplitN(sub, "/", 2)
		id := parts[0]
		if id == "" { writeJSON(w, http.StatusBadRequest, map[string]string{"error": "missing id"}); return }
		if len(parts) == 2 && parts[1] == "escalate" && r.Method == http.MethodPost {
			srv.handleEscalate(w, r, id); return
		}
		switch r.Method {
		case http.MethodGet: srv.handleGet(w, r, id)
		case http.MethodPut, http.MethodPatch: srv.handleUpdate(w, r, id)
		case http.MethodDelete: srv.handleDelete(w, r, id)
		default: writeJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "method not allowed"})
		}
	})

	handler := securityHeaders(mux)
	log.Printf("%s listening on :%s", serviceName, port)
	log.Fatal(http.ListenAndServe(":"+port, handler))
}
