<script setup lang="ts">
/** Generic form built from a field schema. Used by entity CRUD pages
 *  to avoid writing 40 separate form components. */

export interface FormField {
  key: string;
  label: string;
  type?: "text" | "textarea" | "number" | "email" | "password" | "select" | "checkbox" | "tel" | "date";
  options?: { value: any; label: string }[];
  placeholder?: string;
  required?: boolean;
  cols?: 1 | 2 | 3; // grid span
}

const props = defineProps<{
  modelValue: Record<string, any>;
  fields: FormField[];
}>();

const emit = defineEmits<{
  (e: "update:modelValue", v: Record<string, any>): void;
}>();

const update = (key: string, val: any) => {
  emit("update:modelValue", { ...props.modelValue, [key]: val });
};
</script>

<template>
  <div class="grid grid-cols-1 md:grid-cols-2 gap-4">
    <div
      v-for="f in fields"
      :key="f.key"
      :class="['md:col-span-' + (f.cols || 1)]"
    >
      <label class="label">
        {{ f.label }}
        <span v-if="f.required" class="text-red-500">*</span>
      </label>

      <textarea
        v-if="f.type === 'textarea'"
        class="input min-h-[100px]"
        :value="modelValue[f.key]"
        :placeholder="f.placeholder"
        :required="f.required"
        @input="update(f.key, ($event.target as HTMLTextAreaElement).value)"
      />

      <select
        v-else-if="f.type === 'select'"
        class="input"
        :value="modelValue[f.key]"
        :required="f.required"
        @change="update(f.key, ($event.target as HTMLSelectElement).value)"
      >
        <option value="">—</option>
        <option v-for="o in f.options" :key="o.value" :value="o.value">
          {{ o.label }}
        </option>
      </select>

      <label
        v-else-if="f.type === 'checkbox'"
        class="flex items-center gap-2 mt-2 cursor-pointer"
      >
        <input
          type="checkbox"
          class="w-4 h-4 accent-brand"
          :checked="!!modelValue[f.key]"
          @change="update(f.key, ($event.target as HTMLInputElement).checked)"
        />
        <span class="text-sm text-ink-medium">{{ f.placeholder || f.label }}</span>
      </label>

      <input
        v-else
        :type="f.type || 'text'"
        class="input"
        :value="modelValue[f.key]"
        :placeholder="f.placeholder"
        :required="f.required"
        :step="f.type === 'number' ? 'any' : undefined"
        @input="update(f.key, f.type === 'number' ? Number(($event.target as HTMLInputElement).value) : ($event.target as HTMLInputElement).value)"
      />
    </div>
  </div>
</template>
