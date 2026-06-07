<script setup lang="ts">
import type { Task, User } from "~/types";
import type { FormField } from "~/components/EntityForm.vue";

useHead({ title: () => useI18n().t("nav.tasks") });

const api = useApi();
const toast = useToast();
const { t } = useI18n();

const { data: tasks, pending, refresh } = useAsyncData<Task[]>(
  "tasks",
  () => api.get<Task[]>("/api/tasks")
);
const { data: users } = useAsyncData<User[]>(
  "users-list",
  () => api.get<User[]>("/api/users/list")
);

const filter = ref("");
const filtered = computed(() => {
  if (!tasks.value) return [];
  if (!filter.value) return tasks.value;
  return tasks.value.filter((x) => !filter.value || x.status === filter.value);
});

const open = ref(false);
const editing = ref<Partial<Task>>({});
const confirm = ref<{ open: boolean; id?: number }>({ open: false });

const fields = computed<FormField[]>(() => [
  { key: "title", label: t("task.title_field"), required: true, cols: 2 },
  { key: "description", label: t("task.description"), type: "textarea", cols: 2 },
  {
    key: "executor_id",
    label: t("task.executor"),
    type: "select",
    options: (users.value || []).map((u) => ({ value: u.id, label: u.full_name })),
  },
  {
    key: "status",
    label: t("task.status"),
    type: "select",
    options: [
      { value: "new", label: t("task.status_new") },
      { value: "in_progress", label: t("task.status_in_progress") },
      { value: "done", label: t("task.status_done") },
      { value: "cancelled", label: t("task.status_cancelled") },
    ],
  },
]);

const statusBadge = (s: string) =>
  ({
    new: "badge-blue",
    in_progress: "badge-yellow",
    done: "badge-green",
    cancelled: "badge-red",
  })[s] || "badge-gray";

const openCreate = () => {
  editing.value = { status: "new" };
  open.value = true;
};
const openEdit = (t: Task) => {
  editing.value = { ...t };
  open.value = true;
};

const save = async () => {
  try {
    if (editing.value.id) {
      await api.put(`/api/tasks/${editing.value.id}`, editing.value);
      toast.success(t("notify.updated"));
    } else {
      await api.post("/api/tasks", editing.value);
      toast.success(t("notify.created"));
    }
    open.value = false;
    await refresh();
  } catch (e: any) {
    toast.error(e?.data?.error || t("notify.error"));
  }
};

const askDelete = (x: Task) => (confirm.value = { open: true, id: x.id });
const doDelete = async () => {
  if (!confirm.value.id) return;
  await api.del(`/api/tasks/${confirm.value.id}`);
  toast.success(t("notify.deleted"));
  await refresh();
};
</script>

<template>
  <div class="card p-6 lg:p-8">
    <PageHeader :title="$t('nav.tasks')" :subtitle="$t('task.create_task')">
      <select v-model="filter" class="input !py-2 max-w-[180px]">
        <option value="">{{ $t("app.all") }}</option>
        <option value="new">{{ $t("task.status_new") }}</option>
        <option value="in_progress">{{ $t("task.status_in_progress") }}</option>
        <option value="done">{{ $t("task.status_done") }}</option>
        <option value="cancelled">{{ $t("task.status_cancelled") }}</option>
      </select>
      <button class="btn-primary" @click="openCreate">
        <i class="fa-solid fa-plus" /> {{ $t("task.create_task") }}
      </button>
    </PageHeader>

    <UiLoader v-if="pending" />
    <UiEmpty v-else-if="!filtered.length" icon="fa-list-check" />

    <div v-else class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
      <article
        v-for="x in filtered"
        :key="x.id"
        class="card p-5 hover:-translate-y-0.5 hover:shadow-hover transition-all"
      >
        <div class="flex items-start justify-between gap-2 mb-3">
          <h3 class="font-bold text-ink leading-snug flex-1">{{ x.title }}</h3>
          <span :class="statusBadge(x.status)">
            {{ $t("task.status_" + x.status) }}
          </span>
        </div>
        <p
          v-if="x.description"
          class="text-sm text-ink-medium mb-4 line-clamp-3"
        >
          {{ x.description }}
        </p>
        <div class="flex items-center justify-between text-xs text-ink-soft">
          <div v-if="x.executor_name" class="flex items-center gap-1.5">
            <i class="fa-solid fa-user" />
            {{ x.executor_name }}
          </div>
          <div class="flex items-center gap-1">
            <button class="btn-ghost !p-2" @click="openEdit(x)">
              <i class="fa-solid fa-pen" />
            </button>
            <button class="btn-ghost !p-2 text-red-500" @click="askDelete(x)">
              <i class="fa-solid fa-trash" />
            </button>
          </div>
        </div>
      </article>
    </div>

    <UiModal
      v-model="open"
      :title="editing.id ? $t('app.edit') : $t('task.create_task')"
    >
      <EntityForm v-model="editing" :fields="fields" />
      <template #footer>
        <button class="btn-outline" @click="open = false">
          {{ $t("app.cancel") }}
        </button>
        <button class="btn-primary" @click="save">{{ $t("app.save") }}</button>
      </template>
    </UiModal>

    <UiConfirmDialog
      v-model="confirm.open"
      :message="$t('app.delete') + '?'"
      danger
      @confirm="doDelete"
    />
  </div>
</template>
