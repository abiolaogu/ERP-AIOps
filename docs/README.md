# ERP-AIOps Documentation

## Project Context
This documentation set applies to the **ERP-AIOps** Product and its System boundaries, Services, APIs, Tenant and Org operations across each Environment.

## Documentation Index

This index provides a complete reading path from strategy through operations. Every section is implementation-ready and version controlled.

## Reading Order

1. **Strategy and Product Intent**
   - [Project Charter](00-project-strategy/Project-Charter.md)
   - [BRD](00-project-strategy/BRD.md)
   - [PRD](00-project-strategy/PRD.md)
   - [Enterprise Architecture Roadmap](00-project-strategy/EA-Roadmap.md)
   - [Compliance and Regulatory Matrix](00-project-strategy/Compliance-Regulatory-Matrix.md)

2. **Architecture and Technical Design**
   - [SAD (includes HLD + LLD summary)](01-architecture/SAD.md)
   - [HLD](01-architecture/HLD.md)
   - [LLD](01-architecture/LLD.md)
   - [Security Architecture](01-architecture/Security-Architecture.md)
   - [Technical Specifications](01-architecture/Technical-Specifications.md)
   - [Data Architecture](01-architecture/Data-Architecture.md)
   - API:
     - [API Overview](01-architecture/API/API-Overview.md)
     - [AuthN/AuthZ](01-architecture/API/AuthN-AuthZ.md)
     - [OpenAPI](01-architecture/API/openapi.yaml)

3. **Developer Onboarding and Governance**
   - [Repo README](02-developer-onboarding/Repo-README.md)
   - [Environment Setup](02-developer-onboarding/Environment-Setup.md)
   - [Contributing](02-developer-onboarding/CONTRIBUTING.md)
   - [Changelog](02-developer-onboarding/CHANGELOG.md)
   - ADRs:
     - [ADR Template](02-developer-onboarding/ADR/ADR-0001-template.md)
     - [Vitastor vs Ceph ADR](02-developer-onboarding/ADR/ADR-0002-vitastor-vs-ceph.md)

4. **Quality, DevOps, and Reliability**
   - [Test Plan and Strategy](03-quality-devops/Test-Plan-Strategy.md)
   - [CI/CD for Harvester + Rancher + Fleet + Coolify](03-quality-devops/CI-CD.md)
   - [Disaster Recovery Plan](03-quality-devops/DR-Plan.md)
   - [Performance Benchmarks](03-quality-devops/Performance-Benchmarks.md)
   - [Observability](03-quality-devops/Observability.md)

5. **User and Operations Documentation**
   - [End User Guide](04-user-ops/End-User-Guide.md)
   - [Admin Guide](04-user-ops/Admin-Guide.md)
   - [Training Manual](04-user-ops/Training-Manual.md)
   - [Release Notes](04-user-ops/Release-Notes.md)
   - [FAQ and Troubleshooting](04-user-ops/FAQ-Troubleshooting.md)
   - [Operations Manual and Runbooks](04-user-ops/Operations-Manual.md)

6. **Figma Make Prompt Library**
   - [Master Prompts](05-figma-make-prompts/Figma-Make-Master-Prompts.md)
   - [PRD to UX Flows](05-figma-make-prompts/Figma-Make-PRD-to-UX-Flows.md)
   - [Design System](05-figma-make-prompts/Figma-Make-Design-System.md)
   - [Admin Portal Prompts](05-figma-make-prompts/Figma-Make-Admin-Portal.md)
   - [User Portal Prompts](05-figma-make-prompts/Figma-Make-User-Portal.md)

## System Terminology

- **Product**: the commercial SaaS offering.
- **System**: the complete technical solution powering the Product.
- **Tenant**: isolated customer boundary.
- **Org**: a business entity within a Tenant.
- **User**: end user of the Product.
- **Admin**: operational or tenant-level administrator.
- **API**: contract surface for clients and integrations.
- **Service**: deployable backend capability.
- **Environment**: dev, staging, prod deployment slice.

## Documentation Change Policy

- Every architectural change requires an ADR.
- Every feature change requires PRD section updates.
- Every operational change requires runbook updates.
- Every release requires changelog and release notes updates.


## Category-King Extensions

- [Category-King Operating Model](00-project-strategy/Category-King-Operating-Model.md)
- [Category-King Architecture Addendum](01-architecture/Category-King-Architecture-Addendum.md)
- [Category-King Figma Expansion](05-figma-make-prompts/Figma-Make-Category-King-Expansion.md)

## Institutional Fundraising Suite

- [Fundraising Suite Index](08-institutional-fundraising-suite/README.md)
- [Executive Summary](08-institutional-fundraising-suite/01_Executive_Summary.md)
- [5-Year Financial Model](08-institutional-fundraising-suite/05_Five_Year_Financial_Model.md)
- [Investor FAQ](08-institutional-fundraising-suite/13_Investor_FAQ_20_Difficult_Questions.md)
