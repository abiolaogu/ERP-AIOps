"use client";

import { useEffect } from "react";
import type { ReactNode } from "react";
import { Centrifuge } from "centrifuge";
import { useQueryClient } from "@tanstack/react-query";
import { env } from "@/lib/env";

type EntityUpdatedEvent = {
  type: "ENTITY_UPDATED";
  entity: string;
  tenant_id: string;
};

export function RealtimeProvider({ children }: { children: ReactNode }) {
  const queryClient = useQueryClient();

  useEffect(() => {
    const client = new Centrifuge(env.centrifugoUrl);
    const channel = `tenant_updates:tenant_${env.tenantId}`;

    const subscription = client.newSubscription(channel);

    subscription.on("publication", (ctx) => {
      const payload = ctx.data as Partial<EntityUpdatedEvent>;
      if (payload.type === "ENTITY_UPDATED" && payload.entity) {
        queryClient.invalidateQueries({ queryKey: [payload.entity, env.tenantId] });
      }
    });

    subscription.subscribe();
    client.connect();

    return () => {
      subscription.unsubscribe();
      client.disconnect();
    };
  }, [queryClient]);

  return <>{children}</>;
}
