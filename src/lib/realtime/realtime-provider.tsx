"use client";

import { env } from "@/lib/config/env";
import { ensureSession } from "@/lib/auth/session";
import { Centrifuge } from "centrifuge";
import { useEffect } from "react";
import { useQueryClient, type QueryKey } from "@tanstack/react-query";
import { toast } from "sonner";

interface InvalidatePayload {
  queryKeys?: QueryKey[];
  title?: string;
  description?: string;
}

export function RealtimeProvider({ children }: { children: React.ReactNode }) {
  const queryClient = useQueryClient();

  useEffect(() => {
    if (!env.centrifugoUrl) {
      return;
    }

    const session = ensureSession(env.defaultTenant);
    const topic = `${env.environment}.${env.org}.${env.module}.${session.tenantId}.ui.invalidate`;
    const client = new Centrifuge(env.centrifugoUrl);
    const subscription = client.newSubscription(topic);

    subscription.on("publication", (context) => {
      const data = (context.data ?? {}) as InvalidatePayload;
      const keys =
        Array.isArray(data.queryKeys) && data.queryKeys.length > 0
          ? data.queryKeys
          : [["control-center", session.tenantId]];

      keys.forEach((queryKey) => {
        const normalized = Array.isArray(queryKey) ? queryKey : [String(queryKey)];
        queryClient.invalidateQueries({ queryKey: normalized });
      });

      toast.success(data.title ?? "Realtime update received", {
        description: data.description ?? `Invalidated ${keys.length} query key(s).`,
      });
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
