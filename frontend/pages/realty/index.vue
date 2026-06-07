<script setup lang="ts">
import type { RealtyObject } from "~/types";
import type { FormField } from "~/components/EntityForm.vue";
import type { Column } from "~/components/CrudTable.vue";
import { formatNumber } from "~/utils/format";

useHead({ title: () => useI18n().t("nav.realty") });
const { t } = useI18n();
const c = useCrud<RealtyObject>("/api/realty/objects", "realty");

const fields = computed<FormField[]>(() => [
  { key: "name", label: t("house.title_field"), required: true, cols: 2 },
  { key: "address", label: t("house.address"), cols: 2 },
  { key: "lat", label: "Lat", type: "number" },
  { key: "lng", label: "Lng", type: "number" },
  { key: "district", label: t("house.district") },
  { key: "construction_type", label: t("house.construction_type") },
  { key: "total_floors", label: t("house.total_floors"), type: "number" },
  { key: "price_per_m2", label: t("house.price_per_m2"), type: "number" },
  { key: "developer", label: t("house.developer") },
  { key: "description", label: t("post.description"), type: "textarea", cols: 2 },
]);

const cols: Column[] = [
  { key: "name", label: t("house.title_field") },
  { key: "district", label: t("house.district") },
  { key: "address", label: t("house.address") },
  { key: "developer", label: t("house.developer") },
  { key: "price_per_m2", label: t("house.price_per_m2"), format: r => formatNumber(r.price_per_m2) },
];
</script>

<template>
  <div class="card p-6 lg:p-8">
    <PageHeader :title="$t('nav.realty')">
      <button class="btn-primary" @click="c.openCreate({ lat: 38.5598, lng: 68.787 })">
        <i class="fa-solid fa-plus" /> {{ $t("app.create") }}
      </button>
    </PageHeader>

    <div class="bg-page rounded-2xl p-12 text-center mb-6">
      <i class="fa-solid fa-map-location-dot text-5xl text-brand mb-3" />
      <p class="text-ink-medium">
        Намоиш дар харита (OpenStreetMap embed) дар нусхаи навбатӣ илова мешавад.
      </p>
    </div>

    <UiLoader v-if="c.pending.value" />
    <CrudTable
      v-else
      :rows="c.rows.value || []"
      :columns="cols"
      empty-icon="fa-map-location-dot"
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
