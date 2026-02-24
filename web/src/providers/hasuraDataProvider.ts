import type { DataProvider } from "@refinedev/core";
import { gql } from "graphql-request";
import { graphqlClient } from "./graphqlClient";

/**
 * AIOps tables use the "aiops_" prefix in Hasura.
 * E.g. resource "incidents" -> Hasura table "aiops_incidents"
 */
function tableName(resource: string): string {
  return `aiops_${resource}`;
}

/**
 * Convert Refine filter operators to Hasura _bool_exp operators.
 */
function mapOperator(operator: string): string {
  const mapping: Record<string, string> = {
    eq: "_eq",
    ne: "_neq",
    lt: "_lt",
    gt: "_gt",
    lte: "_lte",
    gte: "_gte",
    in: "_in",
    nin: "_nin",
    contains: "_ilike",
    ncontains: "_nilike",
    containss: "_like",
    ncontainss: "_nlike",
    null: "_is_null",
    nnull: "_is_null",
    between: "_gte",
  };
  return mapping[operator] || "_eq";
}

/**
 * Build Hasura _bool_exp from Refine filters.
 */
function buildWhere(
  filters?: Array<{
    field?: string;
    operator?: string;
    value?: unknown;
    key?: string;
  }>,
): Record<string, unknown> {
  if (!filters || filters.length === 0) return {};

  const where: Record<string, unknown> = {};
  for (const filter of filters) {
    if (!("field" in filter) || !filter.field) continue;

    const { field, operator = "eq", value } = filter;

    if (operator === "contains" || operator === "ncontains") {
      where[field] = { [mapOperator(operator)]: `%${value}%` };
    } else if (operator === "null") {
      where[field] = { _is_null: true };
    } else if (operator === "nnull") {
      where[field] = { _is_null: false };
    } else if (operator === "between" && Array.isArray(value)) {
      where[field] = { _gte: value[0], _lte: value[1] };
    } else {
      where[field] = { [mapOperator(operator)]: value };
    }
  }
  return where;
}

/**
 * Build Hasura order_by from Refine sorters.
 */
function buildOrderBy(
  sorters?: Array<{ field: string; order: string }>,
): Array<Record<string, string>> {
  if (!sorters || sorters.length === 0) return [{ created_at: "desc" }];
  return sorters.map((s) => ({
    [s.field]: s.order === "desc" ? "desc" : "asc",
  }));
}

export const dataProvider: DataProvider = {
  getList: async ({ resource, pagination, sorters, filters, meta }) => {
    const current = pagination?.current ?? 1;
    const pageSize = pagination?.pageSize ?? 10;
    const offset = (current - 1) * pageSize;
    const table = meta?.table || tableName(resource);
    const fields = meta?.fields?.join("\n") || "id";
    const where = buildWhere(
      filters as Array<{
        field?: string;
        operator?: string;
        value?: unknown;
      }>,
    );
    const orderBy = buildOrderBy(
      sorters as Array<{ field: string; order: string }>,
    );

    const query = gql`
      query GetList(
        $where: ${table}_bool_exp
        $order_by: [${table}_order_by!]
        $limit: Int
        $offset: Int
      ) {
        ${table}(
          where: $where
          order_by: $order_by
          limit: $limit
          offset: $offset
        ) {
          ${fields}
        }
        ${table}_aggregate(where: $where) {
          aggregate {
            count
          }
        }
      }
    `;

    const response = await graphqlClient.request<Record<string, unknown>>(
      query,
      {
        where,
        order_by: orderBy,
        limit: pageSize,
        offset,
      },
    );

    const data = response[table] as unknown[];
    const aggregate = response[`${table}_aggregate`] as {
      aggregate: { count: number };
    };

    return {
      data: data as never[],
      total: aggregate?.aggregate?.count ?? data.length,
    };
  },

  getOne: async ({ resource, id, meta }) => {
    const table = meta?.table || tableName(resource);
    const fields = meta?.fields?.join("\n") || "id";

    const query = gql`
      query GetOne($id: uuid!) {
        ${table}_by_pk(id: $id) {
          ${fields}
        }
      }
    `;

    const response = await graphqlClient.request<Record<string, unknown>>(
      query,
      { id: String(id) },
    );

    return {
      data: response[`${table}_by_pk`] as never,
    };
  },

  create: async ({ resource, variables, meta }) => {
    const table = meta?.table || tableName(resource);
    const fields = meta?.fields?.join("\n") || "id";

    const mutation = gql`
      mutation Create($object: ${table}_insert_input!) {
        insert_${table}_one(object: $object) {
          ${fields}
        }
      }
    `;

    const response = await graphqlClient.request<Record<string, unknown>>(
      mutation,
      { object: variables },
    );

    return {
      data: response[`insert_${table}_one`] as never,
    };
  },

  update: async ({ resource, id, variables, meta }) => {
    const table = meta?.table || tableName(resource);
    const fields = meta?.fields?.join("\n") || "id";

    const mutation = gql`
      mutation Update($id: uuid!, $set: ${table}_set_input!) {
        update_${table}_by_pk(pk_columns: { id: $id }, _set: $set) {
          ${fields}
        }
      }
    `;

    const response = await graphqlClient.request<Record<string, unknown>>(
      mutation,
      { id: String(id), set: variables },
    );

    return {
      data: response[`update_${table}_by_pk`] as never,
    };
  },

  deleteOne: async ({ resource, id, meta }) => {
    const table = meta?.table || tableName(resource);

    const mutation = gql`
      mutation Delete($id: uuid!) {
        delete_${table}_by_pk(id: $id) {
          id
        }
      }
    `;

    const response = await graphqlClient.request<Record<string, unknown>>(
      mutation,
      { id: String(id) },
    );

    return {
      data: response[`delete_${table}_by_pk`] as never,
    };
  },

  getApiUrl: () => {
    return (
      import.meta.env.VITE_HASURA_URL ||
      "http://localhost:19109/v1/graphql"
    );
  },
};

export default dataProvider;
