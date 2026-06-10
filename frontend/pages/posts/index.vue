<script setup lang="ts">
import type { Post } from "~/types";
import { formatNumber, formatDate } from "~/utils/format";

useHead({ title: () => useI18n().t("nav.networks") });

const { t } = useI18n();
const api = useApi();
const toast = useToast();

const categories = [
  { value: "instagram", label: "Instagram", icon: "fa-instagram", cls: "platform-instagram" },
  { value: "telegram", label: "Telegram", icon: "fa-telegram", cls: "platform-telegram" },
  { value: "tiktok", label: "TikTok", icon: "fa-tiktok", cls: "platform-tiktok" },
  { value: "facebook", label: "Facebook", icon: "fa-facebook", cls: "platform-facebook" },
  { value: "youtube", label: "YouTube", icon: "fa-youtube", cls: "platform-youtube" },
  { value: "other", label: t("post.other"), icon: "fa-globe", cls: "platform-other" },
];
const contentTypes = [
  { value: "reels", label: "Reels" },
  { value: "publication", label: t("post.publication") },
  { value: "stories", label: "Stories" },
  { value: "other", label: t("post.other") },
];
const projects = [
  { value: "nav_xona", label: "Nav Xona" },
  { value: "navsoht", label: "Navsoht" },
];
const catMeta = (c?: string) => categories.find((x) => x.value === c) || categories[5];

const { data: posts, pending, refresh } = useAsyncData<Post[]>("posts", () => api.get<Post[]>("/api/posts"), { lazy: true, default: () => [] });

const filter = ref("all");
const filtered = computed(() => {
  if (filter.value === "all") return posts.value || [];
  return (posts.value || []).filter((p) => p.category === filter.value);
});

// totals per platform
const platformStats = computed(() => {
  const map: Record<string, { count: number; likes: number; views: number; reach: number }> = {};
  (posts.value || []).forEach((p) => {
    const k = p.category || "other";
    if (!map[k]) map[k] = { count: 0, likes: 0, views: 0, reach: 0 };
    map[k].count++;
    map[k].likes += p.likes || 0;
    map[k].views += p.views || 0;
    map[k].reach += p.reach || 0;
  });
  return map;
});

const modal = ref(false);
const blank = (): Partial<Post> => ({ category: "instagram", content_type: "reels", project: "nav_xona", likes: 0, comments: 0, shares: 0, views: 0, reach: 0 });
const editing = ref<Partial<Post>>(blank());
const mediaFile = ref<File | null>(null);
const openCreate = () => { editing.value = blank(); mediaFile.value = null; modal.value = true; };
const openEdit = (p: Post) => { editing.value = { ...p }; mediaFile.value = null; modal.value = true; };

const save = async () => {
  const e = editing.value;
  if (!e.title) { toast.error(t("notify.error")); return; }
  const fd = new FormData();
  const keys: (keyof Post)[] = ["title", "description", "category", "content_type", "project", "link", "likes", "comments", "shares", "views", "reach"];
  keys.forEach((k) => { if (e[k] != null) fd.append(k as string, String(e[k])); });
  if (mediaFile.value) fd.append("media", mediaFile.value);
  if (e.id) { await api.upload(`/api/posts/${e.id}`, fd); toast.success(t("notify.updated")); }
  else { await api.upload("/api/posts", fd); toast.success(t("notify.created")); }
  modal.value = false;
  await refresh();
};
const confirmDel = ref<{ open: boolean; id?: number }>({ open: false });
const askDelete = (p: Post) => (confirmDel.value = { open: true, id: p.id });
const doDelete = async () => {
  if (!confirmDel.value.id) return;
  await api.del(`/api/posts/${confirmDel.value.id}`);
  toast.success(t("notify.deleted"));
  await refresh();
};
const exportXlsx = () => api.download("/api/posts/export/excel", "posts.xlsx");
</script>

<template>
  <div>
    <div class="flex justify-between items-center flex-wrap gap-4 mb-6">
      <div>
        <h1 class="text-[28px] font-bold text-ink mb-1">{{ t("nav.networks") }}</h1>
        <p class="text-ink-soft text-sm">{{ t("post.subtitle") }}</p>
      </div>
      <div class="flex gap-2">
        <button class="btn-outline" @click="exportXlsx"><i class="fas fa-file-excel" /> Excel</button>
        <button class="btn-primary" @click="openCreate"><i class="fas fa-plus" /> {{ t("post.add") }}</button>
      </div>
    </div>

    <!-- platform stats -->
    <div class="grid grid-cols-2 md:grid-cols-3 lg:grid-cols-6 gap-3 mb-6">
      <div v-for="c in categories" :key="c.value" class="bg-white border border-border-soft rounded-2xl p-4 text-center">
        <div class="w-10 h-10 mx-auto rounded-xl flex items-center justify-center mb-2" :class="c.cls">
          <i class="fab text-lg" :class="c.icon" />
        </div>
        <div class="text-xl font-bold text-ink">{{ platformStats[c.value]?.count || 0 }}</div>
        <div class="text-[11px] text-ink-soft">{{ c.label }}</div>
      </div>
    </div>

    <!-- filter pills -->
    <div class="flex gap-2 mb-6 flex-wrap">
      <button class="px-4 py-2 rounded-full text-[13px] font-medium border-[1.5px] transition-all"
        :class="filter === 'all' ? 'bg-brand text-white border-brand' : 'bg-white border-border-soft text-ink-medium hover:border-brand'"
        @click="filter = 'all'">{{ t("app.all") }}</button>
      <button v-for="c in categories" :key="c.value"
        class="px-4 py-2 rounded-full text-[13px] font-medium border-[1.5px] transition-all"
        :class="filter === c.value ? 'bg-brand text-white border-brand' : 'bg-white border-border-soft text-ink-medium hover:border-brand'"
        @click="filter = c.value"><i class="fab mr-1" :class="c.icon" /> {{ c.label }}</button>
    </div>

    <UiLoader v-if="pending" />
    <UiEmpty v-else-if="!filtered.length" icon="fa-share-nodes" />

    <div v-else class="grid gap-5" style="grid-template-columns: repeat(auto-fill, minmax(300px, 1fr))">
      <article v-for="p in filtered" :key="p.id" class="bg-white rounded-[20px] border border-border-soft overflow-hidden transition-all hover:-translate-y-[3px] hover:shadow-hover">
        <div v-if="p.media_path" class="h-[180px] bg-page">
          <img v-if="p.media_type === 'image'" :src="p.media_path" class="w-full h-full object-cover" />
          <video v-else-if="p.media_type === 'video'" :src="p.media_path" class="w-full h-full object-cover" controls />
        </div>
        <div class="p-5">
          <div class="flex items-center justify-between mb-2">
            <span class="platform-badge inline-flex items-center gap-1.5 px-2.5 py-1 rounded-full text-[11px] font-semibold" :class="catMeta(p.category).cls">
              <i class="fab" :class="catMeta(p.category).icon" /> {{ catMeta(p.category).label }}
            </span>
            <span class="text-[11px] text-ink-soft">{{ formatDate(p.created_at) }}</span>
          </div>
          <h3 class="font-bold text-ink mb-1.5 line-clamp-2">{{ p.title }}</h3>
          <p v-if="p.description" class="text-xs text-ink-medium line-clamp-2 mb-3">{{ p.description }}</p>
          <div class="grid grid-cols-4 gap-1.5 mb-3 text-center">
            <div class="bg-page rounded-lg py-1.5"><div class="font-bold text-ink text-sm">{{ formatNumber(p.likes) }}</div><div class="text-[9px] text-ink-soft uppercase">Likes</div></div>
            <div class="bg-page rounded-lg py-1.5"><div class="font-bold text-ink text-sm">{{ formatNumber(p.views) }}</div><div class="text-[9px] text-ink-soft uppercase">Views</div></div>
            <div class="bg-page rounded-lg py-1.5"><div class="font-bold text-ink text-sm">{{ formatNumber(p.shares) }}</div><div class="text-[9px] text-ink-soft uppercase">Shares</div></div>
            <div class="bg-page rounded-lg py-1.5"><div class="font-bold text-ink text-sm">{{ formatNumber(p.reach) }}</div><div class="text-[9px] text-ink-soft uppercase">Reach</div></div>
          </div>
          <div class="flex items-center justify-between">
            <a v-if="p.link" :href="p.link" target="_blank" class="text-xs text-brand font-semibold"><i class="fas fa-link mr-1" />{{ t("post.open_link") }}</a>
            <span v-else />
            <div class="flex gap-1">
              <button class="w-8 h-8 bg-page rounded-lg text-brand hover:bg-brand-soft" @click="openEdit(p)"><i class="fas fa-pen text-xs" /></button>
              <button class="w-8 h-8 bg-page rounded-lg text-red-500 hover:bg-red-100" @click="askDelete(p)"><i class="fas fa-trash text-xs" /></button>
            </div>
          </div>
        </div>
      </article>
    </div>

    <UiModal v-model="modal" :title="editing.id ? t('app.edit') : t('post.add')" size="md">
      <div class="grid grid-cols-1 md:grid-cols-2 gap-4">
        <div><label class="label">{{ t("post.category") }} *</label><select v-model="editing.category" class="input"><option v-for="c in categories" :key="c.value" :value="c.value">{{ c.label }}</option></select></div>
        <div><label class="label">{{ t("post.content_type") }} *</label><select v-model="editing.content_type" class="input"><option v-for="c in contentTypes" :key="c.value" :value="c.value">{{ c.label }}</option></select></div>
        <div><label class="label">{{ t("post.project") }} *</label><select v-model="editing.project" class="input"><option v-for="p in projects" :key="p.value" :value="p.value">{{ p.label }}</option></select></div>
        <div><label class="label">{{ t("post.link") }}</label><input v-model="editing.link" type="url" class="input" placeholder="https://..." /></div>
        <div class="md:col-span-2"><label class="label">{{ t("post.title_field") }} *</label><input v-model="editing.title" class="input" /></div>
        <div class="md:col-span-2"><label class="label">{{ t("post.description") }}</label><textarea v-model="editing.description" rows="2" class="input" /></div>
        <div class="md:col-span-2"><label class="label">{{ t("post.media") }}</label><input type="file" accept="image/*,video/*" class="input !py-2" @change="mediaFile = ($event.target as HTMLInputElement).files?.[0] || null" /></div>
        <div><label class="label">{{ t("post.likes") }}</label><input v-model.number="editing.likes" type="number" class="input" /></div>
        <div><label class="label">{{ t("post.comments") }}</label><input v-model.number="editing.comments" type="number" class="input" /></div>
        <div><label class="label">{{ t("post.shares") }}</label><input v-model.number="editing.shares" type="number" class="input" /></div>
        <div><label class="label">{{ t("post.views") }}</label><input v-model.number="editing.views" type="number" class="input" /></div>
        <div><label class="label">{{ t("post.reach") }}</label><input v-model.number="editing.reach" type="number" class="input" /></div>
      </div>
      <template #footer>
        <button class="btn-outline" @click="modal = false">{{ t("app.cancel") }}</button>
        <button class="btn-primary" @click="save">{{ t("app.save") }}</button>
      </template>
    </UiModal>

    <UiConfirmDialog v-model="confirmDel.open" danger :message="t('app.delete') + '?'" @confirm="doDelete" />
  </div>
</template>

<style scoped>
.platform-instagram { @apply bg-pink-50 text-pink-600; }
.platform-telegram { @apply bg-sky-50 text-sky-600; }
.platform-tiktok { @apply bg-gray-100 text-gray-900; }
.platform-facebook { @apply bg-indigo-50 text-indigo-600; }
.platform-youtube { @apply bg-red-50 text-red-600; }
.platform-other { @apply bg-page text-ink-medium; }
</style>
