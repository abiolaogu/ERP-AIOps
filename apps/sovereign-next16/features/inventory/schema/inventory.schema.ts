import { z } from "zod";

export const inventoryItemSchema = z.object({
  id: z.string().uuid(),
  sku: z.string().min(1),
  status: z.enum(["in_stock", "low_stock", "out_of_stock"]),
  stock_level: z.number().int().nonnegative(),
  tenant_id: z.string().min(1)
});

export const inventoryListSchema = z.array(inventoryItemSchema);

export type InventoryItem = z.infer<typeof inventoryItemSchema>;
