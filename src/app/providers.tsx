"use client";

import { QueryProvider } from "@/lib/providers/query-provider";
import { RealtimeProvider } from "@/lib/realtime/realtime-provider";
import { Toaster } from "sonner";

export function Providers({ children }: { children: React.ReactNode }) {
  return (
    <QueryProvider>
      <RealtimeProvider>
        {children}
        <Toaster richColors position="top-right" />
      </RealtimeProvider>
    </QueryProvider>
  );
}
