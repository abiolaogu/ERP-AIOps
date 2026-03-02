export function useTenant(): string {
  return process.env.NEXT_PUBLIC_TENANT_ID ?? "tenant-01";
}
