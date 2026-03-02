import { env } from "@/lib/config/env";
import type { AppSession } from "@/lib/auth/session";

interface GraphQLResponse<TData> {
  data?: TData;
  errors?: Array<{ message: string }>;
}

export async function requestCosmo<TData, TVariables extends Record<string, unknown>>(
  query: string,
  variables: TVariables,
  session: AppSession,
): Promise<TData> {
  const response = await fetch(env.cosmoUrl, {
    method: "POST",
    headers: {
      "content-type": "application/json",
      ...(session.accessToken
        ? { authorization: `Bearer ${session.accessToken}` }
        : {}),
      "x-tenant-id": session.tenantId,
    },
    body: JSON.stringify({ query, variables }),
    cache: "no-store",
  });

  if (!response.ok) {
    throw new Error(`Cosmo request failed with status ${response.status}`);
  }

  const payload = (await response.json()) as GraphQLResponse<TData>;
  if (payload.errors?.length) {
    throw new Error(payload.errors.map((error) => error.message).join("; "));
  }
  if (!payload.data) {
    throw new Error("Cosmo response did not include data");
  }
  return payload.data;
}
