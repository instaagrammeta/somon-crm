<script setup lang="ts">
import type { Notification } from "~/types";
import { formatDateTime } from "~/utils/format";

useHead({ title: () => useI18n().t("nav.notifications") });
const { t } = useI18n();
const auth = useAuthStore();
const notif = useNotificationsStore();

// Fetch in onMounted (never via useAsyncData) so a cold reload can't turn a
// transient API error into a fatal 500 page.
onMounted(() => notif.fetch());

const icon = (type: string) => {
  switch (type) {
    case "task_assigned":
    case "task_status":
      return { i: "fa-tasks", c: "text-brand bg-brand-soft" };
    case "request_assigned":
      return { i: "fa-file-invoice", c: "text-blue-500 bg-blue-50" };
    case "lead_moved":
      return { i: "fa-fire", c: "text-orange-500 bg-orange-50" };
    case "tariff_expiring":
      return { i: "fa-sim-card", c: "text-purple-500 bg-purple-50" };
    default:
      return { i: "fa-bell", c: "text-ink-medium bg-page" };
  }
};

const onOpen = async (n: Notification) => {
  if (!n.is_read) await notif.markRead(n.id);
  if (n.link) navigateTo(n.link);
};
</script>

<template>
  <div>
    <div class="flex items-start justify-between flex-wrap gap-3 mb-6">
      <div>
        <h1 class="text-[28px] font-bold text-ink mb-1">{{ t("nav.notifications") }}</h1>
        <p class="text-ink-soft text-sm">{{ t("evedem.subtitle") }}</p>
      </div>
      <button
        v-if="notif.unread > 0"
        class="btn-outline !py-2 !px-4"
        @click="notif.markAll()"
      >
        <i class="fas fa-check-double" /> {{ t("evedem.mark_all_read") }}
      </button>
    </div>

    <!-- Telegram notification channel -->
    <div class="bg-white border border-border-soft rounded-2xl p-6 mb-4 flex items-start gap-4">
      <div class="w-14 h-14 rounded-2xl bg-sky-50 text-sky-500 flex items-center justify-center text-2xl shrink-0">
        <i class="fab fa-telegram" />
      </div>
      <div class="flex-1">
        <h3 class="font-bold text-ink mb-1">{{ t("evedem.telegram_title") }}</h3>
        <p class="text-sm text-ink-medium mb-3">{{ t("evedem.telegram_desc") }}</p>
        <div v-if="auth.user?.has_telegram" class="badge-green">
          <i class="fas fa-check-circle" /> {{ t("telegram.linked", { username: "@" + (auth.user?.telegram_username || "?") }) }}
        </div>
        <NuxtLink v-else to="/me/telegram" class="btn-primary !py-2 !px-4 inline-flex">
          <i class="fab fa-telegram" /> {{ t("evedem.connect") }}
        </NuxtLink>
      </div>
    </div>

    <!-- In-app notifications feed -->
    <div class="bg-white border border-border-soft rounded-2xl overflow-hidden">
      <div class="px-5 py-4 border-b border-border-soft flex items-center justify-between">
        <h3 class="font-bold text-ink">
          <i class="fas fa-bell text-brand mr-2" />{{ t("evedem.feed_title") }}
        </h3>
        <span v-if="notif.unread > 0" class="badge-green">{{ notif.unread }}</span>
      </div>

      <UiLoader v-if="notif.loading && !notif.items.length" />

      <div v-else-if="!notif.items.length" class="py-12">
        <UiEmpty icon="fa-bell-slash" :title="t('evedem.empty')" />
      </div>

      <ul v-else class="divide-y divide-border-soft">
        <li
          v-for="n in notif.items"
          :key="n.id"
          class="px-5 py-4 flex items-start gap-4 transition-colors cursor-pointer hover:bg-page/60"
          :class="{ 'bg-brand-soft/40': !n.is_read }"
          @click="onOpen(n)"
        >
          <div class="w-11 h-11 rounded-xl flex items-center justify-center text-base shrink-0" :class="icon(n.type).c">
            <i class="fas" :class="icon(n.type).i" />
          </div>
          <div class="flex-1 min-w-0">
            <div class="flex items-center gap-2">
              <span class="font-semibold text-ink text-sm">{{ n.title }}</span>
              <span v-if="!n.is_read" class="w-2 h-2 rounded-full bg-brand shrink-0" />
            </div>
            <p v-if="n.body" class="text-sm text-ink-medium truncate">{{ n.body }}</p>
            <span class="text-[11px] text-ink-soft">{{ formatDateTime(n.created_at) }}</span>
          </div>
          <button
            class="btn-ghost !p-2 text-ink-soft hover:text-red-500 shrink-0"
            :title="t('app.delete')"
            @click.stop="notif.remove(n.id)"
          >
            <i class="fas fa-xmark" />
          </button>
        </li>
      </ul>
    </div>
  </div>
</template>
