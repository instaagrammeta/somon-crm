<script setup lang="ts">
import type { User } from "~/types";
import type { FormField } from "~/components/EntityForm.vue";

useHead({ title: () => useI18n().t("nav.users") });

const api = useApi();
const toast = useToast();
const { t } = useI18n();

const { data: users, pending, refresh } = useAsyncData<User[]>(
  "users",
  () => api.get<User[]>("/api/users")
);

const open = ref(false);
const editing = ref<Partial<User & { password: string }>>({});
const tgInfo = ref<{ deep_link?: string; code?: string } | null>(null);
const confirm = ref<{ open: boolean; id?: number }>({ open: false });

const fields = computed<FormField[]>(() => [
  { key: "full_name", label: t("user.full_name"), required: true, cols: 2 },
  { key: "login", label: t("user.login"), required: true },
  { key: "password", label: t("user.password"), type: "password", placeholder: "••••" },
  { key: "age", label: t("user.age"), type: "number" },
  { key: "category", label: t("user.category") },
  {
    key: "role",
    label: t("user.role"),
    type: "select",
    options: [
      { value: "admin", label: t("user.admin") },
      { value: "manager", label: t("user.manager") },
      { value: "employee", label: t("user.employee") },
    ],
  },
  { key: "telegram_username", label: t("user.telegram_username"), placeholder: "@username" },
]);

const openCreate = () => {
  editing.value = { role: "employee" };
  tgInfo.value = null;
  open.value = true;
};

const openEdit = (u: User) => {
  editing.value = { ...u, password: "" };
  tgInfo.value = null;
  open.value = true;
};

const save = async () => {
  try {
    if (editing.value.id) {
      await api.put(`/api/users/${editing.value.id}`, editing.value);
      toast.success(t("notify.updated"));
    } else {
      const res = await api.post<{ user: User; telegram_link_code?: string; telegram_deep_link?: string }>(
        "/api/users",
        editing.value
      );
      toast.success(t("notify.created"));
      if (res.telegram_link_code) {
        tgInfo.value = { code: res.telegram_link_code, deep_link: res.telegram_deep_link };
        await refresh();
        return; // keep modal open to show the link
      }
    }
    open.value = false;
    await refresh();
  } catch (e: any) {
    toast.error(e?.data?.error || t("notify.error"));
  }
};

const askDelete = (u: User) => (confirm.value = { open: true, id: u.id });
const doDelete = async () => {
  if (!confirm.value.id) return;
  try {
    await api.del(`/api/users/${confirm.value.id}`);
    toast.success(t("notify.deleted"));
    await refresh();
  } catch (e: any) {
    toast.error(e?.data?.error || t("notify.error"));
  }
};

const copyLink = async () => {
  if (!tgInfo.value?.deep_link) return;
  await navigator.clipboard.writeText(tgInfo.value.deep_link);
  toast.success(t("telegram.code_copied"));
};
</script>

<template>
  <div class="card p-6 lg:p-8">
    <PageHeader :title="$t('nav.users')" :subtitle="$t('user.create_user')">
      <button class="btn-primary" @click="openCreate">
        <i class="fa-solid fa-plus" /> {{ $t("user.create_user") }}
      </button>
    </PageHeader>

    <UiLoader v-if="pending" />
    <UiEmpty v-else-if="!users?.length" icon="fa-users" />

    <div v-else class="overflow-x-auto -mx-6 lg:-mx-8">
      <table class="table-modern">
        <thead>
          <tr>
            <th>{{ $t("user.full_name") }}</th>
            <th>{{ $t("user.login") }}</th>
            <th>{{ $t("user.role") }}</th>
            <th>{{ $t("user.category") }}</th>
            <th>Telegram</th>
            <th class="text-right">{{ $t("app.actions") }}</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="u in users" :key="u.id">
            <td>
              <div class="flex items-center gap-3">
                <UiAvatar :src="u.photo" :name="u.full_name" />
                <div>
                  <div class="font-semibold text-ink">{{ u.full_name }}</div>
                  <div class="text-xs text-ink-soft">{{ u.personal_phones?.[0] }}</div>
                </div>
              </div>
            </td>
            <td class="font-mono text-xs">{{ u.login }}</td>
            <td>
              <span
                :class="
                  u.role === 'admin'
                    ? 'badge-green'
                    : u.role === 'manager'
                    ? 'badge-blue'
                    : 'badge-gray'
                "
              >
                {{ $t("user." + u.role) }}
              </span>
            </td>
            <td>{{ u.category || "—" }}</td>
            <td>
              <span v-if="u.has_telegram" class="badge-green">
                <i class="fa-brands fa-telegram" /> @{{ u.telegram_username }}
              </span>
              <span v-else class="badge-gray">—</span>
            </td>
            <td class="text-right whitespace-nowrap">
              <button class="btn-ghost !p-2" @click="openEdit(u)">
                <i class="fa-solid fa-pen" />
              </button>
              <button class="btn-ghost !p-2 text-red-500" @click="askDelete(u)">
                <i class="fa-solid fa-trash" />
              </button>
            </td>
          </tr>
        </tbody>
      </table>
    </div>

    <UiModal
      v-model="open"
      :title="editing.id ? $t('app.edit') : $t('user.create_user')"
      size="lg"
    >
      <EntityForm v-model="editing" :fields="fields" />

      <div v-if="tgInfo" class="mt-6 bg-brand-soft rounded-2xl p-5">
        <div class="font-bold text-ink mb-2 flex items-center gap-2">
          <i class="fa-brands fa-telegram text-brand" />
          {{ $t("user.telegram_link") }}
        </div>
        <p class="text-sm text-ink-medium mb-3">
          {{ $t("user.telegram_link_help") }}
        </p>
        <div class="bg-white rounded-xl px-4 py-3 mb-3 font-mono text-brand select-all break-all">
          {{ tgInfo.deep_link }}
        </div>
        <button class="btn-primary w-full" @click="copyLink">
          <i class="fa-solid fa-copy" /> {{ $t("telegram.copy_link") }}
        </button>
      </div>

      <template #footer>
        <button class="btn-outline" @click="open = false">
          {{ $t("app.cancel") }}
        </button>
        <button class="btn-primary" @click="save">{{ $t("app.save") }}</button>
      </template>
    </UiModal>

    <UiConfirmDialog
      v-model="confirm.open"
      :title="$t('app.confirm')"
      :message="$t('app.delete') + '?'"
      danger
      @confirm="doDelete"
    />
  </div>
</template>
