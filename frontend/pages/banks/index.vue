<script setup lang="ts">
import type { Bank } from "~/types";
import type { FormField } from "~/components/EntityForm.vue";

useHead({ title: () => useI18n().t("nav.banks") });
const { t } = useI18n();
const c = useCrud<Bank>("/api/banks", "banks");

const fields = computed<FormField[]>(() => [
  { key: "name", label: t("bank.name"), required: true, cols: 2 },
  { key: "description", label: t("bank.description"), type: "textarea", cols: 2 },
  { key: "order_index", label: "Order", type: "number" },
  { key: "is_active", label: t("user.active"), type: "checkbox" },
]);
</script>

<template>
  <div class="card p-6 lg:p-8">
    <PageHeader :title="$t('nav.banks')">
      <button class="btn-primary" @click="c.openCreate({ is_active: true })">
        <i class="fa-solid fa-plus" /> {{ $t("app.create") }}
      </button>
    </PageHeader>

    <UiLoader v-if="c.pending.value" />
    <UiEmpty v-else-if="!c.rows.value?.length" icon="fa-building-columns" />

    <div v-else class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
      <article v-for="b in c.rows.value" :key="b.id" class="card p-6 hover:-translate-y-0.5 hover:shadow-hover transition-all">
        <div class="flex items-center gap-4 mb-4">
          <div v-if="b.logo" class="w-14 h-14 rounded-2xl bg-page p-2 flex items-center justify-center overflow-hidden">
            <img :src="b.logo" class="w-full h-full object-contain" :alt="b.name" />
          </div>
          <div v-else class="w-14 h-14 rounded-2xl bg-brand-soft text-brand flex items-center justify-center text-xl">
            <i class="fa-solid fa-building-columns" />
          </div>
          <div class="flex-1">
            <h3 class="font-bold text-ink">{{ b.name }}</h3>
            <span :class="b.is_active ? 'badge-green' : 'badge-gray'">
              {{ b.is_active ? "Active" : "Inactive" }}
            </span>
          </div>
        </div>
        <p v-if="b.description" class="text-sm text-ink-medium mb-4 line-clamp-2">{{ b.description }}</p>
        <div class="flex justify-end gap-1">
          <button class="btn-ghost !p-2" @click="c.openEdit(b)">
            <i class="fa-solid fa-pen" />
          </button>
          <button class="btn-ghost !p-2 text-red-500" @click="c.askDelete(b)">
            <i class="fa-solid fa-trash" />
          </button>
        </div>
      </article>
    </div>

    <UiModal v-model="c.open.value" :title="c.editing.value.id ? $t('app.edit') : $t('app.create')" size="lg">
      <EntityForm v-model="c.editing.value" :fields="fields" />
      <template #footer>
        <button class="btn-outline" @click="c.open.value = false">{{ $t('app.cancel') }}</button>
        <button class="btn-primary" @click="c.save">{{ $t('app.save') }}</button>
      </template>
    </UiModal>

    <UiConfirmDialog v-model="c.confirmState.value.open" danger :message="$t('app.delete') + '?'" @confirm="c.doDelete" />
  </div>
</template>
