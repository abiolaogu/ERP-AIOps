"use client";

import { DataTable, type Column } from "@/ui/data-table";
import { type InventoryItem } from "@/features/inventory/schema/inventory.schema";

const columns: Column<InventoryItem>[] = [
  { key: "sku", header: "SKU", render: (row) => row.sku },
  { key: "status", header: "Status", render: (row) => row.status },
  { key: "stock_level", header: "Stock", render: (row) => row.stock_level }
];

export function InventoryTable({ data }: { data: InventoryItem[] }) {
  return <DataTable columns={columns} data={data} />;
}
