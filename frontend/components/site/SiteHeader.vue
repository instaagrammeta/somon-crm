<script setup lang="ts">
import type { SiteConfig } from "~/types/public";

defineProps<{ config: SiteConfig }>();

const links = [
  { to: "/", label: "Хона" },
  { to: "/properties", label: "Манзилҳо" },
  { to: "/about", label: "Дар бораи мо" },
  { to: "/contact", label: "Тамос" },
  { to: "/install", label: "Замимаи мобилӣ" },
];

const open = ref(false);
const route = useRoute();
watch(() => route.path, () => (open.value = false));
</script>

<template>
  <header class="sticky top-0 z-20 bg-white/85 backdrop-blur border-b border-border-soft">
    <div class="max-w-6xl mx-auto px-4 h-16 flex items-center justify-between">
      <NuxtLink to="/" class="flex items-center gap-2 font-bold text-lg text-brand">
        <span class="w-8 h-8 rounded-xl bg-brand text-white grid place-items-center">
          <i class="fas fa-house-chimney text-sm" />
        </span>
        {{ config?.brand_name || "Somon Real Estate" }}
      </NuxtLink>

      <nav class="hidden md:flex items-center gap-1">
        <NuxtLink
          v-for="l in links"
          :key="l.to"
          :to="l.to"
          class="px-3 py-2 rounded-lg text-sm font-medium hover:bg-brand-soft hover:text-brand transition-colors"
          active-class="bg-brand-soft text-brand"
        >
          {{ l.label }}
        </NuxtLink>
        <a
          v-if="config?.phone"
          :href="`tel:${config.phone}`"
          class="ml-2 inline-flex items-center gap-2 px-4 py-2 rounded-full bg-brand text-white text-sm font-semibold hover:bg-brand-dark"
        >
          <i class="fas fa-phone" /> {{ config.phone }}
        </a>
      </nav>

      <button
        class="md:hidden w-10 h-10 rounded-lg hover:bg-brand-soft text-ink"
        @click="open = !open"
      >
        <i :class="open ? 'fas fa-xmark' : 'fas fa-bars'" />
      </button>
    </div>

    <div v-if="open" class="md:hidden border-t border-border-soft bg-white">
      <NuxtLink
        v-for="l in links"
        :key="l.to"
        :to="l.to"
        class="block px-4 py-3 hover:bg-brand-soft hover:text-brand"
      >
        {{ l.label }}
      </NuxtLink>
    </div>
  </header>
</template>
