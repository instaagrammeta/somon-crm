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

  // Lazy-load the current user once. This must never throw — otherwise the
  // whole navigation aborts and Nuxt renders the fatal 500 error page (which
  // is exactly what happened on every cold reload). fetchMe already guards
  // itself, but we wrap here as well for defense in depth.
  if (auth.token && !auth.booted) {
    try {
      await auth.fetchMe();
    } catch {
      auth.booted = true;
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
