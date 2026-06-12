<script setup lang="ts">
import type { PublicHouse } from "~/types/public";

definePageMeta({ layout: "site" });

const route = useRoute();
const api = usePublicApi();

const { data: house, pending } = useAsyncData<PublicHouse | null>(
  () => `house-${route.params.id}`,
  () => api.get<PublicHouse>(`/api/public/houses/${route.params.id}`),
  { default: () => null }
);

const fmt = (n: number) =>
  n ? new Intl.NumberFormat("ru-RU").format(Math.round(n)) : "—";

// Lightweight gallery state.
const activeIdx = ref(0);

// Inline contact form (variant of the popup with the lead source = "house_card").
const sending = ref(false);
const sent = ref(false);
const lead = reactive({ name: "", phone: "", message: "" });

const sendLead = async () => {
  if (!lead.phone) return;
  sending.value = true;
  try {
    await api.post("/api/public/leads", {
      ...lead,
      source: "house_card",
      house_id: house.value?.id,
      page: window.location.pathname,
      referrer: document.referrer,
    });
    sent.value = true;
    lead.name = lead.phone = lead.message = "";
  } finally {
    sending.value = false;
  }
};
</script>

<template>
  <div class="max-w-6xl mx-auto px-4 py-8">
    <NuxtLink to="/properties" class="text-sm text-brand hover:underline">
      <i class="fas fa-arrow-left mr-1" />Ҳама манзилҳо
    </NuxtLink>

    <div v-if="pending" class="py-16 text-center text-ink-soft">
      <i class="fas fa-spinner fa-spin text-2xl" />
    </div>

    <div v-else-if="house" class="mt-3 grid lg:grid-cols-[1fr_360px] gap-8">
      <div>
        <!-- gallery -->
        <div class="rounded-2xl overflow-hidden bg-page aspect-[16/10]">
          <img
            v-if="house.gallery?.[activeIdx]"
            :src="house.gallery[activeIdx]"
            class="w-full h-full object-cover"
          />
          <div v-else class="w-full h-full grid place-items-center text-ink-soft">
            <i class="fas fa-image text-4xl" />
          </div>
        </div>
        <div v-if="house.gallery && house.gallery.length > 1" class="mt-3 flex gap-2 overflow-x-auto">
          <button
            v-for="(p, i) in house.gallery"
            :key="i"
            class="w-20 h-20 rounded-xl overflow-hidden flex-shrink-0 border-2 transition-all"
            :class="i === activeIdx ? 'border-brand' : 'border-transparent opacity-70 hover:opacity-100'"
            @click="activeIdx = i"
          >
            <img :src="p" class="w-full h-full object-cover" />
          </button>
        </div>

        <h1 class="mt-6 text-2xl md:text-3xl font-bold text-ink">{{ house.title }}</h1>
        <p class="text-ink-soft mt-1">
          <i class="fas fa-location-dot mr-1" />{{ house.address || house.district }}
        </p>

        <div class="mt-6 grid grid-cols-2 sm:grid-cols-4 gap-3">
          <div class="p-3 rounded-xl bg-page text-center">
            <div class="text-xs text-ink-soft">Ҳуҷраҳо</div>
            <div class="font-bold text-ink mt-0.5">{{ house.rooms || "—" }}</div>
          </div>
          <div class="p-3 rounded-xl bg-page text-center">
            <div class="text-xs text-ink-soft">Масоҳат</div>
            <div class="font-bold text-ink mt-0.5">{{ house.area || "—" }} м²</div>
          </div>
          <div class="p-3 rounded-xl bg-page text-center">
            <div class="text-xs text-ink-soft">Ошёна</div>
            <div class="font-bold text-ink mt-0.5">{{ house.floor || "—" }}/{{ house.total_floors || "—" }}</div>
          </div>
          <div class="p-3 rounded-xl bg-page text-center">
            <div class="text-xs text-ink-soft">Аз нархи</div>
            <div class="font-bold text-brand mt-0.5">{{ fmt(house.price_from) }}</div>
          </div>
        </div>

        <div v-if="house.full_desc" class="mt-6 prose prose-sm max-w-none whitespace-pre-line">
          {{ house.full_desc }}
        </div>

        <div v-if="house.features && house.features.length" class="mt-6 grid grid-cols-2 md:grid-cols-3 gap-2">
          <span
            v-for="(f, i) in house.features"
            :key="i"
            class="inline-flex items-center gap-2 px-3 py-2 rounded-lg bg-brand-soft text-brand text-sm"
          >
            <i class="fas fa-check" /> {{ f }}
          </span>
        </div>

        <NuxtLink
          v-if="house.has_tour && house.tour_slug"
          :to="`/tour/${house.tour_slug}`"
          class="mt-8 inline-flex items-center gap-2 px-6 py-3 rounded-xl bg-brand text-white font-semibold hover:bg-brand-dark"
        >
          <i class="fas fa-vr-cardboard" /> Тури виртуалӣ 360° кушоед
        </NuxtLink>
      </div>

      <!-- contact card -->
      <aside class="lg:sticky lg:top-20 h-fit bg-white rounded-2xl border border-border-soft p-5">
        <h3 class="font-bold text-ink">Дархости занг</h3>
        <p class="text-sm text-ink-soft mt-1">Дар 5 дақиқа занг мезанем</p>

        <form v-if="!sent" class="mt-4 space-y-3" @submit.prevent="sendLead">
          <input
            v-model="lead.name"
            placeholder="Ном"
            class="w-full px-4 py-3 rounded-xl border border-border-soft focus:outline-none focus:border-brand"
          />
          <input
            v-model="lead.phone"
            type="tel"
            required
            placeholder="+992 ___ ___ ___"
            class="w-full px-4 py-3 rounded-xl border border-border-soft focus:outline-none focus:border-brand"
          />
          <textarea
            v-model="lead.message"
            rows="3"
            placeholder="Савол (ихтиёрӣ)"
            class="w-full px-4 py-3 rounded-xl border border-border-soft focus:outline-none focus:border-brand"
          />
          <button
            :disabled="sending"
            class="w-full py-3 rounded-xl bg-brand text-white font-semibold hover:bg-brand-dark disabled:opacity-60"
          >
            <i v-if="sending" class="fas fa-spinner fa-spin mr-2" />
            {{ sending ? "Ирсол…" : "Дархост ирсол кунед" }}
          </button>
        </form>
        <div v-else class="mt-4 p-4 rounded-xl bg-brand-soft text-brand text-sm">
          <i class="fas fa-circle-check mr-1" />Раҳмат! Ба зудӣ занг мезанем.
        </div>
      </aside>
    </div>
  </div>
</template>
