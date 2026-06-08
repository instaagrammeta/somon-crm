<script setup lang="ts">
/**
 * Topbar — port of base.html topbar: search, envelope→chat, bell→evedem,
 * user chip. Plus a locale switcher (tg/ru) which the original lacked but
 * is required by the multilingual spec.
 */
const ui = useUiStore();
const auth = useAuthStore();
const { locale, locales, setLocale, t } = useI18n();

const search = ref("");
const localeOpen = ref(false);
const localesList = computed(() => locales.value as { code: string; name: string }[]);

const onSearch = (e: KeyboardEvent) => {
  if (e.key === "Enter" && search.value.trim()) {
    navigateTo(`/zayavka?search=${encodeURIComponent(search.value.trim())}`);
  }
};

const switchLocale = (code: string) => {
  setLocale(code as "tg" | "ru");
  localeOpen.value = false;
};
</script>

<template>
  <header class="bg-white rounded-3xl px-[22px] py-3.5 flex items-center gap-[18px] shadow-card">
    <button
      class="lg:hidden w-[42px] h-[42px] bg-page rounded-xl flex items-center justify-center text-ink text-lg"
      @click="ui.openMobileSidebar()"
    >
      <i class="fas fa-bars" />
    </button>

    <!-- Search -->
    <div class="flex-1 max-w-[480px] relative flex items-center">
      <i class="fas fa-search absolute left-4 text-ink-soft text-sm" />
      <input
        v-model="search"
        type="text"
        :placeholder="t('app.search_placeholder')"
        class="w-full pl-[42px] pr-3.5 py-3 bg-page border-[1.5px] border-transparent rounded-xl text-[13px] focus:outline-none focus:border-brand focus:bg-white transition-all"
        @keydown="onSearch"
      />
    </div>

    <div class="flex items-center gap-2.5 ml-auto">
      <!-- Locale switcher -->
      <div class="relative">
        <button
          class="h-[42px] px-3 bg-page rounded-xl flex items-center gap-2 text-[13px] font-semibold text-ink-medium hover:bg-brand-soft hover:text-brand transition-all"
          @click="localeOpen = !localeOpen"
        >
          <i class="fas fa-globe text-brand" />
          {{ locale === "tg" ? "TG" : "RU" }}
        </button>
        <Transition name="dd">
          <div
            v-if="localeOpen"
            class="absolute right-0 top-full mt-1.5 bg-white rounded-2xl shadow-hover py-1.5 min-w-[150px] z-40"
          >
            <button
              v-for="l in localesList"
              :key="l.code"
              class="w-full text-left px-4 py-2 text-sm hover:bg-brand-soft transition-colors"
              :class="locale === l.code ? 'text-brand font-semibold' : 'text-ink-medium'"
              @click="switchLocale(l.code)"
            >
              {{ l.name }}
            </button>
          </div>
        </Transition>
      </div>

      <!-- Envelope → chat -->
      <NuxtLink
        to="/chat"
        class="w-[42px] h-[42px] bg-page rounded-xl flex items-center justify-center text-ink-medium hover:bg-brand-soft hover:text-brand transition-all relative"
        :title="t('nav.chat')"
      >
        <i class="fas fa-envelope" />
        <span class="absolute top-2.5 right-[11px] w-2 h-2 bg-red-500 rounded-full border-2 border-white" />
      </NuxtLink>

      <!-- Bell → notifications -->
      <NuxtLink
        to="/evedem"
        class="w-[42px] h-[42px] bg-page rounded-xl flex items-center justify-center text-ink-medium hover:bg-brand-soft hover:text-brand transition-all relative"
        :title="t('nav.notifications')"
      >
        <i class="fas fa-bell" />
        <span class="absolute top-2.5 right-[11px] w-2 h-2 bg-red-500 rounded-full border-2 border-white" />
      </NuxtLink>

      <!-- User -->
      <NuxtLink
        to="/me/telegram"
        class="flex items-center gap-2.5 pl-1 pr-3 py-1 bg-page rounded-full hover:bg-brand-soft transition-all"
      >
        <UiAvatar :src="auth.user?.photo" :name="auth.user?.full_name" size="sm" />
        <div class="hidden sm:flex flex-col leading-[1.2]">
          <span class="font-semibold text-[13px] text-ink">{{ auth.user?.full_name }}</span>
          <span class="text-[11px] text-ink-soft">{{ auth.user?.category }}</span>
        </div>
      </NuxtLink>
    </div>
  </header>
</template>

<style scoped>
.dd-enter-active,
.dd-leave-active {
  transition: all 0.15s ease;
}
.dd-enter-from,
.dd-leave-to {
  opacity: 0;
  transform: translateY(-4px);
}
</style>
