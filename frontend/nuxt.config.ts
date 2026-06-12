// https://nuxt.com/docs/api/configuration/nuxt-config
//
// v-2 changes vs v-1-соз:
//  • added @vite-pwa/nuxt → installable PWA + offline + push notifications
//  • added Marzipano + Three.js (virtual tour viewer); both are client-only
//  • the public marketing website lives under /site/* and is reachable
//    without auth (handled by middleware/auth.global.ts).
export default defineNuxtConfig({
  compatibilityDate: "2024-11-01",
  devtools: { enabled: true },

  // SPA mode: the CRM is fully behind auth, the public site uses no SSR
  // either (Marzipano + most of the panorama controls require window).
  ssr: false,

  modules: [
    "@nuxtjs/tailwindcss",
    "@nuxtjs/i18n",
    "@pinia/nuxt",
    "@vueuse/nuxt",
    "@vite-pwa/nuxt",
  ],

  css: ["~/assets/css/main.css"],

  app: {
    head: {
      htmlAttrs: { lang: "tg" },
      title: "Somon Real Estate",
      meta: [
        { charset: "utf-8" },
        { name: "viewport", content: "width=device-width, initial-scale=1" },
        { name: "theme-color", content: "#1f7a4d" },
        {
          name: "description",
          content:
            "Somon Real Estate — биноҳои нав, виртуалии турҳои 360°, лоиҳаҳои сохтмонии Тоҷикистон.",
        },
      ],
      link: [
        { rel: "icon", type: "image/svg+xml", href: "/favicon.svg" },
        { rel: "manifest", href: "/manifest.webmanifest" },
        {
          rel: "stylesheet",
          href: "https://fonts.googleapis.com/css2?family=Inter:wght@300;400;500;600;700;800&display=swap",
        },
        {
          rel: "stylesheet",
          href: "https://cdnjs.cloudflare.com/ajax/libs/font-awesome/6.5.1/css/all.min.css",
        },
      ],
    },
  },

  runtimeConfig: {
    public: {
      apiBase: process.env.NUXT_PUBLIC_API_BASE || "",
      defaultLocale: process.env.NUXT_PUBLIC_DEFAULT_LOCALE || "tg",
      vapidKey: process.env.NUXT_PUBLIC_VAPID_KEY || "",
    },
  },

  i18n: {
    strategy: "no_prefix",
    defaultLocale: "tg",
    locales: [
      { code: "tg", name: "Тоҷикӣ", file: "tg.json" },
      { code: "ru", name: "Русский", file: "ru.json" },
    ],
    detectBrowserLanguage: {
      useCookie: true,
      cookieKey: "i18n_redirected",
      redirectOn: "root",
      alwaysRedirect: false,
      fallbackLocale: "tg",
    },
    bundle: { optimizeTranslationDirective: false },
  },

  pinia: { storesDirs: ["./stores/**"] },

  typescript: { strict: false, typeCheck: false },

  nitro: { compressPublicAssets: true },

  // PWA: installable + offline shell + push notifications.
  // The dev mode is disabled to keep HMR simple; on `nuxt build` the SW is
  // emitted into `.output/public/sw.js`.
  pwa: {
    registerType: "autoUpdate",
    strategies: "generateSW",
    manifest: {
      name: "Somon Real Estate",
      short_name: "Somon",
      description: "CRM ва вебсайти расмии Somon Real Estate",
      lang: "tg",
      theme_color: "#1f7a4d",
      background_color: "#ffffff",
      display: "standalone",
      start_url: "/",
      scope: "/",
      icons: [
        { src: "/icons/icon-192.png", sizes: "192x192", type: "image/png" },
        { src: "/icons/icon-512.png", sizes: "512x512", type: "image/png" },
        {
          src: "/icons/icon-maskable.png",
          sizes: "512x512",
          type: "image/png",
          purpose: "maskable",
        },
      ],
    },
    workbox: {
      navigateFallback: "/",
      // Cache the public-site read API + uploaded panorama tiles + assets.
      runtimeCaching: [
        {
          urlPattern: ({ url }) =>
            url.pathname.startsWith("/api/public/houses") ||
            url.pathname.startsWith("/api/public/tours") ||
            url.pathname.startsWith("/api/public/site"),
          handler: "NetworkFirst",
          options: {
            cacheName: "public-api",
            networkTimeoutSeconds: 5,
            expiration: { maxEntries: 200, maxAgeSeconds: 60 * 60 * 24 },
          },
        },
        {
          urlPattern: ({ url }) => url.pathname.startsWith("/uploads/"),
          handler: "CacheFirst",
          options: {
            cacheName: "uploads",
            expiration: { maxEntries: 500, maxAgeSeconds: 60 * 60 * 24 * 30 },
          },
        },
      ],
      globPatterns: ["**/*.{js,css,html,svg,png,ico,webp}"],
    },
    client: { installPrompt: true },
    devOptions: { enabled: false },
  },

  vite: {
    // Marzipano ships its own UMD-style file that doesn't always tree-shake
    // cleanly; ensure it's bundled as commonjs but only on the client.
    optimizeDeps: { include: ["marzipano"] },
    css: { preprocessorOptions: {} },
    ssr: { noExternal: ["marzipano", "three"] },
  },
});
