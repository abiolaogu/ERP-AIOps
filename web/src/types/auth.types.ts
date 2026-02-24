export interface AuthUser {
  id: string;
  email: string;
  name: string;
  firstName: string;
  lastName: string;
  avatar?: string;
  role: string;
  permissions: string[];
  tenantId: string;
}

export interface LoginRequest {
  email: string;
  password: string;
}

export interface LoginResponse {
  token: string;
  refreshToken: string;
  user: AuthUser;
  expiresAt: number;
}

export interface TokenPayload {
  sub: string;
  email: string;
  name: string;
  role: string;
  permissions: string[];
  tenantId: string;
  exp: number;
  iat: number;
}

export interface AuthState {
  isAuthenticated: boolean;
  user: AuthUser | null;
  token: string | null;
  loading: boolean;
}
