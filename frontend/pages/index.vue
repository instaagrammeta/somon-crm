<script setup lang="ts">
/**
 * Public marketing home. Reachable to anyone (no auth) via auth.global.ts.
 * Loads featured houses + the brand config; the lead-capture popup is wired
 * into the layout so we don't repeat it on every page.
 */
import type { PublicHouse } from "~/types/public";

definePageMeta({ layout: "site" });

const api = usePublicApi();

const { data: featured } = useAsyncData<PublicHouse[]>(
  "home-featured",
  () => api.get<PublicHouse[]>("/api/public/houses?limit=6"),
  { default: () => [] }
);

const search = reactive({
  q: "",
  district: "",
});

const router = useRouter();
const goSearch = () => {
  const params = new URLSearchParams();
  for (const [k, v] of Object.entries(search)) {
    if (v) params.set(k, String(v));
  }
  router.push("/properties" + (params.toString() ? "?" + params.toString() : ""));
};
</script>

<template>
  <div>
    <!-- Hero -->
    <section class="relative overflow-hidden">
      <div class="absolute inset-0 bg-gradient-to-br from-brand/95 via-brand to-brand-dark" />
      <div class="relative max-w-6xl mx-auto px-4 py-20 md:py-28 text-white">
        <h1 class="text-3xl md:text-5xl font-extrabold leading-tight max-w-2xl">
          Манзилҳои нав бо <span class="underline decoration-white/40">Тури 360°</span>
        </h1>
        <p class="mt-4 text-white/85 max-w-xl">
          Хонаро аз ҳар нуқтаи ҷаҳон бубинед — бо панораҳои воқеӣ ва харитаи интерактивӣ.
        </p>

        <form
          class="mt-8 grid sm:grid-cols-4 gap-2 bg-white/95 p-2 rounded-2xl shadow-2xl max-w-3xl"
          @submit.prevent="goSearch"
        >
          <input
            v-model="search.q"
            type="text"
            placeholder="Унвон, кӯча…"
            class="px-4 py-3 rounded-xl bg-page border-0 text-ink placeholder:text-ink-soft focus:outline-none focus:ring-2 focus:ring-brand sm:col-span-2"
          />
          <input
            v-model="search.district"
            type="text"
            placeholder="Ноҳия"
            class="px-4 py-3 rounded-xl bg-page border-0 text-ink placeholder:text-ink-soft focus:outline-none focus:ring-2 focus:ring-brand"
          />
          <button
            type="submit"
            class="px-6 py-3 rounded-xl bg-brand text-white font-semibold hover:bg-brand-dark"
          >
            <i class="fas fa-search mr-2" />Ҷустуҷӯ
          </button>
        </form>

        <div class="mt-8 flex flex-wrap gap-3 text-sm">
          <span class="inline-flex items-center gap-2 px-3 py-1.5 rounded-full bg-white/15 backdrop-blur">
            <i class="fas fa-vr-cardboard" /> Турҳои виртуалӣ
          </span>
          <span class="inline-flex items-center gap-2 px-3 py-1.5 rounded-full bg-white/15 backdrop-blur">
            <i class="fas fa-mobile-screen-button" /> Замимаи Android
          </span>
          <span class="inline-flex items-center gap-2 px-3 py-1.5 rounded-full bg-white/15 backdrop-blur">
            <i class="fas fa-robot" /> Дастёри AI
          </span>
        </div>
      </div>
    </section>

    <!-- Featured houses -->
    <section class="max-w-6xl mx-auto px-4 py-16">
      <div class="flex items-end justify-between mb-6">
        <div>
          <h2 class="text-2xl font-bold text-ink">Лоиҳаҳои нав</h2>
          <p class="text-ink-soft text-sm">Аз сохтмонгарони боэътимод</p>
        </div>
        <NuxtLink to="/properties" class="text-sm font-medium text-brand hover:underline">
          Ҳама манзилҳо <i class="fas fa-arrow-right text-xs ml-1" />
        </NuxtLink>
      </div>

      <div v-if="featured && featured.length" class="grid gap-5 sm:grid-cols-2 lg:grid-cols-3">
        <SiteHouseCard v-for="h in featured" :key="h.id" :house="h" />
      </div>
      <div v-else class="text-center py-16 text-ink-soft">
        <i class="fas fa-house text-4xl mb-3 text-brand/30" />
        <p>Дар наздикии вақт лоиҳаҳои нав илова мекунем.</p>
      </div>
    </section>

    <!-- 3 features -->
    <section class="bg-white border-y border-border-soft">
      <div class="max-w-6xl mx-auto px-4 py-16 grid gap-8 md:grid-cols-3 text-center">
        <div>
          <div class="w-14 h-14 mx-auto rounded-2xl bg-brand-soft text-brand grid place-items-center text-xl">
            <i class="fas fa-vr-cardboard" />
          </div>
          <h3 class="mt-4 font-bold text-lg">Турҳои 360°</h3>
          <p class="mt-1 text-sm text-ink-soft">
            Ошхона → Меҳмонхона → Хонаи хоб бо як зер кардан.
          </p>
        </div>
        <div>
          <div class="w-14 h-14 mx-auto rounded-2xl bg-brand-soft text-brand grid place-items-center text-xl">
            <i class="fab fa-whatsapp" />
          </div>
          <h3 class="mt-4 font-bold text-lg">WhatsApp бевосита</h3>
          <p class="mt-1 text-sm text-ink-soft">
            Дар бораи манзил аз WhatsApp пурсиш кунед — фавран ҷавоб мегиред.
          </p>
        </div>
        <div>
          <div class="w-14 h-14 mx-auto rounded-2xl bg-brand-soft text-brand grid place-items-center text-xl">
            <i class="fas fa-mobile-screen" />
          </div>
          <h3 class="mt-4 font-bold text-lg">Замимаи Android</h3>
          <p class="mt-1 text-sm text-ink-soft">
            APK-ро мустақим аз вебсайт зер карда тавонед.
            <NuxtLink to="/install" class="text-brand">Зер кашед</NuxtLink>
          </p>
        </div>
      </div>
    </section>
  </div>
</template>
