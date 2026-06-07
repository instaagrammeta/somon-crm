<script setup lang="ts">
const { toasts } = useToast();

const icons = {
  success: "fa-circle-check",
  error: "fa-circle-exclamation",
  info: "fa-circle-info",
} as const;

const styles = {
  success: "bg-brand text-white",
  error: "bg-red-500 text-white",
  info: "bg-ink text-white",
} as const;
</script>

<template>
  <Teleport to="body">
    <div class="fixed top-5 right-5 z-[200] flex flex-col gap-2 max-w-sm">
      <TransitionGroup name="toast" tag="div" class="flex flex-col gap-2">
        <div
          v-for="t in toasts"
          :key="t.id"
          :class="[
            'rounded-2xl px-5 py-3.5 shadow-hover flex items-center gap-3 text-sm font-medium',
            styles[t.type],
          ]"
        >
          <i class="fa-solid" :class="icons[t.type]" />
          <span>{{ t.text }}</span>
        </div>
      </TransitionGroup>
    </div>
  </Teleport>
</template>

<style scoped>
.toast-enter-active,
.toast-leave-active {
  transition: all 0.25s ease;
}
.toast-enter-from {
  opacity: 0;
  transform: translateX(20px);
}
.toast-leave-to {
  opacity: 0;
  transform: translateX(20px);
}
</style>
