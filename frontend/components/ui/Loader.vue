<script setup lang="ts">
defineProps<{ inline?: boolean; text?: string }>();
</script>

<template>
  <div
    :class="inline ? 'inline-flex' : 'flex'"
    class="items-center justify-center gap-3 py-10 text-ink-soft"
  >
    <div
      class="w-6 h-6 border-[3px] border-brand-soft border-t-brand rounded-full animate-spin"
    />
    <span class="text-sm font-medium">{{ text || $t("app.loading_data") }}</span>
  </div>
</template>
