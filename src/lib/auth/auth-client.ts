import { env } from "@/lib/config/env";

export function buildAuthentikAuthorizeUrl() {
  const callbackUrl =
    typeof window === "undefined"
      ? "http://localhost"
      : `${window.location.origin}/auth/callback`;

  const issuer = env.authentikIssuer.replace(/\/$/, "");
  const params = new URLSearchParams({
    response_type: "code",
    client_id: env.authentikClientId,
    redirect_uri: callbackUrl,
    scope: "openid profile email",
  });

  return `${issuer}/authorize/?${params.toString()}`;
}
