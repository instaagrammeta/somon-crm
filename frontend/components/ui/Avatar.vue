<script setup lang="ts">
import { initials, stringToColor } from "~/utils/format";

const props = defineProps<{
  src?: string;
  name?: string;
  size?: "xs" | "sm" | "md" | "lg";
}>();

const sizes: Record<string, string> = {
  xs: "w-7 h-7 text-[10px]",
  sm: "w-9 h-9 text-xs",
  md: "w-11 h-11 text-sm",
  lg: "w-16 h-16 text-base",
};
const ini = computed(() => initials(props.name));
const col = computed(() => stringToColor(props.name || "?"));
</script>

<template>
  <div
    class="relative inline-flex items-center justify-center font-bold rounded-full text-white shrink-0 overflow-hidden"
    :class="sizes[size || 'sm']"
    :style="!src ? { background: col } : {}"
  >
    <img v-if="src" :src="src" :alt="name" class="absolute inset-0 w-full h-full object-cover" />
    <span v-else>{{ ini }}</span>
  </div>
</template>
