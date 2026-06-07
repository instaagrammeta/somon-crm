<script setup lang="ts">
useHead({ title: () => useI18n().t("telegram.title") });
const api = useApi();
const auth = useAuthStore();
const toast = useToast();
const { t } = useI18n();

const linkData = ref<{
  code?: string;
  deep_link?: string;
  bot_username?: string;
  valid_minutes?: number;
}>({});
const generating = ref(false);

const generate = async () => {
  generating.value = true;
  try {
    linkData.value = await api.post("/api/me/telegram-link");
  } catch (e: any) {
    toast.error(e?.data?.error || "Telegram бот фаъол нест");
  } finally {
    generating.value = false;
  }
};

const copy = async () => {
  if (!linkData.value.deep_link) return;
  await navigator.clipboard.writeText(linkData.value.deep_link);
  toast.success(t("telegram.code_copied"));
};

const unlink = async () => {
  await api.del("/api/me/telegram-link");
  await auth.fetchMe();
  toast.success(t("notify.updated"));
};
</script>

<template>
  <div class="card p-6 lg:p-10 max-w-3xl mx-auto">
    <PageHeader :title="$t('telegram.title')" :subtitle="$t('telegram.subtitle')" />

    <!-- Status block -->
    <div
      class="rounded-2xl p-5 mb-6 flex items-start gap-4"
      :class="auth.user?.has_telegram ? 'bg-brand-soft' : 'bg-page'"
    >
      <div
        class="w-14 h-14 rounded-2xl flex items-center justify-center text-2xl"
        :class="auth.user?.has_telegram ? 'bg-brand text-white' : 'bg-white text-ink-soft'"
      >
        <i class="fa-solid fa-paper-plane" />
      </div>
      <div class="flex-1">
        <div v-if="auth.user?.has_telegram" class="font-semibold text-ink mb-1">
          {{ $t("telegram.linked", { username: auth.user?.telegram_username || "?" }) }}
        </div>
        <div v-else class="font-semibold text-ink mb-1">
          {{ $t("telegram.not_linked") }}
        </div>
        <p class="text-sm text-ink-soft">{{ $t("telegram.instructions") }}</p>
      </div>
      <button
        v-if="auth.user?.has_telegram"
        class="btn-danger"
        @click="unlink"
      >
        {{ $t("telegram.unlink") }}
      </button>
    </div>

    <!-- Generate code -->
    <div v-if="!linkData.code" class="text-center py-6">
      <button class="btn-primary !py-3 !px-8" :disabled="generating" @click="generate">
        <i v-if="generating" class="fa-solid fa-spinner animate-spin" />
        <i v-else class="fa-solid fa-key" />
        {{ $t("telegram.generate_code") }}
      </button>
    </div>

    <div v-else class="bg-page rounded-2xl p-6">
      <div class="text-xs uppercase tracking-wider text-ink-soft font-bold mb-2">
        Рамзи пайвасткунӣ
      </div>
      <div
        class="text-3xl font-mono font-bold text-brand mb-4 tracking-widest select-all break-all"
      >
        {{ linkData.code }}
      </div>

      <div class="text-xs uppercase tracking-wider text-ink-soft font-bold mb-2">
        Ҳаволаи мустақим
      </div>
      <div class="flex items-center gap-2 bg-white border border-border-soft rounded-xl px-4 py-3 mb-4">
        <input
          readonly
          :value="linkData.deep_link"
          class="bg-transparent outline-none text-sm flex-1 font-mono text-ink-medium select-all"
        />
        <button class="btn-secondary !py-2" @click="copy">
          <i class="fa-solid fa-copy" /> Нусха
        </button>
      </div>

      <a
        v-if="linkData.deep_link"
        :href="linkData.deep_link"
        target="_blank"
        rel="noopener"
        class="btn-primary w-full !py-3"
      >
        <i class="fa-brands fa-telegram" />
        Кушодан дар Telegram
      </a>

      <p class="text-xs text-ink-soft text-center mt-4">
        Рамз барои {{ linkData.valid_minutes || 15 }} дақиқа эътибор дорад
      </p>
    </div>
  </div>
</template>
