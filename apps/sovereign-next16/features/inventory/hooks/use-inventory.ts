"use client";

import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { InventoryService } from "@/features/inventory/services/inventory-service";
import { type InventoryItem } from "@/features/inventory/schema/inventory.schema";

const key = (tenantId: string) => ["inventory", tenantId] as const;

export function useInventory(tenantId: string) {
  return useQuery({
    queryKey: key(tenantId),
    queryFn: () => InventoryService.getInventory(tenantId),
    staleTime: 1000 * 60 * 5
  });
}

export function useUpdateInventoryStock(tenantId: string) {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: ({ id, newLevel }: { id: string; newLevel: number }) =>
      InventoryService.updateStock(tenantId, id, newLevel),
    onMutate: async ({ id, newLevel }) => {
      await queryClient.cancelQueries({ queryKey: key(tenantId) });
      const previous = queryClient.getQueryData<InventoryItem[]>(key(tenantId));
      queryClient.setQueryData<InventoryItem[]>(key(tenantId), (items = []) =>
        items.map((item) => (item.id === id ? { ...item, stock_level: newLevel } : item))
      );
      return { previous };
    },
    onError: (_error, _vars, context) => {
      if (context?.previous) {
        queryClient.setQueryData(key(tenantId), context.previous);
      }
    },
    onSettled: () => {
      queryClient.invalidateQueries({ queryKey: key(tenantId) });
    }
  });
}
