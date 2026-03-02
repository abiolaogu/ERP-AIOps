package aiddguardrail

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestEvaluateBlocksCrossTenant(t *testing.T) {
	p := DefaultPolicy()
	decision := p.Evaluate(Action{
		Name:         "tenant.export",
		TenantID:     "tenant-a",
		Confidence:   0.99,
		BlastRadius:  1,
		CrossTenant:  true,
		Destructive:  false,
		EstimatedCostUSD: 100,
	})

	if decision.Allowed {
		t.Fatalf("expected denied decision")
	}
	if decision.Mode != ModeProtected {
		t.Fatalf("expected protected mode, got %s", decision.Mode)
	}
}

func TestEvaluateRequiresApproval(t *testing.T) {
	p := DefaultPolicy()
	decision := p.Evaluate(Action{
		Name:             "bulk.update",
		TenantID:         "tenant-a",
		Confidence:       0.8,
		BlastRadius:      800,
		EstimatedCostUSD: 12000,
	})

	if !decision.Allowed {
		t.Fatalf("expected action to be allowed in supervised mode")
	}
	if !decision.RequiresApproval {
		t.Fatalf("expected approval requirement")
	}
}

func TestMiddlewareRejectsMissingTenant(t *testing.T) {
	p := DefaultPolicy()
	handler := Middleware(p)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}))

	req := httptest.NewRequest(http.MethodPost, "/dangerous", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusForbidden {
		t.Fatalf("expected 403, got %d", rec.Code)
	}
}

func TestMiddlewareRequiresApproval(t *testing.T) {
	p := DefaultPolicy()
	handler := Middleware(p)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}))

	req := httptest.NewRequest(http.MethodPost, "/bulk", nil)
	req.Header.Set("X-Tenant-ID", "tenant-a")
	req.Header.Set("X-AIDD-Confidence", "0.8")
	req.Header.Set("X-AIDD-Blast-Radius", "800")
	req.Header.Set("X-AIDD-Estimated-Cost-USD", "12000")
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusPreconditionFailed {
		t.Fatalf("expected 412, got %d", rec.Code)
	}
}

func TestMiddlewareAllowsApproved(t *testing.T) {
	p := DefaultPolicy()
	handler := Middleware(p)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}))

	req := httptest.NewRequest(http.MethodPost, "/bulk", nil)
	req.Header.Set("X-Tenant-ID", "tenant-a")
	req.Header.Set("X-AIDD-Confidence", "0.8")
	req.Header.Set("X-AIDD-Blast-Radius", "800")
	req.Header.Set("X-AIDD-Estimated-Cost-USD", "12000")
	req.Header.Set("X-AIDD-Approval", "approved")
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusNoContent {
		t.Fatalf("expected 204, got %d", rec.Code)
	}
}
