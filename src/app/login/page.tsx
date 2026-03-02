"use client";

import { useMemo, useState } from "react";
import { useRouter } from "next/navigation";
import { Button } from "@/components/ui/button";
import { env } from "@/lib/config/env";
import { modulePolicy } from "@/lib/config/module-policy";
import { buildAuthentikAuthorizeUrl } from "@/lib/auth/auth-client";
import { saveSession } from "@/lib/auth/session";

export default function LoginPage() {
  const router = useRouter();
  const [email, setEmail] = useState(env.prefillEmail);
  const [password, setPassword] = useState(env.prefillPassword);

  const fallbackLabel = useMemo(() => {
    if (modulePolicy.authPolicy === "demo-token-fallback") {
      return "Use Demo Token";
    }
    if (modulePolicy.authPolicy === "dev-token-fallback") {
      return "Use Dev Token";
    }
    return ""
  }, []);

  const onFallback = () => {
    const mode = modulePolicy.authPolicy === "demo-token-fallback" ? "demo-token" : "dev-token";
    saveSession({
      tenantId: env.defaultTenant,
      accessToken: `${mode}-${modulePolicy.slug}`,
      mode,
    });
    router.push("/dashboard");
  };

  const onIam = () => {
    window.location.href = buildAuthentikAuthorizeUrl();
  };

  return (
    <main className="mx-auto flex min-h-screen w-full max-w-md items-center px-6">
      <section className="w-full space-y-5 rounded-xl border bg-[var(--card)] p-6 shadow-sm">
        <header className="space-y-1">
          <h1 className="text-xl font-semibold">Sign in to {modulePolicy.name}</h1>
          <p className="text-sm text-[var(--muted-foreground)]">
            Auth policy: {modulePolicy.authPolicy}. OIDC is routed through ERP-IAM/Authentik.
          </p>
        </header>

        <label className="space-y-1 text-sm">
          <span>Email</span>
          <input
            value={email}
            onChange={(event) => setEmail(event.target.value)}
            className="h-10 w-full rounded-md border bg-transparent px-3"
            placeholder="you@example.com"
          />
        </label>

        <label className="space-y-1 text-sm">
          <span>Password</span>
          <input
            type="password"
            value={password}
            onChange={(event) => setPassword(event.target.value)}
            className="h-10 w-full rounded-md border bg-transparent px-3"
            placeholder="********"
          />
        </label>

        <div className="space-y-2">
          <Button className="w-full" onClick={onIam}>
            Continue with ERP-IAM
          </Button>
          {modulePolicy.authPolicy !== "iam-only" ? (
            <Button variant="outline" className="w-full" onClick={onFallback}>
              {fallbackLabel}
            </Button>
          ) : null}
        </div>

        {modulePolicy.notes ? (
          <p className="rounded-md bg-[var(--muted)] px-3 py-2 text-xs text-[var(--muted-foreground)]">
            {modulePolicy.notes}
          </p>
        ) : null}
      </section>
    </main>
  );
}
