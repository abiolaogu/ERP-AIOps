import { GraphQLClient } from "graphql-request";
import { env } from "@/lib/env";

export function createCosmoClient(tenantId: string, token?: string): GraphQLClient {
  return new GraphQLClient(env.cosmoUrl, {
    headers: {
      "x-tenant-id": tenantId,
      ...(token ? { Authorization: `Bearer ${token}` } : {})
    }
  });
}
