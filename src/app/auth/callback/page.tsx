"use client";

import { useEffect } from "react";
import { useRouter } from "next/navigation";
import { env } from "@/lib/config/env";
import { saveSession } from "@/lib/auth/session";

export default function AuthCallbackPage() {
  const router = useRouter();

  useEffect(() => {
    saveSession({
      tenantId: env.defaultTenant,
      accessToken: "oidc-session-token",
      mode: "oidc",
    });
    router.replace("/dashboard");
  }, [router]);

  return (
    <main className="flex min-h-screen items-center justify-center">
      <p className="text-sm text-[var(--muted-foreground)]">Completing sign-in...</p>
    </main>
  );
}
