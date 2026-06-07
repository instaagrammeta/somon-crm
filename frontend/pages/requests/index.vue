<script setup lang="ts">
import type { RequestsBoard, RequestsColumn, RequestItem, User } from "~/types";

useHead({ title: () => useI18n().t("nav.requests") });

const { t } = useI18n();
const api = useApi();
const toast = useToast();

const { data: boards, refresh: refreshBoards } = useAsyncData<RequestsBoard[]>(
  "req-boards",
  () => api.get<RequestsBoard[]>("/api/requests-boards")
);
const { data: users } = useAsyncData<User[]>("users-list", () => api.get<User[]>("/api/users/list"));
const activeId = ref<number | null>(null);
watch(boards, (b) => { if (b?.length && !activeId.value) activeId.value = b[0].id; }, { immediate: true });

const { data: cols, refresh: refreshCols } = useAsyncData<RequestsColumn[]>(
  () => `req-cols-${activeId.value}`,
  () => activeId.value ? api.get<RequestsColumn[]>(`/api/requests-boards/${activeId.value}/columns`) : Promise.resolve([]),
  { watch: [activeId] }
);
const { data: items, refresh: refreshItems } = useAsyncData<RequestItem[]>(
  () => `req-items-${activeId.value}`,
  () => activeId.value ? api.get<RequestItem[]>(`/api/requests-board/${activeId.value}/requests`) : Promise.resolve([]),
  { watch: [activeId] }
);

const newBoardOpen = ref(false);
const newBoard = ref({ title: "", color: "#0f172a", is_public: true });
const createBoard = async () => {
  await api.post("/api/requests-boards", newBoard.value);
  newBoardOpen.value = false;
  toast.success(t("notify.created"));
  await refreshBoards();
};

const colOpen = ref(false);
const editingCol = ref<Partial<RequestsColumn>>({});
const openNewCol = () => { editingCol.value = { board_id: activeId.value!, color: "#3b82f6" }; colOpen.value = true; };
const openEditCol = (c: RequestsColumn) => { editingCol.value = { ...c }; colOpen.value = true; };
const saveCol = async () => {
  if (editingCol.value.id) await api.put(`/api/requests-columns/${editingCol.value.id}`, editingCol.value);
  else await api.post("/api/requests-columns", editingCol.value);
  colOpen.value = false;
  await refreshCols();
};
const deleteCol = async (c: RequestsColumn) => {
  if (!confirm(t("app.delete") + "?")) return;
  await api.del(`/api/requests-columns/${c.id}`);
  await Promise.all([refreshCols(), refreshItems()]);
};

const itemOpen = ref(false);
const editingItem = ref<Partial<RequestItem>>({});
const openNewItem = (columnId: number) => {
  editingItem.value = { board_id: activeId.value!, column_id: columnId };
  itemOpen.value = true;
};
const openEditItem = (i: RequestItem) => { editingItem.value = { ...i }; itemOpen.value = true; };
const saveItem = async () => {
  if (editingItem.value.id) await api.put(`/api/requests-update/${editingItem.value.id}`, editingItem.value);
  else await api.post("/api/requests-new", editingItem.value);
  itemOpen.value = false;
  await refreshItems();
};
const deleteItem = async (i: RequestItem) => {
  if (!confirm(t("app.delete") + "?")) return;
  await api.del(`/api/requests-delete/${i.id}`);
  await refreshItems();
};

const moveCard = async (e: { cardId: number; toColumnId: number; orderIndex: number }) => {
  await api.post("/api/requests-move", { request_id: e.cardId, column_id: e.toColumnId, order_index: e.orderIndex });
  await refreshItems();
};
const exportXlsx = () =>
  api.download(`/api/requests-export/excel?board_id=${activeId.value}`, `requests_${activeId.value}.xlsx`);
</script>

<template>
  <div class="card p-6 lg:p-8">
    <PageHeader :title="$t('nav.requests')">
      <select v-model="activeId" class="input !py-2 max-w-[260px]">
        <option v-for="b in boards" :key="b.id" :value="b.id">{{ b.title }}</option>
      </select>
      <button v-if="activeId" class="btn-outline" @click="exportXlsx">
        <i class="fa-solid fa-file-excel" /> Excel
      </button>
      <button class="btn-primary" @click="newBoardOpen = true">
        <i class="fa-solid fa-plus" /> {{ $t("kanban.create_board") }}
      </button>
    </PageHeader>

    <KanbanBoard
      v-if="cols?.length"
      :columns="cols || []"
      :cards="items || []"
      @move-card="moveCard"
      @add-card="openNewItem"
      @edit-column="openEditCol"
      @delete-column="deleteCol"
      @add-column="openNewCol"
    >
      <template #card="{ card }">
        <div class="flex justify-between mb-1.5">
          <div class="font-semibold text-sm text-ink">
            {{ (card as RequestItem).client_name || "—" }}
          </div>
          <div class="flex items-center gap-0.5">
            <button class="btn-ghost !p-1 text-[10px]" @click.stop="openEditItem(card as RequestItem)">
              <i class="fa-solid fa-pen" />
            </button>
            <button class="btn-ghost !p-1 text-[10px] text-red-500" @click.stop="deleteItem(card as RequestItem)">
              <i class="fa-solid fa-trash" />
            </button>
          </div>
        </div>
        <div v-if="(card as RequestItem).phone" class="text-xs text-ink-medium">
          <i class="fa-solid fa-phone text-brand text-[10px] mr-1" />{{ (card as RequestItem).phone }}
        </div>
        <div v-if="(card as RequestItem).address" class="text-xs text-ink-soft line-clamp-2 mt-1">
          <i class="fa-solid fa-location-dot text-[10px] mr-1" />{{ (card as RequestItem).address }}
        </div>
        <div v-if="(card as RequestItem).total_price" class="text-xs font-bold text-brand mt-2">
          {{ Number((card as RequestItem).total_price).toLocaleString("ru-RU") }}
        </div>
      </template>
    </KanbanBoard>

    <UiEmpty v-else icon="fa-clipboard-list" />

    <UiModal v-model="newBoardOpen" :title="$t('kanban.create_board')">
      <div class="space-y-4">
        <div><label class="label">Title</label><input v-model="newBoard.title" class="input" /></div>
        <div><label class="label">Color</label><input v-model="newBoard.color" type="color" class="input h-12" /></div>
      </div>
      <template #footer>
        <button class="btn-outline" @click="newBoardOpen = false">{{ $t('app.cancel') }}</button>
        <button class="btn-primary" @click="createBoard">{{ $t('app.save') }}</button>
      </template>
    </UiModal>

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

    <UiModal v-model="itemOpen" :title="editingItem.id ? $t('app.edit') : $t('app.create')" size="lg">
      <EntityForm
        v-model="editingItem"
        :fields="[
          { key: 'client_name', label: $t('kanban.client_name'), required: true },
          { key: 'phone', label: $t('kanban.phone') },
          { key: 'address', label: $t('house.address'), cols: 2 },
          { key: 'property_type', label: $t('house.construction_type') },
          { key: 'rooms', label: $t('house.rooms'), type: 'number' },
          { key: 'area', label: $t('house.area'), type: 'number' },
          { key: 'floor', label: $t('house.floor'), type: 'number' },
          { key: 'total_price', label: $t('house.total_price'), type: 'number' },
          { key: 'price_per_m2', label: $t('house.price_per_m2'), type: 'number' },
          {
            key: 'executor_id', label: $t('task.executor'), type: 'select',
            options: (users || []).map(u => ({ value: u.id, label: u.full_name }))
          },
          { key: 'comment', label: $t('kanban.comment'), type: 'textarea', cols: 2 },
        ]"
      />
      <template #footer>
        <button class="btn-outline" @click="itemOpen = false">{{ $t('app.cancel') }}</button>
        <button class="btn-primary" @click="saveItem">{{ $t('app.save') }}</button>
      </template>
    </UiModal>
  </div>
</template>
