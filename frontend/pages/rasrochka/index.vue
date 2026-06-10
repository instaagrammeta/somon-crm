<script setup lang="ts">
import type { InstallmentObject } from "~/types";

interface InstallmentCondition {
  id: number;
  object_id: number;
  title: string;
  currency: string;
  min_price: number;
  max_price: number;
  down_payment_percent: number;
  months: number;
  monthly_payment_rule: string;
  interest_percent: number;
  extra_conditions: string;
}

useHead({ title: () => useI18n().t("nav.installments") });

const { t } = useI18n();
const api = useApi();
const auth = useAuthStore();
const toast = useToast();

const { data: objects, pending, refresh } = useAsyncData<InstallmentObject[]>(
  "installments",
  () => api.get<InstallmentObject[]>("/api/installment-objects"),
  { lazy: true, default: () => [] }
);

const conditions = ref<Record<number, InstallmentCondition[]>>({});
const expanded = ref<number | null>(null);
const loadConditions = async (objId: number) => {
  conditions.value[objId] = await api.get<InstallmentCondition[]>(`/api/installment-objects/${objId}/conditions`);
};
const toggle = async (objId: number) => {
  if (expanded.value === objId) { expanded.value = null; return; }
  expanded.value = objId;
  if (!conditions.value[objId]) await loadConditions(objId);
};

// object modal
const objModal = ref(false);
const editing = ref<Partial<InstallmentObject>>({});
const imageFile = ref<File | null>(null);
const openNew = () => { editing.value = { is_active: true, order_index: 0 }; imageFile.value = null; objModal.value = true; };
const openEdit = (o: InstallmentObject) => { editing.value = { ...o }; imageFile.value = null; objModal.value = true; };
const saveObj = async () => {
  const e = editing.value;
  if (!e.name) { toast.error(t("notify.error")); return; }
  const fd = new FormData();
  fd.append("name", e.name || "");
  fd.append("description", e.description || "");
  fd.append("address", e.address || "");
  fd.append("developer", e.developer || "");
  fd.append("order_index", String(e.order_index || 0));
  fd.append("is_active", String(e.is_active ?? true));
  if (imageFile.value) fd.append("image", imageFile.value);
  if (e.id) { await api.upload(`/api/installment-objects/${e.id}`, fd); toast.success(t("notify.updated")); }
  else { await api.upload("/api/installment-objects", fd); toast.success(t("notify.created")); }
  objModal.value = false;
  await refresh();
};
const deleteObj = async (o: InstallmentObject) => {
  if (!confirm(t("app.delete") + "?")) return;
  await api.del(`/api/installment-objects/${o.id}`);
  toast.success(t("notify.deleted"));
  await refresh();
};

// condition modal
const condModal = ref(false);
const editingCond = ref<Partial<InstallmentCondition>>({});
const condObjId = ref(0);
const blankCond = (): Partial<InstallmentCondition> => ({ currency: "TJS", months: 12, down_payment_percent: 30, interest_percent: 0 });
const openNewCond = (objId: number) => { condObjId.value = objId; editingCond.value = blankCond(); condModal.value = true; };
const openEditCond = (objId: number, c: InstallmentCondition) => { condObjId.value = objId; editingCond.value = { ...c }; condModal.value = true; };
const saveCond = async () => {
  const e = editingCond.value;
  if (e.id) await api.put(`/api/installment-conditions/${e.id}`, e);
  else await api.post(`/api/installment-objects/${condObjId.value}/conditions`, e);
  condModal.value = false;
  toast.success(t("notify.saved"));
  await loadConditions(condObjId.value);
};
const deleteCond = async (objId: number, c: InstallmentCondition) => {
  if (!confirm(t("app.delete") + "?")) return;
  await api.del(`/api/installment-conditions/${c.id}`);
  await loadConditions(objId);
};
</script>

<template>
  <div>
    <div class="flex justify-between items-center flex-wrap gap-4 mb-6">
      <div>
        <h1 class="text-[28px] font-bold text-ink mb-1">{{ t("nav.installments") }}</h1>
        <p class="text-ink-soft text-sm">{{ t("installment.subtitle") }}</p>
      </div>
      <button v-if="auth.isAdmin" class="btn-primary" @click="openNew"><i class="fas fa-plus" /> {{ t("installment.add") }}</button>
    </div>

    <UiLoader v-if="pending" />
    <UiEmpty v-else-if="!objects?.length" icon="fa-store" />

    <div v-else class="grid grid-cols-1 lg:grid-cols-2 gap-4">
      <div v-for="o in objects" :key="o.id" class="bg-white border border-border-soft rounded-2xl overflow-hidden">
        <div class="flex gap-4 p-4">
          <div class="w-24 h-24 rounded-xl bg-page overflow-hidden shrink-0 flex items-center justify-center">
            <img v-if="o.image" :src="o.image" class="w-full h-full object-cover" />
            <i v-else class="fas fa-store text-brand text-2xl" />
          </div>
          <div class="flex-1 min-w-0">
            <div class="flex items-start justify-between gap-2">
              <h3 class="font-bold text-ink">{{ o.name }}</h3>
              <div v-if="auth.isAdmin" class="flex gap-1 shrink-0">
                <button class="w-7 h-7 bg-page rounded-lg text-brand hover:bg-brand-soft" @click="openEdit(o)"><i class="fas fa-pen text-xs" /></button>
                <button class="w-7 h-7 bg-page rounded-lg text-red-500 hover:bg-red-100" @click="deleteObj(o)"><i class="fas fa-trash text-xs" /></button>
              </div>
            </div>
            <p v-if="o.address" class="text-xs text-ink-soft mt-1"><i class="fas fa-map-marker-alt mr-1" />{{ o.address }}</p>
            <p v-if="o.developer" class="text-xs text-ink-soft"><i class="fas fa-industry mr-1" />{{ o.developer }}</p>
            <button class="text-xs text-brand font-semibold mt-2" @click="toggle(o.id)">
              <i class="fas" :class="expanded === o.id ? 'fa-chevron-up' : 'fa-chevron-down'" /> {{ t("installment.conditions") }}
            </button>
          </div>
        </div>

        <div v-if="expanded === o.id" class="border-t border-border-soft p-4 bg-page/40">
          <div class="flex justify-between items-center mb-3">
            <span class="font-semibold text-sm text-ink">{{ t("installment.conditions") }}</span>
            <button v-if="auth.isAdmin" class="btn-secondary !py-1.5 !px-3 text-xs" @click="openNewCond(o.id)"><i class="fas fa-plus" /> {{ t("app.add") }}</button>
          </div>
          <div v-if="!(conditions[o.id] || []).length" class="text-sm text-ink-soft py-3 text-center">{{ t("installment.no_conditions") }}</div>
          <div v-else class="space-y-2">
            <div v-for="c in conditions[o.id]" :key="c.id" class="bg-white rounded-xl p-3 border border-border-soft">
              <div class="flex items-center justify-between mb-1">
                <span class="font-semibold text-sm text-ink">{{ c.title || t("installment.condition") }}</span>
                <div v-if="auth.isAdmin" class="flex gap-2">
                  <button class="text-brand hover:opacity-70" @click="openEditCond(o.id, c)"><i class="fas fa-pen text-xs" /></button>
                  <button class="text-red-500 hover:opacity-70" @click="deleteCond(o.id, c)"><i class="fas fa-trash text-xs" /></button>
                </div>
              </div>
              <div class="flex flex-wrap gap-x-3 gap-y-0.5 text-xs text-ink-medium">
                <span class="px-2 py-0.5 rounded bg-brand-soft text-brand font-semibold">{{ c.currency }}</span>
                <span>{{ c.months }} {{ t("bank.months") }}</span>
                <span>{{ t("bank.down_payment_percent") }}: <b class="text-ink">{{ c.down_payment_percent }}%</b></span>
                <span v-if="c.interest_percent">%: <b class="text-ink">{{ c.interest_percent }}</b></span>
              </div>
              <p v-if="c.monthly_payment_rule" class="text-xs text-ink-soft mt-1">{{ c.monthly_payment_rule }}</p>
            </div>
          </div>
        </div>
      </div>
    </div>

    <!-- object modal -->
    <UiModal v-model="objModal" :title="editing.id ? t('app.edit') : t('installment.add')" size="md">
      <div class="space-y-4">
        <div><label class="label">{{ t("installment.name") }} *</label><input v-model="editing.name" class="input" /></div>
        <div><label class="label">{{ t("house.address") }}</label><input v-model="editing.address" class="input" /></div>
        <div><label class="label">{{ t("house.developer") }}</label><input v-model="editing.developer" class="input" /></div>
        <div><label class="label">{{ t("bank.description") }}</label><textarea v-model="editing.description" rows="2" class="input" /></div>
        <div><label class="label">{{ t("installment.image") }}</label>
          <input type="file" accept="image/*" class="input !py-2" @change="imageFile = ($event.target as HTMLInputElement).files?.[0] || null" />
        </div>
        <label class="flex items-center gap-2 cursor-pointer"><input v-model="editing.is_active" type="checkbox" class="w-4 h-4 accent-brand" /><span class="text-sm text-ink-medium">{{ t("user.active") }}</span></label>
      </div>
      <template #footer>
        <button class="btn-outline" @click="objModal = false">{{ t("app.cancel") }}</button>
        <button class="btn-primary" @click="saveObj">{{ t("app.save") }}</button>
      </template>
    </UiModal>

    <!-- condition modal -->
    <UiModal v-model="condModal" :title="editingCond.id ? t('app.edit') : t('bank.add_condition')" size="md">
      <div class="grid grid-cols-1 md:grid-cols-2 gap-4">
        <div class="md:col-span-2"><label class="label">{{ t("installment.cond_title") }}</label><input v-model="editingCond.title" class="input" /></div>
        <div><label class="label">{{ t("bank.currency") }} *</label><select v-model="editingCond.currency" class="input"><option value="TJS">TJS</option><option value="USD">USD</option></select></div>
        <div><label class="label">{{ t("installment.months") }} *</label><input v-model.number="editingCond.months" type="number" class="input" /></div>
        <div><label class="label">{{ t("bank.down_payment_percent") }} *</label><input v-model.number="editingCond.down_payment_percent" type="number" step="0.1" class="input" /></div>
        <div><label class="label">{{ t("installment.interest") }} %</label><input v-model.number="editingCond.interest_percent" type="number" class="input" /></div>
        <div><label class="label">{{ t("installment.min_price") }}</label><input v-model.number="editingCond.min_price" type="number" class="input" /></div>
        <div><label class="label">{{ t("installment.max_price") }}</label><input v-model.number="editingCond.max_price" type="number" class="input" /></div>
        <div class="md:col-span-2"><label class="label">{{ t("installment.rule") }}</label><textarea v-model="editingCond.monthly_payment_rule" rows="2" class="input" /></div>
        <div class="md:col-span-2"><label class="label">{{ t("bank.extra") }}</label><textarea v-model="editingCond.extra_conditions" rows="2" class="input" /></div>
      </div>
      <template #footer>
        <button class="btn-outline" @click="condModal = false">{{ t("app.cancel") }}</button>
        <button class="btn-primary" @click="saveCond">{{ t("app.save") }}</button>
      </template>
    </UiModal>
  </div>
</template>
