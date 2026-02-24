import type { LiveProvider, LiveEvent } from "@refinedev/core";
import { createClient, type Client } from "graphql-ws";
import { HASURA_WS_URL } from "./graphqlClient";

const TOKEN_KEY = "erp_aiops_auth_token";

let wsClient: Client | null = null;

function getWsClient(): Client {
  if (!wsClient) {
    wsClient = createClient({
      url: HASURA_WS_URL,
      connectionParams: () => {
        const token = localStorage.getItem(TOKEN_KEY);
        return token
          ? { headers: { Authorization: `Bearer ${token}` } }
          : {};
      },
      retryAttempts: Infinity,
      shouldRetry: () => true,
      retryWait: async (retryCount) => {
        const delay = Math.min(1000 * 2 ** retryCount, 30000);
        await new Promise((resolve) => setTimeout(resolve, delay));
      },
      on: {
        connected: () => {
          console.log("[Hasura WS] Connected");
        },
        closed: () => {
          console.log("[Hasura WS] Disconnected");
        },
        error: (error) => {
          console.error("[Hasura WS] Error:", error);
        },
      },
    });
  }
  return wsClient;
}

export const liveProvider: LiveProvider = {
  subscribe: ({ channel, types, params, callback }) => {
    const resource = params?.resource || channel;
    const table = `aiops_${resource}`;

    const subscriptionQuery = `
      subscription LiveSubscription {
        ${table}(order_by: { updated_at: desc }, limit: 50) {
          id
          updated_at
        }
      }
    `;

    const client = getWsClient();
    let unsubscribed = false;

    const iterate = async () => {
      const iterable = client.iterate({ query: subscriptionQuery });
      try {
        for await (const result of iterable) {
          if (unsubscribed) break;
          if (result.data) {
            const event: LiveEvent = {
              channel,
              type: "created",
              payload: {
                ids: ((result.data as Record<string, Array<{ id: string }>>)[table] || []).map(
                  (item) => item.id,
                ),
              },
              date: new Date(),
            };
            callback(event);
          }
        }
      } catch (error) {
        if (!unsubscribed) {
          console.error("[Hasura WS] Subscription error:", error);
        }
      }
    };

    iterate();

    return {
      unsubscribe: () => {
        unsubscribed = true;
      },
    };
  },

  unsubscribe: (subscription) => {
    if (subscription && typeof subscription.unsubscribe === "function") {
      subscription.unsubscribe();
    }
  },
};

export default liveProvider;
