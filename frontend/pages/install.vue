<script setup lang="ts">
/**
 * APK download landing page. The actual file is served by the backend at
 * /api/public/apk which redirects to whatever SITE_APK_URL points at —
 * usually a static file under /static/apk/somon-crm.apk.
 *
 * On iOS we simply show the PWA install instructions because there is no APK.
 */
import type { SiteConfig } from "~/types/public";

definePageMeta({ layout: "site" });

const api = usePublicApi();
const apkURL = computed(() => api.baseURL + "/api/public/apk");

const config = inject<Ref<SiteConfig>>("siteConfig", ref({} as SiteConfig));

// Detect platform for adaptive instructions.
const isIOS = ref(false);
const isAndroid = ref(false);
onMounted(() => {
  const ua = navigator.userAgent;
  isIOS.value = /iPad|iPhone|iPod/.test(ua);
  isAndroid.value = /Android/.test(ua);
});

// PWA install prompt (Chrome / Edge fire `beforeinstallprompt`).
const installEvt = ref<any>(null);
onMounted(() => {
  window.addEventListener("beforeinstallprompt", (e: any) => {
    e.preventDefault();
    installEvt.value = e;
  });
});
const installPWA = async () => {
  if (!installEvt.value) return;
  installEvt.value.prompt();
  await installEvt.value.userChoice;
  installEvt.value = null;
};
</script>

<template>
  <div class="max-w-3xl mx-auto px-4 py-12">
    <h1 class="text-3xl font-bold text-ink">Замимаи мобилӣ</h1>
    <p class="mt-2 text-ink-soft">
      Somon Real Estate-ро ба телефони худ зер кунед — манзилҳо ҳамеша ҳамроҳи шумо.
    </p>

    <div class="mt-8 grid md:grid-cols-2 gap-5">
      <!-- Android APK -->
      <div class="p-6 rounded-2xl bg-white border border-border-soft">
        <div class="w-12 h-12 rounded-xl bg-brand-soft text-brand grid place-items-center text-xl">
          <i class="fab fa-android" />
        </div>
        <h3 class="mt-3 font-bold text-lg">Android (APK)</h3>
        <p class="mt-1 text-sm text-ink-soft">
          Версияи {{ config?.apk_version || "1.0.0" }} — мустақим аз вебсайт.
        </p>
        <a
          :href="apkURL"
          class="mt-4 inline-flex items-center gap-2 px-5 py-3 rounded-xl bg-brand text-white font-semibold hover:bg-brand-dark"
        >
          <i class="fas fa-download" /> APK-ро зер кашед
        </a>
        <p class="mt-3 text-xs text-ink-soft">
          Маслиҳат: дар Андроид аз "Танзимот → Бехатарӣ → Манбаи номаълум" иҷозат
          диҳед, то насб шавад.
        </p>
      </div>

      <!-- PWA / iOS -->
      <div class="p-6 rounded-2xl bg-white border border-border-soft">
        <div class="w-12 h-12 rounded-xl bg-brand-soft text-brand grid place-items-center text-xl">
          <i :class="isIOS ? 'fab fa-apple' : 'fas fa-globe'" />
        </div>
        <h3 class="mt-3 font-bold text-lg">{{ isIOS ? "iOS (Safari)" : "Web App (PWA)" }}</h3>
        <p class="mt-1 text-sm text-ink-soft">
          Барои iOS — насби APK имконнопазир аст; ин вебсайтро ҳамчун Web App насб кунед.
        </p>
        <button
          v-if="installEvt"
          class="mt-4 inline-flex items-center gap-2 px-5 py-3 rounded-xl bg-brand text-white font-semibold hover:bg-brand-dark"
          @click="installPWA"
        >
          <i class="fas fa-circle-down" /> Web App-ро насб кунед
        </button>
        <ol v-else class="mt-4 text-sm text-ink-soft list-decimal pl-4 space-y-1">
          <li>Дар Safari тугмаи "Share" (▲)-ро зер кунед</li>
          <li>"Add to Home Screen"-ро интихоб намоед</li>
          <li>"Add"-ро зер кунед</li>
        </ol>
      </div>
    </div>

    <div class="mt-10 p-5 rounded-2xl bg-brand-soft/40 border border-brand/20 text-sm text-ink-medium">
      <i class="fas fa-shield-halved text-brand mr-1" />
      Ҳамаи нусхаҳо аз сервери расмии Somon аз тариқи HTTPS дода мешаванд.
    </div>
  </div>
</template>
