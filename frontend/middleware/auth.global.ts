/** Auto-redirect rules:
 *  - if not logged in and route !== /login → /login
 *  - if logged in and route === /login → /
 *  - admin-only routes are checked here as well
 */
export default defineNuxtRouteMiddleware(async (to) => {
  const auth = useAuthStore();
  // Hydrate token from cookie on every nav (works in both SSR and client).
  auth.hydrate();

  // Lazily load the current user once.
  if (import.meta.client && auth.token && !auth.booted) {
    await auth.fetchMe();
  }

  const isPublic = to.path === "/login";

  if (!auth.token && !isPublic) {
    return navigateTo("/login");
  }
  if (auth.token && isPublic) {
    return navigateTo("/");
  }

  // Admin-only paths (kept simple — fine for v1).
  const adminPaths = ["/users"];
  if (
    auth.token &&
    auth.user &&
    auth.user.role !== "admin" &&
    adminPaths.some((p) => to.path.startsWith(p))
  ) {
    return navigateTo("/");
  }
});
