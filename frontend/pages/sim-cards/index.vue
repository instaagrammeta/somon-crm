<script setup lang="ts">
import type { SimCard, CompanyPhone, User } from "~/types";
import type { FormField } from "~/components/EntityForm.vue";
import type { Column } from "~/components/CrudTable.vue";

useHead({ title: () => useI18n().t("nav.sim_cards") });

const { t } = useI18n();
const api = useApi();

const c = useCrud<SimCard>("/api/sim-cards", "sim-cards");
const ph = useCrud<CompanyPhone>("/api/company-phones", "phones");
const { data: users } = useAsyncData<User[]>("users-list", () => api.get<User[]>("/api/users/list"), { lazy: true, default: () => [] });

const tab = ref<"sims" | "phones">("sims");

const simFields = computed<FormField[]>(() => [
  { key: "phone_number", label: t("sim.phone_number"), required: true },
  { key: "operator", label: t("sim.operator"), required: true },
  { key: "assigned_to", label: t("sim.assigned_to"), type: "select",
    options: (users.value || []).map((u) => ({ value: u.id, label: u.full_name })) },
  { key: "phone_id", label: t("sim.phone_model"), type: "select",
    options: (ph.rows.value || []).map((p) => ({ value: p.id, label: p.model })) },
  { key: "status", label: t("sim.tariff"), type: "select",
    options: [{ value: "active", label: t("user.active") }, { value: "blocked", label: "Blocked" }] },
  { key: "description", label: t("post.description"), type: "textarea", cols: 2 },
]);

const phoneFields = computed<FormField[]>(() => [
  { key: "model", label: t("sim.phone_model"), required: true },
  { key: "phone_id", label: "ID", required: true },
  { key: "assigned_to", label: t("sim.assigned_to"), type: "select",
    options: (users.value || []).map((u) => ({ value: u.id, label: u.full_name })) },
  { key: "status", label: "Status", type: "select",
    options: [{ value: "free", label: "Free" }, { value: "in_use", label: "In use" }] },
  { key: "description", label: t("post.description"), type: "textarea", cols: 2 },
]);

const simCols: Column[] = [
  { key: "phone_number", label: t("sim.phone_number") },
  { key: "operator", label: t("sim.operator") },
  { key: "assigned_name", label: t("sim.assigned_to") },
  { key: "phone_model", label: t("sim.phone_model") },
  { key: "status", label: "Status",
    badge: (r) => ({ text: r.status, cls: r.status === "active" ? "badge-green" : "badge-red" }) },
];
const phoneCols: Column[] = [
  { key: "model", label: t("sim.phone_model") },
  { key: "phone_id", label: "ID" },
  { key: "assigned_name", label: t("sim.assigned_to") },
  { key: "status", label: "Status",
    badge: (r) => ({ text: r.status, cls: r.status === "free" ? "badge-blue" : "badge-yellow" }) },
];

const exportXlsx = () => api.download("/api/sim-phones/export/excel", "sim-cards.xlsx");
const sendNotifications = async () => {
  try {
    const res = await api.post<{ sent: number }>("/api/sim-tariffs/notify-expiring");
    useToast().success(`Огоҳсозиҳо равон шуданд: ${res.sent}`);
  } catch (e: any) {
    useToast().error(e?.data?.error || t("notify.error"));
  }
};
</script>

<template>
  <div class="card p-6 lg:p-8">
    <PageHeader :title="$t('nav.sim_cards')">
      <button class="btn-outline" @click="exportXlsx">
        <i class="fa-solid fa-file-excel" /> Excel
      </button>
      <button class="btn-outline" @click="sendNotifications">
        <i class="fa-brands fa-telegram" /> Telegram
      </button>
      <button v-if="tab === 'sims'" class="btn-primary" @click="c.openCreate({ status: 'active' })">
        <i class="fa-solid fa-plus" /> SIM
      </button>
      <button v-else class="btn-primary" @click="ph.openCreate({ status: 'free' })">
        <i class="fa-solid fa-plus" /> Phone
      </button>
    </PageHeader>

    <div class="flex gap-2 mb-6">
      <button :class="tab === 'sims' ? 'btn-primary' : 'btn-outline'" @click="tab = 'sims'">
        <i class="fa-solid fa-sim-card" /> SIM-кортҳо
      </button>
      <button :class="tab === 'phones' ? 'btn-primary' : 'btn-outline'" @click="tab = 'phones'">
        <i class="fa-solid fa-mobile" /> Телефонҳо
      </button>
    </div>

    <UiLoader v-if="c.pending.value && tab === 'sims'" />
    <UiLoader v-else-if="ph.pending.value && tab === 'phones'" />
    <CrudTable
      v-else-if="tab === 'sims'"
      :rows="c.rows.value || []"
      :columns="simCols"
      @edit="c.openEdit"
      @delete="c.askDelete"
    />
    <CrudTable
      v-else
      :rows="ph.rows.value || []"
      :columns="phoneCols"
      @edit="ph.openEdit"
      @delete="ph.askDelete"
    />

    <UiModal v-model="c.open.value" :title="c.editing.value.id ? $t('app.edit') : 'SIM'" size="lg">
      <EntityForm v-model="c.editing.value" :fields="simFields" />
      <template #footer>
        <button class="btn-outline" @click="c.open.value = false">{{ $t('app.cancel') }}</button>
        <button class="btn-primary" @click="c.save">{{ $t('app.save') }}</button>
      </template>
    </UiModal>

    <UiModal v-model="ph.open.value" :title="ph.editing.value.id ? $t('app.edit') : 'Phone'" size="lg">
      <EntityForm v-model="ph.editing.value" :fields="phoneFields" />
      <template #footer>
        <button class="btn-outline" @click="ph.open.value = false">{{ $t('app.cancel') }}</button>
        <button class="btn-primary" @click="ph.save">{{ $t('app.save') }}</button>
      </template>
    </UiModal>

    <UiConfirmDialog v-model="c.confirmState.value.open" danger :message="$t('app.delete') + '?'" @confirm="c.doDelete" />
    <UiConfirmDialog v-model="ph.confirmState.value.open" danger :message="$t('app.delete') + '?'" @confirm="ph.doDelete" />
  </div>
</template>
