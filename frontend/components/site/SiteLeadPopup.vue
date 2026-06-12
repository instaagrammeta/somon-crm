<script setup lang="ts">
/**
 * Auto-shown popup that asks for name + phone after the visitor has spent at
 * least `delaySec` seconds on the public site. Shown ONCE per session — once
 * dismissed or submitted we set sessionStorage so it does not nag.
 *
 * Body fields are kept tiny on purpose: the more questions, the lower the
 * conversion. Just name + phone, no email, no captcha.
 */
const delaySec = 30;
const open = ref(false);
const sent = ref(false);
const submitting = ref(false);
const error = ref<string | null>(null);

const name = ref("");
const phone = ref("");

const STORAGE_KEY = "somon_lead_popup_seen";

let timer: ReturnType<typeof setTimeout> | null = null;

const start = () => {
  if (typeof window === "undefined") return;
  if (sessionStorage.getItem(STORAGE_KEY) === "1") return;
  if (timer) clearTimeout(timer);
  timer = setTimeout(() => (open.value = true), delaySec * 1000);
};

const dismiss = () => {
  open.value = false;
  if (typeof window !== "undefined") sessionStorage.setItem(STORAGE_KEY, "1");
};

const api = usePublicApi();
const route = useRoute();

const submit = async () => {
  if (submitting.value) return;
  if (!phone.value || phone.value.replace(/\D/g, "").length < 6) {
    error.value = "Рақами телефон лозим";
    return;
  }
  submitting.value = true;
  error.value = null;
  try {
    const params = new URLSearchParams(window.location.search);
    await api.post("/api/public/leads", {
      name: name.value.trim(),
      phone: phone.value.trim(),
      source: "popup",
      page: window.location.pathname,
      referrer: document.referrer,
      utm_source: params.get("utm_source") || "",
      utm_medium: params.get("utm_medium") || "",
      utm_campaign: params.get("utm_campaign") || "",
    });
    sent.value = true;
    if (typeof window !== "undefined") sessionStorage.setItem(STORAGE_KEY, "1");
    setTimeout(() => (open.value = false), 2500);
  } catch (e: any) {
    error.value = e?.data?.error || "Хатогии ирсол";
  } finally {
    submitting.value = false;
  }
};

onMounted(start);
onBeforeUnmount(() => timer && clearTimeout(timer));

// If the route changes we keep the timer running (single global popup).
watch(() => route.fullPath, () => {
  if (!open.value && !sent.value) start();
});
</script>

<template>
  <Transition name="fade">
    <div v-if="open" class="fixed inset-0 z-40 flex items-end md:items-center justify-center p-4 bg-black/40 backdrop-blur-sm">
      <div class="w-full max-w-md bg-white rounded-2xl shadow-2xl overflow-hidden">
        <div v-if="!sent" class="p-6">
          <div class="flex items-start justify-between mb-3">
            <div>
              <div class="w-12 h-12 rounded-full bg-brand-soft text-brand grid place-items-center text-xl">
                <i class="fas fa-house-circle-check" />
              </div>
              <h3 class="mt-3 font-bold text-lg text-ink">Биёед бо шумо тамос гирем</h3>
              <p class="text-sm text-ink-soft">
                Танҳо ном ва рақами шумо. Пас аз 5 дақиқа занг мезанем.
              </p>
            </div>
            <button class="text-ink-soft hover:text-ink" @click="dismiss">
              <i class="fas fa-xmark text-lg" />
            </button>
          </div>

          <form class="space-y-3" @submit.prevent="submit">
            <input
              v-model="name"
              type="text"
              placeholder="Ном (ихтиёрӣ)"
              class="w-full px-4 py-3 rounded-xl border border-border-soft focus:outline-none focus:border-brand focus:ring-2 focus:ring-brand/15"
            />
            <input
              v-model="phone"
              type="tel"
              required
              placeholder="+992 ___ ___ ___"
              class="w-full px-4 py-3 rounded-xl border border-border-soft focus:outline-none focus:border-brand focus:ring-2 focus:ring-brand/15"
            />
            <p v-if="error" class="text-sm text-red-500">{{ error }}</p>
            <button
              type="submit"
              :disabled="submitting"
              class="w-full py-3 rounded-xl bg-brand text-white font-semibold hover:bg-brand-dark disabled:opacity-60"
            >
              <i v-if="submitting" class="fas fa-spinner fa-spin mr-2" />
              {{ submitting ? "Ирсол шуда истодааст…" : "Тамос гиред бо ман" }}
            </button>
            <button
              type="button"
              class="w-full py-2 text-sm text-ink-soft hover:text-ink"
              @click="dismiss"
            >
              Шояд дертар
            </button>
          </form>
        </div>

        <div v-else class="p-6 text-center">
          <div class="w-14 h-14 rounded-full bg-brand-soft text-brand grid place-items-center text-2xl mx-auto">
            <i class="fas fa-circle-check" />
          </div>
          <h3 class="mt-3 font-bold text-lg text-ink">Раҳмат!</h3>
          <p class="text-sm text-ink-soft mt-1">Дар наздиктарин вақт бо шумо тамос мегирем.</p>
        </div>
      </div>
    </div>
  </Transition>
</template>

<style scoped>
.fade-enter-active,
.fade-leave-active {
  transition: opacity 0.25s ease;
}
.fade-enter-from,
.fade-leave-to {
  opacity: 0;
}
</style>
