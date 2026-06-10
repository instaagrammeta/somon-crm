<script setup lang="ts">
import type { ObjektProject, ObjektBlock, ObjektApartment } from "~/types";
import type { FormField } from "~/components/EntityForm.vue";
import { formatNumber } from "~/utils/format";

useHead({ title: () => useI18n().t("nav.objekt") });

const { t } = useI18n();
const api = useApi();
const toast = useToast();

const { data: projects, refresh: refreshProjects } = useAsyncData<ObjektProject[]>(
  "objekt-projects",
  () => api.get<ObjektProject[]>("/api/objekt/projects"),
  { lazy: true, default: () => [] }
);
const activeId = ref<number | null>(null);
watch(projects, (p) => { if (p?.length && !activeId.value) activeId.value = p[0].id; }, { immediate: true });

const { data: blocks, refresh: refreshBlocks } = useAsyncData<ObjektBlock[]>(
  () => `objekt-blocks-${activeId.value}`,
  () => activeId.value ? api.get<ObjektBlock[]>(`/api/objekt/projects/${activeId.value}/blocks`) : Promise.resolve([]),
  { watch: [activeId], lazy: true, default: () => [] }
);
const { data: apartments, refresh: refreshApts } = useAsyncData<ObjektApartment[]>(
  () => `objekt-apts-${activeId.value}`,
  () => activeId.value ? api.get<ObjektApartment[]>(`/api/objekt/projects/${activeId.value}/apartments`) : Promise.resolve([]),
  { watch: [activeId], lazy: true, default: () => [] }
);

// === project CRUD ===
const projectOpen = ref(false);
const editingProject = ref<Partial<ObjektProject>>({});
const openNewProject = () => { editingProject.value = {}; projectOpen.value = true; };
const saveProject = async () => {
  if (editingProject.value.id) await api.put(`/api/objekt/projects/${editingProject.value.id}`, editingProject.value);
  else {
    const r = await api.post<{ id: number }>("/api/objekt/projects", editingProject.value);
    activeId.value = r.id;
  }
  projectOpen.value = false;
  await refreshProjects();
};
const deleteProject = async () => {
  if (!activeId.value || !confirm(t("app.delete") + "?")) return;
  await api.del(`/api/objekt/projects/${activeId.value}`);
  activeId.value = null;
  await refreshProjects();
};

// === block CRUD ===
const blockOpen = ref(false);
const editingBlock = ref<Partial<ObjektBlock>>({});
const openNewBlock = () => {
  editingBlock.value = {
    project_id: activeId.value!, floor_from: 1, floor_to: 9,
    default_area: 60, default_rooms: 2, default_windows: 2,
    default_price_per_m2: 8000,
  };
  blockOpen.value = true;
};
const openEditBlock = (b: ObjektBlock) => { editingBlock.value = { ...b }; blockOpen.value = true; };
const saveBlock = async () => {
  const fd = new FormData();
  Object.entries(editingBlock.value).forEach(([k, v]) => {
    if (v != null) fd.append(k, String(v));
  });
  if (editingBlock.value.id) await api.upload(`/api/objekt/blocks/${editingBlock.value.id}`, fd);
  else await api.upload("/api/objekt/blocks", fd);
  blockOpen.value = false;
  await Promise.all([refreshBlocks(), refreshApts()]);
};
const deleteBlock = async (b: ObjektBlock) => {
  if (!confirm(t("app.delete") + "?")) return;
  await api.del(`/api/objekt/blocks/${b.id}`);
  await Promise.all([refreshBlocks(), refreshApts()]);
};

// === apartment status ===
const aptOpen = ref(false);
const aptEditing = ref<Partial<ObjektApartment>>({});
const openApt = (a: ObjektApartment) => { aptEditing.value = { ...a }; aptOpen.value = true; };
const saveAptStatus = async (status: ObjektApartment["status"]) => {
  if (!aptEditing.value.id) return;
  await api.patch(`/api/objekt/apartments/${aptEditing.value.id}/status`, {
    status,
    client_name: aptEditing.value.client_name || "",
    client_phone: aptEditing.value.client_phone || "",
  });
  aptOpen.value = false;
  await refreshApts();
};

const exportXlsx = () => api.download(`/api/objekt/export/excel?project_id=${activeId.value}`, `shahmatka_${activeId.value}.xlsx`);

const blockFields = computed<FormField[]>(() => [
  { key: "name", label: t("objekt.blocks"), required: true, cols: 2 },
  { key: "floor_from", label: t("objekt.floor_from"), type: "number", required: true },
  { key: "floor_to", label: t("objekt.floor_to"), type: "number", required: true },
  { key: "default_area", label: t("objekt.default_area"), type: "number" },
  { key: "default_rooms", label: t("objekt.default_rooms"), type: "number" },
  { key: "default_windows", label: t("house.windows"), type: "number" },
  { key: "default_price_per_m2", label: t("house.price_per_m2"), type: "number" },
  { key: "default_balcony", label: t("objekt.balcony"), type: "checkbox" },
  { key: "default_bathroom_count", label: t("objekt.bathroom"), type: "number" },
  { key: "description", label: t("post.description"), type: "textarea", cols: 2 },
]);

// Group apartments by block
const blockGroups = computed(() => {
  const m = new Map<number, { block: ObjektBlock; apts: ObjektApartment[] }>();
  for (const b of blocks.value || []) m.set(b.id, { block: b, apts: [] });
  for (const a of apartments.value || []) {
    const g = m.get(a.block_id);
    if (g) g.apts.push(a);
  }
  for (const g of m.values()) g.apts.sort((a, b) => b.floor - a.floor);
  return Array.from(m.values()).sort((a, b) => a.block.order_index - b.block.order_index);
});

const statusBg = (s: string) =>
  ({ free: "bg-brand text-white", reserved: "bg-yellow-400 text-ink", sold: "bg-red-500 text-white" })[s] || "bg-gray-300 text-ink";
</script>

<template>
  <div class="card p-6 lg:p-8">
    <PageHeader :title="$t('nav.objekt')">
      <select v-model="activeId" class="input !py-2 max-w-[260px]">
        <option v-for="p in projects" :key="p.id" :value="p.id">{{ p.name }}</option>
      </select>
      <button v-if="activeId" class="btn-outline" @click="exportXlsx">
        <i class="fa-solid fa-file-excel" /> Excel
      </button>
      <button v-if="activeId" class="btn-outline" @click="openNewBlock">
        <i class="fa-solid fa-plus" /> {{ $t("objekt.create_block") }}
      </button>
      <button v-if="activeId" class="btn-danger" @click="deleteProject">
        <i class="fa-solid fa-trash" />
      </button>
      <button class="btn-primary" @click="openNewProject">
        <i class="fa-solid fa-plus" /> {{ $t("objekt.create_project") }}
      </button>
    </PageHeader>

    <!-- Status legend -->
    <div class="flex gap-3 mb-6 text-xs font-medium">
      <span class="flex items-center gap-1.5"><span class="w-3 h-3 bg-brand rounded" />{{ $t("objekt.free") }}</span>
      <span class="flex items-center gap-1.5"><span class="w-3 h-3 bg-yellow-400 rounded" />{{ $t("objekt.reserved") }}</span>
      <span class="flex items-center gap-1.5"><span class="w-3 h-3 bg-red-500 rounded" />{{ $t("objekt.sold") }}</span>
    </div>

    <UiEmpty v-if="!projects?.length" icon="fa-table-cells-large" />

    <div v-else-if="!blocks?.length" class="text-center py-10 text-ink-soft">
      Блокҳо мавҷуд нестанд
    </div>

    <div v-else class="space-y-8">
      <section v-for="g in blockGroups" :key="g.block.id">
        <div class="flex items-center justify-between mb-3">
          <h3 class="font-bold text-ink">
            {{ g.block.name }}
            <span class="text-xs text-ink-soft font-normal ml-2">
              {{ g.block.floor_from }}—{{ g.block.floor_to }} {{ $t("house.floor") }}
            </span>
          </h3>
          <div class="flex gap-1">
            <button class="btn-ghost !p-2" @click="openEditBlock(g.block)">
              <i class="fa-solid fa-pen" />
            </button>
            <button class="btn-ghost !p-2 text-red-500" @click="deleteBlock(g.block)">
              <i class="fa-solid fa-trash" />
            </button>
          </div>
        </div>
        <div class="grid gap-1.5" style="grid-template-columns: repeat(auto-fill, minmax(110px, 1fr))">
          <button
            v-for="apt in g.apts"
            :key="apt.id"
            class="rounded-xl p-3 text-xs font-semibold transition-transform hover:scale-105 text-left"
            :class="statusBg(apt.status)"
            @click="openApt(apt)"
          >
            <div class="font-bold mb-0.5">{{ apt.floor }} {{ $t('house.floor') }}</div>
            <div class="opacity-90 text-[10px]">
              {{ apt.rooms }} {{ $t("house.rooms") }} · {{ formatNumber(apt.area, 1) }} м²
            </div>
            <div v-if="apt.client_name" class="opacity-90 text-[10px] mt-1 truncate">
              {{ apt.client_name }}
            </div>
            <div class="opacity-90 text-[10px] mt-1">
              {{ formatNumber(apt.total_price) }}
            </div>
          </button>
        </div>
      </section>
    </div>

    <UiModal v-model="projectOpen" :title="editingProject.id ? $t('app.edit') : $t('objekt.create_project')">
      <div class="space-y-4">
        <div><label class="label">Name</label><input v-model="editingProject.name" class="input" /></div>
        <div><label class="label">Address</label><input v-model="editingProject.address" class="input" /></div>
        <div><label class="label">Developer</label><input v-model="editingProject.developer" class="input" /></div>
      </div>
      <template #footer>
        <button class="btn-outline" @click="projectOpen = false">{{ $t('app.cancel') }}</button>
        <button class="btn-primary" @click="saveProject">{{ $t('app.save') }}</button>
      </template>
    </UiModal>

    <UiModal v-model="blockOpen" :title="editingBlock.id ? $t('app.edit') : $t('objekt.create_block')" size="lg">
      <EntityForm v-model="editingBlock" :fields="blockFields" />
      <template #footer>
        <button class="btn-outline" @click="blockOpen = false">{{ $t('app.cancel') }}</button>
        <button class="btn-primary" @click="saveBlock">{{ $t('app.save') }}</button>
      </template>
    </UiModal>

    <UiModal v-model="aptOpen" title="Хона" size="md">
      <div class="space-y-4">
        <div class="grid grid-cols-2 gap-3 text-sm">
          <div><span class="text-ink-soft">Ошёна:</span> <b>{{ aptEditing.floor }}</b></div>
          <div><span class="text-ink-soft">Майдон:</span> <b>{{ aptEditing.area }} м²</b></div>
          <div><span class="text-ink-soft">Ҳуҷра:</span> <b>{{ aptEditing.rooms }}</b></div>
          <div><span class="text-ink-soft">Нарх:</span> <b>{{ formatNumber(aptEditing.total_price) }}</b></div>
        </div>
        <div>
          <label class="label">{{ $t("kanban.client_name") }}</label>
          <input v-model="aptEditing.client_name" class="input" />
        </div>
        <div>
          <label class="label">{{ $t("kanban.phone") }}</label>
          <input v-model="aptEditing.client_phone" class="input" />
        </div>
        <div class="flex gap-2 mt-4">
          <button class="btn-outline flex-1" @click="saveAptStatus('free')">
            <i class="fa-solid fa-check" /> {{ $t("objekt.free") }}
          </button>
          <button class="btn-secondary flex-1" @click="saveAptStatus('reserved')">
            <i class="fa-solid fa-clock" /> {{ $t("objekt.reserved") }}
          </button>
          <button class="btn-danger flex-1" @click="saveAptStatus('sold')">
            <i class="fa-solid fa-coins" /> {{ $t("objekt.sold") }}
          </button>
        </div>
      </div>
    </UiModal>
  </div>
</template>
