<script setup lang="ts">
import type { DashboardStats } from "~/types";

useHead({ title: () => useI18n().t("dashboard.title") });

const api = useApi();
const auth = useAuthStore();
const { t } = useI18n();

const { data: stats, pending } = useAsyncData<DashboardStats>(
  "dashboard-stats",
  () => api.get<DashboardStats>("/api/dashboard/stats")
);

const cards = computed(() => {
  const s = stats.value;
  if (!s) return [];
  return [
    {
      label: t("dashboard.users"),
      value: s.users,
      icon: "fa-users",
      featured: false,
      href: "/users",
    },
    {
      label: t("dashboard.open_requests"),
      value: s.requests_open,
      icon: "fa-clipboard-list",
      featured: true,
      href: "/requests",
    },
    {
      label: t("dashboard.leads"),
      value: s.leads,
      icon: "fa-bullseye",
      featured: false,
      href: "/lids",
    },
    {
      label: t("dashboard.my_tasks"),
      value: s.my_tasks,
      icon: "fa-list-check",
      featured: false,
      href: "/tasks",
    },
    {
      label: t("dashboard.apartments"),
      value: s.apartments,
      icon: "fa-building",
      featured: false,
      href: "/objekt",
    },
    {
      label: t("dashboard.free"),
      value: s.apartments_free,
      icon: "fa-circle-check",
      featured: false,
      href: "/objekt",
    },
    {
      label: t("dashboard.sold"),
      value: s.apartments_sold,
      icon: "fa-coins",
      featured: false,
      href: "/objekt",
    },
    {
      label: t("dashboard.sim_cards"),
      value: s.sim_cards,
      icon: "fa-sim-card",
      featured: false,
      href: "/sim-cards",
    },
  ];
});

const formatChartDays = (rows: { day: string; count: number }[]) => {
  // Make sure all 7 days are represented
  const today = new Date();
  const map = new Map<string, number>();
  rows?.forEach((r) => map.set(new Date(r.day).toDateString(), r.count));
  const out: { label: string; count: number }[] = [];
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

const leadsChart = computed(() => formatChartDays(stats.value?.leads_7d || []));
const requestsChart = computed(() =>
  formatChartDays(stats.value?.requests_7d || [])
);

const maxCount = (rows: { count: number }[]) =>
  Math.max(1, ...rows.map((r) => r.count));
</script>

<template>
  <div class="card p-6 lg:p-8">
    <PageHeader
      :title="$t('dashboard.title')"
      :subtitle="$t('dashboard.subtitle')"
    >
      <NuxtLink to="/tasks" class="btn-outline">
        <i class="fa-solid fa-list-check" />
        {{ $t("nav.tasks") }}
      </NuxtLink>
      <NuxtLink to="/requests" class="btn-primary">
        <i class="fa-solid fa-plus" />
        {{ $t("nav.requests") }}
      </NuxtLink>
    </PageHeader>

    <UiLoader v-if="pending" />

    <template v-else>
      <!-- Stats grid -->
      <div class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4 mb-8">
        <StatCard
          v-for="(c, i) in cards"
          :key="i"
          :label="c.label"
          :value="c.value"
          :icon="c.icon"
          :featured="c.featured"
          :href="c.href"
        />
      </div>

      <!-- Charts -->
      <div class="grid grid-cols-1 lg:grid-cols-2 gap-4">
        <div class="card p-6">
          <h3 class="text-base font-bold text-ink mb-4 flex items-center gap-2">
            <i class="fa-solid fa-bullseye text-brand" />
            {{ $t("dashboard.leads_chart") }}
          </h3>
          <div class="flex items-end gap-3 h-44 mt-3">
            <div
              v-for="(d, i) in leadsChart"
              :key="i"
              class="flex flex-col items-center gap-2 flex-1"
            >
              <div
                class="w-full bg-gradient-to-t from-brand to-brand-light rounded-t-lg transition-all"
                :style="{
                  height: `${(d.count / maxCount(leadsChart)) * 100}%`,
                  minHeight: '6px',
                }"
              />
              <div class="text-[10px] font-semibold text-ink-soft uppercase">
                {{ d.label }}
              </div>
              <div class="text-xs font-bold text-ink">{{ d.count }}</div>
            </div>
          </div>
        </div>

        <div class="card p-6">
          <h3 class="text-base font-bold text-ink mb-4 flex items-center gap-2">
            <i class="fa-solid fa-clipboard-list text-brand" />
            {{ $t("dashboard.requests_chart") }}
          </h3>
          <div class="flex items-end gap-3 h-44 mt-3">
            <div
              v-for="(d, i) in requestsChart"
              :key="i"
              class="flex flex-col items-center gap-2 flex-1"
            >
              <div
                class="w-full bg-gradient-to-t from-ink to-ink-medium rounded-t-lg transition-all"
                :style="{
                  height: `${(d.count / maxCount(requestsChart)) * 100}%`,
                  minHeight: '6px',
                }"
              />
              <div class="text-[10px] font-semibold text-ink-soft uppercase">
                {{ d.label }}
              </div>
              <div class="text-xs font-bold text-ink">{{ d.count }}</div>
            </div>
          </div>
        </div>
      </div>
    </template>
  </div>
</template>
