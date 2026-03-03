# Release Notes

## Format

Each release includes:

- summary of Product changes
- architecture or operational impact
- migration notes
- known issues and mitigations

## 2026-03-02 - Documentation Factory Baseline

### Added

- Full docs-as-code structure with strategy, architecture, onboarding, quality/devops, and operations sections
- OpenAPI contract for auth, tenants, users, audit, billing, notifications
- Compliance matrix for GDPR, SOC2, HIPAA, and PCI-DSS
- CI/CD guidance for Harvester HCI + Rancher + Fleet + Coolify
- Figma Make prompt framework from BRD/PRD through handoff

### Operational Impact

- PR process now enforces docs and ADR updates via templates/checklists
- Docs lint workflow introduced to prevent broken links and formatting drift

### Known Issues

- Some teams may need local Mermaid renderer updates for advanced diagrams
- OpenAPI examples require environment-specific URLs for live testing

### Mitigations

- Environment setup guide includes tooling versions and validation commands
- API overview includes standard base URL patterns by Environment

