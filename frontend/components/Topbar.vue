<script setup lang="ts">
const ui = useUiStore();
const { locale, locales, setLocale } = useI18n();
const auth = useAuthStore();

const localesList = computed(() => locales.value as { code: string; name: string }[]);
const open = ref(false);

const switchLocale = (code: string) => {
  setLocale(code as any);
  open.value = false;
};
</script>

<template>
  <header
    class="card flex items-center justify-between px-5 h-[68px] sticky top-3.5 z-30"
  >
    <div class="flex items-center gap-3">
      <button
        class="lg:hidden btn-ghost !p-2"
        @click="ui.openMobileSidebar()"
      >
        <i class="fa-solid fa-bars text-lg" />
      </button>

      <div class="hidden md:flex items-center gap-3 bg-page rounded-xl px-3.5 py-2 w-[320px]">
        <i class="fa-solid fa-magnifying-glass text-ink-soft text-sm" />
        <input
          type="text"
          :placeholder="$t('app.search')"
          class="bg-transparent border-0 outline-none w-full text-sm placeholder:text-ink-soft"
        />
      </div>
    </div>

    <div class="flex items-center gap-2">
      <!-- Locale switcher -->
      <div class="relative">
        <button
          class="btn-outline !py-2 !px-3 gap-2 text-xs font-semibold"
          @click="open = !open"
        >
          <i class="fa-solid fa-globe text-brand" />
          {{ locale === 'tg' ? 'TG' : 'RU' }}
          <i class="fa-solid fa-chevron-down text-[10px] text-ink-soft" />
        </button>
        <Transition name="dd">
          <div
            v-if="open"
            class="absolute right-0 top-full mt-1.5 bg-white rounded-2xl shadow-hover py-1.5 min-w-[160px] z-40"
          >
            <button
              v-for="l in localesList"
              :key="l.code"
              class="w-full text-left px-4 py-2 text-sm hover:bg-brand-soft hover:text-brand transition-colors"
              :class="locale === l.code ? 'text-brand font-semibold' : 'text-ink-medium'"
              @click="switchLocale(l.code)"
            >
              {{ l.name }}
            </button>
          </div>
        </Transition>
      </div>

      <!-- Notifications (placeholder) -->
      <button class="btn-ghost !p-2.5">
        <i class="fa-solid fa-bell text-base" />
      </button>

      <!-- Avatar -->
      <NuxtLink to="/me/telegram" class="hidden sm:flex items-center gap-2 ml-1">
        <UiAvatar :src="auth.user?.photo" :name="auth.user?.full_name" size="sm" />
        <div class="text-xs leading-tight">
          <div class="font-semibold text-ink">{{ auth.user?.full_name?.split(' ')[0] }}</div>
          <div class="text-ink-soft">{{ auth.user?.role }}</div>
        </div>
      </NuxtLink>
    </div>
  </header>
</template>

<style scoped>
.dd-enter-active, .dd-leave-active {
  transition: all .15s ease;
}
.dd-enter-from, .dd-leave-to {
  opacity: 0;
  transform: translateY(-4px);
}
</style>
