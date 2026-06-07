<script setup lang="ts">
defineProps<{ error: { statusCode: number; message: string } }>();
const handleError = () => clearError({ redirect: "/" });
</script>

<template>
  <div class="min-h-screen flex items-center justify-center bg-page p-4">
    <div class="max-w-md w-full bg-white rounded-3xl shadow-card p-10 text-center">
      <div class="text-7xl font-black text-brand mb-2">{{ error.statusCode }}</div>
      <p class="text-ink-medium mb-6">{{ error.message || "Хатогии номаълум" }}</p>
      <button
        class="px-6 py-3 bg-brand hover:bg-brand-dark text-white rounded-xl font-semibold transition-colors"
        @click="handleError"
      >
        Бозгашт
      </button>
    </div>
  </div>
</template>
