/**
 * Thin wrapper around `$fetch` that automatically:
 *  - chooses the right base URL: SSR → backend container, browser → nginx
 *  - attaches the JWT bearer from Pinia store
 *  - handles 401 → forced logout + redirect
 *  - exposes a `download()` helper for binary endpoints (Excel)
 */
import type { FetchOptions } from "ofetch";

export function useApi() {
  const cfg = useRuntimeConfig();
  const auth = useAuthStore();
  const router = useRouter();
  const { locale } = useI18n();

  // SSR (Nitro inside the frontend container) uses Docker DNS to reach the
  // backend; the browser hits nginx (relative URL).
  const baseURL = import.meta.server
    ? (cfg.apiBaseInternal as string) || "http://backend:8080"
    : (cfg.public.apiBase as string) || "";

  const buildOpts = (opts: FetchOptions = {}): FetchOptions => {
    const headers: Record<string, string> = {
      Accept: "application/json",
      "Accept-Language": locale.value || "tg",
      ...((opts.headers as Record<string, string>) || {}),
    };
    if (auth.token) {
      headers.Authorization = `Bearer ${auth.token}`;
    }
    return {
      baseURL,
      credentials: "include",
      ...opts,
      headers,
      onResponseError: async (ctx) => {
        if (ctx.response?.status === 401 && auth.token) {
          await auth.softLogout();
          if (import.meta.client) router.push("/login");
        }
        if (opts.onResponseError) await opts.onResponseError(ctx);
      },
    };
  };

  const get = <T = any>(url: string, opts?: FetchOptions) =>
    $fetch<T>(url, { method: "GET", ...buildOpts(opts) });

  const post = <T = any>(url: string, body?: any, opts?: FetchOptions) =>
    $fetch<T>(url, { method: "POST", body, ...buildOpts(opts) });

  const put = <T = any>(url: string, body?: any, opts?: FetchOptions) =>
    $fetch<T>(url, { method: "PUT", body, ...buildOpts(opts) });

  const patch = <T = any>(url: string, body?: any, opts?: FetchOptions) =>
    $fetch<T>(url, { method: "PATCH", body, ...buildOpts(opts) });

  const del = <T = any>(url: string, opts?: FetchOptions) =>
    $fetch<T>(url, { method: "DELETE", ...buildOpts(opts) });

  /** Download a binary file (Excel/PDF). Triggers browser save dialog.
   *  Always client-only (binary ops aren't useful in SSR). */
  const download = async (url: string, filename: string) => {
    if (!import.meta.client) return;
    const blob = await $fetch<Blob>(url, {
      method: "GET",
      ...buildOpts({}),
      baseURL: cfg.public.apiBase as string || "",
      responseType: "blob",
    });
    const u = URL.createObjectURL(blob);
    const a = document.createElement("a");
    a.href = u;
    a.download = filename;
    a.click();
    URL.revokeObjectURL(u);
  };

  /** Upload a multipart form. */
  const upload = <T = any>(url: string, form: FormData) =>
    post<T>(url, form);

  return { get, post, put, patch, del, download, upload, baseURL };
}
