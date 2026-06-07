<script setup lang="ts">
import type { Folder, FolderFile } from "~/types";
import { formatBytes } from "~/utils/format";

useHead({ title: () => useI18n().t("nav.folders") });

const { t } = useI18n();
const api = useApi();
const toast = useToast();

const parentId = ref(0);
const { data: folders, refresh: refreshFolders } = useAsyncData<Folder[]>(
  () => `folders-${parentId.value}`,
  () => api.get<Folder[]>(`/api/folders?parent_id=${parentId.value}`),
  { watch: [parentId] }
);
const breadcrumbs = ref<Folder[]>([]);

const openFolder = (f: Folder) => {
  breadcrumbs.value.push(f);
  parentId.value = f.id;
  selectedFolder.value = f;
};
const goRoot = () => {
  breadcrumbs.value = [];
  parentId.value = 0;
  selectedFolder.value = null;
};

const selectedFolder = ref<Folder | null>(null);
const { data: files, refresh: refreshFiles } = useAsyncData<FolderFile[]>(
  () => `files-${selectedFolder.value?.id}`,
  () => selectedFolder.value
    ? api.get<FolderFile[]>(`/api/folders/${selectedFolder.value.id}/files`)
    : Promise.resolve([]),
  { watch: [selectedFolder] }
);

const newFolderName = ref("");
const createFolder = async () => {
  if (!newFolderName.value.trim()) return;
  await api.post("/api/folders", { name: newFolderName.value, parent_id: parentId.value });
  newFolderName.value = "";
  toast.success(t("notify.created"));
  await refreshFolders();
};
const deleteFolder = async (f: Folder) => {
  if (!confirm(t("app.delete") + "?")) return;
  await api.del(`/api/folders/${f.id}`);
  await refreshFolders();
};

const fileInput = ref<HTMLInputElement | null>(null);
const onFile = async (e: Event) => {
  const files = (e.target as HTMLInputElement).files;
  if (!files?.length || !selectedFolder.value) return;
  const fd = new FormData();
  fd.append("file", files[0]);
  await api.upload(`/api/folders/${selectedFolder.value.id}/files`, fd);
  toast.success(t("notify.created"));
  if (fileInput.value) fileInput.value.value = "";
  await refreshFiles();
};
const deleteFile = async (f: FolderFile) => {
  if (!confirm(t("app.delete") + "?")) return;
  await api.del(`/api/files/${f.id}`);
  await refreshFiles();
};
</script>

<template>
  <div class="card p-6 lg:p-8">
    <PageHeader :title="$t('nav.folders')">
      <input v-model="newFolderName" class="input !py-2" :placeholder="$t('app.create')" @keyup.enter="createFolder" />
      <button class="btn-primary" @click="createFolder">
        <i class="fa-solid fa-folder-plus" /> {{ $t("app.create") }}
      </button>
    </PageHeader>

    <!-- Breadcrumbs -->
    <nav class="flex items-center gap-2 text-sm mb-6 text-ink-soft">
      <button class="hover:text-brand transition-colors" @click="goRoot">
        <i class="fa-solid fa-house" /> Root
      </button>
      <template v-for="(b, i) in breadcrumbs" :key="b.id">
        <i class="fa-solid fa-chevron-right text-[10px]" />
        <button
          class="hover:text-brand transition-colors"
          :class="{ 'text-ink font-semibold': i === breadcrumbs.length - 1 }"
          @click="breadcrumbs = breadcrumbs.slice(0, i + 1); parentId = b.id; selectedFolder = b"
        >
          {{ b.name }}
        </button>
      </template>
    </nav>

    <div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-3">
      <div
        v-for="f in folders"
        :key="f.id"
        class="card-hover p-4 cursor-pointer group"
        @click="openFolder(f)"
      >
        <div class="flex items-center gap-3">
          <i class="fa-solid fa-folder text-2xl text-brand" />
          <div class="flex-1 min-w-0">
            <div class="font-semibold text-ink truncate">{{ f.name }}</div>
            <div class="text-xs text-ink-soft truncate">{{ f.author_name }}</div>
          </div>
          <button class="btn-ghost !p-1.5 opacity-0 group-hover:opacity-100 text-red-500" @click.stop="deleteFolder(f)">
            <i class="fa-solid fa-trash" />
          </button>
        </div>
      </div>
    </div>

    <!-- Files inside selected folder -->
    <section v-if="selectedFolder" class="mt-8">
      <div class="flex items-center justify-between mb-4">
        <h3 class="font-bold text-ink">
          <i class="fa-solid fa-folder-open text-brand mr-2" />{{ selectedFolder.name }}
        </h3>
        <label class="btn-primary cursor-pointer">
          <i class="fa-solid fa-upload" /> Upload
          <input ref="fileInput" type="file" hidden @change="onFile" />
        </label>
      </div>

      <UiEmpty v-if="!files?.length" icon="fa-file" />

      <div v-else class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-3">
        <div v-for="ff in files" :key="ff.id" class="card p-4 flex items-center gap-3">
          <div class="w-10 h-10 rounded-xl bg-brand-soft text-brand flex items-center justify-center">
            <i class="fa-solid fa-file" />
          </div>
          <div class="flex-1 min-w-0">
            <div class="text-sm font-semibold text-ink truncate">{{ ff.original_name }}</div>
            <div class="text-xs text-ink-soft">{{ formatBytes(ff.filesize) }} · {{ ff.filetype }}</div>
          </div>
          <a :href="`/api/download/${ff.id}`" class="btn-ghost !p-2">
            <i class="fa-solid fa-download" />
          </a>
          <button class="btn-ghost !p-2 text-red-500" @click="deleteFile(ff)">
            <i class="fa-solid fa-trash" />
          </button>
        </div>
      </div>
    </section>
  </div>
</template>
