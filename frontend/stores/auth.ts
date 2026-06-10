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
        path: "/",
      });
      this.token = c.value;
    },
    /** Persist the token both in state and cookie. */
    setToken(token: string | null) {
      this.token = token;
      const cookie = useCookie<string | null>("access_token", {
        maxAge: 60 * 60 * 24 * 7,
        sameSite: "lax",
        path: "/",
      });
      cookie.value = token;
      // When logging out, remove the cookie synchronously from document.cookie
      // as well. useCookie writes can be deferred to the next flush, and if the
      // very next navigation re-reads a stale cookie in hydrate() the user gets
      // bounced straight back into the CRM (i.e. "logout doesn't work").
      if (import.meta.client && !token) {
        document.cookie = "access_token=; path=/; max-age=0; SameSite=Lax";
      }
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
      // Best-effort server-side blacklist; never block local logout on it.
      try {
        const api = useApi();
        await api.post("/api/logout");
      } catch {
        /* ignore network/parse errors — we log out locally regardless */
      }
      this.softLogout();
      this.booted = true;
    },
    softLogout() {
      this.setToken(null);
      this.user = null;
    },
  },
});
