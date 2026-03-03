# Compliance and Regulatory Matrix

This matrix maps regulatory obligations to practical controls, implementation notes, evidence artifacts, ownership, and status.

## Legend

- **Status**: Implemented, In Progress, Planned
- **Owner**: role responsible for operating control

| Regulation | Control ID | Control Objective | Implementation Notes | Evidence Artifacts | Owner | Status |
|---|---|---|---|---|---|---|
| GDPR | GDPR-01 | Lawful processing and consent records | Capture consent event with timestamp and policy version per User | Consent logs, policy version history | Product Owner | Implemented |
| GDPR | GDPR-02 | Right to access/export | API endpoint for data export by Tenant + User scope | Export job logs, signed export manifest | Engineering Lead | Implemented |
| GDPR | GDPR-03 | Right to erasure | Soft-delete workflow plus irreversible purge job with approval | Erasure request tickets, purge job audit events | Compliance Lead | In Progress |
| GDPR | GDPR-04 | Data minimization | Data dictionary marks required vs optional PII fields | Data dictionary, schema review reports | Data Architect | Implemented |
| SOC2 | SOC2-CC1 | Access control governance | RBAC with least privilege templates and quarterly review | Access review reports, role diff logs | Security Lead | Implemented |
| SOC2 | SOC2-CC2 | Change management | PR approvals, CI checks, release sign-off checklist | PR history, pipeline logs, release records | Engineering Lead | Implemented |
| SOC2 | SOC2-CC3 | Monitoring and incident response | Alerting with on-call runbooks and incident postmortems | Incident tickets, postmortems, on-call timeline | Platform Lead | Implemented |
| SOC2 | SOC2-CC4 | Vendor risk management | Third-party inventory and annual assessment process | Vendor register, assessment templates | Compliance Lead | In Progress |
| HIPAA | HIPAA-164.312(a) | Unique user identification | OIDC identity and unique User IDs with MFA policy | IAM config snapshots, MFA policy logs | Security Lead | Implemented |
| HIPAA | HIPAA-164.312(b) | Audit controls | Immutable audit stream for ePHI-related operations | Audit logs, retention policy evidence | Security Architect | Implemented |
| HIPAA | HIPAA-164.312(c) | Integrity controls | Signed events and tamper-evident audit pipeline | Hash validation reports, event signatures | Platform Lead | In Progress |
| HIPAA | HIPAA-164.312(e) | Transmission security | TLS 1.2+ enforced across API and Service mesh | TLS scans, certificate inventory | Platform Lead | Implemented |
| PCI-DSS | PCI-1 | Network segmentation | Segmented card-data paths, deny-by-default policy | Network policy manifests, firewall rule exports | Platform Lead | Implemented |
| PCI-DSS | PCI-2 | Secure configurations | Hardened baselines and image scanning gates | CIS benchmark reports, image scan logs | DevOps Lead | In Progress |
| PCI-DSS | PCI-3 | Protect stored account data | Tokenization and encrypted storage; no PAN in app DB | Tokenization architecture, encryption key policy | Security Lead | Implemented |
| PCI-DSS | PCI-4 | Encrypt transmission of cardholder data | mTLS internal and TLS external for payment flows | mTLS cert inventory, transport test evidence | Platform Lead | Implemented |
| PCI-DSS | PCI-10 | Track and monitor access | Centralized logs and privileged access monitoring | SIEM correlation reports, privileged access alerts | Security Operations | Implemented |

## Control-Evidence Operating Cadence

- Daily: automated pipeline evidence collection (CI logs, scan outputs)
- Weekly: incident and access review
- Monthly: compliance owner attestations
- Quarterly: executive compliance posture review

## Evidence Repository Structure

```text
compliance-evidence/
  gdpr/
  soc2/
  hipaa/
  pci-dss/
  release-artifacts/
  access-reviews/
  incident-postmortems/
```

## Escalation Rules

- Any missing critical evidence older than 7 days escalates to Compliance Lead.
- Any control marked failed in production Environment escalates to Security Lead and Engineering Lead.

