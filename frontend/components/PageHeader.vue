<script setup lang="ts">
defineProps<{ title: string; subtitle?: string }>();
</script>

<template>
  <div class="flex items-start flex-wrap justify-between gap-4 mb-6">
    <div>
      <h1
        class="text-3xl md:text-[32px] font-bold text-ink leading-tight tracking-tight"
      >
        {{ title }}
      </h1>
      <p v-if="subtitle" class="text-ink-soft mt-1.5 text-sm">{{ subtitle }}</p>
    </div>
    <div class="flex flex-wrap items-center gap-2">
      <slot />
    </div>
  </div>
</template>
