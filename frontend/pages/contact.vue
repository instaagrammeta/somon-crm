<script setup lang="ts">
import type { SiteConfig } from "~/types/public";

definePageMeta({ layout: "site" });

const api = usePublicApi();
const config = inject<Ref<SiteConfig>>("siteConfig", ref({} as SiteConfig));

const sending = ref(false);
const sent = ref(false);
const error = ref<string | null>(null);
const form = reactive({ name: "", phone: "", message: "" });

const submit = async () => {
  error.value = null;
  if (!form.phone) {
    error.value = "Рақами телефонро ворид кунед";
    return;
  }
  sending.value = true;
  try {
    await api.post("/api/public/leads", {
      ...form,
      source: "contact_form",
      page: "/contact",
      referrer: document.referrer,
    });
    sent.value = true;
  } catch (e: any) {
    error.value = e?.data?.error || "Хатогии ирсол";
  } finally {
    sending.value = false;
  }
};
</script>

<template>
  <div class="max-w-3xl mx-auto px-4 py-12">
    <h1 class="text-3xl font-bold text-ink">Тамос бо мо</h1>
    <p class="mt-2 text-ink-soft">
      Шумо метавонед бо роҳи зерин бо мо тамос гиред — ё формаро пур кунед.
    </p>

    <div class="mt-8 grid md:grid-cols-2 gap-8">
      <div class="space-y-4">
        <a
          v-if="config?.phone"
          :href="`tel:${config.phone}`"
          class="flex items-center gap-3 p-4 rounded-2xl bg-white border border-border-soft hover:border-brand"
        >
          <span class="w-10 h-10 rounded-xl bg-brand-soft text-brand grid place-items-center"><i class="fas fa-phone" /></span>
          <div>
            <div class="text-xs text-ink-soft">Телефон</div>
            <div class="font-semibold text-ink">{{ config.phone }}</div>
          </div>
        </a>
        <a
          v-if="config?.email"
          :href="`mailto:${config.email}`"
          class="flex items-center gap-3 p-4 rounded-2xl bg-white border border-border-soft hover:border-brand"
        >
          <span class="w-10 h-10 rounded-xl bg-brand-soft text-brand grid place-items-center"><i class="fas fa-envelope" /></span>
          <div>
            <div class="text-xs text-ink-soft">Email</div>
            <div class="font-semibold text-ink">{{ config.email }}</div>
          </div>
        </a>
        <div
          v-if="config?.address"
          class="flex items-center gap-3 p-4 rounded-2xl bg-white border border-border-soft"
        >
          <span class="w-10 h-10 rounded-xl bg-brand-soft text-brand grid place-items-center"><i class="fas fa-location-dot" /></span>
          <div>
            <div class="text-xs text-ink-soft">Суроға</div>
            <div class="font-semibold text-ink">{{ config.address }}</div>
          </div>
        </div>
      </div>

      <form
        v-if="!sent"
        class="bg-white p-6 rounded-2xl border border-border-soft space-y-3"
        @submit.prevent="submit"
      >
        <input
          v-model="form.name"
          placeholder="Ном"
          class="w-full px-4 py-3 rounded-xl border border-border-soft focus:outline-none focus:border-brand"
        />
        <input
          v-model="form.phone"
          type="tel"
          required
          placeholder="+992 ___ ___ ___"
          class="w-full px-4 py-3 rounded-xl border border-border-soft focus:outline-none focus:border-brand"
        />
        <textarea
          v-model="form.message"
          rows="4"
          placeholder="Паём…"
          class="w-full px-4 py-3 rounded-xl border border-border-soft focus:outline-none focus:border-brand"
        />
        <p v-if="error" class="text-sm text-red-500">{{ error }}</p>
        <button
          :disabled="sending"
          class="w-full py-3 rounded-xl bg-brand text-white font-semibold hover:bg-brand-dark disabled:opacity-60"
        >
          <i v-if="sending" class="fas fa-spinner fa-spin mr-2" />
          {{ sending ? "Ирсол…" : "Ирсол" }}
        </button>
      </form>
      <div v-else class="bg-white p-6 rounded-2xl border border-border-soft">
        <div class="w-12 h-12 rounded-full bg-brand-soft text-brand grid place-items-center text-xl">
          <i class="fas fa-circle-check" />
        </div>
        <h3 class="mt-3 font-bold text-lg">Раҳмат!</h3>
        <p class="text-sm text-ink-soft">Дар наздиктарин вақт бо шумо тамос мегирем.</p>
      </div>
    </div>
  </div>
</template>
