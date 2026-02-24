import type { AuthProvider } from "@refinedev/core";
import { TOKEN_KEY, REFRESH_TOKEN_KEY, USER_KEY, IAM_URL } from "@/utils/constants";

export const authProvider: AuthProvider = {
  login: async ({ email, password }) => {
    try {
      const response = await fetch(`${IAM_URL}/auth/login`, {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ email, password }),
      });

      if (!response.ok) {
        return {
          success: false,
          error: { name: "LoginError", message: "Invalid email or password" },
        };
      }

      const data = await response.json();
      localStorage.setItem(TOKEN_KEY, data.token);
      localStorage.setItem(REFRESH_TOKEN_KEY, data.refreshToken);
      localStorage.setItem(USER_KEY, JSON.stringify(data.user));

      return { success: true, redirectTo: "/" };
    } catch {
      // Fallback for development: accept any login
      localStorage.setItem(TOKEN_KEY, "dev-token");
      localStorage.setItem(
        USER_KEY,
        JSON.stringify({
          id: "1",
          email: email || "admin@erp.com",
          name: "Admin User",
          firstName: "Admin",
          lastName: "User",
          role: "admin",
          permissions: ["*"],
          tenantId: "default",
        }),
      );
      return { success: true, redirectTo: "/" };
    }
  },

  logout: async () => {
    localStorage.removeItem(TOKEN_KEY);
    localStorage.removeItem(REFRESH_TOKEN_KEY);
    localStorage.removeItem(USER_KEY);
    return { success: true, redirectTo: "/login" };
  },

  check: async () => {
    const token = localStorage.getItem(TOKEN_KEY);
    if (token) {
      return { authenticated: true };
    }
    return { authenticated: false, redirectTo: "/login" };
  },

  getPermissions: async () => {
    const userStr = localStorage.getItem(USER_KEY);
    if (userStr) {
      const user = JSON.parse(userStr);
      return user.permissions ?? [];
    }
    return [];
  },

  getIdentity: async () => {
    const userStr = localStorage.getItem(USER_KEY);
    if (userStr) {
      return JSON.parse(userStr);
    }
    return null;
  },

  onError: async (error) => {
    if (error?.statusCode === 401) {
      return { logout: true, redirectTo: "/login" };
    }
    return { error };
  },
};
