import { defineStore } from "pinia";
import type { User } from "~/types";

interface LoginResponse {
  success: boolean;
  token: string;
  expires_at: string;
  user: User;
}

export const useAuthStore = defineStore("auth", {
  state: () => ({
    token: null as string | null,
    user: null as User | null,
    booted: false,
  }),
  getters: {
    isAuthenticated: (s) => !!s.token,
    isAdmin: (s) => s.user?.role === "admin",
  },
  actions: {
    /** Hydrate token from cookie. Called from middleware on every navigation. */
    hydrate() {
      if (this.token !== null) return;
      const c = useCookie<string | null>("access_token", {
        maxAge: 60 * 60 * 24 * 7,
        sameSite: "lax",
      });
      this.token = c.value;
    },
    /** Persist the token both in state and cookie. */
    setToken(token: string | null) {
      this.token = token;
      const cookie = useCookie<string | null>("access_token", {
        maxAge: 60 * 60 * 24 * 7,
        sameSite: "lax",
      });
      cookie.value = token;
    },
    async login(login: string, password: string) {
      // IMPORTANT: do NOT use useApi() here — login is the very first
      // request and we want to bypass any wrapper subtleties (interceptors,
      // SSR baseURL, etc). Login always runs in the browser.
      const cfg = useRuntimeConfig();
      const baseURL = (cfg.public.apiBase as string) || "";
      const res = await $fetch<LoginResponse>("/api/login", {
        method: "POST",
        baseURL,
        body: { login, password },
        headers: { "Content-Type": "application/json" },
      });
      this.setToken(res.token);
      this.user = res.user;
      // Mark as booted so the global middleware does NOT immediately re-run
      // fetchMe() on the post-login navigation. We already have the user.
      this.booted = true;
      return res.user;
    },
    async fetchMe() {
      if (!this.token) return null;
      try {
        const api = useApi();
        const res = await api.get<{ authenticated: boolean; user?: User }>(
          "/api/check-auth"
        );
        if (res.authenticated && res.user) {
          this.user = res.user;
        } else {
          this.setToken(null);
          this.user = null;
        }
      } catch {
        this.setToken(null);
        this.user = null;
      }
      this.booted = true;
      return this.user;
    },
    async logout() {
      const api = useApi();
      try {
        await api.post("/api/logout");
      } catch {
        /* ignore */
      }
      this.softLogout();
    },
    softLogout() {
      this.setToken(null);
      this.user = null;
    },
  },
});
