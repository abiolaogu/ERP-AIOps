# Repository Onboarding Guide

This repository is the authoritative documentation System for Product and platform delivery.

## What You Get

- End-to-end strategy, architecture, and operational documentation
- API contracts and data model standards
- CI/CD, DR, observability, and compliance controls
- Figma Make prompt framework for design execution

## Getting Started in 15 Minutes

1. Clone repository and open this file.
2. Read [Project Charter](../00-project-strategy/Project-Charter.md).
3. Review [SAD](../01-architecture/SAD.md) and [OpenAPI](../01-architecture/API/openapi.yaml).
4. Configure local tools from [Environment Setup](Environment-Setup.md).
5. Read [Contributing](CONTRIBUTING.md) before opening a PR.

## Repository Map

```text
/docs
  /00-project-strategy
  /01-architecture
  /02-developer-onboarding
  /03-quality-devops
  /04-user-ops
  /05-figma-make-prompts
/.github
  /ISSUE_TEMPLATE
  /workflows
```

## Typical Workflows

- **Feature planning**: update BRD/PRD, then link architecture deltas.
- **Architecture updates**: update HLD/LLD and create ADR.
- **Operational updates**: update runbooks and SLO docs in the same PR.
- **Compliance changes**: update matrix with owner and evidence artifact.

## Required Checks Before Merge

- Docs lint and link checks pass.
- PR template checklist completed.
- Relevant cross-links updated in [docs index](../README.md).
- ADR added for architectural changes.

## Common Commands

```bash
# markdown lint
npx markdownlint-cli2 "**/*.md"

# link check
npx lychee --no-progress --verbose docs README.md
```

