import { GraphQLClient } from "graphql-request";

const HASURA_URL = import.meta.env.VITE_HASURA_URL || "http://localhost:19109/v1/graphql";
const TOKEN_KEY = "erp_aiops_auth_token";

export const graphqlClient = new GraphQLClient(HASURA_URL, {
  headers: () => {
    const token = localStorage.getItem(TOKEN_KEY);
    return token ? { Authorization: `Bearer ${token}` } : {};
  },
});

export const HASURA_WS_URL = import.meta.env.VITE_HASURA_WS_URL || "ws://localhost:19109/v1/graphql";
