<script setup lang="ts">
const props = defineProps<{
  modelValue: boolean;
  title?: string;
  size?: "sm" | "md" | "lg" | "xl";
  hideClose?: boolean;
}>();

const emit = defineEmits<{ (e: "update:modelValue", v: boolean): void }>();

const close = () => emit("update:modelValue", false);

const widths: Record<string, string> = {
  sm: "max-w-md",
  md: "max-w-xl",
  lg: "max-w-3xl",
  xl: "max-w-5xl",
};

watch(
  () => props.modelValue,
  (v) => {
    if (process.client) {
      document.body.style.overflow = v ? "hidden" : "";
    }
  }
);

onUnmounted(() => {
  if (process.client) document.body.style.overflow = "";
});
</script>

<template>
  <Teleport to="body">
    <Transition name="modal">
      <div
        v-if="modelValue"
        class="fixed inset-0 z-[150] flex items-center justify-center p-4 bg-ink/40 backdrop-blur-sm"
        @click.self="close"
      >
        <div
          :class="[
            'bg-white rounded-3xl shadow-hover w-full max-h-[92vh] overflow-hidden flex flex-col',
            widths[size || 'md'],
          ]"
        >
          <header
            v-if="title || !hideClose"
            class="flex items-center justify-between px-6 py-4 border-b border-border-soft"
          >
            <h3 class="font-bold text-lg text-ink">{{ title }}</h3>
            <button
              v-if="!hideClose"
              class="w-8 h-8 rounded-xl bg-page hover:bg-brand-soft flex items-center justify-center text-ink-medium hover:text-brand"
              @click="close"
            >
              <i class="fa-solid fa-xmark" />
            </button>
          </header>
          <div class="overflow-y-auto p-6">
            <slot />
          </div>
          <footer
            v-if="$slots.footer"
            class="px-6 py-4 border-t border-border-soft flex justify-end gap-2"
          >
            <slot name="footer" />
          </footer>
        </div>
      </div>
    </Transition>
  </Teleport>
</template>

<style scoped>
.modal-enter-active,
.modal-leave-active {
  transition: opacity 0.2s ease;
}
.modal-enter-active > div,
.modal-leave-active > div {
  transition: transform 0.2s ease;
}
.modal-enter-from,
.modal-leave-to {
  opacity: 0;
}
.modal-enter-from > div,
.modal-leave-to > div {
  transform: scale(0.95);
}
</style>
