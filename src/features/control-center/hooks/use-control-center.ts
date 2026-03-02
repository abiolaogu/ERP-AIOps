"use client";

import { useQuery } from "@tanstack/react-query";
import { ensureSession } from "@/lib/auth/session";
import { env } from "@/lib/config/env";
import { ControlCenterService } from "@/features/control-center/services/control-center-service";

export function useControlCenter() {
  const session = ensureSession(env.defaultTenant);

  return useQuery({
    queryKey: ["control-center", session.tenantId],
    queryFn: () => ControlCenterService.getCards(session),
    staleTime: 1000 * 60 * 5,
  });
}
