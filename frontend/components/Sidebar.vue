<script setup lang="ts">
const auth = useAuthStore();
const ui = useUiStore();
const { t } = useI18n();
const route = useRoute();

interface NavItem {
  to: string;
  label: string;
  icon: string;
  adminOnly?: boolean;
}

const sections = computed(() => [
  {
    title: t("nav.main"),
    items: [
      { to: "/", label: t("nav.dashboard"), icon: "fa-house" },
      { to: "/tasks", label: t("nav.tasks"), icon: "fa-list-check" },
      { to: "/requests", label: t("nav.requests"), icon: "fa-clipboard-list" },
      { to: "/lids", label: t("nav.lids"), icon: "fa-bullseye" },
      { to: "/chat", label: t("nav.chat"), icon: "fa-comments" },
    ] as NavItem[],
  },
  {
    title: t("nav.work"),
    items: [
      { to: "/houses", label: t("nav.houses"), icon: "fa-building" },
      { to: "/realty", label: t("nav.realty"), icon: "fa-map-location-dot" },
      { to: "/objekt", label: t("nav.objekt"), icon: "fa-table-cells-large" },
      { to: "/posts", label: t("nav.posts"), icon: "fa-image" },
    ] as NavItem[],
  },
  {
    title: t("nav.infrastructure"),
    items: [
      { to: "/sim-cards", label: t("nav.sim_cards"), icon: "fa-sim-card" },
      { to: "/folders", label: t("nav.folders"), icon: "fa-folder-open" },
    ] as NavItem[],
  },
  {
    title: t("nav.finance"),
    items: [
      { to: "/banks", label: t("nav.banks"), icon: "fa-building-columns" },
      { to: "/installments", label: t("nav.installments"), icon: "fa-handshake" },
    ] as NavItem[],
  },
  {
    title: t("nav.admin"),
    items: [
      { to: "/users", label: t("nav.users"), icon: "fa-users-gear", adminOnly: true },
      { to: "/me/telegram", label: t("nav.telegram"), icon: "fa-paper-plane" },
    ] as NavItem[],
  },
]);

const visibleSections = computed(() =>
  sections.value
    .map((s) => ({
      ...s,
      items: s.items.filter((i) => !i.adminOnly || auth.isAdmin),
    }))
    .filter((s) => s.items.length > 0)
);

function isActive(path: string) {
  if (path === "/") return route.path === "/";
  return route.path.startsWith(path);
}

const onLogout = async () => {
  await auth.logout();
  navigateTo("/login");
};
</script>

<template>
  <!-- Mobile overlay -->
  <Transition name="fade">
    <div
      v-if="ui.mobileSidebarOpen"
      class="fixed inset-0 z-[110] bg-ink/50 backdrop-blur-sm lg:hidden"
      @click="ui.closeMobileSidebar()"
    />
  </Transition>

  <aside
    class="fixed top-3.5 left-3.5 lg:flex flex-col w-[260px] bg-white rounded-3xl shadow-card overflow-hidden z-[120]"
    :class="[
      'transition-transform duration-300 ease-smooth',
      ui.mobileSidebarOpen ? 'translate-x-0' : '-translate-x-[110%] lg:translate-x-0',
    ]"
    style="height: calc(100vh - 28px); display: flex"
  >
    <!-- Header -->
    <div class="px-5 pt-6 pb-4">
      <NuxtLink to="/" class="flex items-center gap-2.5">
        <span class="logo-icon" />
        <span class="font-extrabold text-[19px] tracking-tight text-ink">
          Somon CRM
        </span>
      </NuxtLink>
    </div>

    <!-- Nav -->
    <nav class="flex-1 px-3.5 overflow-y-auto no-scrollbar">
      <div v-for="(s, si) in visibleSections" :key="si" class="mb-1">
        <div
          class="text-[10px] font-bold text-ink-soft tracking-[1.5px] px-3 pt-3 pb-2"
        >
          {{ s.title }}
        </div>
        <NuxtLink
          v-for="item in s.items"
          :key="item.to"
          :to="item.to"
          class="nav-item"
          :class="{ active: isActive(item.to) }"
          @click="ui.closeMobileSidebar()"
        >
          <i class="fa-solid w-[18px] text-[15px] text-center" :class="item.icon" />
          <span>{{ item.label }}</span>
        </NuxtLink>
      </div>
    </nav>

    <!-- Footer (user + logout) -->
    <div class="p-3.5 pt-2">
      <div class="bg-ink rounded-2xl p-4 mb-3 relative overflow-hidden">
        <span
          class="absolute -right-7 -bottom-7 w-28 h-28 rounded-full opacity-40"
          style="background: linear-gradient(135deg, #1f7a4d, transparent)"
        />
        <div class="relative">
          <div
            class="w-8 h-8 rounded-lg bg-white text-brand flex items-center justify-center text-sm mb-3"
          >
            <i class="fa-solid fa-paper-plane" />
          </div>
          <div class="text-white text-sm font-bold mb-1">Telegram bot</div>
          <div class="text-white/70 text-[11px] mb-3 leading-snug">
            Огоҳсозиҳои дархостҳо ва вазифаҳо
          </div>
          <NuxtLink to="/me/telegram" class="inline-flex items-center gap-2 bg-brand hover:bg-brand-dark text-white px-4 py-1.5 rounded-lg text-xs font-semibold transition-colors">
            <i class="fa-solid fa-arrow-right text-[10px]" />
            {{ $t("telegram.title") }}
          </NuxtLink>
        </div>
      </div>

      <div class="flex items-center gap-3 px-1 py-2 mb-2">
        <UiAvatar :src="auth.user?.photo" :name="auth.user?.full_name" size="md" />
        <div class="flex-1 min-w-0">
          <div class="font-semibold text-[13px] text-ink truncate">
            {{ auth.user?.full_name }}
          </div>
          <div class="text-[11px] text-ink-soft capitalize">
            {{ auth.user ? $t(`user.${auth.user.role}`) : "" }}
          </div>
        </div>
      </div>

      <button class="logout-btn w-full" @click="onLogout">
        <i class="fa-solid fa-right-from-bracket" />
        <span>{{ $t("app.logout") }}</span>
      </button>
    </div>
  </aside>
</template>

<style scoped>
.logout-btn {
  @apply px-4 py-2.5 bg-red-50 rounded-xl text-red-500 font-semibold flex items-center gap-3 text-[13px] hover:bg-red-100 transition-colors;
}
.fade-enter-active, .fade-leave-active {
  transition: opacity .25s ease;
}
.fade-enter-from, .fade-leave-to { opacity: 0; }
</style>
