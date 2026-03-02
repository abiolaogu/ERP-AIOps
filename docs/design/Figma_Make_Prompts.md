# Figma & Make Prompts — ERP-AIOps
> Version: 1.1 | Last Updated: 2026-03-01 | Status: Active


<!-- SOVEREIGN_FIGMA_MAKE_EXPANSION_2026_03 -->
## 2026-03 Shared Infra + AIDD + GA Expansion Pack

### Scope
This section upgrades the prompt pack to align with:
- Shared infrastructure (`Hasura + ERP-DBaaS + ERP-IAM + ERP-Observability`)
- AIDD guardrails (Protected, Supervised, Autonomous execution modes)
- GA deployment quality expectations (accessibility, performance, observability, rollback-safe UX)

### Global Prompt Rules (Apply To Every Screen)
- Use production-safe copy and deterministic states for loading/error/empty/success.
- Include tenant context, role context, and policy context in all critical admin flows.
- Require explicit confirmation UX for destructive actions; show blast radius and rollback hint.
- Include instrumentation notes: event name, trace ID propagation, KPI target, and SLO linkage.
- Ensure WCAG 2.1 AA contrast and keyboard focus maps in every frame set.

### Prompt F-900: Shared Infra Control Plane Screen
```
Design a control plane view for ERP-AIOps that shows:
1. Hasura GraphQL connectivity (latency, success %, schema drift status)
2. ERP-DBaaS health (connection pool, replica lag, failover posture)
3. ERP-IAM integration (OIDC issuer health, token validation errors, session invalidations)
4. ERP-Observability status (OTLP export rate, tracing availability, alert health)

Include: command surface, timeline, runbook links, and one-click diagnostics panel.
Add states for: fully healthy, degraded IAM, degraded DB, degraded GraphQL, global outage.
```

### Prompt F-901: Role-Aware Workspace + Guardrail Surface
```
Generate a role-aware workspace for ERP-AIOps with:
- Role lenses: Operator, Manager, Auditor, Admin
- Guardrail panel with mode badge (Protected / Supervised / Autonomous)
- Inline policy rationale for blocked/supervised actions
- Approval request flow for supervised actions

Ensure each role sees distinct navigation and action priorities.
```

### Prompt F-902: Incident + Recovery UX
```
Design a full incident-response flow for ERP-AIOps:
- Alert ingestion to triage board
- Impacted tenant view and blast-radius visualization
- Action timeline with runbook steps and rollback controls
- Post-incident report generator modal

Must include: trace-id copy, audit evidence export, and SLA breach indicators.
```

### Prompt F-903: Mobile-Responsive Executive Snapshot
```
Create mobile and tablet executive dashboards for ERP-AIOps:
- KPI cards (availability, p95 latency, error budget burn, throughput)
- AI-generated narrative summary with confidence score
- Risk highlights and next-best-action CTA

Use adaptive layout breakpoints and preserve semantic heading structure.
```

### Prompt F-904: Figma Make -> Engineering Handoff Packet
```
For each completed flow, generate a handoff packet containing:
- Component inventory and token usage map
- API contract touchpoints (GraphQL operations and IAM claims)
- Test matrix (unit/integration/e2e/accessibility/performance)
- Observability event catalog and dashboard mapping
- GA release checklist references
```

### Prompt M-900: Make Automation Blueprint (Shared Infra)
```
Create a Make scenario for ERP-AIOps that:
1. Triggers on deployment/webhook events
2. Validates GraphQL, IAM, DB, and OTLP health in sequence
3. Creates incident ticket when thresholds are breached
4. Posts status updates to operations channels
5. Writes audit trail to persistent store

Add branching for: transient failures, policy violations, and hard-stop conditions.
```

### Prompt M-901: Tenant Onboarding Orchestration
```
Generate a Make scenario for tenant onboarding in ERP-AIOps:
- Validate tenant metadata and compliance profile
- Provision IAM roles and claims mapping
- Apply Hasura metadata/migrations checks
- Seed tenant defaults and dashboards
- Emit onboarding completion event and SLA timer
```

### Completion Criteria
- Every screen has loading/error/empty/success/access-denied variants.
- Every critical action includes telemetry mapping and guardrail annotation.
- Every flow includes desktop + tablet + mobile-responsive frames.
- Every high-risk flow includes rollback and operator escalation UX.

