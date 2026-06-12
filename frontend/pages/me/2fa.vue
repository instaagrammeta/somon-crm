<script setup lang="ts">
/**
 * Self-service 2FA management for the current user.
 *
 * Three states:
 *   1. Not enrolled at all          → "Setup" button starts enrollment
 *   2. Enrolled but not verified    → show otpauth QR + 6-digit input
 *   3. Enabled (verified)           → show "Disable" form (requires password)
 */
import QRCode from "qrcode";

useHead({ title: "Two-factor authentication" });

const api = useApi();
const toast = useToast();

const status = ref<{ enrolled: boolean; enabled: boolean } | null>(null);
const setup = ref<{ secret: string; otpauth_url: string } | null>(null);
const qrDataURL = ref<string | null>(null);
const code = ref("");
const password = ref("");
const recoveryCodes = ref<string[]>([]);
const busy = ref(false);

const refresh = async () => {
  status.value = await api.get("/api/2fa/status");
};
onMounted(refresh);

const startSetup = async () => {
  busy.value = true;
  try {
    setup.value = await api.post("/api/2fa/setup", {});
    qrDataURL.value = await QRCode.toDataURL(setup.value!.otpauth_url, { width: 220, margin: 1 });
    await refresh();
  } finally {
    busy.value = false;
  }
};

const verify = async () => {
  if (code.value.length < 6) return;
  busy.value = true;
  try {
    const res = await api.post<{ recovery_codes: string[] }>("/api/2fa/verify", { code: code.value });
    recoveryCodes.value = res.recovery_codes || [];
    code.value = "";
    setup.value = null;
    qrDataURL.value = null;
    toast.success("2FA фаъол шуд");
    await refresh();
  } catch (e: any) {
    toast.error(e?.data?.error || "Коди нодуруст");
  } finally {
    busy.value = false;
  }
};

const disable = async () => {
  if (!password.value) {
    toast.error("Парол лозим");
    return;
  }
  busy.value = true;
  try {
    await api.post("/api/2fa/disable", { password: password.value });
    password.value = "";
    toast.success("2FA нест шуд");
    await refresh();
  } catch (e: any) {
    toast.error(e?.data?.error || "Парол нодуруст");
  } finally {
    busy.value = false;
  }
};
</script>

<template>
  <div class="max-w-xl">
    <h1 class="text-2xl font-bold text-ink">Тасдиқи дузарбагӣ (2FA)</h1>
    <p class="text-ink-soft mt-1 text-sm">
      Барои бехатарӣ — иловаи код аз аппликатсия (Google Authenticator, Authy, 1Password).
    </p>

    <div class="mt-6 bg-white rounded-2xl border border-border-soft p-6">
      <!-- State 1: not enrolled -->
      <div v-if="status && !status.enrolled">
        <p class="text-ink">Ҳоло 2FA фаъол нест.</p>
        <button class="mt-4 btn-primary" :disabled="busy" @click="startSetup">
          <i class="fas fa-shield-halved" /> Фаъол кардан
        </button>
      </div>

      <!-- State 2: enrolled but not enabled -->
      <div v-else-if="status && status.enrolled && !status.enabled">
        <h3 class="font-bold text-ink">Қадами 1. QR-ро скан кунед</h3>
        <p class="text-sm text-ink-soft mt-1">
          Аппликатсияи Google Authenticator-ро дар телефон кушоед ва ин QR-ро скан кунед.
        </p>
        <div v-if="qrDataURL" class="mt-4 p-3 rounded-2xl bg-page inline-block">
          <img :src="qrDataURL" alt="QR" class="w-44 h-44" />
        </div>
        <p v-if="setup?.secret" class="mt-3 text-xs text-ink-soft">
          Ё дастӣ ворид кунед:
          <code class="px-2 py-1 rounded bg-page font-mono">{{ setup.secret }}</code>
        </p>

        <h3 class="mt-6 font-bold text-ink">Қадами 2. Коди 6-рақама ворид кунед</h3>
        <input
          v-model="code"
          maxlength="6"
          class="input mt-2 text-center text-2xl tracking-widest font-mono"
          placeholder="••••••"
        />
        <button class="mt-3 btn-primary w-full" :disabled="busy || code.length < 6" @click="verify">
          <i class="fas fa-check" /> Тасдиқ
        </button>

        <button class="mt-2 w-full text-sm text-ink-soft hover:text-ink" @click="startSetup">
          QR-и навро эҷод кунед
        </button>
      </div>

      <!-- State 3: enabled -->
      <div v-else-if="status && status.enabled">
        <div class="flex items-center gap-3">
          <span class="w-10 h-10 rounded-full bg-brand-soft text-brand grid place-items-center">
            <i class="fas fa-shield-halved" />
          </span>
          <div>
            <div class="font-bold text-ink">2FA фаъол аст</div>
            <div class="text-xs text-ink-soft">Ҳангоми вуруд код пурсида мешавад</div>
          </div>
        </div>

        <div class="mt-6 border-t border-border-soft pt-4">
          <h3 class="font-bold text-ink">Хомӯш кардан</h3>
          <p class="text-sm text-ink-soft mt-1">Барои тасдиқ паролатонро ворид кунед.</p>
          <input
            v-model="password"
            type="password"
            class="input mt-2"
            placeholder="••••••••"
          />
          <button class="mt-3 px-4 py-2 rounded-xl bg-red-500 text-white hover:bg-red-600" :disabled="busy" @click="disable">
            <i class="fas fa-shield-xmark" /> Хомӯш кардан
          </button>
        </div>
      </div>
    </div>

    <!-- Recovery codes (shown ONCE) -->
    <div v-if="recoveryCodes.length" class="mt-6 bg-amber-50 border-2 border-amber-300 p-5 rounded-2xl">
      <h3 class="font-bold text-amber-900 flex items-center gap-2">
        <i class="fas fa-key" /> Кодҳои бозгашт (як бор нишон дода мешавад!)
      </h3>
      <p class="text-sm text-amber-800 mt-1">
        Чопи онҳоро бигиред ё ҷое нигоҳ доред. Агар телефон гум шавад, бо инҳо
        ворид мешавед.
      </p>
      <div class="mt-3 grid grid-cols-2 sm:grid-cols-4 gap-2 font-mono text-sm">
        <code v-for="c in recoveryCodes" :key="c" class="px-3 py-2 rounded-lg bg-white border border-amber-200">
          {{ c }}
        </code>
      </div>
    </div>
  </div>
</template>
