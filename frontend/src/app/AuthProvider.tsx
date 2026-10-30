import { createContext, useCallback, useContext, useEffect, useMemo, useState, type ReactNode } from "react";
import { getCurrentUser, login as loginRequest } from "../features/auth/api";
import type { AuthUser, LoginInput } from "../features/auth/schemas";

type AuthContextValue = {
  user: AuthUser | null;
  isLoading: boolean;
  login: (input: LoginInput) => Promise<void>;
  logout: () => void;
};

const AuthContext = createContext<AuthContextValue | null>(null);
const TOKEN_KEY = "sigo.access_token";
const EMAIL_KEY = "sigo.user_email";
const PASSWORD_CHANGE_KEY = "sigo.must_change_password";

function clearStoredSession() {
  window.localStorage.removeItem(TOKEN_KEY);
  window.localStorage.removeItem(EMAIL_KEY);
  window.localStorage.removeItem(PASSWORD_CHANGE_KEY);
}

export function AuthProvider({ children }: { children: ReactNode }) {
  const [user, setUser] = useState<AuthUser | null>(null);
  const [isLoading, setIsLoading] = useState(() => Boolean(window.localStorage.getItem(TOKEN_KEY)));

  const logout = useCallback(() => {
    clearStoredSession();
    setUser(null);
    setIsLoading(false);
  }, []);

  useEffect(() => {
    const onUnauthorized = () => logout();
    window.addEventListener("sigo:unauthorized", onUnauthorized);
    return () => window.removeEventListener("sigo:unauthorized", onUnauthorized);
  }, [logout]);

  useEffect(() => {
    if (!window.localStorage.getItem(TOKEN_KEY)) return;

    let active = true;
    getCurrentUser()
      .then((currentUser) => {
        if (!active) return;
        setUser({
          ...currentUser,
          email: window.localStorage.getItem(EMAIL_KEY) ?? "",
          mustChangePassword: window.localStorage.getItem(PASSWORD_CHANGE_KEY) === "true",
        });
      })
      .catch(() => {
        if (active) logout();
      })
      .finally(() => {
        if (active) setIsLoading(false);
      });

    return () => {
      active = false;
    };
  }, [logout]);

  const login = useCallback(async (input: LoginInput) => {
    const response = await loginRequest(input);
    window.localStorage.setItem(TOKEN_KEY, response.access_token);
    window.localStorage.setItem(EMAIL_KEY, input.email.trim().toLowerCase());
    window.localStorage.setItem(PASSWORD_CHANGE_KEY, String(response.must_change_password));

    try {
      const currentUser = await getCurrentUser();
      setUser({
        ...currentUser,
        email: input.email.trim().toLowerCase(),
        mustChangePassword: response.must_change_password,
      });
    } catch (error) {
      logout();
      throw error;
    }
  }, [logout]);

  const value = useMemo(() => ({ user, isLoading, login, logout }), [user, isLoading, login, logout]);

  return <AuthContext.Provider value={value}>{children}</AuthContext.Provider>;
}

export function useAuth() {
  const context = useContext(AuthContext);
  if (!context) throw new Error("useAuth must be used inside AuthProvider");
  return context;
}
