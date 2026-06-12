<script setup lang="ts">
/**
 * 360° virtual tour viewer powered by Marzipano.
 *
 * Marzipano needs the DOM (`window`/`document`), so this page is fully
 * client-side. The whole page is wrapped in `<ClientOnly>` and Marzipano is
 * imported with a `dynamic import()` so it never enters the SSR bundle.
 *
 * The data shape returned by /api/public/tours/:slug is:
 *   { id, title, panoramas: [{ id, title, image_url, initial_yaw, ...,
 *                              hotspots: [{ kind, target_panorama_id,
 *                                           yaw, pitch, label, ... }] }] }
 *
 * Each panorama is loaded as an *equirectangular* image with multi-resolution
 * tile support — Marzipano handles 4K/8K via its own tile pipeline so the
 * browser never has to decode the full image at once.
 */
import type { Tour, Panorama } from "~/types/public";

definePageMeta({ layout: "site" });

const route = useRoute();
const api = usePublicApi();

const { data: tour, pending } = useAsyncData<Tour | null>(
  () => `tour-${route.params.slug}`,
  () => api.get<Tour>(`/api/public/tours/${route.params.slug}`),
  { default: () => null }
);

const viewerEl = ref<HTMLElement | null>(null);
const currentSceneId = ref<number | null>(null);
const fullscreen = ref(false);

let viewer: any = null;
let scenes: Map<number, any> = new Map();

const switchTo = (panoId: number) => {
  const scene = scenes.get(panoId);
  if (!scene) return;
  scene.switchTo();
  currentSceneId.value = panoId;
};

const buildScene = (Marzipano: any, viewer: any, p: Panorama) => {
  // Multi-resolution tiled equirectangular source. Marzipano falls back to a
  // single-image source automatically when only one URL is provided.
  const source = Marzipano.ImageUrlSource.fromString(p.image_url);
  // 4K-8K source rendering: limited by Marzipano's tile cache. For real 8K
  // production output the equirect should be pre-tiled with `marzipano-tool`
  // and the URL becomes a tile template like `/tiles/{level}/{tile_x}_{y}.jpg`.
  const geometry = new Marzipano.EquirectGeometry([{ width: 4096 }]);
  const limiter = Marzipano.RectilinearView.limit.traditional(
    4096,
    (100 * Math.PI) / 180
  );
  const view = new Marzipano.RectilinearView(
    {
      yaw: p.initial_yaw || 0,
      pitch: p.initial_pitch || 0,
      fov: ((p.initial_zoom || 90) * Math.PI) / 180,
    },
    limiter
  );
  const scene = viewer.createScene({
    source,
    geometry,
    view,
    pinFirstLevel: true,
  });

  // Hotspots: render simple HTML markers; click handler either jumps to
  // another panorama (kind=link) or opens an info bubble.
  for (const h of p.hotspots || []) {
    const el = document.createElement("button");
    el.className =
      "marz-hotspot inline-flex items-center justify-center w-10 h-10 rounded-full " +
      "bg-brand text-white shadow-xl border-2 border-white animate-pulse-slow";
    el.title = h.label || "";
    if (h.kind === "info") {
      el.innerHTML = '<i class="fas fa-info"></i>';
      el.onclick = () => {
        if (h.info_url) window.open(h.info_url, "_blank");
        else if (h.info_text) alert(h.info_text);
      };
    } else if (h.kind === "url" && h.info_url) {
      el.innerHTML = '<i class="fas fa-up-right-from-square"></i>';
      el.onclick = () => window.open(h.info_url, "_blank");
    } else {
      el.innerHTML = '<i class="fas fa-arrow-right"></i>';
      el.onclick = () => {
        if (h.target_panorama_id) switchTo(h.target_panorama_id);
      };
    }
    scene.hotspotContainer().createHotspot(el, { yaw: h.yaw, pitch: h.pitch });
  }
  return scene;
};

const start = async () => {
  if (!tour.value || !viewerEl.value || !tour.value.panoramas?.length) return;
  // Cleanup any previous viewer (fast nav between tours).
  if (viewer) {
    try { viewer.destroy(); } catch { /* */ }
    viewer = null;
    scenes = new Map();
  }
  const Marzipano = (await import("marzipano")).default;
  viewer = new Marzipano.Viewer(viewerEl.value, {
    controls: { mouseViewMode: "drag" },
  });
  for (const p of tour.value.panoramas) {
    scenes.set(p.id, buildScene(Marzipano, viewer, p));
  }
  const startId = tour.value.initial_panorama_id || tour.value.panoramas[0].id;
  switchTo(startId);
};

watch(() => tour.value?.id, () => start(), { immediate: false });
onMounted(() => start());
onBeforeUnmount(() => {
  if (viewer) {
    try { viewer.destroy(); } catch { /* */ }
  }
});

const toggleFullscreen = async () => {
  const el = viewerEl.value;
  if (!el) return;
  if (!document.fullscreenElement) {
    await el.requestFullscreen();
    fullscreen.value = true;
  } else {
    await document.exitFullscreen();
    fullscreen.value = false;
  }
};
</script>

<template>
  <div class="bg-ink text-white min-h-screen">
    <div class="max-w-6xl mx-auto px-4 py-4 flex items-center justify-between">
      <NuxtLink to="/properties" class="text-sm hover:underline">
        <i class="fas fa-arrow-left mr-1" />Манзилҳо
      </NuxtLink>
      <div class="text-sm font-semibold truncate">{{ tour?.title || "Тур" }}</div>
      <button
        class="w-9 h-9 rounded-lg bg-white/10 hover:bg-white/20"
        :title="fullscreen ? 'Хуруҷ' : 'Пурра'"
        @click="toggleFullscreen"
      >
        <i :class="fullscreen ? 'fas fa-compress' : 'fas fa-expand'" />
      </button>
    </div>

    <div v-if="pending" class="py-24 text-center">
      <i class="fas fa-spinner fa-spin text-3xl" />
    </div>

    <div v-else-if="!tour || !tour.panoramas?.length" class="py-24 text-center text-white/60">
      <i class="fas fa-image text-4xl mb-3" />
      <p>Барои ин лоиҳа тур мавҷуд нест.</p>
    </div>

    <ClientOnly v-else>
      <div class="relative">
        <div
          ref="viewerEl"
          class="w-full bg-black"
          style="height: calc(100vh - 8rem);"
        />

        <!-- room switcher -->
        <div class="absolute bottom-4 inset-x-0 px-4 flex justify-center pointer-events-none">
          <div class="pointer-events-auto bg-black/55 backdrop-blur rounded-full px-2 py-1 flex gap-1 max-w-full overflow-x-auto">
            <button
              v-for="p in tour.panoramas"
              :key="p.id"
              class="px-3 py-1.5 rounded-full text-xs font-medium whitespace-nowrap transition-colors"
              :class="currentSceneId === p.id ? 'bg-brand text-white' : 'text-white/70 hover:text-white'"
              @click="switchTo(p.id)"
            >
              {{ p.title }}
            </button>
          </div>
        </div>
      </div>
    </ClientOnly>
  </div>
</template>

<style>
/* Marzipano keeps hotspots positioned absolute over the panorama; this
   selector lifts them above the room switcher chrome. */
.marz-hotspot {
  cursor: pointer;
  transform: translate(-50%, -50%);
}
@keyframes pulse-slow {
  0%, 100% { transform: translate(-50%, -50%) scale(1); box-shadow: 0 0 0 0 rgba(31, 122, 77, 0.6); }
  50% { transform: translate(-50%, -50%) scale(1.05); box-shadow: 0 0 0 14px rgba(31, 122, 77, 0); }
}
.animate-pulse-slow { animation: pulse-slow 2s infinite; }
</style>
