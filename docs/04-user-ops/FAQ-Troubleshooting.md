# FAQ and Troubleshooting

## Frequently Asked Questions

### Q1: Why can a User not access a feature after role assignment?

- Confirm role assignment completed successfully.
- Check Tenant and Org scope alignment.
- Validate token refresh occurred after role update.

### Q2: Why is billing summary missing recent usage?

- Confirm usage ingestion job status.
- Check delayed processing queue metrics.
- Validate Environment-specific timezone cutoffs.

### Q3: Why are audit logs not visible for a recent action?

- Verify query filters (`from`, `to`, `action`, `actor`).
- Check audit pipeline health and lag metrics.
- Ensure caller role includes audit read permissions.

### Q4: Why are notification webhooks failing?

- Validate endpoint reachability and TLS certificate.
- Verify webhook signature secret matches receiver config.
- Review retry queue and dead-letter events.

## Troubleshooting Playbook

| Symptom | First Check | Next Step | Escalate To |
|---|---|---|---|
| Login failures spike | Auth error rate dashboard | Verify identity provider health | Security + Platform |
| Elevated API latency | Gateway and DB latency charts | Scale read replicas or cache | Platform |
| Billing charge failures | Payment gateway status | Retry idempotent charges | Finance + Engineering |
| Missing audit events | Audit ingestion lag | Run reconciliation job | Security |

## Escalation Matrix

- Sev-1: Incident commander + Platform on-call within 5 minutes
- Sev-2: Service owner within 15 minutes
- Sev-3: Next business-hour response

