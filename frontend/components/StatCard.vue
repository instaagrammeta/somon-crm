<script setup lang="ts">
defineProps<{
  label: string;
  value: number | string;
  icon?: string;
  trend?: string;
  featured?: boolean;
  href?: string;
}>();
</script>

<template>
  <component
    :is="href ? 'NuxtLink' : 'div'"
    :to="href"
    class="card relative overflow-hidden p-6 transition-all duration-200 hover:-translate-y-0.5 hover:shadow-hover block"
    :class="featured ? '!bg-gradient-to-br from-brand to-brand-dark text-white !border-0' : ''"
  >
    <template v-if="featured">
      <span
        class="absolute -right-10 -top-10 w-36 h-36 rounded-full bg-white/[0.06]"
      />
      <span
        class="absolute -right-20 -bottom-20 w-48 h-48 rounded-full bg-white/[0.04]"
      />
    </template>

    <div class="relative flex items-center justify-between mb-4">
      <span
        :class="featured ? 'text-white/90' : 'text-ink-medium'"
        class="text-sm font-medium"
      >
        {{ label }}
      </span>
      <div
        :class="
          featured
            ? 'bg-white/20 text-white'
            : 'bg-brand-soft text-brand'
        "
        class="w-9 h-9 rounded-xl flex items-center justify-center text-sm transition-transform hover:rotate-[-45deg]"
      >
        <i class="fa-solid" :class="icon || 'fa-arrow-right'" />
      </div>
    </div>

    <div
      class="relative text-[42px] leading-none font-bold mb-3"
      :class="featured ? 'text-white' : 'text-ink'"
    >
      {{ value }}
    </div>

    <span
      v-if="trend"
      class="relative inline-flex items-center gap-1.5 px-2.5 py-1 text-[11px] font-semibold rounded-full"
      :class="
        featured ? 'bg-white/20 text-white' : 'bg-brand-soft text-brand'
      "
    >
      <i class="fa-solid fa-arrow-trend-up" />
      {{ trend }}
    </span>
  </component>
</template>
