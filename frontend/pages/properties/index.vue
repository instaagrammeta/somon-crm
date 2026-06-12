<script setup lang="ts">
import type { PublicHouse } from "~/types/public";

definePageMeta({ layout: "site" });

const api = usePublicApi();
const route = useRoute();
const router = useRouter();

// Filter state mirrors the URL so the page is shareable.
const filters = reactive({
  q: (route.query.q as string) || "",
  district: (route.query.district as string) || "",
  rooms: (route.query.rooms as string) || "",
  min_price: (route.query.min_price as string) || "",
  max_price: (route.query.max_price as string) || "",
});

const queryString = computed(() => {
  const p = new URLSearchParams();
  for (const [k, v] of Object.entries(filters)) {
    if (v) p.set(k, String(v));
  }
  p.set("limit", "60");
  return p.toString();
});

const { data: houses, pending } = useAsyncData<PublicHouse[]>(
  "houses-list",
  () => api.get<PublicHouse[]>("/api/public/houses?" + queryString.value),
  { watch: [queryString], default: () => [] as PublicHouse[] }
);

const apply = () => {
  router.replace({ query: Object.fromEntries(new URLSearchParams(queryString.value)) });
};
const reset = () => {
  for (const k of Object.keys(filters)) (filters as any)[k] = "";
  apply();
};
</script>

<template>
  <div class="max-w-6xl mx-auto px-4 py-10">
    <h1 class="text-2xl md:text-3xl font-bold text-ink">Манзилҳо</h1>
    <p class="text-ink-soft text-sm">{{ houses?.length || 0 }} натиҷа</p>

    <div class="mt-6 grid lg:grid-cols-[280px_1fr] gap-6">
      <aside class="bg-white rounded-2xl border border-border-soft p-5 h-fit lg:sticky lg:top-20">
        <h3 class="font-semibold mb-3 text-ink">Филтрҳо</h3>
        <div class="space-y-3">
          <input
            v-model="filters.q"
            placeholder="Ҷустуҷӯ…"
            class="w-full px-3 py-2.5 rounded-xl border border-border-soft focus:outline-none focus:border-brand"
          />
          <input
            v-model="filters.district"
            placeholder="Ноҳия"
            class="w-full px-3 py-2.5 rounded-xl border border-border-soft focus:outline-none focus:border-brand"
          />
          <select
            v-model="filters.rooms"
            class="w-full px-3 py-2.5 rounded-xl border border-border-soft bg-white"
          >
            <option value="">Ҳама ҳуҷраҳо</option>
            <option v-for="n in 5" :key="n" :value="n">{{ n }}-ҳуҷра</option>
          </select>
          <div class="grid grid-cols-2 gap-2">
            <input
              v-model="filters.min_price"
              placeholder="Аз"
              type="number"
              class="px-3 py-2.5 rounded-xl border border-border-soft focus:outline-none focus:border-brand"
            />
            <input
              v-model="filters.max_price"
              placeholder="То"
              type="number"
              class="px-3 py-2.5 rounded-xl border border-border-soft focus:outline-none focus:border-brand"
            />
          </div>
          <div class="flex gap-2">
            <button
              class="flex-1 py-2.5 rounded-xl bg-brand text-white font-semibold hover:bg-brand-dark"
              @click="apply"
            >
              Татбиқ
            </button>
            <button
              class="px-4 py-2.5 rounded-xl border border-border-soft hover:bg-page"
              @click="reset"
            >
              Тоза
            </button>
          </div>
        </div>
      </aside>

      <section>
        <div v-if="pending" class="text-center py-16 text-ink-soft">
          <i class="fas fa-spinner fa-spin text-2xl" />
        </div>
        <div v-else-if="!houses?.length" class="text-center py-16 text-ink-soft">
          <i class="fas fa-magnifying-glass text-3xl mb-3" />
          <p>Натиҷае ёфт нашуд.</p>
        </div>
        <div v-else class="grid gap-5 sm:grid-cols-2 xl:grid-cols-3">
          <SiteHouseCard v-for="h in houses" :key="h.id" :house="h" />
        </div>
      </section>
    </div>
  </div>
</template>
