function required(name: string, fallback?: string): string {
  const value = process.env[name] ?? fallback;
  if (!value) throw new Error(`Missing required env var: ${name}`);
  return value;
}

export const env = {
  cosmoUrl: required("ERP_SHARED_COSMO_URL", "http://localhost:3002/graphql"),
  centrifugoUrl: required("ERP_SHARED_CENTRIFUGO_URL", "ws://localhost:8000/connection/websocket"),
  tenantId: required("NEXT_PUBLIC_TENANT_ID", "tenant-01")
};
