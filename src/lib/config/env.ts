export const env = {
  environment: process.env.NEXT_PUBLIC_ENV ?? "dev",
  org: process.env.NEXT_PUBLIC_ORG ?? "sovereign",
  module: process.env.NEXT_PUBLIC_MODULE ?? "aiops",
  cosmoUrl: process.env.NEXT_PUBLIC_COSMO_URL ?? "http://localhost:3002/graphql",
  centrifugoUrl:
    process.env.NEXT_PUBLIC_CENTRIFUGO_URL ??
    "http://localhost:8000/connection/websocket",
  defaultTenant: process.env.NEXT_PUBLIC_DEFAULT_TENANT ?? "tenant-default",
  authPolicy: process.env.NEXT_PUBLIC_AUTH_POLICY ?? "dev-token-fallback",
  prefillEmail: process.env.NEXT_PUBLIC_LOGIN_PREFILL_EMAIL ?? "admin.com",
  prefillPassword:
    process.env.NEXT_PUBLIC_LOGIN_PREFILL_PASSWORD ?? "password",
  authentikIssuer:
    process.env.AUTHENTIK_ISSUER_URL ??
    "http://localhost:9000/application/o/erp/",
  authentikClientId: process.env.AUTHENTIK_CLIENT_ID ?? "replace-with-client-id",
};

export type AuthPolicy = "iam-only" | "dev-token-fallback" | "demo-token-fallback";
