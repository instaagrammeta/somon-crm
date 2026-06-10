<script setup lang="ts">
import type { Bank, MortgageCondition } from "~/types";

useHead({ title: () => useI18n().t("nav.mortgage") });

const { t } = useI18n();
const api = useApi();
const auth = useAuthStore();
const toast = useToast();

const { data: banks, pending, refresh } = useAsyncData<Bank[]>("banks", () => api.get<Bank[]>("/api/banks"), { lazy: true, default: () => [] });

// conditions cache per bank
const conditions = ref<Record<number, MortgageCondition[]>>({});
const expanded = ref<number | null>(null);

const loadConditions = async (bankId: number) => {
  conditions.value[bankId] = await api.get<MortgageCondition[]>(`/api/banks/${bankId}/conditions`);
};
const toggle = async (bankId: number) => {
  if (expanded.value === bankId) { expanded.value = null; return; }
  expanded.value = bankId;
  if (!conditions.value[bankId]) await loadConditions(bankId);
};

// bank modal
const bankModal = ref(false);
const editingBank = ref<Partial<Bank>>({});
const logoFile = ref<File | null>(null);
const openNewBank = () => { editingBank.value = { is_active: true, order_index: 0 }; logoFile.value = null; bankModal.value = true; };
const openEditBank = (b: Bank) => { editingBank.value = { ...b }; logoFile.value = null; bankModal.value = true; };
const saveBank = async () => {
  const e = editingBank.value;
  if (!e.name) { toast.error(t("notify.error")); return; }
  const fd = new FormData();
  fd.append("name", e.name || "");
  fd.append("description", e.description || "");
  fd.append("order_index", String(e.order_index || 0));
  fd.append("is_active", String(e.is_active ?? true));
  if (logoFile.value) fd.append("logo", logoFile.value);
  if (e.id) { await api.upload(`/api/banks/${e.id}`, fd); toast.success(t("notify.updated")); }
  else { await api.upload("/api/banks", fd); toast.success(t("notify.created")); }
  bankModal.value = false;
  await refresh();
};
const deleteBank = async (b: Bank) => {
  if (!confirm(t("app.delete") + "?")) return;
  await api.del(`/api/banks/${b.id}`);
  toast.success(t("notify.deleted"));
  await refresh();
};

// condition modal
const condModal = ref(false);
const editingCond = ref<Partial<MortgageCondition>>({});
const condBankId = ref<number>(0);
const blankCond = (): Partial<MortgageCondition> => ({ currency: "TJS", interest_yearly: 0, interest_monthly: 0, min_months: 12, max_months: 240, down_payment_percent: 30, guarantor_required: false, collateral_required: false });
const openNewCond = (bankId: number) => { condBankId.value = bankId; editingCond.value = blankCond(); condModal.value = true; };
const openEditCond = (bankId: number, c: MortgageCondition) => { condBankId.value = bankId; editingCond.value = { ...c }; condModal.value = true; };
const saveCond = async () => {
  const e = editingCond.value;
  if (e.id) await api.put(`/api/conditions/${e.id}`, e);
  else await api.post(`/api/banks/${condBankId.value}/conditions`, e);
  condModal.value = false;
  toast.success(t("notify.saved"));
  await loadConditions(condBankId.value);
};
const deleteCond = async (bankId: number, c: MortgageCondition) => {
  if (!confirm(t("app.delete") + "?")) return;
  await api.del(`/api/conditions/${c.id}`);
  await loadConditions(bankId);
};
</script>

<template>
  <div>
    <div class="flex justify-between items-center flex-wrap gap-4 mb-6">
      <div>
        <h1 class="text-[28px] font-bold text-ink mb-1">{{ t("nav.mortgage") }}</h1>
        <p class="text-ink-soft text-sm">{{ t("bank.subtitle") }}</p>
      </div>
      <button v-if="auth.isAdmin" class="btn-primary" @click="openNewBank"><i class="fas fa-plus" /> {{ t("bank.add") }}</button>
    </div>

    <UiLoader v-if="pending" />
    <UiEmpty v-else-if="!banks?.length" icon="fa-building-columns" />

    <div v-else class="space-y-3">
      <div v-for="b in banks" :key="b.id" class="bg-white border border-border-soft rounded-2xl overflow-hidden">
        <!-- bank row -->
        <div class="flex items-center gap-4 p-4 cursor-pointer hover:bg-page/50" @click="toggle(b.id)">
          <div class="w-14 h-14 rounded-2xl bg-page flex items-center justify-center overflow-hidden shrink-0">
            <img v-if="b.logo" :src="b.logo" class="w-full h-full object-contain" />
            <i v-else class="fas fa-building-columns text-brand text-xl" />
          </div>
          <div class="flex-1 min-w-0">
            <div class="font-bold text-ink">{{ b.name }}</div>
            <div class="text-xs text-ink-soft truncate">{{ b.description }}</div>
          </div>
          <span class="text-xs text-ink-soft hidden sm:flex items-center gap-1.5">
            <i class="fas fa-list-check" /> {{ (conditions[b.id] || []).length }} {{ t("bank.conditions") }}
          </span>
          <div v-if="auth.isAdmin" class="flex gap-1" @click.stop>
            <button class="w-8 h-8 bg-page rounded-lg text-brand hover:bg-brand-soft" @click="openEditBank(b)"><i class="fas fa-pen text-xs" /></button>
            <button class="w-8 h-8 bg-page rounded-lg text-red-500 hover:bg-red-100" @click="deleteBank(b)"><i class="fas fa-trash text-xs" /></button>
          </div>
          <i class="fas fa-chevron-down text-ink-soft transition-transform" :class="{ 'rotate-180': expanded === b.id }" />
        </div>

        <!-- conditions -->
        <div v-if="expanded === b.id" class="border-t border-border-soft p-4 bg-page/40">
          <div class="flex justify-between items-center mb-3">
            <span class="font-semibold text-sm text-ink">{{ t("bank.conditions") }}</span>
            <button v-if="auth.isAdmin" class="btn-secondary !py-1.5 !px-3 text-xs" @click="openNewCond(b.id)"><i class="fas fa-plus" /> {{ t("app.add") }}</button>
          </div>
          <div v-if="!(conditions[b.id] || []).length" class="text-sm text-ink-soft py-4 text-center">{{ t("bank.no_conditions") }}</div>
          <div v-else class="grid grid-cols-1 md:grid-cols-2 gap-2">
            <div v-for="c in conditions[b.id]" :key="c.id" class="flex items-center gap-3 bg-white rounded-xl p-3 border border-border-soft">
              <span class="px-2 py-1 rounded-lg text-xs font-bold" :class="c.currency === 'USD' ? 'bg-blue-50 text-blue-600' : 'bg-brand-soft text-brand'">{{ c.currency }}</span>
              <div class="flex-1 text-xs text-ink-medium flex flex-wrap gap-x-3 gap-y-0.5">
                <span><b class="text-ink">{{ c.interest_yearly }}%</b> {{ t("bank.yearly") }}</span>
                <span><b class="text-ink">{{ c.interest_monthly }}%</b> {{ t("bank.monthly") }}</span>
                <span>{{ t("bank.down_payment_percent") }}: <b class="text-ink">{{ c.down_payment_percent }}%</b></span>
                <span>{{ c.min_months }}–{{ c.max_months }} {{ t("bank.months") }}</span>
              </div>
              <div v-if="auth.isAdmin" class="flex gap-1">
                <button class="text-brand hover:opacity-70" @click="openEditCond(b.id, c)"><i class="fas fa-pen text-xs" /></button>
                <button class="text-red-500 hover:opacity-70" @click="deleteCond(b.id, c)"><i class="fas fa-trash text-xs" /></button>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>

    <!-- bank modal -->
    <UiModal v-model="bankModal" :title="editingBank.id ? t('app.edit') : t('bank.add')" size="md">
      <div class="space-y-4">
        <div><label class="label">{{ t("bank.name") }} *</label><input v-model="editingBank.name" class="input" placeholder="Алиф Бонк" /></div>
        <div><label class="label">{{ t("bank.description") }}</label><textarea v-model="editingBank.description" rows="2" class="input" /></div>
        <div><label class="label">{{ t("bank.logo") }}</label>
          <input type="file" accept="image/*" class="input !py-2" @change="logoFile = ($event.target as HTMLInputElement).files?.[0] || null" />
        </div>
        <label class="flex items-center gap-2 cursor-pointer">
          <input v-model="editingBank.is_active" type="checkbox" class="w-4 h-4 accent-brand" />
          <span class="text-sm text-ink-medium">{{ t("user.active") }}</span>
        </label>
      </div>
      <template #footer>
        <button class="btn-outline" @click="bankModal = false">{{ t("app.cancel") }}</button>
        <button class="btn-primary" @click="saveBank">{{ t("app.save") }}</button>
      </template>
    </UiModal>

    <!-- condition modal -->
    <UiModal v-model="condModal" :title="editingCond.id ? t('app.edit') : t('bank.add_condition')" size="md">
      <div class="grid grid-cols-1 md:grid-cols-2 gap-4">
        <div class="md:col-span-2">
          <label class="label">{{ t("bank.currency") }} *</label>
          <select v-model="editingCond.currency" class="input"><option value="TJS">Сомонӣ (TJS)</option><option value="USD">Доллар (USD)</option></select>
        </div>
        <div><label class="label">{{ t("bank.yearly") }} % *</label><input v-model.number="editingCond.interest_yearly" type="number" step="0.01" class="input" /></div>
        <div><label class="label">{{ t("bank.monthly") }} % *</label><input v-model.number="editingCond.interest_monthly" type="number" step="0.01" class="input" /></div>
        <div><label class="label">{{ t("bank.min_amount") }}</label><input v-model.number="editingCond.min_amount" type="number" class="input" /></div>
        <div><label class="label">{{ t("bank.max_amount") }}</label><input v-model.number="editingCond.max_amount" type="number" class="input" /></div>
        <div><label class="label">{{ t("bank.min_months") }}</label><input v-model.number="editingCond.min_months" type="number" class="input" /></div>
        <div><label class="label">{{ t("bank.max_months") }}</label><input v-model.number="editingCond.max_months" type="number" class="input" /></div>
        <div class="md:col-span-2"><label class="label">{{ t("bank.down_payment_percent") }} *</label><input v-model.number="editingCond.down_payment_percent" type="number" step="0.1" class="input" /></div>
        <label class="flex items-center gap-2 cursor-pointer"><input v-model="editingCond.guarantor_required" type="checkbox" class="w-4 h-4 accent-brand" /><span class="text-sm text-ink-medium">{{ t("bank.guarantor") }}</span></label>
        <label class="flex items-center gap-2 cursor-pointer"><input v-model="editingCond.collateral_required" type="checkbox" class="w-4 h-4 accent-brand" /><span class="text-sm text-ink-medium">{{ t("bank.collateral") }}</span></label>
        <div class="md:col-span-2"><label class="label">{{ t("bank.extra") }}</label><textarea v-model="editingCond.extra_conditions" rows="2" class="input" /></div>
      </div>
      <template #footer>
        <button class="btn-outline" @click="condModal = false">{{ t("app.cancel") }}</button>
        <button class="btn-primary" @click="saveCond">{{ t("app.save") }}</button>
      </template>
    </UiModal>
  </div>
</template>
