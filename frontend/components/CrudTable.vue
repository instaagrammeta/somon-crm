<script setup lang="ts">
/** Reusable list with edit/delete actions, designed to keep most CRUD
 *  pages around 30 lines of code each. */
export interface Column {
  key: string;
  label: string;
  format?: (row: any) => string;
  badge?: (row: any) => { text: string; cls: string } | null;
}

defineProps<{
  rows: any[];
  columns: Column[];
  emptyIcon?: string;
  hideActions?: boolean;
}>();

const emit = defineEmits<{
  (e: "edit", row: any): void;
  (e: "delete", row: any): void;
}>();
</script>

<template>
  <UiEmpty v-if="!rows?.length" :icon="emptyIcon || 'fa-database'" />
  <div v-else class="overflow-x-auto -mx-6 lg:-mx-8">
    <table class="table-modern">
      <thead>
        <tr>
          <th v-for="c in columns" :key="c.key">{{ c.label }}</th>
          <th v-if="!hideActions" class="text-right">{{ $t("app.actions") }}</th>
        </tr>
      </thead>
      <tbody>
        <tr v-for="r in rows" :key="r.id">
          <td v-for="c in columns" :key="c.key">
            <span v-if="c.badge && c.badge(r)" :class="c.badge(r)?.cls">{{ c.badge(r)?.text }}</span>
            <template v-else>{{ c.format ? c.format(r) : (r[c.key] ?? "—") }}</template>
          </td>
          <td v-if="!hideActions" class="text-right whitespace-nowrap">
            <button class="btn-ghost !p-2" @click="emit('edit', r)">
              <i class="fa-solid fa-pen" />
            </button>
            <button class="btn-ghost !p-2 text-red-500" @click="emit('delete', r)">
              <i class="fa-solid fa-trash" />
            </button>
          </td>
        </tr>
      </tbody>
    </table>
  </div>
</template>
