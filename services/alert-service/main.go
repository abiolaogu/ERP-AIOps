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
	serviceName = "alert-service"
	moduleName  = "ERP-AIOps"
	basePath    = "/v1/aiops/alerts"
	dbName      = "erp_aiops"
	tableName   = "aiops_alerts"
	eventTopic  = "erp.aiops.alert"
	cacheTTL    = 30 * time.Second
)

var (
	validSeverities = map[string]bool{"critical": true, "warning": true, "info": true}
	validStatuses   = map[string]bool{"firing": true, "acknowledged": true, "silenced": true, "resolved": true}
)

type alertRule struct {
	ID          string  `json:"id"`
	TenantID    string  `json:"tenant_id"`
	Name        string  `json:"name"`
	Description *string `json:"description,omitempty"`
	Metric      string  `json:"metric"`
	Condition   string  `json:"condition"`
	Threshold   float64 `json:"threshold"`
	Duration    string  `json:"duration"`
	Severity    string  `json:"severity"`
	Status      string  `json:"status"`
	Labels      *string `json:"labels,omitempty"`
	Channel     *string `json:"notification_channel,omitempty"`
	ServiceName *string `json:"service_name,omitempty"`
	Environment *string `json:"environment,omitempty"`
	FiredAt     *string `json:"fired_at,omitempty"`
	AckedAt     *string `json:"acknowledged_at,omitempty"`
	SilencedAt  *string `json:"silenced_at,omitempty"`
	ResolvedAt  *string `json:"resolved_at,omitempty"`
	SilenceUntil *string `json:"silence_until,omitempty"`
	CreatedAt   string  `json:"created_at"`
	UpdatedAt   string  `json:"updated_at"`
}

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
	if err != nil { return nil, err }
	defer r.Body.Close()
	if len(body) == 0 { return nil, errors.New("empty body") }
	var m map[string]any
	if err := json.Unmarshal(body, &m); err != nil { return nil, err }
	return m, nil
}

func strPtr(v any) *string {
	if v == nil { return nil }
	s := fmt.Sprintf("%v", v)
	return &s
}

func strVal(p *string) string {
	if p == nil { return "" }
	return *p
}

type cacheEntry struct { data any; expiresAt time.Time }
type ttlCache struct { mu sync.RWMutex; entries map[string]cacheEntry }

func newCache() *ttlCache {
	c := &ttlCache{entries: make(map[string]cacheEntry)}
	go func() { t := time.NewTicker(30 * time.Second); defer t.Stop(); for range t.C { c.evict() } }()
	return c
}
func (c *ttlCache) get(key string) (any, bool) { c.mu.RLock(); defer c.mu.RUnlock(); e, ok := c.entries[key]; if !ok || time.Now().After(e.expiresAt) { return nil, false }; return e.data, true }
func (c *ttlCache) set(key string, data any) { c.mu.Lock(); defer c.mu.Unlock(); c.entries[key] = cacheEntry{data: data, expiresAt: time.Now().Add(cacheTTL)} }
func (c *ttlCache) invalidate(prefix string) { c.mu.Lock(); defer c.mu.Unlock(); for k := range c.entries { if strings.HasPrefix(k, prefix) { delete(c.entries, k) } } }
func (c *ttlCache) evict() { c.mu.Lock(); defer c.mu.Unlock(); now := time.Now(); for k, e := range c.entries { if now.After(e.expiresAt) { delete(c.entries, k) } } }

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
	List(ctx context.Context, tenantID, cursor string, limit int, filters map[string]string) ([]alertRule, string, error)
	GetByID(ctx context.Context, tenantID, id string) (*alertRule, error)
	Create(ctx context.Context, a *alertRule) error
	Update(ctx context.Context, a *alertRule) error
	Delete(ctx context.Context, tenantID, id string) error
}

type memoryStore struct { mu sync.RWMutex; records map[string]alertRule }

func newMemoryStore() *memoryStore { return &memoryStore{records: make(map[string]alertRule)} }

func (m *memoryStore) List(_ context.Context, tenantID, cursor string, limit int, filters map[string]string) ([]alertRule, string, error) {
	m.mu.RLock(); defer m.mu.RUnlock()
	var all []alertRule
	for _, a := range m.records {
		if a.TenantID != tenantID { continue }
		if v, ok := filters["severity"]; ok && a.Severity != v { continue }
		if v, ok := filters["status"]; ok && a.Status != v { continue }
		if v, ok := filters["metric"]; ok && a.Metric != v { continue }
		all = append(all, a)
	}
	sort.Slice(all, func(i, j int) bool { return all[i].CreatedAt < all[j].CreatedAt })
	start := 0
	if cursor != "" { for idx, a := range all { if a.ID == cursor { start = idx + 1; break } } }
	if start >= len(all) { return []alertRule{}, "", nil }
	end := start + limit; if end > len(all) { end = len(all) }
	result := all[start:end]
	nextCursor := ""; if end < len(all) { nextCursor = result[len(result)-1].ID }
	return result, nextCursor, nil
}

func (m *memoryStore) GetByID(_ context.Context, tenantID, id string) (*alertRule, error) {
	m.mu.RLock(); defer m.mu.RUnlock()
	a, ok := m.records[id]; if !ok || a.TenantID != tenantID { return nil, errors.New("not found") }
	return &a, nil
}

func (m *memoryStore) Create(_ context.Context, a *alertRule) error { m.mu.Lock(); defer m.mu.Unlock(); m.records[a.ID] = *a; return nil }
func (m *memoryStore) Update(_ context.Context, a *alertRule) error { m.mu.Lock(); defer m.mu.Unlock(); if _, ok := m.records[a.ID]; !ok { return errors.New("not found") }; m.records[a.ID] = *a; return nil }
func (m *memoryStore) Delete(_ context.Context, tenantID, id string) error { m.mu.Lock(); defer m.mu.Unlock(); a, ok := m.records[id]; if !ok || a.TenantID != tenantID { return errors.New("not found") }; delete(m.records, id); return nil }

// ---------------------------------------------------------------------------
// Postgres store
// ---------------------------------------------------------------------------

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
	id TEXT PRIMARY KEY,
	tenant_id TEXT NOT NULL,
	name TEXT NOT NULL,
	description TEXT,
	metric TEXT NOT NULL,
	condition TEXT NOT NULL,
	threshold DOUBLE PRECISION NOT NULL DEFAULT 0,
	duration TEXT DEFAULT '5m',
	severity TEXT CHECK (severity IN ('critical','warning','info')) DEFAULT 'warning',
	status TEXT CHECK (status IN ('firing','acknowledged','silenced','resolved')) DEFAULT 'resolved',
	labels TEXT,
	notification_channel TEXT,
	service_name TEXT,
	environment TEXT,
	fired_at TIMESTAMPTZ,
	acknowledged_at TIMESTAMPTZ,
	silenced_at TIMESTAMPTZ,
	resolved_at TIMESTAMPTZ,
	silence_until TIMESTAMPTZ,
	created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
	updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_` + tableName + `_tenant ON ` + tableName + ` (tenant_id);
CREATE INDEX IF NOT EXISTS idx_` + tableName + `_status ON ` + tableName + ` (tenant_id, status);`

func (p *postgresStore) List(ctx context.Context, tenantID, cursor string, limit int, filters map[string]string) ([]alertRule, string, error) {
	query := `SELECT id,tenant_id,name,description,metric,condition,threshold,duration,severity,status,labels,notification_channel,service_name,environment,fired_at,acknowledged_at,silenced_at,resolved_at,silence_until,created_at,updated_at FROM ` + tableName + ` WHERE tenant_id=$1`
	args := []any{tenantID}; idx := 2
	for _, f := range []struct{ k, c string }{{"severity", "severity"}, {"status", "status"}, {"metric", "metric"}} {
		if v, ok := filters[f.k]; ok { query += fmt.Sprintf(" AND %s=$%d", f.c, idx); args = append(args, v); idx++ }
	}
	if cursor != "" { query += fmt.Sprintf(" AND created_at>(SELECT created_at FROM "+tableName+" WHERE id=$%d)", idx); args = append(args, cursor); idx++ }
	query += " ORDER BY created_at ASC" + fmt.Sprintf(" LIMIT $%d", idx); args = append(args, limit+1)
	rows, err := p.db.QueryContext(ctx, query, args...); if err != nil { return nil, "", fmt.Errorf("query: %w", err) }
	defer rows.Close()
	var results []alertRule
	for rows.Next() {
		var a alertRule
		var desc, labels, channel, svcName, env sql.NullString
		var firedAt, ackedAt, silencedAt, resolvedAt, silenceUntil, createdAt, updatedAt sql.NullTime
		if err := rows.Scan(&a.ID, &a.TenantID, &a.Name, &desc, &a.Metric, &a.Condition, &a.Threshold, &a.Duration, &a.Severity, &a.Status, &labels, &channel, &svcName, &env, &firedAt, &ackedAt, &silencedAt, &resolvedAt, &silenceUntil, &createdAt, &updatedAt); err != nil {
			return nil, "", fmt.Errorf("scan: %w", err)
		}
		if desc.Valid { a.Description = &desc.String }
		if labels.Valid { a.Labels = &labels.String }
		if channel.Valid { a.Channel = &channel.String }
		if svcName.Valid { a.ServiceName = &svcName.String }
		if env.Valid { a.Environment = &env.String }
		if firedAt.Valid { s := firedAt.Time.Format(time.RFC3339); a.FiredAt = &s }
		if ackedAt.Valid { s := ackedAt.Time.Format(time.RFC3339); a.AckedAt = &s }
		if silencedAt.Valid { s := silencedAt.Time.Format(time.RFC3339); a.SilencedAt = &s }
		if resolvedAt.Valid { s := resolvedAt.Time.Format(time.RFC3339); a.ResolvedAt = &s }
		if silenceUntil.Valid { s := silenceUntil.Time.Format(time.RFC3339); a.SilenceUntil = &s }
		if createdAt.Valid { a.CreatedAt = createdAt.Time.Format(time.RFC3339) }
		if updatedAt.Valid { a.UpdatedAt = updatedAt.Time.Format(time.RFC3339) }
		results = append(results, a)
	}
	nextCursor := ""; if len(results) > limit { nextCursor = results[limit-1].ID; results = results[:limit] }
	return results, nextCursor, nil
}

func (p *postgresStore) GetByID(ctx context.Context, tenantID, id string) (*alertRule, error) {
	query := `SELECT id,tenant_id,name,description,metric,condition,threshold,duration,severity,status,labels,notification_channel,service_name,environment,fired_at,acknowledged_at,silenced_at,resolved_at,silence_until,created_at,updated_at FROM ` + tableName + ` WHERE id=$1 AND tenant_id=$2`
	var a alertRule
	var desc, labels, channel, svcName, env sql.NullString
	var firedAt, ackedAt, silencedAt, resolvedAt, silenceUntil, createdAt, updatedAt sql.NullTime
	err := p.db.QueryRowContext(ctx, query, id, tenantID).Scan(&a.ID, &a.TenantID, &a.Name, &desc, &a.Metric, &a.Condition, &a.Threshold, &a.Duration, &a.Severity, &a.Status, &labels, &channel, &svcName, &env, &firedAt, &ackedAt, &silencedAt, &resolvedAt, &silenceUntil, &createdAt, &updatedAt)
	if err != nil { if errors.Is(err, sql.ErrNoRows) { return nil, errors.New("not found") }; return nil, fmt.Errorf("query row: %w", err) }
	if desc.Valid { a.Description = &desc.String }
	if labels.Valid { a.Labels = &labels.String }
	if channel.Valid { a.Channel = &channel.String }
	if svcName.Valid { a.ServiceName = &svcName.String }
	if env.Valid { a.Environment = &env.String }
	if firedAt.Valid { s := firedAt.Time.Format(time.RFC3339); a.FiredAt = &s }
	if ackedAt.Valid { s := ackedAt.Time.Format(time.RFC3339); a.AckedAt = &s }
	if silencedAt.Valid { s := silencedAt.Time.Format(time.RFC3339); a.SilencedAt = &s }
	if resolvedAt.Valid { s := resolvedAt.Time.Format(time.RFC3339); a.ResolvedAt = &s }
	if silenceUntil.Valid { s := silenceUntil.Time.Format(time.RFC3339); a.SilenceUntil = &s }
	if createdAt.Valid { a.CreatedAt = createdAt.Time.Format(time.RFC3339) }
	if updatedAt.Valid { a.UpdatedAt = updatedAt.Time.Format(time.RFC3339) }
	return &a, nil
}

func (p *postgresStore) Create(ctx context.Context, a *alertRule) error {
	query := `INSERT INTO ` + tableName + ` (id,tenant_id,name,description,metric,condition,threshold,duration,severity,status,labels,notification_channel,service_name,environment,fired_at,acknowledged_at,silenced_at,resolved_at,silence_until,created_at,updated_at) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17,$18,$19,$20,$21)`
	_, err := p.db.ExecContext(ctx, query, a.ID, a.TenantID, a.Name, a.Description, a.Metric, a.Condition, a.Threshold, a.Duration, a.Severity, a.Status, a.Labels, a.Channel, a.ServiceName, a.Environment, parseTimePtr(a.FiredAt), parseTimePtr(a.AckedAt), parseTimePtr(a.SilencedAt), parseTimePtr(a.ResolvedAt), parseTimePtr(a.SilenceUntil), parseTime(a.CreatedAt), parseTime(a.UpdatedAt))
	return err
}

func (p *postgresStore) Update(ctx context.Context, a *alertRule) error {
	query := `UPDATE ` + tableName + ` SET name=$1,description=$2,metric=$3,condition=$4,threshold=$5,duration=$6,severity=$7,status=$8,labels=$9,notification_channel=$10,service_name=$11,environment=$12,fired_at=$13,acknowledged_at=$14,silenced_at=$15,resolved_at=$16,silence_until=$17,updated_at=$18 WHERE id=$19 AND tenant_id=$20`
	res, err := p.db.ExecContext(ctx, query, a.Name, a.Description, a.Metric, a.Condition, a.Threshold, a.Duration, a.Severity, a.Status, a.Labels, a.Channel, a.ServiceName, a.Environment, parseTimePtr(a.FiredAt), parseTimePtr(a.AckedAt), parseTimePtr(a.SilencedAt), parseTimePtr(a.ResolvedAt), parseTimePtr(a.SilenceUntil), parseTime(a.UpdatedAt), a.ID, a.TenantID)
	if err != nil { return err }; n, _ := res.RowsAffected(); if n == 0 { return errors.New("not found") }; return nil
}

func (p *postgresStore) Delete(ctx context.Context, tenantID, id string) error {
	res, err := p.db.ExecContext(ctx, `DELETE FROM `+tableName+` WHERE id=$1 AND tenant_id=$2`, id, tenantID)
	if err != nil { return err }; n, _ := res.RowsAffected(); if n == 0 { return errors.New("not found") }; return nil
}

func parseTime(s string) time.Time { t, err := time.Parse(time.RFC3339, s); if err != nil { return time.Now() }; return t }
func parseTimePtr(s *string) *time.Time { if s == nil { return nil }; t, err := time.Parse(time.RFC3339, *s); if err != nil { return nil }; return &t }

// ---------------------------------------------------------------------------
// Handlers
// ---------------------------------------------------------------------------

type server struct { store store; cache *ttlCache }

func (s *server) handleList(w http.ResponseWriter, r *http.Request) {
	tenantID := r.Header.Get("X-Tenant-ID")
	cursor := r.URL.Query().Get("cursor")
	limit := 20; if v, err := strconv.Atoi(r.URL.Query().Get("limit")); err == nil && v > 0 && v <= 100 { limit = v }
	filters := make(map[string]string)
	for _, key := range []string{"severity", "status", "metric"} { if v := r.URL.Query().Get(key); v != "" { filters[key] = v } }
	cacheKey := fmt.Sprintf("list:%s:%s:%d:%v", tenantID, cursor, limit, filters)
	if cached, ok := s.cache.get(cacheKey); ok { writeJSON(w, http.StatusOK, cached); return }
	items, nextCursor, err := s.store.List(r.Context(), tenantID, cursor, limit, filters)
	if err != nil { writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()}); return }
	resp := map[string]any{"items": items, "next_cursor": nextCursor, "limit": limit, "count": len(items), "event_topic": eventTopic + ".listed"}
	s.cache.set(cacheKey, resp); writeJSON(w, http.StatusOK, resp)
}

func (s *server) handleGet(w http.ResponseWriter, r *http.Request, id string) {
	tenantID := r.Header.Get("X-Tenant-ID")
	cacheKey := fmt.Sprintf("get:%s:%s", tenantID, id)
	if cached, ok := s.cache.get(cacheKey); ok { writeJSON(w, http.StatusOK, cached); return }
	item, err := s.store.GetByID(r.Context(), tenantID, id)
	if err != nil { if err.Error() == "not found" { writeJSON(w, http.StatusNotFound, map[string]string{"error": "not found"}); return }; writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()}); return }
	resp := map[string]any{"item": item, "event_topic": eventTopic + ".read"}
	s.cache.set(cacheKey, resp); writeJSON(w, http.StatusOK, resp)
}

func (s *server) handleCreate(w http.ResponseWriter, r *http.Request) {
	tenantID := r.Header.Get("X-Tenant-ID")
	body, err := readJSON(r)
	if err != nil { writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid JSON: " + err.Error()}); return }
	name, _ := body["name"].(string); metric, _ := body["metric"].(string); cond, _ := body["condition"].(string)
	if name == "" || metric == "" || cond == "" { writeJSON(w, http.StatusBadRequest, map[string]string{"error": "name, metric, and condition are required"}); return }
	now := time.Now().UTC().Format(time.RFC3339)
	a := &alertRule{ID: newID(), TenantID: tenantID, Name: name, Description: strPtr(body["description"]), Metric: metric, Condition: cond, Threshold: 0, Duration: "5m", Severity: "warning", Status: "resolved", Labels: strPtr(body["labels"]), Channel: strPtr(body["notification_channel"]), ServiceName: strPtr(body["service_name"]), Environment: strPtr(body["environment"]), CreatedAt: now, UpdatedAt: now}
	if v, ok := body["threshold"].(float64); ok { a.Threshold = v }
	if v, ok := body["duration"].(string); ok && v != "" { a.Duration = v }
	if v, ok := body["severity"].(string); ok && validSeverities[v] { a.Severity = v }
	if v, ok := body["status"].(string); ok && validStatuses[v] { a.Status = v }
	if err := s.store.Create(r.Context(), a); err != nil { writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()}); return }
	s.cache.invalidate("list:" + tenantID)
	writeJSON(w, http.StatusCreated, map[string]any{"item": a, "event_topic": eventTopic + ".created"})
}

func (s *server) handleUpdate(w http.ResponseWriter, r *http.Request, id string) {
	tenantID := r.Header.Get("X-Tenant-ID")
	existing, err := s.store.GetByID(r.Context(), tenantID, id)
	if err != nil { if err.Error() == "not found" { writeJSON(w, http.StatusNotFound, map[string]string{"error": "not found"}); return }; writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()}); return }
	body, err := readJSON(r)
	if err != nil { writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid JSON: " + err.Error()}); return }
	if v, ok := body["name"].(string); ok && v != "" { existing.Name = v }
	if v, exists := body["description"]; exists { existing.Description = strPtr(v) }
	if v, ok := body["metric"].(string); ok && v != "" { existing.Metric = v }
	if v, ok := body["condition"].(string); ok && v != "" { existing.Condition = v }
	if v, ok := body["threshold"].(float64); ok { existing.Threshold = v }
	if v, ok := body["duration"].(string); ok && v != "" { existing.Duration = v }
	if v, ok := body["severity"].(string); ok && validSeverities[v] { existing.Severity = v }
	if v, ok := body["status"].(string); ok && validStatuses[v] { existing.Status = v }
	if v, exists := body["labels"]; exists { existing.Labels = strPtr(v) }
	if v, exists := body["notification_channel"]; exists { existing.Channel = strPtr(v) }
	if v, exists := body["service_name"]; exists { existing.ServiceName = strPtr(v) }
	if v, exists := body["environment"]; exists { existing.Environment = strPtr(v) }
	now := time.Now().UTC().Format(time.RFC3339)
	if existing.Status == "firing" && existing.FiredAt == nil { existing.FiredAt = &now }
	if existing.Status == "acknowledged" && existing.AckedAt == nil { existing.AckedAt = &now }
	if existing.Status == "silenced" && existing.SilencedAt == nil { existing.SilencedAt = &now }
	if existing.Status == "resolved" && existing.ResolvedAt == nil { existing.ResolvedAt = &now }
	existing.UpdatedAt = now
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
	writeJSON(w, http.StatusOK, map[string]any{
		"service": serviceName, "module": moduleName, "base_path": basePath, "database": dbName, "table": tableName, "event_topic": eventTopic, "entity": "alertRule",
		"fields": []string{"id", "tenant_id", "name", "description", "metric", "condition", "threshold", "duration", "severity", "status", "labels", "notification_channel", "service_name", "environment", "fired_at", "acknowledged_at", "silenced_at", "resolved_at", "silence_until", "created_at", "updated_at"},
		"filters": []string{"severity", "status", "metric"},
	})
}

func main() {
	port := os.Getenv("PORT"); if port == "" { port = "8080" }
	var st store
	dsn := os.Getenv("DATABASE_URL")
	if dsn != "" { pg, err := newPostgresStore(dsn); if err != nil { log.Fatalf("postgres: %v", err) }; st = pg; log.Println("Using PostgreSQL store") } else { st = newMemoryStore(); log.Println("Using in-memory store (set DATABASE_URL for PostgreSQL)") }
	cache := newCache(); srv := &server{store: st, cache: cache}
	mux := http.NewServeMux()
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
	handler := securityHeaders(mux)
	log.Printf("%s listening on :%s", serviceName, port)
	log.Fatal(http.ListenAndServe(":"+port, handler))
}
