# Environment Setup

## Local Requirements

- Git 2.44+
- Node.js 20+
- npm 10+
- Docker Desktop or compatible runtime (optional for tooling)
- Mermaid-compatible Markdown preview extension

## Clone and Bootstrap

```bash
git clone <repo-url>
cd Documentation
npm init -y
npm install --save-dev markdownlint-cli2 lychee
```

## Recommended Local Tooling

- Markdown linting: `markdownlint-cli2`
- Link checking: `lychee`
- OpenAPI validation: `swagger-cli` or `redocly` (optional)

```bash
npm install --save-dev @redocly/cli
npx redocly lint docs/01-architecture/API/openapi.yaml
```

## Environment Variables (Optional)

For local scripts that target internal tools:

- `DOCS_PORTAL_URL`
- `GITHUB_TOKEN` (for API-driven link checks or PR automation)

## Troubleshooting

### Markdown lint failures

- Ensure heading levels are sequential.
- Break long lines where configured.

### Broken links

- Run `npx lychee docs README.md`.
- Prefer relative links within docs directory.

### Mermaid rendering issues

- Validate fenced block starts with ` ```mermaid `.
- Avoid unsupported diagram syntax for your renderer version.

### OpenAPI parse errors

- Validate indentation and schema references.
- Ensure all `$ref` paths exist.

