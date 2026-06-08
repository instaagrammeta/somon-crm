<script setup lang="ts">
useHead({ title: () => useI18n().t("nav.notifications") });
const { t } = useI18n();
const auth = useAuthStore();
</script>

<template>
  <div>
    <div class="mb-6">
      <h1 class="text-[28px] font-bold text-ink mb-1">{{ t("nav.notifications") }}</h1>
      <p class="text-ink-soft text-sm">{{ t("evedem.subtitle") }}</p>
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

    <!-- Notification types info -->
    <div class="grid grid-cols-1 md:grid-cols-2 gap-4">
      <div v-for="(item, i) in [
          { icon: 'fa-tasks', color: 'text-brand bg-brand-soft', key: 'tasks' },
          { icon: 'fa-file-invoice', color: 'text-blue-500 bg-blue-50', key: 'requests' },
          { icon: 'fa-fire', color: 'text-orange-500 bg-orange-50', key: 'leads' },
          { icon: 'fa-sim-card', color: 'text-purple-500 bg-purple-50', key: 'sim' },
        ]" :key="i"
        class="bg-white border border-border-soft rounded-2xl p-5 flex items-center gap-4">
        <div class="w-12 h-12 rounded-xl flex items-center justify-center text-lg" :class="item.color">
          <i class="fas" :class="item.icon" />
        </div>
        <div>
          <div class="font-semibold text-ink text-sm">{{ t(`evedem.type_${item.key}`) }}</div>
          <div class="text-xs text-ink-soft">{{ t(`evedem.type_${item.key}_desc`) }}</div>
        </div>
      </div>
    </div>
  </div>
</template>
