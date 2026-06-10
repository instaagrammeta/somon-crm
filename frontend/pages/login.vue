<script setup lang="ts">
definePageMeta({ layout: "auth" });
useHead({ title: () => useI18n().t("auth.login") });

const auth = useAuthStore();
const toast = useToast();
const { t } = useI18n();

const form = reactive({ login: "", password: "", remember: false });
const showPassword = ref(false);
const loading = ref(false);
const error = ref("");

const submit = async () => {
  if (!form.login || !form.password) return;
  loading.value = true;
  error.value = "";
  let user;
  try {
    user = await auth.login(form.login.trim(), form.password);
  } catch (e: any) {
    error.value = e?.data?.error || t("auth.invalid");
    loading.value = false;
    return;
  }
  // Login succeeded — show welcome and go to the dashboard. A navigation
  // error here must NOT be reported as "invalid credentials".
  toast.success(t("auth.welcome", { name: user.full_name }));
  loading.value = false;
  await navigateTo("/");
};
</script>

<template>
  <div class="w-full max-w-[980px] grid grid-cols-1 md:grid-cols-2 bg-white rounded-[32px] shadow-[0_25px_50px_-12px_rgba(0,0,0,0.12)] overflow-hidden relative z-10">
    <!-- Left: green illustration -->
    <div class="relative overflow-hidden bg-gradient-to-br from-brand to-brand-dark text-white p-8 sm:p-[50px] flex flex-col justify-between min-h-[220px]">
      <span class="absolute -right-20 -top-20 w-[250px] h-[250px] rounded-full bg-white/[0.08]" />
      <span class="absolute -left-[100px] -bottom-[100px] w-[300px] h-[300px] rounded-full bg-white/[0.05]" />

      <div class="flex items-center gap-3 relative z-10">
        <div class="w-12 h-12 bg-white rounded-[14px] flex items-center justify-center relative">
          <span class="absolute w-7 h-7 rounded-full border-[3px] border-brand" />
          <span class="absolute w-2.5 h-2.5 rounded-full bg-brand" />
        </div>
        <span class="text-[22px] font-extrabold tracking-[-0.3px]">RealEstate</span>
      </div>

      <div class="relative z-10 hidden md:block">
        <h2 class="text-3xl font-bold leading-tight mb-3.5">
          {{ t("auth.hero_title_1") }}<br />{{ t("auth.hero_title_2") }}
        </h2>
        <p class="text-sm opacity-85 leading-relaxed mb-7">{{ t("auth.hero_subtitle") }}</p>
        <div class="flex flex-col gap-3.5">
          <div v-for="i in 3" :key="i" class="flex items-center gap-3 text-[13px]">
            <div class="w-8 h-8 bg-white/[0.18] rounded-[10px] flex items-center justify-center text-sm shrink-0">
              <i class="fas fa-check" />
            </div>
            <span>{{ t(`auth.feature_${i}`) }}</span>
          </div>
        </div>
      </div>

      <div class="text-[11px] opacity-60 relative z-10 hidden md:block">{{ t("auth.rights") }}</div>
    </div>

    <!-- Right: form -->
    <div class="p-8 sm:p-[50px] flex flex-col justify-center">
      <div class="mb-8">
        <h1 class="text-[28px] font-bold text-ink mb-2 tracking-[-0.3px]">{{ t("auth.page_title") }}</h1>
        <p class="text-ink-soft text-sm">{{ t("auth.page_subtitle") }}</p>
      </div>

      <form @submit.prevent="submit">
        <Transition name="err">
          <div v-if="error" class="bg-red-50 text-red-700 px-4 py-3 rounded-xl mb-[18px] text-[13px] flex items-center gap-2">
            <i class="fas fa-exclamation-circle" />
            <span>{{ error }}</span>
          </div>
        </Transition>

        <div class="mb-[18px]">
          <label class="flex items-center gap-2 mb-2 font-semibold text-ink-medium text-[13px]">
            <i class="fas fa-user text-brand text-xs" /> {{ t("auth.login_field") }}
          </label>
          <div class="relative">
            <i class="fas fa-user absolute left-4 top-1/2 -translate-y-1/2 text-ink-soft text-sm" />
            <input
              v-model="form.login"
              type="text"
              required
              autocomplete="username"
              :placeholder="t('auth.login_ph')"
              class="w-full pl-[46px] pr-4 py-3.5 bg-page border-[1.5px] border-transparent rounded-[14px] text-sm focus:outline-none focus:border-brand focus:bg-white focus:ring-[3px] focus:ring-brand/10 transition-all"
            />
          </div>
        </div>

        <div class="mb-[18px]">
          <label class="flex items-center gap-2 mb-2 font-semibold text-ink-medium text-[13px]">
            <i class="fas fa-lock text-brand text-xs" /> {{ t("auth.password") }}
          </label>
          <div class="relative">
            <i class="fas fa-lock absolute left-4 top-1/2 -translate-y-1/2 text-ink-soft text-sm" />
            <input
              v-model="form.password"
              :type="showPassword ? 'text' : 'password'"
              required
              autocomplete="current-password"
              :placeholder="t('auth.password_ph')"
              class="w-full pl-[46px] pr-12 py-3.5 bg-page border-[1.5px] border-transparent rounded-[14px] text-sm focus:outline-none focus:border-brand focus:bg-white focus:ring-[3px] focus:ring-brand/10 transition-all"
            />
            <button
              type="button"
              class="absolute right-4 top-1/2 -translate-y-1/2 text-ink-soft hover:text-brand text-sm"
              @click="showPassword = !showPassword"
            >
              <i class="fas" :class="showPassword ? 'fa-eye-slash' : 'fa-eye'" />
            </button>
          </div>
        </div>

        <div class="flex justify-between items-center mb-6 text-[13px]">
          <label class="flex items-center gap-2 cursor-pointer text-ink-medium">
            <input v-model="form.remember" type="checkbox" class="w-4 h-4 accent-brand cursor-pointer" />
            <span>{{ t("auth.remember_me") }}</span>
          </label>
          <a href="#" class="text-brand font-semibold hover:underline">{{ t("auth.forgot") }}</a>
        </div>

        <button
          type="submit"
          :disabled="loading"
          class="w-full py-3.5 bg-brand hover:bg-brand-dark text-white rounded-[14px] text-[15px] font-semibold flex items-center justify-center gap-2 transition-all hover:-translate-y-0.5 hover:shadow-[0_8px_20px_rgba(31,122,77,0.25)] disabled:opacity-70 disabled:translate-y-0"
        >
          <i class="fas" :class="loading ? 'fa-spinner fa-spin' : 'fa-arrow-right-to-bracket'" />
          <span>{{ loading ? t("auth.logging_in") : t("auth.submit") }}</span>
        </button>
      </form>

      <div class="text-center mt-6 text-xs text-ink-soft">{{ t("auth.version") }}</div>
    </div>
  </div>
</template>

<style scoped>
.err-enter-active,
.err-leave-active {
  transition: all 0.2s ease;
}
.err-enter-from,
.err-leave-to {
  opacity: 0;
  transform: translateY(-4px);
}
</style>
