<script setup lang="ts">
import type { Post } from "~/types";
import type { FormField } from "~/components/EntityForm.vue";
import { formatDate } from "~/utils/format";

useHead({ title: () => useI18n().t("nav.posts") });

const { t } = useI18n();
const c = useCrud<Post>("/api/posts", "posts");
const api = useApi();

const fields = computed<FormField[]>(() => [
  { key: "title", label: t("post.title_field"), required: true, cols: 2 },
  { key: "description", label: t("post.description"), type: "textarea", cols: 2 },
  { key: "category", label: t("post.category") },
  { key: "project", label: t("post.project") },
  { key: "likes", label: t("post.likes"), type: "number" },
  { key: "comments", label: t("post.comments"), type: "number" },
  { key: "shares", label: t("post.shares"), type: "number" },
  { key: "views", label: t("post.views"), type: "number" },
  { key: "reach", label: t("post.reach"), type: "number" },
]);

const exportXlsx = () => api.download("/api/posts/export/excel", "posts.xlsx");
</script>

<template>
  <div class="card p-6 lg:p-8">
    <PageHeader :title="$t('nav.posts')">
      <button class="btn-outline" @click="exportXlsx">
        <i class="fa-solid fa-file-excel" /> Excel
      </button>
      <button class="btn-primary" @click="c.openCreate({})">
        <i class="fa-solid fa-plus" /> {{ $t("app.create") }}
      </button>
    </PageHeader>

    <UiLoader v-if="c.pending.value" />
    <UiEmpty v-else-if="!c.rows.value?.length" icon="fa-image" />

    <div v-else class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
      <article v-for="p in c.rows.value" :key="p.id" class="card p-5 hover:-translate-y-0.5 hover:shadow-hover transition-all">
        <h3 class="font-bold text-ink mb-2 line-clamp-2">{{ p.title }}</h3>
        <p v-if="p.description" class="text-sm text-ink-medium line-clamp-3 mb-3">{{ p.description }}</p>
        <div class="flex flex-wrap gap-2 mb-4">
          <span v-if="p.category" class="badge-blue">{{ p.category }}</span>
          <span v-if="p.project" class="badge-green">{{ p.project }}</span>
        </div>
        <div class="grid grid-cols-3 gap-2 mb-4 text-center">
          <div class="bg-page rounded-xl p-2">
            <div class="font-bold text-ink">{{ p.likes }}</div>
            <div class="text-[10px] text-ink-soft uppercase">Likes</div>
          </div>
          <div class="bg-page rounded-xl p-2">
            <div class="font-bold text-ink">{{ p.views }}</div>
            <div class="text-[10px] text-ink-soft uppercase">Views</div>
          </div>
          <div class="bg-page rounded-xl p-2">
            <div class="font-bold text-ink">{{ p.shares }}</div>
            <div class="text-[10px] text-ink-soft uppercase">Shares</div>
          </div>
        </div>
        <div class="flex items-center justify-between text-xs text-ink-soft">
          <span>{{ formatDate(p.created_at) }}</span>
          <div>
            <button class="btn-ghost !p-1.5" @click="c.openEdit(p)">
              <i class="fa-solid fa-pen" />
            </button>
            <button class="btn-ghost !p-1.5 text-red-500" @click="c.askDelete(p)">
              <i class="fa-solid fa-trash" />
            </button>
          </div>
        </div>
      </article>
    </div>

    <UiModal v-model="c.open.value" :title="c.editing.value.id ? $t('app.edit') : $t('app.create')" size="lg">
      <EntityForm v-model="c.editing.value" :fields="fields" />
      <template #footer>
        <button class="btn-outline" @click="c.open.value = false">{{ $t('app.cancel') }}</button>
        <button class="btn-primary" @click="c.save">{{ $t('app.save') }}</button>
      </template>
    </UiModal>

    <UiConfirmDialog v-model="c.confirmState.value.open" danger :message="$t('app.delete') + '?'" @confirm="c.doDelete" />
  </div>
</template>
