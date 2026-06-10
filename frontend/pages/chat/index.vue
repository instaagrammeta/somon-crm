<script setup lang="ts">
import type { Message } from "~/types";
import { formatDateTime } from "~/utils/format";

useHead({ title: () => useI18n().t("nav.chat") });
const api = useApi();
const auth = useAuthStore();
const { t } = useI18n();
const toast = useToast();

const { data: messages, refresh } = useAsyncData<Message[]>("messages",
  () => api.get<Message[]>("/api/messages?limit=200"), { lazy: true, default: () => [] });

const text = ref("");
const fileInput = ref<HTMLInputElement | null>(null);
const file = ref<File | null>(null);

const send = async () => {
  if (!text.value.trim() && !file.value) return;
  const fd = new FormData();
  fd.append("message", text.value);
  if (file.value) fd.append("file", file.value);
  await api.upload("/api/messages", fd);
  text.value = "";
  file.value = null;
  if (fileInput.value) fileInput.value.value = "";
  await refresh();
  await nextTick();
  scrollBottom();
};

const remove = async (id: number) => {
  if (!confirm(t("app.delete") + "?")) return;
  await api.del(`/api/messages/${id}`);
  await refresh();
};

const chatRef = ref<HTMLDivElement | null>(null);
const scrollBottom = () => { if (chatRef.value) chatRef.value.scrollTop = chatRef.value.scrollHeight; };
onMounted(() => setTimeout(scrollBottom, 100));

// Auto refresh every 5s (light polling)
let timer: any = null;
onMounted(() => { timer = setInterval(refresh, 5000); });
onUnmounted(() => clearInterval(timer));
</script>

<template>
  <div class="card p-6 lg:p-8 flex flex-col" style="height: calc(100vh - 130px)">
    <PageHeader :title="$t('nav.chat')" />

    <div ref="chatRef" class="flex-1 overflow-y-auto space-y-3 pr-2">
      <div
        v-for="m in messages"
        :key="m.id"
        :class="m.user_id === auth.user?.id ? 'flex flex-row-reverse gap-3' : 'flex gap-3'"
      >
        <UiAvatar :name="m.user_name" />
        <div class="max-w-[70%]">
          <div class="text-xs text-ink-soft mb-1" :class="m.user_id === auth.user?.id ? 'text-right' : ''">
            {{ m.user_name }} · {{ formatDateTime(m.created_at) }}
          </div>
          <div
            class="rounded-2xl px-4 py-2.5 text-sm relative group"
            :class="m.user_id === auth.user?.id ? 'bg-brand text-white' : 'bg-page text-ink'"
          >
            <span v-if="m.message" class="whitespace-pre-wrap break-words">{{ m.message }}</span>
            <a
              v-if="m.file_path"
              :href="m.file_path"
              target="_blank"
              class="block mt-1 underline text-xs"
            >
              📎 {{ m.file_name }}
            </a>
            <button
              v-if="m.user_id === auth.user?.id"
              class="opacity-0 group-hover:opacity-100 absolute -top-2 -right-2 bg-red-500 text-white w-6 h-6 rounded-full text-[10px] transition-opacity"
              @click="remove(m.id)"
            >
              <i class="fa-solid fa-xmark" />
            </button>
          </div>
        </div>
      </div>
    </div>

    <form class="mt-4 flex items-center gap-2" @submit.prevent="send">
      <label class="btn-ghost !p-3 cursor-pointer">
        <i class="fa-solid fa-paperclip" />
        <input ref="fileInput" type="file" hidden @change="file = ($event.target as HTMLInputElement).files?.[0] || null" />
      </label>
      <input
        v-model="text"
        :placeholder="$t('app.search')"
        class="input flex-1"
      />
      <button class="btn-primary !px-5" :disabled="!text.trim() && !file">
        <i class="fa-solid fa-paper-plane" />
      </button>
    </form>
    <div v-if="file" class="text-xs text-ink-soft mt-2">📎 {{ file.name }}</div>
  </div>
</template>
