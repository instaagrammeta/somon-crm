<script setup lang="ts">
import type { SiteConfig } from "~/types/public";
defineProps<{ config: SiteConfig }>();
</script>

<template>
  <footer class="mt-16 bg-ink text-white/85">
    <div class="max-w-6xl mx-auto px-4 py-10 grid gap-8 md:grid-cols-4">
      <div>
        <div class="flex items-center gap-2 font-bold text-lg text-white">
          <span class="w-8 h-8 rounded-xl bg-brand text-white grid place-items-center">
            <i class="fas fa-house-chimney text-sm" />
          </span>
          {{ config?.brand_name || "Somon Real Estate" }}
        </div>
        <p class="mt-3 text-sm leading-relaxed text-white/60">
          Биноҳои нав, манзилҳо ва турҳои виртуалии 360°. Шумо метавонед хонаро аз
          ҳар нуқтаи ҷаҳон бубинед пеш аз харид.
        </p>
      </div>

      <div>
        <h4 class="font-semibold text-white mb-3">Маҳсулот</h4>
        <ul class="space-y-2 text-sm text-white/70">
          <li><NuxtLink to="/properties" class="hover:text-white">Манзилҳо</NuxtLink></li>
          <li><NuxtLink to="/about" class="hover:text-white">Дар бораи мо</NuxtLink></li>
          <li><NuxtLink to="/install" class="hover:text-white">Замимаи мобилӣ</NuxtLink></li>
        </ul>
      </div>

      <div>
        <h4 class="font-semibold text-white mb-3">Тамос</h4>
        <ul class="space-y-2 text-sm text-white/70">
          <li v-if="config?.phone"><i class="fas fa-phone mr-2" />{{ config.phone }}</li>
          <li v-if="config?.email"><i class="fas fa-envelope mr-2" />{{ config.email }}</li>
          <li v-if="config?.address"><i class="fas fa-location-dot mr-2" />{{ config.address }}</li>
        </ul>
      </div>

      <div>
        <h4 class="font-semibold text-white mb-3">Барои корфармоён</h4>
        <NuxtLink to="/login" class="text-sm text-white/70 hover:text-white">
          <i class="fas fa-arrow-right-to-bracket mr-2" />Вуруд ба CRM
        </NuxtLink>
      </div>
    </div>

    <div class="border-t border-white/10">
      <div class="max-w-6xl mx-auto px-4 py-4 text-xs text-white/50 flex justify-between">
        <span>© {{ new Date().getFullYear() }} {{ config?.brand_name || "Somon" }}</span>
        <span>Сохтаи Somon Tech</span>
      </div>
    </div>
  </footer>
</template>
