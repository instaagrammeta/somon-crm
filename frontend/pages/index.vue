<script setup lang="ts">
import type { DashboardStats } from "~/types";

useHead({ title: () => useI18n().t("nav.dashboard") });

const api = useApi();
const auth = useAuthStore();
const { t } = useI18n();

const { data: stats, pending } = useAsyncData<DashboardStats>(
  "dashboard-stats",
  () => api.get<DashboardStats>("/api/dashboard/stats"),
  { lazy: true, default: () => ({} as DashboardStats) }
);

const cards = computed(() => {
  const s = stats.value;
  if (!s) return [];
  return [
    { label: t("dashboard.open_requests"), value: s.requests_open ?? 0, icon: "fa-file-invoice", trend: t("dashboard.active"), featured: true, href: "/zayavka" },
    { label: t("dashboard.leads"), value: s.leads ?? 0, icon: "fa-fire", trend: t("dashboard.total_label"), href: "/lids" },
    { label: t("dashboard.my_tasks"), value: s.my_tasks ?? 0, icon: "fa-tasks", trend: t("dashboard.active"), href: "/zadacha" },
    { label: t("dashboard.users"), value: s.users ?? 0, icon: "fa-users", trend: t("dashboard.total_label"), href: "/admin" },
    { label: t("dashboard.apartments"), value: s.apartments ?? 0, icon: "fa-building", trend: t("dashboard.total_label"), href: "/obiekt" },
    { label: t("dashboard.free"), value: s.apartments_free ?? 0, icon: "fa-door-open", trend: t("objekt.free"), href: "/obiekt" },
    { label: t("dashboard.sold"), value: s.apartments_sold ?? 0, icon: "fa-coins", trend: t("objekt.sold"), href: "/obiekt" },
    { label: t("dashboard.sim_cards"), value: s.sim_cards ?? 0, icon: "fa-sim-card", trend: t("dashboard.active"), href: "/sim-cards" },
  ];
});

const buildDays = (rows?: { day: string; count: number }[]) => {
  const map = new Map<string, number>();
  (rows || []).forEach((r) => map.set(new Date(r.day).toDateString(), r.count));
  const out: { label: string; count: number }[] = [];
  const today = new Date();
  for (let i = 6; i >= 0; i--) {
    const d = new Date(today);
    d.setDate(today.getDate() - i);
    out.push({
      label: d.toLocaleDateString("ru-RU", { weekday: "short" }),
      count: map.get(d.toDateString()) || 0,
    });
  }
  return out;
};
const leads7 = computed(() => buildDays(stats.value?.leads_7d));
const req7 = computed(() => buildDays(stats.value?.requests_7d));
const maxOf = (rows: { count: number }[]) => Math.max(1, ...rows.map((r) => r.count));
</script>

<template>
  <div>
    <div class="mb-7">
      <h1 class="text-[28px] font-bold text-ink tracking-[-0.3px] mb-1.5">
        {{ t("dashboard.hello", { name: auth.user?.full_name?.split(" ")[0] || "" }) }}
      </h1>
      <p class="text-ink-soft text-sm">{{ t("dashboard.subtitle") }}</p>
    </div>

    <UiLoader v-if="pending" />

    <template v-else>
      <!-- stats grid -->
      <div class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4 mb-6">
        <NuxtLink
          v-for="(c, i) in cards"
          :key="i"
          :to="c.href"
          class="rounded-[20px] p-[22px] border transition-all duration-200 hover:-translate-y-[3px] hover:shadow-hover relative overflow-hidden"
          :class="c.featured ? 'bg-gradient-to-br from-brand to-brand-dark text-white border-transparent' : 'bg-white border-border-soft'"
        >
          <template v-if="c.featured">
            <span class="absolute -right-10 -top-10 w-[140px] h-[140px] rounded-full bg-white/[0.06]" />
            <span class="absolute -right-20 -bottom-20 w-[180px] h-[180px] rounded-full bg-white/[0.04]" />
          </template>
          <div class="flex justify-between items-center mb-4 relative z-10">
            <span class="text-sm font-medium" :class="c.featured ? 'text-white/90' : 'text-ink-medium'">{{ c.label }}</span>
            <span
              class="w-8 h-8 rounded-[10px] flex items-center justify-center text-[13px]"
              :class="c.featured ? 'bg-white/20 text-white' : 'bg-brand-soft text-brand'"
            >
              <i class="fas" :class="c.icon" />
            </span>
          </div>
          <div class="text-[42px] leading-none font-bold mb-3 relative z-10" :class="c.featured ? 'text-white' : 'text-ink'">
            {{ c.value }}
          </div>
          <span
            class="inline-flex items-center gap-1.5 text-[11px] font-medium px-2.5 py-1 rounded-full relative z-10"
            :class="c.featured ? 'bg-white/20 text-white' : 'bg-brand-soft text-brand'"
          >
            <i class="fas fa-arrow-trend-up" /> {{ c.trend }}
          </span>
        </NuxtLink>
      </div>

      <!-- charts -->
      <div class="grid grid-cols-1 lg:grid-cols-2 gap-4">
        <div class="bg-white border border-border-soft rounded-[20px] p-6">
          <div class="font-bold text-ink mb-4 flex items-center gap-2">
            <i class="fas fa-chart-line text-brand" /> {{ t("dashboard.leads_chart") }}
          </div>
          <div class="flex items-end gap-3 h-44">
            <div v-for="(d, i) in leads7" :key="i" class="flex flex-col items-center gap-2 flex-1">
              <div class="w-full bg-gradient-to-t from-brand to-brand-light rounded-t-lg transition-all" :style="{ height: `${(d.count / maxOf(leads7)) * 100}%`, minHeight: '6px' }" />
              <div class="text-[10px] font-semibold text-ink-soft uppercase">{{ d.label }}</div>
              <div class="text-xs font-bold text-ink">{{ d.count }}</div>
            </div>
          </div>
        </div>
        <div class="bg-white border border-border-soft rounded-[20px] p-6">
          <div class="font-bold text-ink mb-4 flex items-center gap-2">
            <i class="fas fa-chart-bar text-brand" /> {{ t("dashboard.requests_chart") }}
          </div>
          <div class="flex items-end gap-3 h-44">
            <div v-for="(d, i) in req7" :key="i" class="flex flex-col items-center gap-2 flex-1">
              <div class="w-full bg-gradient-to-t from-ink to-ink-medium rounded-t-lg transition-all" :style="{ height: `${(d.count / maxOf(req7)) * 100}%`, minHeight: '6px' }" />
              <div class="text-[10px] font-semibold text-ink-soft uppercase">{{ d.label }}</div>
              <div class="text-xs font-bold text-ink">{{ d.count }}</div>
            </div>
          </div>
        </div>
      </div>
    </template>
  </div>
</template>
