# Figma Make Prompts: PRD to UX Flows

## Objective

Translate PRD feature requirements into deterministic UX flows with complete state coverage.

## Prompt Sequence

### Prompt A: Persona-to-Task Matrix

```text
From the attached PRD, build a matrix with rows=personas and columns=primary tasks.
For each task include business value, frequency, risk if failed, and required permissions.
```

### Prompt B: Information Architecture

```text
Create site maps for:
1) End User portal
2) Admin portal
Each map must show global nav, local nav, and deep-link paths.
Annotate which screens require Tenant context and Org context.
```

### Prompt C: Flow Definition Pack

```text
Create flow definitions for these modules:
- Onboarding
- User management
- RBAC editor
- Billing
- Audit logs
- Notifications
For each flow, include nodes for success, validation error, authorization error, service timeout, and recovery.
```

### Prompt D: Screen Inventory

```text
Create a screen inventory with IDs and descriptions.
Format: MODULE-SCREEN-STATE (example: RBAC-EDITOR-PERMISSION-DENIED)
For each screen define required components and data dependencies.
```

### Prompt E: UX Copy and Microcopy

```text
Generate concise microcopy for:
- form labels
- inline validation errors
- empty states
- destructive confirmations
- toast notifications
Tone: professional, calm, explicit.
```

## Flow Coverage Checklist

- Login + refresh + logout
- Create Tenant + invite Org Admin
- Add User + assign role + revoke role
- View billing summary + execute charge action + failure recovery
- Query and export audit events
- Create and test notification webhook

## Accessibility Prompt Add-On

```text
Annotate each critical flow with keyboard traversal order, focus indicators, aria-label hints, and color-contrast notes.
```

