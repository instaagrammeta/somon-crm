<script setup lang="ts">
import type { Task, User } from "~/types";

useHead({ title: () => useI18n().t("nav.tasks") });

const api = useApi();
const toast = useToast();
const { t } = useI18n();

const status = ref<"all" | "new" | "in_progress" | "done" | "cancelled">("all");

const { data: tasks, pending, refresh } = useAsyncData<Task[]>(
  "tasks",
  () => api.get<Task[]>(`/api/tasks${status.value !== "all" ? `?status=${status.value}` : ""}`),
  { watch: [status], lazy: true, default: () => [] }
);
const { data: users } = useAsyncData<User[]>("users-list", () => api.get<User[]>("/api/users/list"), { lazy: true, default: () => [] });

const filters = computed(() => [
  { key: "all", label: t("app.all") },
  { key: "new", label: "🟢 " + t("task.status_new") },
  { key: "in_progress", label: "🟡 " + t("task.status_in_progress") },
  { key: "done", label: "✅ " + t("task.status_done") },
  { key: "cancelled", label: "🔴 " + t("task.status_cancelled") },
]);

const statusOptions = computed(() => [
  { value: "new", label: t("task.status_new") },
  { value: "in_progress", label: t("task.status_in_progress") },
  { value: "done", label: t("task.status_done") },
  { value: "cancelled", label: t("task.status_cancelled") },
]);

// ===== card helpers =====
const photosOf = (tk: Task): string[] =>
  tk.photos?.length ? tk.photos : tk.photo ? [tk.photo] : [];

const userName = (id: number) => users.value?.find((u) => u.id === id)?.full_name || "";

const executorNamesOf = (tk: Task): string[] => {
  if (tk.executor_names?.length) return tk.executor_names;
  if (tk.executor_ids?.length) return tk.executor_ids.map(userName).filter(Boolean);
  if (tk.executor_name) return [tk.executor_name];
  if (tk.executor_id) {
    const n = userName(tk.executor_id);
    return n ? [n] : [];
  }
  return [];
};

// ===== modal / form state =====
interface PhotoItem {
  uid: string;
  kind: "existing" | "new";
  url: string; // remote url (existing) — used as `existing_photos`
  preview: string; // <img> src (object url for new files)
  file?: File;
}

let uidSeq = 0;
const nextUid = () => `p${Date.now()}_${uidSeq++}`;

const open = ref(false);
const editing = ref<Partial<Task>>({});
const executorIds = ref<number[]>([]);
const photoItems = ref<PhotoItem[]>([]);
const saving = ref(false);
const uploadPct = ref(0);
const fileInput = ref<HTMLInputElement | null>(null);

const clearNewPreviews = () => {
  for (const p of photoItems.value) {
    if (p.kind === "new" && p.preview.startsWith("blob:")) URL.revokeObjectURL(p.preview);
  }
};

const openCreate = () => {
  clearNewPreviews();
  editing.value = { status: "new" };
  executorIds.value = [];
  photoItems.value = [];
  uploadPct.value = 0;
  open.value = true;
};

const openEdit = (tk: Task) => {
  clearNewPreviews();
  editing.value = { ...tk };
  executorIds.value = tk.executor_ids?.length
    ? [...tk.executor_ids]
    : tk.executor_id
      ? [tk.executor_id]
      : [];
  photoItems.value = photosOf(tk).map((u) => ({ uid: nextUid(), kind: "existing", url: u, preview: u }));
  uploadPct.value = 0;
  open.value = true;
};

watch(open, (v) => {
  if (!v) clearNewPreviews();
});

const onPhotos = (e: Event) => {
  const input = e.target as HTMLInputElement;
  const files = Array.from(input.files || []);
  for (const f of files) {
    photoItems.value.push({ uid: nextUid(), kind: "new", url: "", preview: URL.createObjectURL(f), file: f });
  }
  input.value = ""; // allow re-selecting the same file
};

const removePhoto = (uid: string) => {
  const idx = photoItems.value.findIndex((p) => p.uid === uid);
  if (idx === -1) return;
  const [removed] = photoItems.value.splice(idx, 1);
  if (removed.kind === "new" && removed.preview.startsWith("blob:")) URL.revokeObjectURL(removed.preview);
};

const movePhoto = (idx: number, dir: -1 | 1) => {
  const to = idx + dir;
  if (to < 0 || to >= photoItems.value.length) return;
  const arr = photoItems.value;
  [arr[idx], arr[to]] = [arr[to], arr[idx]];
};

const toggleExecutor = (id: number) => {
  const i = executorIds.value.indexOf(id);
  if (i === -1) executorIds.value.push(id);
  else executorIds.value.splice(i, 1);
};

const selectedExecutorLabel = computed(() => {
  if (!executorIds.value.length) return t("task.no_executor");
  return executorIds.value.map(userName).filter(Boolean).join(", ");
});

const save = async () => {
  if (saving.value) return; // guard against double-submit / impatient clicks
  if (!editing.value.title?.trim()) {
    toast.error(t("task.title_required"));
    return;
  }
  saving.value = true;
  uploadPct.value = 0;

  const fd = new FormData();
  fd.append("title", editing.value.title.trim());
  fd.append("description", editing.value.description || "");
  if (editing.value.status) fd.append("status", editing.value.status);

  fd.append("manage_executors", "1");
  for (const id of executorIds.value) fd.append("executor_ids", String(id));

  fd.append("manage_photos", "1");
  for (const p of photoItems.value) {
    if (p.kind === "existing") fd.append("existing_photos", p.url);
    else if (p.file) fd.append("photos", p.file);
  }

  try {
    if (editing.value.id) {
      await api.uploadProgress(`/api/tasks/${editing.value.id}`, fd, "PUT", (pct) => (uploadPct.value = pct));
      toast.success(t("notify.updated"));
    } else {
      await api.uploadProgress("/api/tasks", fd, "POST", (pct) => (uploadPct.value = pct));
      toast.success(t("notify.created"));
    }
    open.value = false;
    await refresh();
  } catch (e: any) {
    toast.error(e?.data?.error || t("notify.error"));
  } finally {
    saving.value = false;
  }
};

const changeStatus = async (tk: Task, newStatus: string) => {
  const fd = new FormData();
  fd.append("status", newStatus);
  try {
    await api.upload(`/api/tasks/${tk.id}`, fd, "PUT");
    await refresh();
  } catch (e: any) {
    toast.error(e?.data?.error || t("notify.error"));
  }
};

const confirmDel = ref<{ open: boolean; id?: number }>({ open: false });
const askDelete = (tk: Task) => (confirmDel.value = { open: true, id: tk.id });
const doDelete = async () => {
  if (!confirmDel.value.id) return;
  await api.del(`/api/tasks/${confirmDel.value.id}`);
  toast.success(t("notify.deleted"));
  await refresh();
};

const statusBadge = (s: string) =>
  ({ new: "status-new", in_progress: "status-progress", done: "status-done", cancelled: "status-cancel" })[s] || "status-new";
</script>

<template>
  <div>
    <!-- header -->
    <div class="flex justify-between items-center flex-wrap gap-4 mb-6">
      <div>
        <h1 class="text-[28px] font-bold text-ink mb-1">{{ t("nav.tasks") }}</h1>
        <p class="text-ink-soft text-sm">{{ t("task.subtitle") }}</p>
      </div>
      <button class="btn-primary" @click="openCreate">
        <i class="fas fa-plus" /> {{ t("task.create_task") }}
      </button>
    </div>

    <!-- filter pills -->
    <div class="flex gap-3 mb-6 flex-wrap">
      <button
        v-for="f in filters"
        :key="f.key"
        class="px-[18px] py-[9px] rounded-full border-[1.5px] text-[13px] font-medium transition-all"
        :class="status === f.key ? 'bg-brand text-white border-brand' : 'bg-white border-border-soft text-ink-medium hover:border-brand hover:text-brand'"
        @click="status = f.key as any"
      >
        {{ f.label }}
      </button>
    </div>

    <UiLoader v-if="pending" />
    <UiEmpty v-else-if="!tasks?.length" icon="fa-tasks" />

    <!-- grid -->
    <div v-else class="grid gap-5" style="grid-template-columns: repeat(auto-fill, minmax(340px, 1fr))">
      <div
        v-for="tk in tasks"
        :key="tk.id"
        class="bg-white rounded-[20px] p-5 border border-border-soft transition-all hover:-translate-y-[3px] hover:shadow-hover hover:border-brand"
      >
        <div class="flex justify-between items-start gap-2.5 mb-3.5">
          <h3 class="text-base font-bold text-ink leading-[1.3]">{{ tk.title }}</h3>
          <span class="status-badge" :class="statusBadge(tk.status)">{{ t("task.status_" + tk.status) }}</span>
        </div>

        <!-- photo gallery -->
        <template v-if="photosOf(tk).length">
          <img :src="photosOf(tk)[0]" class="w-full rounded-[14px] mb-2 max-h-[180px] object-cover" />
          <div v-if="photosOf(tk).length > 1" class="flex gap-2 mb-3 overflow-x-auto pb-1">
            <img
              v-for="(p, i) in photosOf(tk).slice(1)"
              :key="i"
              :src="p"
              class="w-14 h-14 rounded-[10px] object-cover flex-shrink-0 border border-border-soft"
            />
          </div>
        </template>

        <p v-if="tk.description" class="text-ink-medium text-[13px] leading-relaxed mb-3.5">{{ tk.description }}</p>

        <div class="bg-brand-soft p-3 rounded-[14px] my-3 text-xs text-ink-medium flex flex-col gap-1.5">
          <div v-if="executorNamesOf(tk).length" class="flex items-start gap-2">
            <i class="fas fa-user-check w-4 text-brand mt-0.5" />
            <span class="flex flex-wrap gap-1">
              <span
                v-for="(n, i) in executorNamesOf(tk)"
                :key="i"
                class="bg-white/70 rounded-full px-2 py-0.5"
                >{{ n }}</span
              >
            </span>
          </div>
          <div v-if="tk.author_name" class="flex items-center gap-2">
            <i class="fas fa-user-pen w-4 text-brand" /> {{ tk.author_name }}
          </div>
        </div>

        <div class="mt-3.5 flex gap-2.5 items-center">
          <select
            class="flex-1 px-3 py-2.5 bg-page border-[1.5px] border-transparent rounded-xl text-xs cursor-pointer focus:outline-none focus:border-brand focus:bg-white"
            :value="tk.status"
            @change="changeStatus(tk, ($event.target as HTMLSelectElement).value)"
          >
            <option v-for="o in statusOptions" :key="o.value" :value="o.value">{{ o.label }}</option>
          </select>
          <button class="w-[34px] h-[34px] bg-page rounded-[10px] text-brand hover:bg-brand-soft transition-all" @click="openEdit(tk)">
            <i class="fas fa-pen text-[13px]" />
          </button>
          <button class="w-[34px] h-[34px] bg-page rounded-[10px] text-red-500 hover:bg-red-100 transition-all" @click="askDelete(tk)">
            <i class="fas fa-trash text-[13px]" />
          </button>
        </div>
      </div>
    </div>

    <!-- modal -->
    <UiModal v-model="open" :title="editing.id ? t('app.edit') : t('task.create_task')" size="lg">
      <div class="space-y-4">
        <div>
          <label class="label"><i class="fas fa-heading text-brand mr-1.5" /> {{ t("task.title_field") }} *</label>
          <input v-model="editing.title" class="input" required />
        </div>
        <div>
          <label class="label"><i class="fas fa-align-left text-brand mr-1.5" /> {{ t("task.description") }}</label>
          <textarea v-model="editing.description" rows="3" class="input" />
        </div>

        <!-- multi-executor -->
        <div>
          <label class="label">
            <i class="fas fa-user-check text-brand mr-1.5" /> {{ t("task.executors") }}
            <span v-if="executorIds.length" class="ml-1 text-brand font-semibold">({{ executorIds.length }})</span>
          </label>
          <div class="text-xs text-ink-soft mb-2 truncate">{{ selectedExecutorLabel }}</div>
          <div class="max-h-44 overflow-y-auto rounded-xl border border-border-soft divide-y divide-border-soft">
            <label
              v-for="u in users"
              :key="u.id"
              class="flex items-center gap-3 px-3 py-2.5 cursor-pointer hover:bg-brand-soft transition-colors"
            >
              <input
                type="checkbox"
                class="accent-brand w-4 h-4"
                :checked="executorIds.includes(u.id)"
                @change="toggleExecutor(u.id)"
              />
              <span class="text-sm text-ink">{{ u.full_name }}</span>
              <span v-if="u.category" class="text-[11px] text-ink-soft ml-auto">{{ u.category }}</span>
            </label>
            <div v-if="!users?.length" class="px-3 py-3 text-sm text-ink-soft">—</div>
          </div>
        </div>

        <!-- multi-photo manager -->
        <div>
          <label class="label"><i class="fas fa-image text-brand mr-1.5" /> {{ t("task.photo") }}</label>
          <div class="flex items-center gap-2 mb-3">
            <button
              type="button"
              class="inline-flex items-center gap-2 px-4 py-2.5 bg-brand-soft text-brand rounded-[10px] cursor-pointer text-[13px] font-semibold hover:bg-brand-light transition-colors"
              @click="fileInput?.click()"
            >
              <i class="fas fa-upload" /> {{ t("task.add_photos") }}
            </button>
            <input ref="fileInput" type="file" accept="image/*" multiple hidden @change="onPhotos" />
          </div>

          <div v-if="photoItems.length" class="grid grid-cols-3 sm:grid-cols-4 gap-3">
            <div
              v-for="(p, idx) in photoItems"
              :key="p.uid"
              class="relative group rounded-[12px] overflow-hidden border border-border-soft aspect-square"
            >
              <img :src="p.preview" class="w-full h-full object-cover" />
              <span
                v-if="p.kind === 'new'"
                class="absolute top-1 left-1 bg-brand text-white text-[10px] px-1.5 py-0.5 rounded-full"
                >{{ t("task.new_badge") }}</span
              >
              <!-- controls -->
              <div class="absolute inset-x-0 bottom-0 flex justify-between items-center px-1.5 py-1 bg-ink/45 opacity-0 group-hover:opacity-100 transition-opacity">
                <button
                  type="button"
                  class="w-6 h-6 rounded-md bg-white/85 text-ink disabled:opacity-30"
                  :disabled="idx === 0"
                  :title="t('task.move_left')"
                  @click="movePhoto(idx, -1)"
                >
                  <i class="fas fa-chevron-left text-[11px]" />
                </button>
                <button
                  type="button"
                  class="w-6 h-6 rounded-md bg-white/85 text-red-600"
                  :title="t('task.remove_photo')"
                  @click="removePhoto(p.uid)"
                >
                  <i class="fas fa-trash text-[11px]" />
                </button>
                <button
                  type="button"
                  class="w-6 h-6 rounded-md bg-white/85 text-ink disabled:opacity-30"
                  :disabled="idx === photoItems.length - 1"
                  :title="t('task.move_right')"
                  @click="movePhoto(idx, 1)"
                >
                  <i class="fas fa-chevron-right text-[11px]" />
                </button>
              </div>
            </div>
          </div>
          <p v-else class="text-xs text-ink-soft">{{ t("task.no_photos") }}</p>
        </div>

        <!-- upload progress -->
        <div v-if="saving" class="pt-1">
          <div class="flex justify-between text-xs text-ink-medium mb-1">
            <span>{{ uploadPct < 100 ? t("task.uploading") : t("task.saving") }}</span>
            <span class="font-semibold text-brand">{{ uploadPct }}%</span>
          </div>
          <div class="h-2 rounded-full bg-page overflow-hidden">
            <div class="h-full bg-brand transition-all duration-150" :style="{ width: uploadPct + '%' }" />
          </div>
        </div>
      </div>

      <template #footer>
        <button class="btn-outline" :disabled="saving" @click="open = false">{{ t("app.cancel") }}</button>
        <button class="btn-primary" :disabled="saving" @click="save">
          <i v-if="saving" class="fas fa-spinner fa-spin mr-1.5" />
          {{ saving ? t("task.saving") : t("app.save") }}
        </button>
      </template>
    </UiModal>

    <UiConfirmDialog v-model="confirmDel.open" danger :message="t('app.delete') + '?'" @confirm="doDelete" />
  </div>
</template>
