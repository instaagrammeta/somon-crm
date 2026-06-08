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
  { watch: [status], default: () => [] }
);
const { data: users } = useAsyncData<User[]>("users-list", () => api.get<User[]>("/api/users/list"), { default: () => [] });

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

// modal
const open = ref(false);
const editing = ref<Partial<Task>>({});
const photoFile = ref<File | null>(null);
const photoPreview = ref<string>("");

const openCreate = () => {
  editing.value = { status: "new" };
  photoFile.value = null;
  photoPreview.value = "";
  open.value = true;
};
const openEdit = (tk: Task) => {
  editing.value = { ...tk };
  photoFile.value = null;
  photoPreview.value = tk.photo || "";
  open.value = true;
};
const onPhoto = (e: Event) => {
  const f = (e.target as HTMLInputElement).files?.[0];
  if (f) {
    photoFile.value = f;
    photoPreview.value = URL.createObjectURL(f);
  }
};

const save = async () => {
  if (!editing.value.title) {
    toast.error(t("notify.error"));
    return;
  }
  const fd = new FormData();
  fd.append("title", editing.value.title || "");
  fd.append("description", editing.value.description || "");
  if (editing.value.executor_id) fd.append("executor_id", String(editing.value.executor_id));
  if (editing.value.status) fd.append("status", editing.value.status);
  if (photoFile.value) fd.append("photo", photoFile.value);
  try {
    if (editing.value.id) {
      await api.upload(`/api/tasks/${editing.value.id}`, fd);
      toast.success(t("notify.updated"));
    } else {
      await api.upload("/api/tasks", fd);
      toast.success(t("notify.created"));
    }
    open.value = false;
    await refresh();
  } catch (e: any) {
    toast.error(e?.data?.error || t("notify.error"));
  }
};

const changeStatus = async (tk: Task, newStatus: string) => {
  const fd = new FormData();
  fd.append("status", newStatus);
  await api.upload(`/api/tasks/${tk.id}`, fd);
  await refresh();
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

        <img v-if="tk.photo" :src="tk.photo" class="w-full rounded-[14px] mb-3 max-h-[180px] object-cover" />

        <p v-if="tk.description" class="text-ink-medium text-[13px] leading-relaxed mb-3.5">{{ tk.description }}</p>

        <div class="bg-brand-soft p-3 rounded-[14px] my-3 text-xs text-ink-medium flex flex-col gap-1.5">
          <div v-if="tk.executor_name" class="flex items-center gap-2">
            <i class="fas fa-user-check w-4 text-brand" /> {{ tk.executor_name }}
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
    <UiModal v-model="open" :title="editing.id ? t('app.edit') : t('task.create_task')">
      <div class="space-y-4">
        <div>
          <label class="label"><i class="fas fa-heading text-brand mr-1.5" /> {{ t("task.title_field") }} *</label>
          <input v-model="editing.title" class="input" required />
        </div>
        <div>
          <label class="label"><i class="fas fa-align-left text-brand mr-1.5" /> {{ t("task.description") }}</label>
          <textarea v-model="editing.description" rows="3" class="input" />
        </div>
        <div>
          <label class="label"><i class="fas fa-user-check text-brand mr-1.5" /> {{ t("task.executor") }}</label>
          <select v-model="editing.executor_id" class="input">
            <option :value="undefined">—</option>
            <option v-for="u in users" :key="u.id" :value="u.id">{{ u.full_name }}</option>
          </select>
        </div>
        <div>
          <label class="label"><i class="fas fa-image text-brand mr-1.5" /> {{ t("task.photo") }}</label>
          <label class="inline-flex items-center gap-2 px-4 py-2.5 bg-brand-soft text-brand rounded-[10px] cursor-pointer text-[13px] font-semibold hover:bg-brand-light transition-colors">
            <i class="fas fa-upload" /> {{ t("task.choose_photo") }}
            <input type="file" accept="image/*" hidden @change="onPhoto" />
          </label>
          <img v-if="photoPreview" :src="photoPreview" class="mt-2 max-w-[100px] rounded-[10px]" />
        </div>
      </div>
      <template #footer>
        <button class="btn-outline" @click="open = false">{{ t("app.cancel") }}</button>
        <button class="btn-primary" @click="save">{{ t("app.save") }}</button>
      </template>
    </UiModal>

    <UiConfirmDialog v-model="confirmDel.open" danger :message="t('app.delete') + '?'" @confirm="doDelete" />
  </div>
</template>
