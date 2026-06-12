<script setup lang="ts">
import type { PublicHouse } from "~/types/public";

defineProps<{ house: PublicHouse }>();

const fmt = (n: number) =>
  n ? new Intl.NumberFormat("ru-RU").format(Math.round(n)) : "—";
</script>

<template>
  <NuxtLink
    :to="`/properties/${house.id}`"
    class="group block rounded-2xl overflow-hidden bg-white border border-border-soft hover:shadow-hover hover:-translate-y-1 transition-all"
  >
    <div class="relative aspect-[4/3] bg-page overflow-hidden">
      <img
        v-if="house.gallery?.[0]"
        :src="house.gallery[0]"
        :alt="house.title"
        class="w-full h-full object-cover group-hover:scale-105 transition-transform duration-500"
        loading="lazy"
      />
      <div v-else class="w-full h-full grid place-items-center text-ink-soft">
        <i class="fas fa-image text-3xl" />
      </div>
      <span
        v-if="house.has_tour"
        class="absolute top-3 left-3 inline-flex items-center gap-1.5 px-2.5 py-1 rounded-full bg-brand text-white text-[11px] font-semibold shadow-lg"
      >
        <i class="fas fa-vr-cardboard text-[10px]" /> Тури 360°
      </span>
      <span
        v-if="house.price_from"
        class="absolute bottom-3 left-3 px-3 py-1.5 rounded-full bg-white/95 text-ink font-bold text-sm shadow"
      >
        {{ fmt(house.price_from) }} сомонӣ
      </span>
    </div>

    <div class="p-4">
      <h3 class="font-bold text-base text-ink line-clamp-1">{{ house.title }}</h3>
      <p class="text-xs text-ink-soft mt-0.5 line-clamp-1">
        <i class="fas fa-location-dot mr-1" />{{ house.district || house.address || "—" }}
      </p>
      <div class="mt-3 flex items-center gap-3 text-xs text-ink-medium">
        <span><i class="fas fa-bed mr-1 text-brand" />{{ house.rooms || "—" }}</span>
        <span><i class="fas fa-vector-square mr-1 text-brand" />{{ house.area || "—" }} м²</span>
        <span v-if="house.floor">
          <i class="fas fa-stairs mr-1 text-brand" />{{ house.floor }}/{{ house.total_floors }}
        </span>
      </div>
    </div>
  </NuxtLink>
</template>
