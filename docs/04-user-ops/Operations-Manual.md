# Operations Manual

## Operations Scope

This manual defines routine operations, SLO governance, incident handling, and runbooks for core operational tasks.

## Monitoring and Alert Response

### Daily Operations

- Review SLO dashboard and active alerts.
- Check queue health for notifications and audit streams.
- Validate backup completion report.

### Weekly Operations

- Access review completion checks.
- Error budget trend review.
- Capacity trend and scaling review.

## SLO Governance

| Service Area | SLO | Alert Threshold |
|---|---|---|
| API availability | 99.9% | 99.7% rolling 24h |
| Auth success ratio | 99.95% | 99.8% rolling 1h |
| Billing charge success | 99.5% | 99.0% rolling 1h |
| Audit write success | 99.99% | 99.9% rolling 15m |

## Runbooks and Playbooks

### 1) Rotate API Keys

1. Identify key owner, scope, and dependent Services.
2. Create replacement key with same or reduced scope.
3. Deploy key to dependent Service via secret manager.
4. Run smoke tests for dependent endpoints.
5. Revoke old key and confirm no active usage.
6. Record rotation evidence in compliance artifacts.

### 2) Restore from Backup

1. Confirm incident scope and recovery point requirement.
2. Select latest valid backup satisfying RPO target.
3. Restore to isolated validation environment.
4. Run integrity checks (row counts, checksums, sample records).
5. Promote restored dataset to target Environment.
6. Validate core flows: login, tenant read, billing summary, audit query.
7. Announce recovery completion and post-incident review plan.

### 3) Incident Response

1. Declare incident severity and assign incident commander.
2. Create timeline channel and assign responders.
3. Contain blast radius (feature flag, traffic shift, rate limit).
4. Mitigate and validate recovery against SLO indicators.
5. Communicate status updates every 15 minutes for sev-1.
6. Close incident and publish postmortem with actions.

### 4) Deploy Rollback

1. Identify failing release and affected Services.
2. Revert GitOps manifest commit or Coolify release version.
3. Confirm rollout status in Rancher workload dashboard.
4. Run post-rollback smoke tests.
5. Keep canary disabled until root cause is understood.

### 5) Horizontal Scaling

1. Review autoscaling triggers and current pod saturation.
2. Increase replicas or HPA target for affected Service.
3. Confirm downstream dependencies can absorb load.
4. Re-run load probes and compare p95 latency.
5. Update capacity baseline document.

### 6) Certificate Rotation

1. Inventory certificates nearing expiry.
2. Generate or request new cert chain.
3. Update secret store and workload mounts.
4. Rolling restart workloads with zero-downtime strategy.
5. Validate TLS handshakes and certificate fingerprint.
6. Archive evidence and expiry calendar updates.

## Routine Maintenance

- Monthly secret rotation checks
- Quarterly DR drills
- Quarterly dependency and vulnerability review
- Semi-annual access and permission recertification

## Escalation Contacts (Role-Based)

- Incident Commander: on-call lead
- Platform Owner: cluster/network/runtime issues
- Security Owner: auth, policy, data protection issues
- Compliance Owner: evidence and regulatory impact

