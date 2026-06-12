/**
 * usePublicApi — fetch helper for the marketing website (no auth headers).
 *
 * The CRM-side useApi() forces a Bearer token and aborts navigation on 401.
 * Public pages must not do that — they must work for anonymous visitors who
 * never had a token in the first place. Hence this very thin wrapper around
 * Nuxt's $fetch with the right base URL.
 */
export const usePublicApi = () => {
  const baseURL = useRuntimeConfig().public.apiBase || "";

  const get = <T = any>(url: string, opts: any = {}) =>
    $fetch<T>(url, { method: "GET", baseURL, ...opts });

  const post = <T = any>(url: string, body: any, opts: any = {}) =>
    $fetch<T>(url, { method: "POST", baseURL, body, ...opts });

  return { get, post, baseURL };
};
