package aiddguardrail

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"
)

type Mode string

const (
	ModeAutonomous Mode = "autonomous"
	ModeSupervised Mode = "supervised"
	ModeProtected  Mode = "protected"
)

type Action struct {
	Name                string
	TenantID            string
	Confidence          float64
	BlastRadius         int
	EstimatedCostUSD    float64
	CrossTenant         bool
	PrivilegeEscalation bool
	Destructive         bool
}

type Policy struct {
	AutonomousMinConfidence float64
	AutonomousMaxBlast      int
	AutonomousMaxCostUSD    float64
	SupervisedMinConfidence float64
	SupervisedMaxBlast      int
	SupervisedMaxCostUSD    float64
}

type Decision struct {
	Allowed          bool     `json:"allowed"`
	Mode             Mode     `json:"mode"`
	RequiresApproval bool     `json:"requires_approval"`
	Reasons          []string `json:"reasons"`
}

func DefaultPolicy() Policy {
	return Policy{
		AutonomousMinConfidence: 0.84,
		AutonomousMaxBlast:      400,
		AutonomousMaxCostUSD:    5000,
		SupervisedMinConfidence: 0.72,
		SupervisedMaxBlast:      5000,
		SupervisedMaxCostUSD:    75000,
	}
}

func ValidateTenantID(tenantID string) error {
	if strings.TrimSpace(tenantID) == "" {
		return errors.New("missing tenant id")
	}
	return nil
}

func (p Policy) Evaluate(action Action) Decision {
	reasons := make([]string, 0, 4)

	if err := ValidateTenantID(action.TenantID); err != nil {
		return Decision{Allowed: false, Mode: ModeProtected, Reasons: []string{err.Error()}}
	}

	if action.CrossTenant {
		reasons = append(reasons, "cross-tenant access denied")
	}
	if action.PrivilegeEscalation {
		reasons = append(reasons, "privilege escalation denied")
	}
	if action.Destructive && action.Confidence < 0.90 {
		reasons = append(reasons, "destructive operation below confidence threshold")
	}
	if len(reasons) > 0 {
		return Decision{Allowed: false, Mode: ModeProtected, Reasons: reasons}
	}

	if action.Confidence >= p.AutonomousMinConfidence &&
		action.BlastRadius <= p.AutonomousMaxBlast &&
		action.EstimatedCostUSD <= p.AutonomousMaxCostUSD {
		return Decision{
			Allowed:          true,
			Mode:             ModeAutonomous,
			RequiresApproval: false,
			Reasons:          []string{"autonomous execution allowed"},
		}
	}

	if action.Confidence >= p.SupervisedMinConfidence &&
		action.BlastRadius <= p.SupervisedMaxBlast &&
		action.EstimatedCostUSD <= p.SupervisedMaxCostUSD {
		return Decision{
			Allowed:          true,
			Mode:             ModeSupervised,
			RequiresApproval: true,
			Reasons:          []string{"approval required"},
		}
	}

	return Decision{
		Allowed:          false,
		Mode:             ModeProtected,
		RequiresApproval: false,
		Reasons:          []string{"risk exceeds supervised guardrail"},
	}
}

func Middleware(policy Policy) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			action := Action{
				Name:                defaultString(r.Header.Get("X-AIDD-Action"), r.Method+" "+r.URL.Path),
				TenantID:            strings.TrimSpace(r.Header.Get("X-Tenant-ID")),
				Confidence:          parseFloatHeader(r.Header.Get("X-AIDD-Confidence"), 1.0),
				BlastRadius:         parseIntHeader(r.Header.Get("X-AIDD-Blast-Radius"), 0),
				EstimatedCostUSD:    parseFloatHeader(r.Header.Get("X-AIDD-Estimated-Cost-USD"), 0),
				CrossTenant:         parseBoolHeader(r.Header.Get("X-AIDD-Cross-Tenant"), false),
				PrivilegeEscalation: parseBoolHeader(r.Header.Get("X-AIDD-Privilege-Escalation"), false),
				Destructive:         parseBoolHeader(r.Header.Get("X-AIDD-Destructive"), false),
			}

			decision := policy.Evaluate(action)
			w.Header().Set("X-AIDD-Mode", string(decision.Mode))
			w.Header().Set("X-AIDD-Requires-Approval", strconv.FormatBool(decision.RequiresApproval))

			if !decision.Allowed {
				writeDecision(w, http.StatusForbidden, decision)
				return
			}

			if decision.RequiresApproval && !strings.EqualFold(r.Header.Get("X-AIDD-Approval"), "approved") {
				decision.Allowed = false
				decision.Reasons = append(decision.Reasons, "explicit approval missing")
				writeDecision(w, http.StatusPreconditionFailed, decision)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

func writeDecision(w http.ResponseWriter, status int, decision Decision) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(map[string]any{
		"error":    "aidd_guardrail_denied",
		"decision": decision,
	})
}

func defaultString(v, fallback string) string {
	if strings.TrimSpace(v) == "" {
		return fallback
	}
	return v
}

func parseBoolHeader(v string, fallback bool) bool {
	if strings.TrimSpace(v) == "" {
		return fallback
	}
	parsed, err := strconv.ParseBool(v)
	if err != nil {
		return fallback
	}
	return parsed
}

func parseFloatHeader(v string, fallback float64) float64 {
	if strings.TrimSpace(v) == "" {
		return fallback
	}
	parsed, err := strconv.ParseFloat(v, 64)
	if err != nil {
		return fallback
	}
	return parsed
}

func parseIntHeader(v string, fallback int) int {
	if strings.TrimSpace(v) == "" {
		return fallback
	}
	parsed, err := strconv.Atoi(v)
	if err != nil {
		return fallback
	}
	return parsed
}
