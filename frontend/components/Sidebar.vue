<script setup lang="ts">
/**
 * Sidebar — faithful port of the Flask base.html sidebar.
 * Logo "RealEstate", two sections (Меню / Умумӣ), green Donezo theme.
 */
const auth = useAuthStore();
const ui = useUiStore();
const { t } = useI18n();
const route = useRoute();

interface NavItem {
  to: string;
  label: string;
  icon: string;
  show?: () => boolean;
}

const isSales = computed(
  () => auth.user?.category === "Отдел продаж" || auth.isAdmin
);

const menu = computed<NavItem[]>(() => [
  { to: "/dashboard", label: t("nav.dashboard"), icon: "fa-th-large" },
  { to: "/zadacha", label: t("nav.tasks"), icon: "fa-tasks" },
  { to: "/zayavka", label: t("nav.requests"), icon: "fa-file-invoice" },
  { to: "/lids", label: t("nav.sales"), icon: "fa-fire", show: () => isSales.value },
  { to: "/houses", label: t("nav.houses"), icon: "fa-home" },
  { to: "/obiekt", label: t("nav.buildings"), icon: "fa-building" },
  { to: "/ipoteka", label: t("nav.mortgage"), icon: "fa-building-columns" },
  { to: "/rasrochka", label: t("nav.installments"), icon: "fa-store" },
]);

const general = computed<NavItem[]>(() => [
  { to: "/posts", label: t("nav.networks"), icon: "fa-share-nodes" },
  { to: "/chat", label: t("nav.chat"), icon: "fa-comment-dots" },
  { to: "/sim-cards", label: t("nav.sim_cards"), icon: "fa-sim-card" },
  { to: "/baza", label: t("nav.base"), icon: "fa-folder-open" },
  // v-2 admin tools
  { to: "/admin/website-leads", label: "Лидҳои вебсайт", icon: "fa-globe", show: () => auth.isAdmin },
  { to: "/me/2fa", label: "2FA" , icon: "fa-shield-halved" },
  { to: "/admin", label: t("nav.employees"), icon: "fa-user-shield", show: () => auth.isAdmin },
]);

const visibleMenu = computed(() => menu.value.filter((i) => !i.show || i.show()));
const visibleGeneral = computed(() => general.value.filter((i) => !i.show || i.show()));

function isActive(path: string) {
  if (path === "/dashboard") return route.path === "/dashboard";
  return route.path.startsWith(path);
}

const onLogout = async () => {
  if (!confirm(t("app.logout_confirm"))) return;
  await auth.logout();
  navigateTo("/login");
};
</script>

<template>
  <!-- Mobile overlay -->
  <Transition name="fade">
    <div
      v-if="ui.mobileSidebarOpen"
      class="fixed inset-0 z-[110] bg-black/40 backdrop-blur-sm lg:hidden"
      @click="ui.closeMobileSidebar()"
    />
  </Transition>

  <aside
    class="fixed top-3.5 left-3.5 flex flex-col w-[260px] bg-white rounded-3xl shadow-card overflow-hidden z-[120] transition-transform duration-300 ease-smooth"
    :class="ui.mobileSidebarOpen ? 'translate-x-0' : '-translate-x-[110%] lg:translate-x-0'"
    style="height: calc(100vh - 28px)"
  >
    <!-- Header / Logo -->
    <div class="px-[22px] pt-6 pb-[18px]">
      <NuxtLink to="/dashboard" class="flex items-center gap-2.5">
        <span class="logo-icon" />
        <span class="text-[19px] font-extrabold tracking-[-0.3px] text-ink">RealEstate</span>
      </NuxtLink>
    </div>

    <!-- Nav -->
    <nav class="flex-1 px-3.5 overflow-y-auto no-scrollbar">
      <div class="nav-section-title">{{ t("nav.menu") }}</div>
      <NuxtLink
        v-for="item in visibleMenu"
        :key="item.to"
        :to="item.to"
        class="nav-item"
        :class="{ active: isActive(item.to) }"
        @click="ui.closeMobileSidebar()"
      >
        <i class="fas w-[18px] text-[15px] text-center" :class="item.icon" />
        <span class="flex-1">{{ item.label }}</span>
      </NuxtLink>

      <div class="nav-section-title">{{ t("nav.general") }}</div>
      <NuxtLink
        v-for="item in visibleGeneral"
        :key="item.to"
        :to="item.to"
        class="nav-item"
        :class="{ active: isActive(item.to) }"
        @click="ui.closeMobileSidebar()"
      >
        <i class="fas w-[18px] text-[15px] text-center" :class="item.icon" />
        <span class="flex-1">{{ item.label }}</span>
      </NuxtLink>
    </nav>

    <!-- Footer (user + logout) -->
    <div class="p-3.5">
      <NuxtLink to="/me/telegram" class="flex items-center gap-3 mb-3 px-1 py-2 rounded-xl hover:bg-brand-soft transition-colors">
        <UiAvatar :src="auth.user?.photo" :name="auth.user?.full_name" size="md" />
        <div class="flex-1 min-w-0">
          <div class="font-semibold text-[13px] text-ink truncate">
            {{ auth.user?.full_name || t("app.user") }}
          </div>
          <div class="text-[11px] text-ink-soft truncate">
            {{ auth.user?.category || "" }}
          </div>
        </div>
      </NuxtLink>
      <button class="logout-btn w-full" @click="onLogout">
        <i class="fas fa-arrow-right-from-bracket" />
        <span>{{ t("app.logout") }}</span>
      </button>
    </div>
  </aside>
</template>

<style scoped>
.nav-section-title {
  @apply text-[10px] font-bold text-ink-soft tracking-[1.5px] uppercase px-3 pt-3.5 pb-2;
}
.logout-btn {
  @apply px-4 py-2.5 bg-red-50 rounded-xl text-red-500 font-semibold flex items-center gap-3 text-[13px] hover:bg-red-100 transition-colors;
}
.fade-enter-active,
.fade-leave-active {
  transition: opacity 0.25s ease;
}
.fade-enter-from,
.fade-leave-to {
  opacity: 0;
}
</style>
