# Admin Guide

## Audience

Tenant Admin and Org Admin roles responsible for operational and governance actions.

## Admin Responsibilities

- Tenant and Org setup
- User and role lifecycle
- Billing and plan visibility
- Integration and API key configuration
- Audit and policy oversight

## Tenant Provisioning

1. Go to **Admin > Tenants**.
2. Select **Create Tenant** and enter required metadata.
3. Confirm default Org creation and first Admin invitation.
4. Verify onboarding checklist completion.

## User and RBAC Management

1. Open **Admin > Users**.
2. Add User to Tenant and optional Org.
3. Assign role template.
4. Confirm policy result and audit event creation.

## Integrations

### API Keys

- Create key under **Admin > Integrations > API Keys**.
- Assign least-privilege scopes.
- Set expiration policy and owner metadata.

### Webhooks

- Register endpoint URL and secret.
- Validate using test event and signature check.
- Monitor delivery status and retry logs.

## Billing Administration

- View plan tier, current usage, and invoices.
- Trigger invoice export for finance workflows.
- Review failed charge events and retry actions.

## Audit and Compliance

- Query privileged events by date, actor, and action.
- Export event reports for compliance reviews.
- Validate periodic access reviews are completed.

## Admin Escalation Paths

- Security issue: Security Lead and incident commander.
- Billing reconciliation issue: Finance Manager + Platform on-call.
- Access policy issue: Identity owner + Engineering Lead.

