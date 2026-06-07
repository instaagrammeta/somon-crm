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
    token: useCookie<string | null>("access_token", {
      maxAge: 60 * 60 * 24 * 7,
      sameSite: "lax",
    }).value,
    user: null as User | null,
    booted: false,
  }),
  getters: {
    isAuthenticated: (s) => !!s.token,
    isAdmin: (s) => s.user?.role === "admin",
  },
  actions: {
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
      const cfg = useRuntimeConfig();
      const res = await $fetch<LoginResponse>("/api/login", {
        method: "POST",
        baseURL: cfg.public.apiBase || "",
        body: { login, password },
      });
      this.setToken(res.token);
      this.user = res.user;
      return res.user;
    },
    async fetchMe() {
      if (!this.token) return null;
      const cfg = useRuntimeConfig();
      try {
        const res = await $fetch<{ authenticated: boolean; user?: User }>(
          "/api/check-auth",
          {
            baseURL: cfg.public.apiBase || "",
            headers: { Authorization: `Bearer ${this.token}` },
          }
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
      const cfg = useRuntimeConfig();
      try {
        await $fetch("/api/logout", {
          method: "POST",
          baseURL: cfg.public.apiBase || "",
          headers: { Authorization: `Bearer ${this.token}` },
        });
      } catch {
        /* ignore */
      }
      this.softLogout();
    },
    async softLogout() {
      this.setToken(null);
      this.user = null;
    },
  },
});
