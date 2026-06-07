<script setup lang="ts">
definePageMeta({ layout: "auth" });
useHead({ title: () => useI18n().t("auth.login") });

const auth = useAuthStore();
const toast = useToast();
const { t } = useI18n();

const form = reactive({ login: "", password: "" });
const loading = ref(false);
const error = ref("");

const submit = async () => {
  if (!form.login || !form.password) return;
  loading.value = true;
  error.value = "";
  try {
    const user = await auth.login(form.login, form.password);
    toast.success(t("auth.welcome", { name: user.full_name }));
    await navigateTo("/");
  } catch (e: any) {
    error.value = e?.data?.error || t("auth.invalid");
  } finally {
    loading.value = false;
  }
};
</script>

<template>
  <div class="w-full max-w-md">
    <div class="card p-8 sm:p-10">
      <div class="flex justify-center mb-6">
        <div class="logo-icon !w-14 !h-14" />
      </div>
      <h1 class="text-2xl font-bold text-center text-ink mb-2">
        {{ $t("auth.login_title") }}
      </h1>
      <p class="text-center text-ink-soft text-sm mb-8">
        {{ $t("auth.login_subtitle") }}
      </p>

      <form class="space-y-4" @submit.prevent="submit">
        <div>
          <label class="label">{{ $t("auth.login_field") }}</label>
          <div class="relative">
            <i
              class="fa-solid fa-user absolute left-4 top-1/2 -translate-y-1/2 text-ink-soft text-sm"
            />
            <input
              v-model="form.login"
              type="text"
              class="input pl-11"
              placeholder="admin"
              autocomplete="username"
              required
            />
          </div>
        </div>

        <div>
          <label class="label">{{ $t("auth.password") }}</label>
          <div class="relative">
            <i
              class="fa-solid fa-lock absolute left-4 top-1/2 -translate-y-1/2 text-ink-soft text-sm"
            />
            <input
              v-model="form.password"
              type="password"
              class="input pl-11"
              placeholder="••••••••"
              autocomplete="current-password"
              required
            />
          </div>
        </div>

        <Transition name="err">
          <div
            v-if="error"
            class="text-sm text-red-500 bg-red-50 px-4 py-2.5 rounded-xl flex items-center gap-2"
          >
            <i class="fa-solid fa-circle-exclamation" />
            {{ error }}
          </div>
        </Transition>

        <button class="btn-primary w-full !py-3 mt-2" :disabled="loading">
          <i v-if="loading" class="fa-solid fa-spinner animate-spin" />
          <span>{{ loading ? $t("auth.logging_in") : $t("auth.submit") }}</span>
        </button>
      </form>
    </div>

    <p class="mt-6 text-center text-xs text-ink-soft">
      © {{ new Date().getFullYear() }} Somon CRM
    </p>
  </div>
</template>

<style scoped>
.err-enter-active, .err-leave-active {
  transition: all .2s ease;
}
.err-enter-from, .err-leave-to {
  opacity: 0;
  transform: translateY(-4px);
}
</style>
