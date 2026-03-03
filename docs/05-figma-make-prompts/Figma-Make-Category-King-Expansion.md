# AIOps Figma Make Category-King Expansion

## Purpose

This prompt system pushes ERP-AIOps beyond baseline UX into category-king territory by tightly connecting Product strategy, System constraints, and engineering delivery.

## Required Input Package

Attach these files in every Figma Make run:

- `docs/00-project-strategy/BRD.md`
- `docs/00-project-strategy/PRD.md`
- `docs/01-architecture/HLD.md`
- `docs/01-architecture/LLD.md`
- `docs/01-architecture/Security-Architecture.md`
- `docs/03-quality-devops/Performance-Benchmarks.md`
- `docs/04-user-ops/Operations-Manual.md`

## Master Prompt (Use First)

```text
You are designing ERP-AIOps, a multi-tenant enterprise Product. Build a coherent UX system for User and Admin personas across web and responsive states.

Constraints:
- Preserve Tenant and Org context in every critical workflow.
- Expose policy/audit implications before irreversible actions.
- Prioritize high-density information layouts that remain legible and keyboard-accessible.
- Every screen must define loading, empty, warning, error, and success states.

Deliverables in order:
1) Journey map and information architecture.
2) Low-fidelity wireframes for all key flows.
3) High-fidelity UI with component/token mapping.
4) Interactive prototype including degraded-mode behavior.
5) Handoff spec with component names and behavior contracts.

Tone:
Decisive, trustworthy, data-rich, and execution-focused.
```

## Step-by-Step Prompt Chain

### 1. BRD to Journey Map

```text
From the BRD and PRD, produce 6 critical journeys for ERP-AIOps:
- onboarding
- core workflow execution
- exception handling
- audit review
- billing/usage visibility
- settings and integrations
For each journey, define actor, trigger, success condition, failure branch, and measurable KPI.
```

### 2. UX Flow Diagram Prompt

```text
Generate UX flow diagrams with decision nodes for User and Admin variants.
Include explicit path differences by role permissions, Tenant scope, and Org policy boundaries.
Return flows with labels for API dependencies and realtime update points.
```

### 3. Wireframe Prompt (Desktop + Mobile)

```text
Produce low-fidelity wireframes for dashboard, list/detail, wizard, settings, and audit views.
Rules:
- Desktop: left navigation + command bar + dense data layout.
- Mobile: bottom-sheet actions and context-preserving breadcrumbs.
- Add placeholders for event timelines, policy warnings, and confidence indicators.
```

### 4. High-Fidelity Prompt (Visual Language)

```text
Create high-fidelity screens using a professional enterprise aesthetic:
- strong hierarchy
- restrained but expressive color system
- tabular density with clear spacing rhythm
- accessible contrast and typography
Add visual states for urgency, risk, and confidence.
```

### 5. Design System Prompt

```text
Generate a component and token system for ERP-AIOps:
- semantic color tokens (surface, emphasis, success, warning, critical)
- type scale and spacing scale
- components: table, filter rail, command palette, timeline, side panel, role editor
- variants for compact, comfortable, and mobile touch modes
Include accessibility annotations for keyboard navigation and focus order.
```

### 6. Admin Portal Prompt

```text
Design Admin-only experiences:
- Tenant and Org provisioning
- RBAC policy editor with simulation mode
- audit evidence explorer
- billing plan and entitlement controls
- integration health dashboard
Each screen must include a "blast radius" hint and confirmation guardrails.
```

### 7. End-User Portal Prompt

```text
Design User workflows for fast daily execution:
- personalized home command center
- guided task queue
- exception triage workspace
- notification center
- profile + preferences
Show "next best action" recommendations and confidence labels.
```

### 8. Microcopy Prompt

```text
Write microcopy for critical actions:
- irreversible mutations
- permission denials
- partial failures
- policy conflict resolution
- integration outages
Tone: concise, direct, non-ambiguous, and enterprise-safe.
```

### 9. Motion and Interaction Prompt

```text
Define motion specs:
- page transitions under 250ms
- panel and drawer choreography
- optimistic update feedback
- realtime invalidation cues
- reduced-motion alternative behavior
Provide timing/easing tokens and rationale for each interaction class.
```

### 10. Prototype and Handoff Prompt

```text
Create a clickable prototype covering all critical journeys with edge cases.
For handoff, include:
- component inventory
- token map
- interaction matrix
- API dependency notes
- analytics event map
- acceptance checklist traceable to PRD requirements
```

## Category-King Readiness Rubric

| Dimension | Minimum Gate | Category-King Gate |
|---|---|---|
| Workflow clarity | task path obvious within 5s | next action obvious within 2s |
| Trust posture | warnings shown pre-submit | policy impact and evidence surfaced contextually |
| Data density | readable with filtering | high-density + low cognitive load |
| Responsiveness | functional mobile parity | role-aware adaptive layout and command surfaces |
| Accessibility | WCAG AA baseline | keyboard-first optimization + reduced motion modes |

## Mandatory Screen Inventory

- Onboarding and role bootstrap
- Home command center
- Core domain workspace (list + detail + action drawer)
- RBAC editor with policy simulation
- Billing and usage analytics
- Audit timeline with evidence drill-down
- Notification routing and escalation
- Integration and environment settings
- Incident/degraded-mode communication screen

## Delivery Contract to Engineering

1. Component names in Figma must match code component names.
2. Every component requires state coverage: default, hover, focus, loading, empty, error, success.
3. Every high-risk action requires guardrails and audit trace metadata.
4. Handoff includes analytics and telemetry IDs for critical journey tracking.
5. Prototype acceptance is tied to PRD acceptance criteria and performance constraints.
