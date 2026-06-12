/** Auto-redirect rules:
 *  - public website routes (/, /houses, /tour/...) are accessible to everyone
 *  - if not logged in and route is admin/CRM → /login
 *  - if logged in and route === /login → /admin (CRM home)
 *  - admin-only routes are checked here as well
 *
 * Runs only in the browser (SPA mode).
 */

// Path prefixes that DO NOT require auth. Everything else does.
// /api/* is server-only; we only check the SPA route here.
const PUBLIC_PREFIXES = [
  "/",            // public marketing home
  "/site",        // explicit alias
  "/properties",  // public catalog (CRM keeps /houses for itself)
  "/tour",        // public virtual tour viewer
  "/contact",     // public contact form
  "/about",       // about page
  "/login",       // CRM login form (still public)
  "/install",     // PWA install / APK landing
];

function isPublic(path: string): boolean {
  // Exact match for "/" because startsWith("/") is true for everything.
  if (path === "/") return true;
  return PUBLIC_PREFIXES.some(
    (p) => p !== "/" && (path === p || path.startsWith(p + "/"))
  );
}

export default defineNuxtRouteMiddleware(async (to) => {
  if (import.meta.server) return; // safety: never run on server

  const auth = useAuthStore();
  auth.hydrate();

  // Boot the user profile once if we have a token. fetchMe is best-effort and
  // never throws — but wrap defensively so a network blip can't take down
  // navigation.
  if (auth.token && !auth.booted) {
    try {
      await auth.fetchMe();
    } catch {
      auth.booted = true;
    }
  }

  // Public website + login: no auth checks.
  if (isPublic(to.path)) {
    // If the user is already logged-in and lands on /login, send them to CRM.
    if (auth.token && to.path === "/login") {
      return navigateTo("/dashboard");
    }
    return;
  }

  // Anything else (CRM): require login.
  if (!auth.token) {
    return navigateTo("/login");
  }

  // Admin-only paths.
  const adminPaths = ["/admin"];
  if (
    auth.user &&
    auth.user.role !== "admin" &&
    adminPaths.some((p) => to.path === p || to.path.startsWith(p + "/"))
  ) {
    return navigateTo("/dashboard");
  }
});
