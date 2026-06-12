<script setup lang="ts">
/**
 * Admin dashboard for leads captured by the public website + WhatsApp.
 * Operators triage them (status: new → contacted → qualified | spam) and
 * "promote" qualified ones into proper CRM leads (`Lid`) so they enter
 * the sales pipeline.
 */
useHead({ title: "Лидҳои вебсайт" });

interface WebsiteLead {
  id: number;
  created_at: string;
  name: string;
  phone: string;
  message: string;
  source: string;
  page: string;
  status: string;
  ai_score: number;
  ai_summary: string;
  notes: string;
  promoted_lead_id?: number | null;
}

const api = useApi();
const toast = useToast();

const filters = reactive({ status: "", source: "", q: "" });

const queryString = computed(() => {
  const p = new URLSearchParams();
  for (const [k, v] of Object.entries(filters)) if (v) p.set(k, v as string);
  return p.toString();
});

const { data: rows, pending, refresh } = useAsyncData<WebsiteLead[]>(
  "website-leads",
  () => api.get<WebsiteLead[]>("/api/website-leads?" + queryString.value),
  { watch: [queryString], default: () => [] as WebsiteLead[] }
);

const setStatus = async (r: WebsiteLead, status: string) => {
  await api.put(`/api/website-leads/${r.id}`, { status });
  await refresh();
};

const promote = async (r: WebsiteLead) => {
  if (r.promoted_lead_id) {
    toast.info?.("Аллакай тавсеа дода шуд");
    return;
  }
  if (!confirm("Лидро ба CRM-pipeline тавсеа диҳем?")) return;
  await api.post(`/api/website-leads/${r.id}/promote`, {});
  toast.success("Тавсеа дода шуд");
  await refresh();
};

const rescore = async (r: WebsiteLead) => {
  toast.info?.("AI ҳисоб карда истодааст…");
  try {
    await api.post(`/api/ai/score-lead/${r.id}`, {});
    await refresh();
    toast.success("AI хол гузошт");
  } catch (e: any) {
    toast.error(e?.data?.error || "AI дастрас нест");
  }
};

const remove = async (r: WebsiteLead) => {
  if (!confirm("Ҳазф?")) return;
  await api.del(`/api/website-leads/${r.id}`);
  await refresh();
};

const sourceIcon = (s: string) =>
  ({ popup: "fa-bullhorn", contact_form: "fa-envelope", house_card: "fa-house", whatsapp: "fa-whatsapp", tour: "fa-vr-cardboard" } as any)[s] || "fa-circle";

const scoreColor = (n: number) =>
  n >= 70 ? "text-red-600 bg-red-100" : n >= 40 ? "text-amber-700 bg-amber-100" : "text-ink-soft bg-page";
</script>

<template>
  <div>
    <div class="flex items-end justify-between mb-6 flex-wrap gap-4">
      <div>
        <h1 class="text-[28px] font-bold text-ink">Лидҳои вебсайт</h1>
        <p class="text-ink-soft text-sm">Лидҳо аз popup, формаи тамос, WhatsApp ва турҳо</p>
      </div>
      <button class="btn-outline" @click="refresh()"><i class="fas fa-rotate" /> Навсозӣ</button>
    </div>

    <div class="bg-white p-4 rounded-2xl border border-border-soft mb-5 grid sm:grid-cols-3 gap-3">
      <select v-model="filters.status" class="input">
        <option value="">Ҳама ҳолатҳо</option>
        <option value="new">Нав</option>
        <option value="contacted">Тамос гирифташуда</option>
        <option value="qualified">Тасдиқшуда</option>
        <option value="spam">Спам</option>
        <option value="promoted">Тавсеашуда</option>
      </select>
      <select v-model="filters.source" class="input">
        <option value="">Ҳама манбаъҳо</option>
        <option value="popup">Popup</option>
        <option value="contact_form">Формаи тамос</option>
        <option value="house_card">Картаи манзил</option>
        <option value="tour">Тур 360°</option>
        <option value="whatsapp">WhatsApp</option>
      </select>
      <input v-model="filters.q" placeholder="Ҷустуҷӯ бо ном/телефон…" class="input" />
    </div>

    <UiLoader v-if="pending" />
    <UiEmpty v-else-if="!rows?.length" icon="fa-inbox" />

    <div v-else class="space-y-3">
      <div
        v-for="r in rows"
        :key="r.id"
        class="bg-white rounded-2xl border border-border-soft p-4 grid md:grid-cols-[1fr_auto] gap-3"
      >
        <div>
          <div class="flex items-center gap-2 flex-wrap">
            <span class="font-bold text-ink">{{ r.name || "Бе ном" }}</span>
            <a :href="`tel:${r.phone}`" class="text-brand font-mono text-sm hover:underline">
              <i class="fas fa-phone text-xs mr-1" />{{ r.phone }}
            </a>
            <span class="inline-flex items-center gap-1 text-xs text-ink-soft">
              <i class="fas" :class="sourceIcon(r.source)" /> {{ r.source }}
            </span>
            <span class="text-xs text-ink-soft">{{ new Date(r.created_at).toLocaleString("ru-RU") }}</span>
          </div>
          <p v-if="r.message" class="mt-1 text-sm text-ink-medium">{{ r.message }}</p>
          <p v-if="r.page" class="mt-1 text-xs text-ink-soft truncate">
            <i class="fas fa-link text-[10px] mr-1" />{{ r.page }}
          </p>
          <div v-if="r.ai_score > 0 || r.ai_summary" class="mt-2 flex items-start gap-2 flex-wrap">
            <span class="px-2 py-0.5 rounded-full text-xs font-bold" :class="scoreColor(r.ai_score)">
              <i class="fas fa-robot mr-1" />AI: {{ r.ai_score }}/100
            </span>
            <span v-if="r.ai_summary" class="text-xs text-ink-soft">{{ r.ai_summary }}</span>
          </div>
        </div>

        <div class="flex flex-wrap gap-2 items-start">
          <select :value="r.status" class="input !py-1.5 !w-auto" @change="setStatus(r, ($event.target as HTMLSelectElement).value)">
            <option value="new">🟢 Нав</option>
            <option value="contacted">📞 Тамос</option>
            <option value="qualified">✅ Тасдиқ</option>
            <option value="spam">🚫 Спам</option>
            <option value="promoted">⭐ Тавсеа</option>
          </select>
          <button class="btn-soft" :title="'AI score'" @click="rescore(r)">
            <i class="fas fa-robot" />
          </button>
          <button
            class="btn-primary !py-1.5"
            :disabled="!!r.promoted_lead_id"
            :title="r.promoted_lead_id ? 'Аллакай ' : 'Тавсеа ба CRM'"
            @click="promote(r)"
          >
            <i class="fas fa-arrow-up-right-from-square" /> Тавсеа
          </button>
          <button class="w-9 h-9 rounded-lg text-red-500 hover:bg-red-100" title="Ҳазф" @click="remove(r)">
            <i class="fas fa-trash" />
          </button>
        </div>
      </div>
    </div>
  </div>
</template>
