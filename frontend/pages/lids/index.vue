<script setup lang="ts">
import type { KanbanBoard, KanbanColumn, KanbanLead } from "~/types";

useHead({ title: () => useI18n().t("nav.lids") });

const { t } = useI18n();
const api = useApi();
const toast = useToast();

const { data: boards, refresh: refreshBoards } = useAsyncData<KanbanBoard[]>(
  "kanban-boards",
  () => api.get<KanbanBoard[]>("/api/kanban/boards")
);
const activeBoardId = ref<number | null>(null);

watch(boards, (b) => {
  if (b?.length && !activeBoardId.value) activeBoardId.value = b[0].id;
}, { immediate: true });

const { data: columns, refresh: refreshCols } = useAsyncData<KanbanColumn[]>(
  () => `kanban-cols-${activeBoardId.value}`,
  () => activeBoardId.value
    ? api.get<KanbanColumn[]>(`/api/kanban/boards/${activeBoardId.value}/columns`)
    : Promise.resolve([]),
  { watch: [activeBoardId] }
);
const { data: leads, refresh: refreshLeads } = useAsyncData<KanbanLead[]>(
  () => `kanban-leads-${activeBoardId.value}`,
  () => activeBoardId.value
    ? api.get<KanbanLead[]>(`/api/kanban/leads?board_id=${activeBoardId.value}`)
    : Promise.resolve([]),
  { watch: [activeBoardId] }
);

const refreshAll = async () => { await Promise.all([refreshCols(), refreshLeads()]); };

// === Board ops ===
const newBoardOpen = ref(false);
const newBoard = ref({ title: "", color: "#0079bf", is_public: true });
const createBoard = async () => {
  await api.post("/api/kanban/boards", newBoard.value);
  newBoardOpen.value = false;
  toast.success(t("notify.created"));
  await refreshBoards();
};

// === Column ops ===
const colOpen = ref(false);
const editingCol = ref<Partial<KanbanColumn>>({});
const openNewCol = () => { editingCol.value = { board_id: activeBoardId.value!, color: "#0079bf" }; colOpen.value = true; };
const openEditCol = (c: KanbanColumn) => { editingCol.value = { ...c }; colOpen.value = true; };
const saveCol = async () => {
  if (editingCol.value.id) await api.put(`/api/kanban/columns/${editingCol.value.id}`, editingCol.value);
  else await api.post("/api/kanban/columns", editingCol.value);
  colOpen.value = false;
  await refreshCols();
};
const deleteCol = async (c: KanbanColumn) => {
  if (!confirm(t("app.delete") + "?")) return;
  await api.del(`/api/kanban/columns/${c.id}`);
  await refreshAll();
};

// === Lead ops ===
const leadOpen = ref(false);
const editingLead = ref<Partial<KanbanLead>>({});
const openNewLead = (columnId: number) => {
  editingLead.value = { board_id: activeBoardId.value!, column_id: columnId };
  leadOpen.value = true;
};
const openEditLead = (l: KanbanLead) => { editingLead.value = { ...l }; leadOpen.value = true; };
const saveLead = async () => {
  if (editingLead.value.id) await api.put(`/api/kanban/leads/${editingLead.value.id}`, editingLead.value);
  else await api.post("/api/kanban/leads", editingLead.value);
  leadOpen.value = false;
  await refreshLeads();
};
const deleteLead = async (l: KanbanLead) => {
  if (!confirm(t("app.delete") + "?")) return;
  await api.del(`/api/kanban/leads/${l.id}`);
  await refreshLeads();
};

// === Drag/drop move ===
const moveCard = async (e: { cardId: number; toColumnId: number; orderIndex: number }) => {
  await api.post("/api/kanban/leads/move", {
    lead_id: e.cardId,
    column_id: e.toColumnId,
    order_index: e.orderIndex,
  });
  await refreshLeads();
};

const exportXlsx = () =>
  api.download(`/api/kanban/boards/${activeBoardId.value}/export/excel`, `kanban_${activeBoardId.value}.xlsx`);
</script>

<template>
  <div class="card p-6 lg:p-8">
    <PageHeader :title="$t('nav.lids')">
      <select v-model="activeBoardId" class="input !py-2 max-w-[260px]">
        <option v-for="b in boards" :key="b.id" :value="b.id">{{ b.title }}</option>
      </select>
      <button v-if="activeBoardId" class="btn-outline" @click="exportXlsx">
        <i class="fa-solid fa-file-excel" /> Excel
      </button>
      <button class="btn-primary" @click="newBoardOpen = true">
        <i class="fa-solid fa-plus" /> {{ $t("kanban.create_board") }}
      </button>
    </PageHeader>

    <KanbanBoard
      v-if="columns?.length"
      :columns="columns || []"
      :cards="leads || []"
      @move-card="moveCard"
      @add-card="openNewLead"
      @edit-column="openEditCol"
      @delete-column="deleteCol"
      @add-column="openNewCol"
    >
      <template #card="{ card }">
        <div class="flex items-start justify-between gap-2 mb-1.5">
          <div class="font-semibold text-sm text-ink truncate">
            {{ (card as KanbanLead).client_name || "—" }}
          </div>
          <div class="flex items-center gap-0.5">
            <button class="btn-ghost !p-1 text-[10px]" @click.stop="openEditLead(card as KanbanLead)">
              <i class="fa-solid fa-pen" />
            </button>
            <button class="btn-ghost !p-1 text-[10px] text-red-500" @click.stop="deleteLead(card as KanbanLead)">
              <i class="fa-solid fa-trash" />
            </button>
          </div>
        </div>
        <div v-if="(card as KanbanLead).phone" class="text-xs text-ink-medium mb-1">
          <i class="fa-solid fa-phone text-brand text-[10px] mr-1" />
          {{ (card as KanbanLead).phone }}
        </div>
        <div v-if="(card as KanbanLead).topic" class="text-xs text-ink-soft line-clamp-2 mb-2">
          {{ (card as KanbanLead).topic }}
        </div>
        <div class="flex flex-wrap gap-1">
          <span v-if="(card as KanbanLead).mortgage" class="badge-green text-[10px]">Ипотека</span>
          <span v-if="(card as KanbanLead).box" class="badge-yellow text-[10px]">Каропка</span>
        </div>
      </template>
    </KanbanBoard>

    <UiEmpty v-else icon="fa-bullseye" :title="$t('kanban.no_columns')" />

    <!-- New board -->
    <UiModal v-model="newBoardOpen" :title="$t('kanban.create_board')">
      <div class="space-y-4">
        <div>
          <label class="label">{{ $t("kanban.boards") }}</label>
          <input v-model="newBoard.title" class="input" />
        </div>
        <div>
          <label class="label">Color</label>
          <input v-model="newBoard.color" type="color" class="input h-12" />
        </div>
      </div>
      <template #footer>
        <button class="btn-outline" @click="newBoardOpen = false">{{ $t('app.cancel') }}</button>
        <button class="btn-primary" @click="createBoard">{{ $t('app.save') }}</button>
      </template>
    </UiModal>

    <!-- Column dialog -->
    <UiModal v-model="colOpen" :title="editingCol.id ? $t('app.edit') : $t('kanban.create_column')">
      <div class="space-y-4">
        <div><label class="label">Title</label><input v-model="editingCol.title" class="input" /></div>
        <div><label class="label">Color</label><input v-model="editingCol.color" type="color" class="input h-12" /></div>
      </div>
      <template #footer>
        <button class="btn-outline" @click="colOpen = false">{{ $t('app.cancel') }}</button>
        <button class="btn-primary" @click="saveCol">{{ $t('app.save') }}</button>
      </template>
    </UiModal>

    <!-- Lead dialog -->
    <UiModal v-model="leadOpen" :title="editingLead.id ? $t('app.edit') : $t('kanban.create_lead')" size="lg">
      <EntityForm
        v-model="editingLead"
        :fields="[
          { key: 'client_name', label: $t('kanban.client_name'), required: true },
          { key: 'phone', label: $t('kanban.phone') },
          { key: 'topic', label: $t('kanban.topic'), cols: 2 },
          { key: 'source', label: $t('kanban.source') },
          { key: 'mortgage', label: $t('kanban.mortgage'), type: 'checkbox' },
          { key: 'box', label: $t('kanban.box'), type: 'checkbox' },
          { key: 'comment', label: $t('kanban.comment'), type: 'textarea', cols: 2 },
        ]"
      />
      <template #footer>
        <button class="btn-outline" @click="leadOpen = false">{{ $t('app.cancel') }}</button>
        <button class="btn-primary" @click="saveLead">{{ $t('app.save') }}</button>
      </template>
    </UiModal>
  </div>
</template>
