<script setup lang="ts">
import type { InstallmentObject } from "~/types";
import type { FormField } from "~/components/EntityForm.vue";

useHead({ title: () => useI18n().t("nav.installments") });
const { t } = useI18n();
const c = useCrud<InstallmentObject>("/api/installment-objects", "installments");

const fields = computed<FormField[]>(() => [
  { key: "name", label: "Name", required: true, cols: 2 },
  { key: "address", label: t("house.address"), cols: 2 },
  { key: "developer", label: t("house.developer") },
  { key: "order_index", label: "Order", type: "number" },
  { key: "description", label: t("post.description"), type: "textarea", cols: 2 },
  { key: "is_active", label: t("user.active"), type: "checkbox" },
]);
</script>

<template>
  <div class="card p-6 lg:p-8">
    <PageHeader :title="$t('nav.installments')">
      <button class="btn-primary" @click="c.openCreate({ is_active: true })">
        <i class="fa-solid fa-plus" /> {{ $t("app.create") }}
      </button>
    </PageHeader>

    <UiLoader v-if="c.pending.value" />
    <UiEmpty v-else-if="!c.rows.value?.length" icon="fa-handshake" />

    <div v-else class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
      <article v-for="i in c.rows.value" :key="i.id" class="card overflow-hidden hover:-translate-y-0.5 hover:shadow-hover transition-all">
        <div class="aspect-[4/3] bg-page relative">
          <img v-if="i.image" :src="i.image" class="w-full h-full object-cover" :alt="i.name" />
          <div v-else class="absolute inset-0 flex items-center justify-center text-ink-soft text-3xl">
            <i class="fa-solid fa-handshake" />
          </div>
        </div>
        <div class="p-5">
          <h3 class="font-bold text-ink mb-1">{{ i.name }}</h3>
          <p v-if="i.address" class="text-xs text-ink-soft mb-3">
            <i class="fa-solid fa-location-dot mr-1" />{{ i.address }}
          </p>
          <p v-if="i.description" class="text-sm text-ink-medium line-clamp-2 mb-3">{{ i.description }}</p>
          <div class="flex justify-end gap-1">
            <button class="btn-ghost !p-2" @click="c.openEdit(i)">
              <i class="fa-solid fa-pen" />
            </button>
            <button class="btn-ghost !p-2 text-red-500" @click="c.askDelete(i)">
              <i class="fa-solid fa-trash" />
            </button>
          </div>
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
