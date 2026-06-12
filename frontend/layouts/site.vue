<script setup lang="ts">
import type { SiteConfig } from "~/types/public";

const api = usePublicApi();
const { data: config } = useAsyncData<SiteConfig>("site-config", () =>
  api.get<SiteConfig>("/api/public/site/config"), { default: () => ({} as SiteConfig) });

provide("siteConfig", config);

const { locale, locales, setLocale } = useI18n();
const dark = useDark();
</script>

<template>
  <div class="min-h-screen flex flex-col bg-page text-ink">
    <SiteHeader :config="config" />
    <main class="flex-1">
      <slot />
    </main>
    <SiteFooter :config="config" />

    <!-- Floating language + dark-mode toggles, always visible -->
    <div class="fixed bottom-4 right-4 flex gap-2 z-30">
      <button
        class="w-11 h-11 rounded-full bg-white/90 backdrop-blur shadow-lg border border-border-soft hover:bg-white text-ink"
        :title="dark ? 'Light' : 'Dark'"
        @click="dark = !dark"
      >
        <i :class="dark ? 'fas fa-sun' : 'fas fa-moon'" />
      </button>
      <select
        v-model="locale"
        class="h-11 rounded-full bg-white/90 backdrop-blur shadow-lg border border-border-soft px-3 text-sm"
        @change="setLocale(locale as any)"
      >
        <option v-for="l in (locales as any[])" :key="l.code" :value="l.code">
          {{ l.name }}
        </option>
      </select>
    </div>

    <!-- Auto lead-capture popup: 30s after first visit (per session) -->
    <SiteLeadPopup />
  </div>
</template>
