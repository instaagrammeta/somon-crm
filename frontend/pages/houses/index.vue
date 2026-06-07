<script setup lang="ts">
import type { House } from "~/types";
import type { FormField } from "~/components/EntityForm.vue";
import type { Column } from "~/components/CrudTable.vue";
import { formatNumber } from "~/utils/format";

useHead({ title: () => useI18n().t("nav.houses") });

const { t } = useI18n();
const c = useCrud<House>("/api/houses", "houses");
const api = useApi();

const fields = computed<FormField[]>(() => [
  { key: "title", label: t("house.title_field"), required: true, cols: 2 },
  { key: "construction_type", label: t("house.construction_type") },
  { key: "district", label: t("house.district") },
  { key: "address", label: t("house.address"), cols: 2 },
  { key: "area", label: t("house.area"), type: "number" },
  { key: "rooms", label: t("house.rooms"), type: "number" },
  { key: "windows", label: t("house.windows"), type: "number" },
  { key: "floor", label: t("house.floor"), type: "number" },
  { key: "total_floors", label: t("house.total_floors"), type: "number" },
  { key: "price_per_m2", label: t("house.price_per_m2"), type: "number" },
  { key: "total_price", label: t("house.total_price"), type: "number" },
  { key: "developer", label: t("house.developer") },
  { key: "contact_phone", label: t("house.contact_phone") },
]);

const cols: Column[] = [
  { key: "title", label: t("house.title_field") },
  { key: "district", label: t("house.district") },
  { key: "rooms", label: t("house.rooms") },
  { key: "area", label: t("house.area"), format: (r) => formatNumber(r.area, 1) },
  { key: "total_price", label: t("house.total_price"), format: (r) => formatNumber(r.total_price) },
  { key: "contact_phone", label: t("house.contact_phone") },
];

const exportXlsx = () => api.download("/api/houses/export", "houses.xlsx");
</script>

<template>
  <div class="card p-6 lg:p-8">
    <PageHeader :title="$t('nav.houses')">
      <button class="btn-outline" @click="exportXlsx">
        <i class="fa-solid fa-file-excel" /> Excel
      </button>
      <button class="btn-primary" @click="c.openCreate({})">
        <i class="fa-solid fa-plus" /> {{ $t("app.create") }}
      </button>
    </PageHeader>

    <UiLoader v-if="c.pending.value" />
    <CrudTable
      v-else
      :rows="c.rows.value || []"
      :columns="cols"
      empty-icon="fa-building"
      @edit="c.openEdit"
      @delete="c.askDelete"
    />

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
