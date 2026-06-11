/**
 * Thin wrapper around `$fetch` that automatically:
 *  - prefixes paths with the configured API base
 *  - attaches the JWT bearer from Pinia store
 *  - handles 401 → forced logout + redirect
 *  - exposes a `download()` helper for binary endpoints (Excel)
 *
 * SPA-only mode: every call runs in the browser, so we always hit the
 * relative `/api/*` URL which nginx proxies to the Go backend.
 *
 * IMPORTANT: `useApi()` must be safe to call from ANY context, not just a
 * component `setup()`. It is invoked from `auth.fetchMe()` which runs inside
 * the global route middleware (and on cold reload). Composables that are only
 * valid inside `setup()` — notably `useI18n()` — would throw there and crash
 * navigation (manifesting as a 500 page on every reload and a bogus
 * "invalid credentials" error right after a successful login). We therefore
 * avoid `useI18n()`/`useRouter()` at the top level and read the locale /
 * perform redirects defensively.
 */
import type { FetchOptions } from "ofetch";

export function useApi() {
  const cfg = useRuntimeConfig();
  const auth = useAuthStore();

  const baseURL = (cfg.public.apiBase as string) || "";

  /**
   * Resolve the active locale without `useI18n()` (which is setup-only).
   * Falls back to the i18n cookie and then the configured default so the
   * Accept-Language header is best-effort and never throws.
   */
  const currentLocale = (): string => {
    try {
      const i18n: any = useNuxtApp().$i18n;
      const loc = i18n?.locale;
      const val = typeof loc === "string" ? loc : loc?.value;
      if (val) return val;
    } catch {
      /* not inside an i18n-aware context */
    }
    try {
      const cookie = useCookie<string | null>("i18n_redirected");
      if (cookie.value) return cookie.value;
    } catch {
      /* ignore */
    }
    return (cfg.public.defaultLocale as string) || "tg";
  };

  const buildOpts = (opts: FetchOptions = {}): FetchOptions => {
    const headers: Record<string, string> = {
      Accept: "application/json",
      "Accept-Language": currentLocale(),
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
          auth.softLogout();
          // navigateTo is safe to call from any context (unlike router.push
          // obtained via useRouter() at the top level).
          await navigateTo("/login");
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

  /** Download a binary file (Excel/PDF). Triggers browser save dialog. */
  const download = async (url: string, filename: string) => {
    const blob = await $fetch<Blob>(url, {
      method: "GET",
      ...buildOpts({}),
      responseType: "blob",
    });
    const u = URL.createObjectURL(blob);
    const a = document.createElement("a");
    a.href = u;
    a.download = filename;
    a.click();
    URL.revokeObjectURL(u);
  };

  /**
   * Upload a multipart form. Defaults to POST (create); pass "PUT" to update an
   * existing resource — the Go backend's update routes are registered as PUT.
   */
  const upload = <T = any>(
    url: string,
    form: FormData,
    method: "POST" | "PUT" = "POST"
  ) => $fetch<T>(url, { method, body: form, ...buildOpts() });

  /**
   * Upload a multipart form with real upload progress (0–100). `$fetch`/`fetch`
   * cannot report request-body progress, so this uses XMLHttpRequest directly
   * while still attaching the bearer token + locale header and handling 401 the
   * same way as the rest of the API. The returned promise resolves with the
   * parsed JSON body (or rejects with `{ status, data }` mirroring ofetch).
   */
  const uploadProgress = <T = any>(
    url: string,
    form: FormData,
    method: "POST" | "PUT" = "POST",
    onProgress?: (percent: number) => void
  ) =>
    new Promise<T>((resolve, reject) => {
      const xhr = new XMLHttpRequest();
      xhr.open(method, baseURL + url);
      xhr.responseType = "json";
      xhr.withCredentials = true;
      xhr.setRequestHeader("Accept", "application/json");
      xhr.setRequestHeader("Accept-Language", currentLocale());
      if (auth.token) xhr.setRequestHeader("Authorization", `Bearer ${auth.token}`);

      if (xhr.upload) {
        xhr.upload.onprogress = (e: ProgressEvent) => {
          if (onProgress && e.lengthComputable) {
            onProgress(Math.round((e.loaded / e.total) * 100));
          }
        };
      }
      xhr.onload = () => {
        const body =
          xhr.response ??
          (() => {
            try {
              return JSON.parse(xhr.responseText);
            } catch {
              return null;
            }
          })();
        if (xhr.status >= 200 && xhr.status < 300) {
          if (onProgress) onProgress(100);
          resolve(body as T);
          return;
        }
        if (xhr.status === 401 && auth.token) {
          auth.softLogout();
          navigateTo("/login");
        }
        reject({ status: xhr.status, data: body });
      };
      xhr.onerror = () => reject({ status: 0, data: null });
      xhr.ontimeout = () => reject({ status: 0, data: null });
      xhr.send(form);
    });

  return { get, post, put, patch, del, download, upload, uploadProgress, baseURL };
}
