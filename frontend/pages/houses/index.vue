<script setup lang="ts">
import type { House } from "~/types";
import { formatNumber } from "~/utils/format";

useHead({ title: () => useI18n().t("nav.houses") });

const { t } = useI18n();
const api = useApi();
const toast = useToast();

const constructionTypes = [
  { value: "новостройка", label: "🏗️ Новостройка" },
  { value: "коробка", label: "📦 Коробка" },
  { value: "котлован", label: "🕳️ Котлован" },
  { value: "Вторичка", label: "Вторичка" },
];
const districts = [
  { value: "н.Сино", label: "Сино" },
  { value: "н.Шоҳмансур", label: "Шоҳмансур" },
  { value: "н.Фирдавси", label: "Фирдавси" },
  { value: "н.Сомони", label: "Сомони" },
];
const typeIcon = (t: string) => (t === "новостройка" ? "🏗️" : t === "коробка" ? "📦" : t === "котлован" ? "🕳️" : "🏢");

const { data: houses, pending, refresh } = useAsyncData<House[]>("houses", () => api.get<House[]>("/api/houses"), { lazy: true, default: () => [] });

const fl = reactive({ search: "", type: "all", district: "all" });
const filtered = computed(() => {
  let list = houses.value || [];
  if (fl.search) {
    const q = fl.search.toLowerCase();
    list = list.filter((h) => [h.title, h.address, h.developer].some((x) => (x || "").toLowerCase().includes(q)));
  }
  if (fl.type !== "all") list = list.filter((h) => h.construction_type === fl.type);
  if (fl.district !== "all") list = list.filter((h) => h.district === fl.district);
  return list;
});

const modal = ref(false);
const blank = (): Partial<House> => ({ construction_type: "новостройка", district: "н.Сино", has_tech_passport: "нет", has_renovation_permit: "нет" });
const editing = ref<Partial<House>>(blank());
const files = ref<File[]>([]);

const openCreate = () => { editing.value = blank(); files.value = []; modal.value = true; };
const openEdit = (h: House) => { editing.value = { ...h }; files.value = []; modal.value = true; };
const onFiles = (e: Event) => { files.value = Array.from((e.target as HTMLInputElement).files || []); };

const save = async () => {
  const e = editing.value;
  if (!e.title || !e.address) { toast.error(t("notify.error")); return; }
  const fd = new FormData();
  const fields: (keyof House)[] = ["title", "construction_type", "district", "address", "area", "rooms", "windows", "floor", "total_floors", "price_per_m2", "total_price", "developer", "contact_phone", "has_tech_passport", "has_renovation_permit"];
  fields.forEach((k) => { if (e[k] != null) fd.append(k as string, String(e[k])); });
  files.value.forEach((f) => fd.append("files", f));
  try {
    if (e.id) { await api.upload(`/api/houses/${e.id}`, fd); toast.success(t("notify.updated")); }
    else { await api.upload("/api/houses", fd); toast.success(t("notify.created")); }
    modal.value = false;
    await refresh();
  } catch (err: any) { toast.error(err?.data?.error || t("notify.error")); }
};

const confirmDel = ref<{ open: boolean; id?: number }>({ open: false });
const askDelete = (h: House) => (confirmDel.value = { open: true, id: h.id });
const doDelete = async () => {
  if (!confirmDel.value.id) return;
  await api.del(`/api/houses/${confirmDel.value.id}`);
  toast.success(t("notify.deleted"));
  await refresh();
};
const exportXlsx = () => api.download("/api/houses/export", "houses.xlsx");

const firstImage = (h: House) => (h.files || []).find((f: any) => typeof f === "string" ? /\.(png|jpe?g|webp|gif)$/i.test(f) : false) as string | undefined;
</script>

<template>
  <div>
    <div class="flex justify-between items-center flex-wrap gap-4 mb-6">
      <div>
        <h1 class="text-[28px] font-bold text-ink mb-1">{{ t("nav.houses") }}</h1>
        <p class="text-ink-soft text-sm">{{ t("house.subtitle") }}</p>
      </div>
      <div class="flex gap-2">
        <button class="btn-outline" @click="exportXlsx"><i class="fas fa-file-excel" /> Excel</button>
        <button class="btn-primary" @click="openCreate"><i class="fas fa-plus" /> {{ t("house.add") }}</button>
      </div>
    </div>

    <!-- filters -->
    <div class="flex flex-wrap gap-2.5 mb-6">
      <div class="flex items-center gap-1.5 bg-page rounded-xl px-3 py-2 text-sm flex-1 min-w-[180px]">
        <i class="fas fa-search text-ink-soft" />
        <input v-model="fl.search" :placeholder="t('app.search')" class="flex-1 bg-transparent focus:outline-none" />
      </div>
      <select v-model="fl.type" class="bg-page rounded-xl px-3 py-2 text-sm focus:outline-none">
        <option value="all">{{ t("house.construction_type") }}: {{ t("app.all") }}</option>
        <option v-for="c in constructionTypes" :key="c.value" :value="c.value">{{ c.label }}</option>
      </select>
      <select v-model="fl.district" class="bg-page rounded-xl px-3 py-2 text-sm focus:outline-none">
        <option value="all">{{ t("house.district") }}: {{ t("app.all") }}</option>
        <option v-for="d in districts" :key="d.value" :value="d.value">{{ d.label }}</option>
      </select>
    </div>

    <UiLoader v-if="pending" />
    <UiEmpty v-else-if="!filtered.length" icon="fa-home" />

    <div v-else class="grid gap-5" style="grid-template-columns: repeat(auto-fill, minmax(300px, 1fr))">
      <article v-for="h in filtered" :key="h.id" class="bg-white rounded-[20px] border border-border-soft overflow-hidden transition-all hover:-translate-y-[3px] hover:shadow-hover">
        <div class="relative h-[170px] bg-gradient-to-br from-brand-soft to-brand-bg flex items-center justify-center">
          <img v-if="firstImage(h)" :src="firstImage(h)" class="w-full h-full object-cover" />
          <i v-else class="fas fa-building text-5xl text-brand/30" />
          <span class="absolute top-3 left-3 bg-white/90 backdrop-blur px-2.5 py-1 rounded-full text-[11px] font-semibold text-ink">
            {{ typeIcon(h.construction_type || "") }} {{ h.construction_type }}
          </span>
        </div>
        <div class="p-5">
          <h3 class="font-bold text-ink mb-1.5 truncate">{{ h.title }}</h3>
          <p class="text-xs text-ink-soft mb-3 truncate"><i class="fas fa-map-marker-alt mr-1" />{{ h.district }}, {{ h.address }}</p>
          <div class="flex flex-wrap gap-2 mb-3 text-xs text-ink-medium">
            <span class="bg-page px-2 py-1 rounded-md"><i class="fas fa-arrows-alt mr-1" />{{ h.area }} м²</span>
            <span class="bg-page px-2 py-1 rounded-md"><i class="fas fa-door-open mr-1" />{{ h.rooms }}</span>
            <span class="bg-page px-2 py-1 rounded-md"><i class="fas fa-layer-group mr-1" />{{ h.floor }}/{{ h.total_floors }}</span>
          </div>
          <div class="text-brand font-bold text-lg mb-3">{{ formatNumber(h.total_price) }} C</div>
          <div class="flex items-center justify-between">
            <span class="text-xs text-ink-soft">{{ h.contact_phone }}</span>
            <div class="flex gap-1">
              <button class="w-8 h-8 bg-page rounded-lg text-brand hover:bg-brand-soft" @click="openEdit(h)"><i class="fas fa-pen text-xs" /></button>
              <button class="w-8 h-8 bg-page rounded-lg text-red-500 hover:bg-red-100" @click="askDelete(h)"><i class="fas fa-trash text-xs" /></button>
            </div>
          </div>
        </div>
      </article>
    </div>

    <!-- modal -->
    <UiModal v-model="modal" :title="editing.id ? t('app.edit') : t('house.add')" size="lg">
      <div class="grid grid-cols-1 md:grid-cols-2 gap-4">
        <div class="md:col-span-2"><label class="label">{{ t("house.title_field") }} *</label><input v-model="editing.title" class="input" /></div>
        <div>
          <label class="label">{{ t("house.construction_type") }} *</label>
          <select v-model="editing.construction_type" class="input">
            <option v-for="c in constructionTypes" :key="c.value" :value="c.value">{{ c.label }}</option>
          </select>
        </div>
        <div>
          <label class="label">{{ t("house.district") }} *</label>
          <select v-model="editing.district" class="input">
            <option v-for="d in districts" :key="d.value" :value="d.value">{{ d.label }}</option>
          </select>
        </div>
        <div class="md:col-span-2"><label class="label">{{ t("house.address") }} *</label><input v-model="editing.address" class="input" /></div>
        <div><label class="label">{{ t("house.area") }} *</label><input v-model.number="editing.area" type="number" step="0.01" class="input" /></div>
        <div><label class="label">{{ t("house.rooms") }} *</label><input v-model.number="editing.rooms" type="number" class="input" /></div>
        <div><label class="label">{{ t("house.windows") }} *</label><input v-model.number="editing.windows" type="number" class="input" /></div>
        <div><label class="label">{{ t("house.floor") }} *</label><input v-model.number="editing.floor" type="number" class="input" /></div>
        <div><label class="label">{{ t("house.total_floors") }} *</label><input v-model.number="editing.total_floors" type="number" class="input" /></div>
        <div><label class="label">{{ t("house.price_per_m2") }} *</label><input v-model.number="editing.price_per_m2" type="number" class="input" /></div>
        <div><label class="label">{{ t("house.total_price") }} *</label><input v-model.number="editing.total_price" type="number" class="input" /></div>
        <div><label class="label">{{ t("house.developer") }}</label><input v-model="editing.developer" class="input" /></div>
        <div><label class="label">{{ t("house.contact_phone") }}</label><input v-model="editing.contact_phone" class="input" /></div>
        <div>
          <label class="label">{{ t("house.tech_passport") }}</label>
          <select v-model="editing.has_tech_passport" class="input"><option value="нет">{{ t("app.no") }}</option><option value="да">{{ t("app.yes") }}</option></select>
        </div>
        <div>
          <label class="label">{{ t("house.renovation") }}</label>
          <select v-model="editing.has_renovation_permit" class="input"><option value="нет">{{ t("app.no") }}</option><option value="да">{{ t("app.yes") }}</option></select>
        </div>
        <div class="md:col-span-2">
          <label class="label">{{ t("house.files") }}</label>
          <label class="flex flex-col items-center justify-center gap-1 p-5 border-2 border-dashed border-border-soft rounded-xl cursor-pointer hover:border-brand transition-colors text-ink-soft">
            <i class="fas fa-cloud-upload-alt text-2xl" />
            <span class="text-sm">{{ t("house.choose_files") }}</span>
            <input type="file" multiple accept="image/*,application/pdf" hidden @change="onFiles" />
          </label>
          <div v-if="files.length" class="text-xs text-ink-soft mt-2">{{ files.length }} {{ t("house.files") }}</div>
        </div>
      </div>
      <template #footer>
        <button class="btn-outline" @click="modal = false">{{ t("app.cancel") }}</button>
        <button class="btn-primary" @click="save">{{ t("app.save") }}</button>
      </template>
    </UiModal>

    <UiConfirmDialog v-model="confirmDel.open" danger :message="t('app.delete') + '?'" @confirm="doDelete" />
  </div>
</template>
