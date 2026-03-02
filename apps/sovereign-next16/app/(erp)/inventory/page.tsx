"use client";

import { useTenant } from "@/core/hooks/use-tenant";
import { useInventory } from "@/features/inventory/hooks/use-inventory";
import { InventoryTable } from "@/features/inventory/components/inventory-table";
import { Card, CardContent, CardHeader, CardTitle } from "@/ui/card";
import { Skeleton } from "@/ui/skeleton";

export default function InventoryPage() {
  const tenantId = useTenant();
  const { data = [], isLoading } = useInventory(tenantId);

  return (
    <div className="space-y-6">
      <div>
        <h1 className="text-3xl font-bold tracking-tight">Inventory Management</h1>
        <p className="text-sm text-[var(--muted)]">Tenant: {tenantId}</p>
      </div>

      <Card>
        <CardHeader>
          <CardTitle>Global Stock Overview</CardTitle>
        </CardHeader>
        <CardContent>
          {isLoading ? <Skeleton className="h-72 w-full" /> : <InventoryTable data={data} />}
        </CardContent>
      </Card>
    </div>
  );
}
