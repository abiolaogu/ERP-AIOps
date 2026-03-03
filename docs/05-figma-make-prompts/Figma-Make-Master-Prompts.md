# Figma Make Master Prompts

Use this master sequence to transform Product requirements into production-ready UX deliverables.

## Inputs to Attach Before Prompting

- [PRD](../00-project-strategy/PRD.md)
- [HLD](../01-architecture/HLD.md)
- [LLD](../01-architecture/LLD.md)
- [Data Architecture](../01-architecture/Data-Architecture.md)
- [AuthN/AuthZ](../01-architecture/API/AuthN-AuthZ.md)

## Master Prompt 1: BRD/PRD Synthesis to UX Brief

```text
You are a senior Product Designer generating a UX brief from attached BRD/PRD and architecture docs.
Product context: multi-tenant SaaS with web portal, admin portal, API, RBAC, audit logs, billing, notifications.
Return:
1) UX goals by persona (Tenant Admin, Org Admin, User)
2) task criticality map (high/medium/low)
3) workflow inventory grouped by module
4) UX risks and mitigations
5) information architecture proposal with navigation model
Keep terminology consistent: Product, System, Tenant, Org, User, Admin, API, Service, Environment.
```

## Master Prompt 2: UX Flows and States

```text
Generate complete UX flow maps for:
- Login + token refresh
- Create Tenant + bootstrap Org + invite Admin
- Role assignment with approval policy
- Billing summary and charge action
- Audit log filter and export
For each flow include:
- happy path
- empty state
- validation errors
- permission denied state
- loading skeleton behavior
- success confirmation
Return as screen-by-screen flow table with IDs.
```

## Master Prompt 3: Wireframes and Layout System

```text
Produce low-fidelity wireframes for desktop and mobile.
Include:
- responsive navigation (sidebar desktop, drawer mobile)
- dashboard cards, data tables, filters, forms, toasts
- consistent spacing scale and typographic hierarchy
- accessibility annotations for keyboard and screen reader focus order
Output should include naming convention: module/screen/state.
```

## Master Prompt 4: Hi-Fi UI + Design Tokens

```text
Using the approved wireframes, generate high-fidelity frames with a neutral enterprise visual language.
Constraints:
- color tokens with semantic roles (bg/surface/text/success/warning/error/info)
- component variants for button/input/table/modal/toast/tabs
- WCAG AA contrast and focus states
- states: default/hover/focus/disabled/error/loading
Deliver a token table and component inventory aligned to handoff.
```

## Master Prompt 5: Prototype and Handoff

```text
Create interactive prototype for core journeys:
- Tenant onboarding
- RBAC editor
- Billing dashboard
- Audit explorer
- Notification center
Include transition specs, timing, easing, and trigger events.
Generate handoff package:
- redlines
- component usage rules
- interaction notes
- copy deck for microcopy
- implementation checklist linked to engineering tickets.
```

## Output Quality Checklist

- Persona goals clearly mapped to screens
- State coverage complete for all critical flows
- Components re-used consistently
- Accessibility notes included per major interaction
- Handoff artifacts explicit and implementation-ready

