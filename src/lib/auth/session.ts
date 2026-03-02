import { modulePolicy } from "@/lib/config/module-policy";

const STORAGE_KEY = "sovereign.erp.session";

export type SessionMode = "oidc" | "dev-token" | "demo-token";

export interface AppSession {
  tenantId: string;
  accessToken: string;
  mode: SessionMode;
}

function safeWindow() {
  return typeof window !== "undefined" ? window : undefined;
}

export function getSession(): AppSession | null {
  const win = safeWindow();
  if (!win) {
    return null;
  }

  const raw = win.localStorage.getItem(STORAGE_KEY);
  if (!raw) {
    return null;
  }

  try {
    return JSON.parse(raw) as AppSession;
  } catch {
    win.localStorage.removeItem(STORAGE_KEY);
    return null;
  }
}

export function saveSession(session: AppSession) {
  const win = safeWindow();
  if (!win) {
    return;
  }
  win.localStorage.setItem(STORAGE_KEY, JSON.stringify(session));
}

export function clearSession() {
  const win = safeWindow();
  if (!win) {
    return;
  }
  win.localStorage.removeItem(STORAGE_KEY);
}

export function ensureSession(defaultTenant: string): AppSession {
  const existing = getSession();
  if (existing) {
    return existing;
  }

  if (modulePolicy.authPolicy === "iam-only") {
    return {
      tenantId: defaultTenant,
      accessToken: "",
      mode: "oidc",
    };
  }

  const mode = modulePolicy.authPolicy === "demo-token-fallback" ? "demo-token" : "dev-token";
  const nextSession: AppSession = {
    tenantId: defaultTenant,
    accessToken: `${mode}-${modulePolicy.slug}`,
    mode,
  };
  saveSession(nextSession);
  return nextSession;
}
