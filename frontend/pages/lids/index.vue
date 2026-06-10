<script setup lang="ts">
import type { KanbanBoard, KanbanColumn, KanbanLead } from "~/types";
import { formatDate } from "~/utils/format";

useHead({ title: () => useI18n().t("nav.sales") });

const { t } = useI18n();
const api = useApi();
const toast = useToast();

const sources = [
  { value: "Входящий", label: "📞 Входящий" },
  { value: "Исходящий", label: "📞 Исходящий" },
  { value: "WhatsApp", label: "💬 WhatsApp" },
  { value: "Telegram", label: "✈️ Telegram" },
  { value: "Somon.tj", label: "🏠 Somon.tj" },
  { value: "Instagram", label: "📷 Instagram" },
  { value: "TikTok", label: "🎵 TikTok" },
];

const { data: boards, refresh: refreshBoards } = useAsyncData<KanbanBoard[]>(
  "kanban-boards",
  () => api.get<KanbanBoard[]>("/api/kanban/boards"),
  { lazy: true, default: () => [] }
);
const activeId = ref<number | null>(null);
watch(boards, (b) => { if (b?.length && !activeId.value) activeId.value = b[0].id; }, { immediate: true });

const { data: cols, refresh: refreshCols } = useAsyncData<KanbanColumn[]>(
  () => `kb-cols-${activeId.value}`,
  () => activeId.value ? api.get<KanbanColumn[]>(`/api/kanban/boards/${activeId.value}/columns`) : Promise.resolve([]),
  { watch: [activeId], lazy: true, default: () => [] }
);
const { data: leads, refresh: refreshLeads } = useAsyncData<KanbanLead[]>(
  () => `kb-leads-${activeId.value}`,
  () => activeId.value ? api.get<KanbanLead[]>(`/api/kanban/leads?board_id=${activeId.value}`) : Promise.resolve([]),
  { watch: [activeId], lazy: true, default: () => [] }
);

// filters
const fl = reactive({ search: "", source: "all", mortgage: "all", box: "all" });
const filteredLeads = computed(() => {
  let list = leads.value || [];
  if (fl.search) {
    const q = fl.search.toLowerCase();
    list = list.filter((l) => [l.client_name, l.phone, l.topic].some((x) => (x || "").toLowerCase().includes(q)));
  }
  if (fl.source !== "all") list = list.filter((l) => l.source === fl.source);
  if (fl.mortgage !== "all") list = list.filter((l) => !!l.mortgage === (fl.mortgage === "yes"));
  if (fl.box !== "all") list = list.filter((l) => !!l.box === (fl.box === "yes"));
  return list;
});

// board
const boardModal = ref(false);
const newBoard = reactive({ title: "", color: "#0079bf" });
const createBoard = async () => {
  const r = await api.post<{ id: number }>("/api/kanban/boards", { title: newBoard.title, color: newBoard.color, is_public: true });
  boardModal.value = false; newBoard.title = "";
  toast.success(t("notify.created"));
  await refreshBoards();
  activeId.value = r.id;
};
const deleteBoard = async () => {
  if (!activeId.value || !confirm(t("app.delete") + "?")) return;
  await api.del(`/api/kanban/boards/${activeId.value}`);
  activeId.value = null;
  await refreshBoards();
};
const exportBoard = () => api.download(`/api/kanban/boards/${activeId.value}/export/excel`, `lids_${activeId.value}.xlsx`);

// column
const colModal = ref(false);
const editingCol = ref<Partial<KanbanColumn>>({});
const openNewCol = () => { editingCol.value = { board_id: activeId.value!, color: "#0079bf" }; colModal.value = true; };
const openEditCol = (c: KanbanColumn) => { editingCol.value = { ...c }; colModal.value = true; };
const saveCol = async () => {
  if (editingCol.value.id) await api.put(`/api/kanban/columns/${editingCol.value.id}`, editingCol.value);
  else await api.post("/api/kanban/columns", editingCol.value);
  colModal.value = false;
  await refreshCols();
};
const deleteCol = async (c: KanbanColumn) => {
  if (!confirm(t("app.delete") + "?")) return;
  await api.del(`/api/kanban/columns/${c.id}`);
  await Promise.all([refreshCols(), refreshLeads()]);
};

// lead
const leadModal = ref(false);
const blank = (): Partial<KanbanLead> => ({ source: "Входящий", mortgage: false, box: false });
const editing = ref<Partial<KanbanLead>>(blank());
const openNewLead = (columnId: number) => { editing.value = { ...blank(), board_id: activeId.value!, column_id: columnId }; leadModal.value = true; };
const openEditLead = (l: KanbanLead) => { editing.value = { ...l }; leadModal.value = true; };
const saveLead = async () => {
  const e = editing.value;
  if (!e.client_name || !e.phone) { toast.error(t("notify.error")); return; }
  if (e.id) { await api.put(`/api/kanban/leads/${e.id}`, e); toast.success(t("notify.updated")); }
  else { await api.post("/api/kanban/leads", e); toast.success(t("notify.created")); }
  leadModal.value = false;
  await refreshLeads();
};
const deleteLead = async (l: KanbanLead) => {
  if (!confirm(t("app.delete") + "?")) return;
  await api.del(`/api/kanban/leads/${l.id}`);
  await refreshLeads();
};
const moveCard = async (e: { cardId: number; toColumnId: number; orderIndex: number }) => {
  await api.post("/api/kanban/leads/move", { lead_id: e.cardId, column_id: e.toColumnId, order_index: e.orderIndex });
  await refreshLeads();
};
</script>

<template>
  <div>
    <!-- header -->
    <div class="flex items-center justify-between flex-wrap gap-3 mb-5">
      <div class="flex items-center gap-3">
        <div class="w-11 h-11 bg-brand-soft text-brand rounded-2xl flex items-center justify-center text-lg"><i class="fas fa-fire" /></div>
        <div>
          <select v-model="activeId" class="text-xl font-bold text-ink bg-transparent focus:outline-none cursor-pointer">
            <option v-for="b in boards" :key="b.id" :value="b.id">{{ b.title }}</option>
          </select>
          <div class="text-xs text-ink-soft">{{ t("app.total") }}: {{ filteredLeads.length }}</div>
        </div>
      </div>
      <div class="flex items-center gap-2 flex-wrap">
        <button class="btn-outline !py-2 !px-3" @click="openNewCol"><i class="fas fa-plus" /> {{ t("kanban.column") }}</button>
        <button class="btn-primary !py-2 !px-3" @click="activeId && openNewLead(cols?.[0]?.id || 0)"><i class="fas fa-user-plus" /> {{ t("kanban.create_lead") }}</button>
        <button class="btn-outline !py-2 !px-3" @click="exportBoard"><i class="fas fa-file-excel" /></button>
        <button class="btn-outline !py-2 !px-3" @click="boardModal = true"><i class="fas fa-plus" /> {{ t("kanban.create_board") }}</button>
        <button v-if="activeId" class="btn-danger !py-2 !px-3" @click="deleteBoard"><i class="fas fa-trash" /></button>
      </div>
    </div>

    <!-- filters -->
    <div class="flex flex-wrap gap-2.5 mb-5">
      <div class="flex items-center gap-1.5 bg-page rounded-xl px-3 py-1.5 text-xs flex-1 min-w-[160px]">
        <i class="fas fa-search text-ink-soft" />
        <input v-model="fl.search" type="text" :placeholder="t('app.search')" class="flex-1 bg-transparent focus:outline-none" />
      </div>
      <select v-model="fl.source" class="bg-page rounded-xl px-3 py-1.5 text-xs focus:outline-none">
        <option value="all">{{ t("kanban.source") }}: {{ t("app.all") }}</option>
        <option v-for="s in sources" :key="s.value" :value="s.value">{{ s.label }}</option>
      </select>
      <select v-model="fl.mortgage" class="bg-page rounded-xl px-3 py-1.5 text-xs focus:outline-none">
        <option value="all">🏦 {{ t("kanban.mortgage") }}</option>
        <option value="yes">{{ t("app.yes") }}</option>
        <option value="no">{{ t("app.no") }}</option>
      </select>
      <select v-model="fl.box" class="bg-page rounded-xl px-3 py-1.5 text-xs focus:outline-none">
        <option value="all">📦 {{ t("kanban.box") }}</option>
        <option value="yes">{{ t("app.yes") }}</option>
        <option value="no">{{ t("app.no") }}</option>
      </select>
    </div>

    <!-- kanban -->
    <KanbanBoard
      v-if="cols?.length"
      :columns="cols"
      :cards="filteredLeads"
      @move-card="moveCard"
      @add-card="openNewLead"
      @edit-column="openEditCol"
      @delete-column="deleteCol"
      @add-column="openNewCol"
    >
      <template #card="{ card }">
        <div class="flex items-start justify-between gap-2 mb-1.5">
          <span class="font-semibold text-sm text-ink truncate">{{ (card as KanbanLead).client_name }}</span>
          <div class="flex gap-0.5">
            <button class="text-ink-soft hover:text-brand text-[11px] p-1" @click.stop="openEditLead(card as KanbanLead)"><i class="fas fa-pen" /></button>
            <button class="text-ink-soft hover:text-red-500 text-[11px] p-1" @click.stop="deleteLead(card as KanbanLead)"><i class="fas fa-trash" /></button>
          </div>
        </div>
        <div class="text-xs text-ink-medium mb-1.5"><i class="fas fa-phone-alt text-brand text-[10px] mr-1" />{{ (card as KanbanLead).phone }}</div>
        <div v-if="(card as KanbanLead).source" class="mb-1.5">
          <span class="text-[10px] bg-brand-soft text-brand px-2 py-0.5 rounded-md font-medium">{{ (card as KanbanLead).source }}</span>
        </div>
        <div v-if="(card as KanbanLead).topic" class="text-xs text-ink-soft mb-2"><i class="fas fa-comment text-[10px] mr-1" />{{ (card as KanbanLead).topic }}</div>
        <div class="flex items-center justify-between text-[10px] text-ink-soft pt-1.5 border-t border-border-soft">
          <span><i class="far fa-calendar-alt mr-1" />{{ formatDate((card as KanbanLead).created_at) }}</span>
          <div class="flex gap-1">
            <span v-if="(card as KanbanLead).mortgage" class="bg-page px-1.5 py-0.5 rounded">🏦</span>
            <span v-if="(card as KanbanLead).box" class="bg-page px-1.5 py-0.5 rounded">📦</span>
          </div>
        </div>
      </template>
    </KanbanBoard>
    <UiEmpty v-else icon="fa-fire" />

    <!-- board modal -->
    <UiModal v-model="boardModal" :title="t('kanban.create_board')" size="sm">
      <div class="space-y-4">
        <div><label class="label">{{ t("kanban.boards") }}</label><input v-model="newBoard.title" class="input" /></div>
        <div><label class="label">Color</label><input v-model="newBoard.color" type="color" class="input h-12" /></div>
      </div>
      <template #footer>
        <button class="btn-outline" @click="boardModal = false">{{ t("app.cancel") }}</button>
        <button class="btn-primary" @click="createBoard">{{ t("app.save") }}</button>
      </template>
    </UiModal>

    <!-- column modal -->
    <UiModal v-model="colModal" :title="editingCol.id ? t('app.edit') : t('kanban.create_column')" size="sm">
      <div class="space-y-4">
        <div><label class="label">{{ t("kanban.column") }}</label><input v-model="editingCol.title" class="input" /></div>
        <div><label class="label">Color</label><input v-model="editingCol.color" type="color" class="input h-12" /></div>
      </div>
      <template #footer>
        <button class="btn-outline" @click="colModal = false">{{ t("app.cancel") }}</button>
        <button class="btn-primary" @click="saveCol">{{ t("app.save") }}</button>
      </template>
    </UiModal>

    <!-- lead modal -->
    <UiModal v-model="leadModal" :title="editing.id ? t('app.edit') : t('kanban.create_lead')" size="md">
      <div class="grid grid-cols-1 md:grid-cols-2 gap-4">
        <div><label class="label">{{ t("kanban.client_name") }} *</label><input v-model="editing.client_name" class="input" /></div>
        <div><label class="label">{{ t("kanban.phone") }} *</label><input v-model="editing.phone" class="input" /></div>
        <div>
          <label class="label">{{ t("kanban.source") }}</label>
          <select v-model="editing.source" class="input">
            <option v-for="s in sources" :key="s.value" :value="s.value">{{ s.label }}</option>
          </select>
        </div>
        <div><label class="label">{{ t("kanban.topic") }}</label><input v-model="editing.topic" class="input" /></div>
        <div>
          <label class="label">{{ t("kanban.mortgage") }}</label>
          <select v-model="editing.mortgage" class="input">
            <option :value="false">{{ t("app.no") }}</option>
            <option :value="true">{{ t("app.yes") }}</option>
          </select>
        </div>
        <div>
          <label class="label">{{ t("kanban.box") }}</label>
          <select v-model="editing.box" class="input">
            <option :value="false">{{ t("app.no") }}</option>
            <option :value="true">{{ t("app.yes") }}</option>
          </select>
        </div>
        <div class="md:col-span-2"><label class="label">{{ t("kanban.comment") }}</label><textarea v-model="editing.comment" rows="3" class="input" /></div>
      </div>
      <template #footer>
        <button class="btn-outline" @click="leadModal = false">{{ t("app.cancel") }}</button>
        <button class="btn-primary" @click="saveLead">{{ t("app.save") }}</button>
      </template>
    </UiModal>
  </div>
</template>
