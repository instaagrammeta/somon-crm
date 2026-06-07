<script setup lang="ts">
const props = defineProps<{
  modelValue: boolean;
  title?: string;
  message?: string;
  okText?: string;
  cancelText?: string;
  danger?: boolean;
}>();

const emit = defineEmits<{
  (e: "update:modelValue", v: boolean): void;
  (e: "confirm"): void;
}>();

const close = () => emit("update:modelValue", false);
const confirm = () => {
  emit("confirm");
  close();
};
</script>

<template>
  <UiModal v-model="props.modelValue" size="sm" :title="title">
    <p class="text-ink-medium leading-relaxed">{{ message }}</p>
    <template #footer>
      <button class="btn-outline" @click="close">{{ cancelText || $t("app.cancel") }}</button>
      <button
        :class="danger ? 'btn-danger' : 'btn-primary'"
        @click="confirm"
      >
        {{ okText || $t("app.confirm") }}
      </button>
    </template>
  </UiModal>
</template>
