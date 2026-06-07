// https://nuxt.com/docs/api/configuration/nuxt-config
export default defineNuxtConfig({
  compatibilityDate: "2024-11-01",
  devtools: { enabled: true },

  // Disable server-side rendering: the app runs as SPA.
  // Reasoning: this CRM is fully behind auth — there's nothing to SEO-optimize,
  // and SSR was crashing because some pages call useFetch before login.
  // Nitro still serves the static index.html + handles assets, but pages
  // hydrate fully on the client where the auth cookie is available.
  ssr: false,

  modules: [
    "@nuxtjs/tailwindcss",
    "@nuxtjs/i18n",
    "@pinia/nuxt",
    "@vueuse/nuxt",
  ],

  css: ["~/assets/css/main.css"],

  app: {
    head: {
      htmlAttrs: { lang: "tg" },
      title: "Somon CRM",
      meta: [
        { charset: "utf-8" },
        { name: "viewport", content: "width=device-width, initial-scale=1" },
        { name: "theme-color", content: "#1f7a4d" },
      ],
      link: [
        { rel: "icon", type: "image/svg+xml", href: "/favicon.svg" },
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

  pinia: {
    storesDirs: ["./stores/**"],
  },

  typescript: {
    strict: false,
    typeCheck: false,
  },

  nitro: {
    compressPublicAssets: true,
  },

  vite: {
    css: { preprocessorOptions: {} },
  },
});
