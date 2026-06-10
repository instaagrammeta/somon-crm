/** Auto-redirect rules:
 *  - if not logged in and route !== /login → /login
 *  - if logged in and route === /login → /
 *  - admin-only routes are checked here as well
 *
 * Runs only in the browser (SPA mode).
 */
export default defineNuxtRouteMiddleware(async (to) => {
  if (import.meta.server) return; // safety: never run on server

  const auth = useAuthStore();
  auth.hydrate();

  // Lazy-load the current user once. Must never throw — otherwise Nuxt
  // renders the 500 error page on every reload.
  if (auth.token && !auth.booted) {
    try {
      await auth.fetchMe();
    } catch {
      /* ignore: fetchMe already handles its own failures */
    }
  }

  const isPublic = to.path === "/login";

  if (!auth.token && !isPublic) {
    return navigateTo("/login");
  }
  if (auth.token && isPublic) {
    return navigateTo("/");
  }

  // Admin-only paths.
  const adminPaths = ["/admin"];
  if (
    auth.token &&
    auth.user &&
    auth.user.role !== "admin" &&
    adminPaths.some((p) => to.path.startsWith(p))
  ) {
    return navigateTo("/");
  }
});
