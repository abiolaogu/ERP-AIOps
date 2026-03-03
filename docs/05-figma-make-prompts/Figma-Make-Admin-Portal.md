# Figma Make Prompts: Admin Portal

## Scope

Admin portal for Tenant and Org operations, RBAC governance, billing administration, audit review, and integrations.

## Prompt Set

### Prompt 1: Admin IA and Navigation

```text
Design admin information architecture with modules:
Dashboard, Tenants, Orgs, Users, Roles, Billing, Audit, Notifications, Integrations, Settings.
Create desktop sidebar + mobile drawer navigation with active state and breadcrumb logic.
```

### Prompt 2: Tenant Provisioning Screens

```text
Generate end-to-end screens for creating a Tenant:
- tenant form
- org bootstrap
- admin invite
- success checklist
Include validation, duplicate detection, and permission denied states.
```

### Prompt 3: RBAC Editor

```text
Design RBAC editor with:
- role list
- permission matrix
- user assignment panel
- impact preview before save
- policy conflict warnings
Include audit metadata display (who changed what and when).
```

### Prompt 4: Billing Admin

```text
Create billing admin views:
- subscription summary
- usage trend chart
- invoice list
- manual charge action with confirmation modal
Include finance-safe wording and high-risk action safeguards.
```

### Prompt 5: Audit Log Explorer

```text
Design audit explorer with advanced filters:
actor, action, target type, time range, environment.
Include export workflow, saved views, and sensitive-event highlighting.
```

### Prompt 6: Integrations and API Keys

```text
Create integrations screens:
- API key lifecycle (create/reveal/revoke/rotate)
- webhook management and test delivery results
- integration health status timeline
```

### Prompt 7: Interaction Specs

```text
Annotate interactions and transitions for:
- modals and confirmation dialogs
- table filter chips
- toasts
- optimistic updates and rollback messaging
```

