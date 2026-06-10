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
      return res.user;
    },
    async fetchMe() {
      if (!this.token) return null;
      // IMPORTANT: do NOT use useApi() here. fetchMe() is called from the
      // global route middleware (a Pinia-action context, not a component
      // setup), where useRouter()/useI18n() inside useApi() throw
      // "Nuxt instance unavailable". A throw here escapes the middleware and
      // makes Nuxt render the 500 error page on every page reload.
      // We use $fetch directly, exactly like login(), and never throw out.
      const cfg = useRuntimeConfig();
      const baseURL = (cfg.public.apiBase as string) || "";
      try {
        const res = await $fetch<{ authenticated: boolean; user?: User }>(
          "/api/check-auth",
          {
            baseURL,
            headers: {
              Authorization: `Bearer ${this.token}`,
              Accept: "application/json",
            },
          }
        );
        if (res.authenticated && res.user) {
          this.user = res.user;
        } else {
          // Server says the token is no longer valid.
          this.setToken(null);
          this.user = null;
        }
      } catch (e: any) {
        // Only drop the session on a real auth rejection (401/403).
        // Network errors or a transient 500 must NOT log the user out or
        // crash navigation.
        const status = e?.response?.status ?? e?.statusCode;
        if (status === 401 || status === 403) {
          this.setToken(null);
          this.user = null;
        }
      } finally {
        this.booted = true;
      }
      return this.user;
    },
    async logout() {
      // Same reasoning as fetchMe: avoid useApi() outside of setup context.
      const cfg = useRuntimeConfig();
      const baseURL = (cfg.public.apiBase as string) || "";
      if (this.token) {
        try {
          await $fetch("/api/logout", {
            method: "POST",
            baseURL,
            headers: { Authorization: `Bearer ${this.token}` },
          });
        } catch {
          /* ignore — we log out locally regardless */
        }
      }
      this.softLogout();
    },
    softLogout() {
      this.setToken(null);
      this.user = null;
    },
  },
});
