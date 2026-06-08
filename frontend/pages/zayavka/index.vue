<script setup lang="ts">
import type { RequestsBoard, RequestsColumn, RequestItem, User } from "~/types";
import { formatNumber } from "~/utils/format";

useHead({ title: () => useI18n().t("nav.requests") });

const { t } = useI18n();
const api = useApi();
const toast = useToast();
const routeQuery = useRoute().query;

const propertyTypes = [
  { value: "новостройка", label: "Навсохт" },
  { value: "коробка", label: "Блок (Коробка)" },
  { value: "котлован", label: "Котлован" },
  { value: "Коммерческая", label: "Тиҷоратӣ" },
];

const { data: boards, refresh: refreshBoards } = useAsyncData<RequestsBoard[]>(
  "req-boards",
  () => api.get<RequestsBoard[]>("/api/requests-boards"),
  { default: () => [] }
);
const { data: users } = useAsyncData<User[]>("users-list", () => api.get<User[]>("/api/users/list"), { default: () => [] });
const activeId = ref<number | null>(null);
watch(boards, (b) => { if (b?.length && !activeId.value) activeId.value = b[0].id; }, { immediate: true });

const { data: cols, refresh: refreshCols } = useAsyncData<RequestsColumn[]>(
  () => `req-cols-${activeId.value}`,
  () => activeId.value ? api.get<RequestsColumn[]>(`/api/requests-boards/${activeId.value}/columns`) : Promise.resolve([]),
  { watch: [activeId], default: () => [] }
);
const { data: items, refresh: refreshItems } = useAsyncData<RequestItem[]>(
  () => `req-items-${activeId.value}`,
  () => activeId.value ? api.get<RequestItem[]>(`/api/requests-board/${activeId.value}/requests`) : Promise.resolve([]),
  { watch: [activeId], default: () => [] }
);

const activeBoard = computed(() => boards.value?.find((b) => b.id === activeId.value));

// ---- filters ----
const f = reactive({
  search: (routeQuery.search as string) || "",
  type: "all",
  author: "all",
  priceMin: "", priceMax: "",
  areaMin: "", areaMax: "",
  floorMin: "", floorMax: "",
});
const inRange = (v: number, min: string, max: string) =>
  (!min || v >= Number(min)) && (!max || v <= Number(max));

const filteredItems = computed(() => {
  let list = items.value || [];
  if (f.search) {
    const q = f.search.toLowerCase();
    list = list.filter((r) =>
      [r.client_name, r.phone, r.address, r.property_type].some((x) => (x || "").toLowerCase().includes(q))
    );
  }
  if (f.type !== "all") list = list.filter((r) => r.property_type === f.type);
  if (f.author !== "all") list = list.filter((r) => String(r.executor_id) === f.author);
  list = list.filter((r) =>
    inRange(r.total_price || 0, f.priceMin, f.priceMax) &&
    inRange(r.area || 0, f.areaMin, f.areaMax) &&
    inRange(r.floor || 0, f.floorMin, f.floorMax)
  );
  return list;
});

// ---- board CRUD ----
const boardModal = ref(false);
const newBoard = reactive({ title: "", color: "#0f172a" });
const createBoard = async () => {
  const r = await api.post<{ id: number }>("/api/requests-boards", { title: newBoard.title, color: newBoard.color, is_public: true });
  boardModal.value = false;
  newBoard.title = "";
  toast.success(t("notify.created"));
  await refreshBoards();
  activeId.value = r.id;
};
const deleteBoard = async () => {
  if (!activeId.value || !confirm(t("app.delete") + "?")) return;
  await api.del(`/api/requests-boards/${activeId.value}`);
  activeId.value = null;
  await refreshBoards();
};

// ---- column CRUD ----
const colModal = ref(false);
const editingCol = ref<Partial<RequestsColumn>>({});
const openNewCol = () => { editingCol.value = { board_id: activeId.value!, color: "#3b82f6" }; colModal.value = true; };
const openEditCol = (c: RequestsColumn) => { editingCol.value = { ...c }; colModal.value = true; };
const saveCol = async () => {
  if (editingCol.value.id) await api.put(`/api/requests-columns/${editingCol.value.id}`, editingCol.value);
  else await api.post("/api/requests-columns", editingCol.value);
  colModal.value = false;
  await refreshCols();
};
const deleteCol = async (c: RequestsColumn) => {
  if (!confirm(t("app.delete") + "?")) return;
  await api.del(`/api/requests-columns/${c.id}`);
  await Promise.all([refreshCols(), refreshItems()]);
};

// ---- request item CRUD ----
const reqModal = ref(false);
const blank = (): Partial<RequestItem> => ({ property_type: "новостройка" });
const editing = ref<Partial<RequestItem>>(blank());
const openNewReq = (columnId: number) => { editing.value = { ...blank(), board_id: activeId.value!, column_id: columnId }; reqModal.value = true; };
const openEditReq = (r: RequestItem) => { editing.value = { ...r }; reqModal.value = true; };
const saveReq = async () => {
  const e = editing.value;
  if (!e.client_name || !e.phone) { toast.error(t("notify.error")); return; }
  try {
    if (e.id) {
      await api.put(`/api/requests-update/${e.id}`, e);
      toast.success(t("notify.updated"));
    } else {
      await api.post("/api/requests-new", e);
      toast.success(t("notify.created"));
    }
    reqModal.value = false;
    await refreshItems();
  } catch (err: any) {
    toast.error(err?.data?.error || t("notify.error"));
  }
};
const deleteReq = async (r: RequestItem) => {
  if (!confirm(t("app.delete") + "?")) return;
  await api.del(`/api/requests-delete/${r.id}`);
  await refreshItems();
};
const moveCard = async (e: { cardId: number; toColumnId: number; orderIndex: number }) => {
  await api.post("/api/requests-move", { request_id: e.cardId, column_id: e.toColumnId, order_index: e.orderIndex });
  await refreshItems();
};
const exportBoard = () => api.download(`/api/requests-export/excel?board_id=${activeId.value}`, `zayavka_${activeId.value}.xlsx`);
</script>

<template>
  <div>
    <!-- header -->
    <div class="flex items-center justify-between flex-wrap gap-3 mb-5">
      <div class="flex items-center gap-3">
        <div class="w-11 h-11 bg-brand-soft text-brand rounded-2xl flex items-center justify-center text-lg">
          <i class="fas fa-columns" />
        </div>
        <div>
          <select v-model="activeId" class="text-xl font-bold text-ink bg-transparent focus:outline-none cursor-pointer">
            <option v-for="b in boards" :key="b.id" :value="b.id">{{ b.title }}</option>
          </select>
          <div class="text-xs text-ink-soft">
            <i class="fas fa-chalkboard-teacher" /> {{ t("app.total") }}: {{ filteredItems.length }}
          </div>
        </div>
      </div>
      <div class="flex items-center gap-2 flex-wrap">
        <button class="btn-outline !py-2 !px-3" :title="t('kanban.create_column')" @click="openNewCol">
          <i class="fas fa-plus" /> {{ t("kanban.column") }}
        </button>
        <button class="btn-primary !py-2 !px-3" @click="activeId && openNewReq(cols?.[0]?.id || 0)">
          <i class="fas fa-plus-circle" /> {{ t("requests.new") }}
        </button>
        <button class="btn-outline !py-2 !px-3" :title="t('app.export')" @click="exportBoard">
          <i class="fas fa-file-excel" />
        </button>
        <button class="btn-outline !py-2 !px-3" @click="boardModal = true">
          <i class="fas fa-plus" /> {{ t("kanban.create_board") }}
        </button>
        <button v-if="activeBoard" class="btn-danger !py-2 !px-3" @click="deleteBoard">
          <i class="fas fa-trash" />
        </button>
      </div>
    </div>

    <!-- filters -->
    <div class="flex flex-wrap gap-2.5 mb-5">
      <div class="flex items-center gap-1.5 bg-page rounded-xl px-3 py-1.5 text-xs">
        <i class="fas fa-credit-card text-brand" /> <span class="text-ink-soft">{{ t("house.total_price") }}:</span>
        <input v-model="f.priceMin" type="number" placeholder="аз" class="w-16 bg-transparent focus:outline-none" />
        <span class="text-ink-soft">—</span>
        <input v-model="f.priceMax" type="number" placeholder="то" class="w-16 bg-transparent focus:outline-none" />
      </div>
      <div class="flex items-center gap-1.5 bg-page rounded-xl px-3 py-1.5 text-xs">
        <i class="fas fa-ruler-combined text-brand" /> <span class="text-ink-soft">м²:</span>
        <input v-model="f.areaMin" type="number" placeholder="аз" class="w-14 bg-transparent focus:outline-none" />
        <span class="text-ink-soft">—</span>
        <input v-model="f.areaMax" type="number" placeholder="то" class="w-14 bg-transparent focus:outline-none" />
      </div>
      <div class="flex items-center gap-1.5 bg-page rounded-xl px-3 py-1.5 text-xs">
        <i class="fas fa-layer-group text-brand" /> <span class="text-ink-soft">{{ t("house.floor") }}:</span>
        <input v-model="f.floorMin" type="number" placeholder="аз" class="w-12 bg-transparent focus:outline-none" />
        <span class="text-ink-soft">—</span>
        <input v-model="f.floorMax" type="number" placeholder="то" class="w-12 bg-transparent focus:outline-none" />
      </div>
      <div class="flex items-center gap-1.5 bg-page rounded-xl px-3 py-1.5 text-xs flex-1 min-w-[160px]">
        <i class="fas fa-search text-ink-soft" />
        <input v-model="f.search" type="text" :placeholder="t('app.search')" class="flex-1 bg-transparent focus:outline-none" />
      </div>
      <select v-model="f.type" class="bg-page rounded-xl px-3 py-1.5 text-xs focus:outline-none">
        <option value="all">{{ t("app.all") }}</option>
        <option v-for="p in propertyTypes" :key="p.value" :value="p.value">{{ p.label }}</option>
      </select>
      <select v-model="f.author" class="bg-page rounded-xl px-3 py-1.5 text-xs focus:outline-none">
        <option value="all">{{ t("task.executor") }}</option>
        <option v-for="u in users" :key="u.id" :value="String(u.id)">{{ u.full_name }}</option>
      </select>
    </div>

    <!-- kanban -->
    <KanbanBoard
      v-if="cols?.length"
      :columns="cols"
      :cards="filteredItems"
      @move-card="moveCard"
      @add-card="openNewReq"
      @edit-column="openEditCol"
      @delete-column="deleteCol"
      @add-column="openNewCol"
    >
      <template #card="{ card }">
        <div class="flex items-start justify-between gap-2 mb-1.5">
          <span class="font-semibold text-sm text-ink truncate">{{ (card as RequestItem).client_name }}</span>
          <div class="flex gap-0.5">
            <button class="text-ink-soft hover:text-brand text-[11px] p-1" @click.stop="openEditReq(card as RequestItem)"><i class="fas fa-pen" /></button>
            <button class="text-ink-soft hover:text-red-500 text-[11px] p-1" @click.stop="deleteReq(card as RequestItem)"><i class="fas fa-trash" /></button>
          </div>
        </div>
        <div class="text-xs text-ink-medium mb-2"><i class="fas fa-phone-alt text-brand text-[10px] mr-1" />{{ (card as RequestItem).phone }}</div>
        <div class="flex flex-wrap gap-1 mb-1.5">
          <span class="text-[10px] bg-page px-2 py-0.5 rounded-md text-ink-medium"><i class="fas fa-building mr-1" />{{ (card as RequestItem).property_type }}</span>
          <span class="text-[10px] bg-page px-2 py-0.5 rounded-md text-ink-medium"><i class="fas fa-ruler-combined mr-1" />{{ (card as RequestItem).area }} м²</span>
          <span class="text-[10px] bg-page px-2 py-0.5 rounded-md text-ink-medium"><i class="fas fa-door-open mr-1" />{{ (card as RequestItem).rooms }}</span>
        </div>
        <div class="flex flex-wrap gap-1 mb-1.5">
          <span class="text-[10px] bg-brand-soft text-brand px-2 py-0.5 rounded-md font-semibold">{{ formatNumber((card as RequestItem).total_price) }} C</span>
          <span class="text-[10px] bg-page px-2 py-0.5 rounded-md text-ink-medium">{{ formatNumber((card as RequestItem).price_per_m2) }} C/м²</span>
        </div>
        <div class="flex items-center justify-between text-[10px] text-ink-soft pt-1.5 border-t border-border-soft">
          <span><i class="far fa-calendar-alt mr-1" />{{ new Date((card as RequestItem).created_at).toLocaleDateString("ru-RU") }}</span>
          <span><i class="fas fa-user-check mr-1" />{{ (card as RequestItem).executor_name || "—" }}</span>
        </div>
      </template>
    </KanbanBoard>
    <UiEmpty v-else icon="fa-columns" />

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

    <!-- request modal -->
    <UiModal v-model="reqModal" :title="editing.id ? t('app.edit') : t('requests.new')" size="lg">
      <div class="grid grid-cols-1 md:grid-cols-2 gap-4">
        <div>
          <label class="label">{{ t("house.construction_type") }} *</label>
          <select v-model="editing.property_type" class="input">
            <option v-for="p in propertyTypes" :key="p.value" :value="p.value">{{ p.label }}</option>
          </select>
        </div>
        <div><label class="label">{{ t("house.address") }} *</label><input v-model="editing.address" class="input" /></div>
        <div><label class="label">{{ t("house.area") }} *</label><input v-model.number="editing.area" type="number" step="0.01" class="input" /></div>
        <div><label class="label">{{ t("house.rooms") }} *</label><input v-model.number="editing.rooms" type="number" class="input" /></div>
        <div><label class="label">{{ t("house.windows") }} *</label><input v-model.number="editing.windows" type="number" class="input" /></div>
        <div><label class="label">{{ t("house.floor") }} *</label><input v-model.number="editing.floor" type="number" class="input" /></div>
        <div><label class="label">{{ t("house.total_floors") }} *</label><input v-model.number="editing.total_floors" type="number" class="input" /></div>
        <div><label class="label">{{ t("house.total_price") }} (C) *</label><input v-model.number="editing.total_price" type="number" class="input" /></div>
        <div><label class="label">{{ t("house.price_per_m2") }} (C) *</label><input v-model.number="editing.price_per_m2" type="number" class="input" /></div>
        <div><label class="label">{{ t("kanban.phone") }} *</label><input v-model="editing.phone" class="input" /></div>
        <div><label class="label">{{ t("kanban.client_name") }} *</label><input v-model="editing.client_name" class="input" /></div>
        <div>
          <label class="label">{{ t("task.executor") }}</label>
          <select v-model="editing.executor_id" class="input">
            <option :value="undefined">—</option>
            <option v-for="u in users" :key="u.id" :value="u.id">{{ u.full_name }}</option>
          </select>
        </div>
        <div class="md:col-span-2"><label class="label">{{ t("kanban.comment") }}</label><textarea v-model="editing.comment" rows="2" class="input" /></div>
      </div>
      <template #footer>
        <button class="btn-outline" @click="reqModal = false">{{ t("app.cancel") }}</button>
        <button class="btn-primary" @click="saveReq">{{ t("app.save") }}</button>
      </template>
    </UiModal>
  </div>
</template>
