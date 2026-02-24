import { useGetIdentity, useLogout, useIsAuthenticated } from "@refinedev/core";
import type { AuthUser } from "@/types/auth.types";

export function useAuth() {
  const { data: identity, isLoading: identityLoading } =
    useGetIdentity<AuthUser>();
  const { data: authData, isLoading: authLoading } = useIsAuthenticated();
  const { mutate: logout } = useLogout();

  return {
    user: identity ?? null,
    isAuthenticated: authData?.authenticated ?? false,
    isLoading: identityLoading || authLoading,
    logout: () => logout(),
  };
}
