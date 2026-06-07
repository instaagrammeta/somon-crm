<script setup lang="ts">
defineProps<{ icon?: string; title?: string; subtitle?: string }>();
</script>

<template>
  <div class="flex flex-col items-center justify-center py-16 px-4 text-center">
    <div
      class="w-20 h-20 rounded-3xl bg-brand-soft text-brand flex items-center justify-center text-3xl mb-4"
    >
      <i class="fa-solid" :class="icon || 'fa-inbox'" />
    </div>
    <h3 class="text-lg font-semibold text-ink mb-1">{{ title || $t("app.no_data") }}</h3>
    <p class="text-ink-soft text-sm max-w-xs">{{ subtitle }}</p>
    <slot />
  </div>
</template>
