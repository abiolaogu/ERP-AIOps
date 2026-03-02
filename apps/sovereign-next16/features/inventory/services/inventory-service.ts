import { inventoryListSchema, inventoryItemSchema, type InventoryItem } from "@/features/inventory/schema/inventory.schema";
import { createCosmoClient } from "@/core/api/cosmo-client";

const GET_INVENTORY = /* GraphQL */ `
  query GetInventory($tenantId: String!, $limit: Int!) {
    inventory(where: { tenant_id: { _eq: $tenantId } }, limit: $limit) {
      id
      sku
      status
      stock_level
      tenant_id
    }
  }
`;

const UPDATE_STOCK = /* GraphQL */ `
  mutation UpdateStock($id: uuid!, $newLevel: Int!) {
    update_inventory_by_pk(pk_columns: { id: $id }, _set: { stock_level: $newLevel }) {
      id
      sku
      status
      stock_level
      tenant_id
    }
  }
`;

export class InventoryService {
  static async getInventory(tenantId: string, limit = 25): Promise<InventoryItem[]> {
    const client = createCosmoClient(tenantId);
    const response = await client.request<{ inventory: unknown[] }>(GET_INVENTORY, { tenantId, limit });
    return inventoryListSchema.parse(response.inventory);
  }

  static async updateStock(tenantId: string, id: string, newLevel: number): Promise<InventoryItem> {
    if (newLevel < 0) throw new Error("ERP_ERR_NEGATIVE_STOCK");

    const client = createCosmoClient(tenantId);
    const response = await client.request<{ update_inventory_by_pk: unknown }>(UPDATE_STOCK, {
      id,
      newLevel
    });
    return inventoryItemSchema.parse(response.update_inventory_by_pk);
  }
}
