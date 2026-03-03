# Contributing

## Branching Model

- `main`: protected branch, release-ready docs
- `feature/*`: functional documentation updates
- `chore/*`: maintenance and standards updates
- `fix/*`: urgent corrections

## Pull Request Rules

1. Keep scope focused and explain why change is needed.
2. Update affected docs and cross-links in the same PR.
3. Include ADR when architectural decisions change.
4. Attach screenshots only when updating user-facing flows.
5. Add changelog entry for meaningful behavior/process updates.

## Commit Style

Use conventional commits:

- `feat(docs): ...`
- `fix(docs): ...`
- `chore(docs): ...`
- `docs(adr): ...`

## Required Reviewers

- Product documentation owner for BRD/PRD edits
- Architecture owner for HLD/LLD/SAD edits
- Security owner for security/compliance edits
- Platform owner for CI/CD and operations edits

## Documentation Quality Gates

- No empty sections
- No unresolved placeholders
- Valid Mermaid syntax
- OpenAPI schema remains valid YAML
- Internal links resolve

## Release Process for Documentation

1. Merge approved PRs to `main`.
2. Tag version in changelog and release notes.
3. Publish docs package to internal portal if required.

## Definition of Done

- The intended audience can execute actions without external clarification.
- The document has at least one owner and clear review cadence.
- Related docs are linked and terminology remains consistent.

