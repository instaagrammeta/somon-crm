<script setup lang="ts">
/**
 * Floating "?" button bottom-right that opens a small chat panel powered by
 * /api/ai/assistant. The CRM context is built server-side so the model has
 * up-to-date numbers (open leads/tasks per status, top sources).
 *
 * Local state only — chat history is per-tab and is wiped when the panel is
 * closed. No persistence (yet).
 */
const api = useApi();
const open = ref(false);
const sending = ref(false);
const input = ref("");

interface Msg { role: "user" | "assistant"; text: string; }
const history = ref<Msg[]>([]);

const ask = async () => {
  const q = input.value.trim();
  if (!q || sending.value) return;
  history.value.push({ role: "user", text: q });
  input.value = "";
  sending.value = true;
  try {
    const res = await api.post<{ answer: string }>("/api/ai/assistant", { question: q });
    history.value.push({ role: "assistant", text: res.answer });
  } catch (e: any) {
    history.value.push({
      role: "assistant",
      text: e?.data?.error === "ai.disabled"
        ? "AI ҳозир дастрас нест. Админ AI_API_KEY-ро гузорад."
        : (e?.data?.error || "Хатогии AI"),
    });
  } finally {
    sending.value = false;
  }
};
</script>

<template>
  <div>
    <button
      class="fixed bottom-4 right-4 z-30 w-14 h-14 rounded-full bg-brand text-white shadow-2xl hover:bg-brand-dark grid place-items-center"
      :title="'AI Дастёр'"
      @click="open = !open"
    >
      <i :class="open ? 'fas fa-xmark text-xl' : 'fas fa-question text-xl'" />
    </button>

    <Transition name="fade">
      <div
        v-if="open"
        class="fixed bottom-24 right-4 z-30 w-[min(92vw,420px)] h-[min(70vh,540px)] bg-white rounded-2xl shadow-2xl border border-border-soft flex flex-col"
      >
        <header class="px-4 py-3 border-b border-border-soft flex items-center gap-2">
          <span class="w-8 h-8 rounded-full bg-brand-soft text-brand grid place-items-center">
            <i class="fas fa-robot" />
          </span>
          <div>
            <div class="font-bold text-ink text-sm">AI Дастёр</div>
            <div class="text-[11px] text-ink-soft">Дар бораи CRM пурсиш диҳед</div>
          </div>
        </header>

        <main class="flex-1 overflow-y-auto p-4 space-y-3 bg-page/40">
          <div v-if="!history.length" class="text-sm text-ink-soft text-center py-8">
            <i class="fas fa-robot text-2xl text-brand/40 mb-2 block" />
            Намунаи саволҳо:
            <ul class="mt-2 space-y-1">
              <li><button class="hover:text-brand" @click="input = 'Лидҳои нав имрӯз чанд?'">Лидҳои нав имрӯз чанд?</button></li>
              <li><button class="hover:text-brand" @click="input = 'Кадом манбаъ беҳтарин аст?'">Кадом манбаъ беҳтарин аст?</button></li>
              <li><button class="hover:text-brand" @click="input = 'Якчанд task кушод аст?'">Якчанд task кушод аст?</button></li>
            </ul>
          </div>
          <div
            v-for="(m, i) in history"
            :key="i"
            :class="m.role === 'user' ? 'flex justify-end' : 'flex justify-start'"
          >
            <div
              class="max-w-[85%] rounded-2xl px-3.5 py-2.5 text-sm whitespace-pre-line leading-relaxed"
              :class="m.role === 'user' ? 'bg-brand text-white' : 'bg-white border border-border-soft text-ink'"
            >{{ m.text }}</div>
          </div>
          <div v-if="sending" class="flex justify-start">
            <div class="rounded-2xl px-3.5 py-2.5 bg-white border border-border-soft text-sm text-ink-soft">
              <i class="fas fa-spinner fa-spin mr-2" />Фикр карда истодаам…
            </div>
          </div>
        </main>

        <form class="p-3 border-t border-border-soft flex gap-2" @submit.prevent="ask">
          <input
            v-model="input"
            placeholder="Саволатонро нависед…"
            class="flex-1 px-3 py-2 rounded-xl border border-border-soft focus:outline-none focus:border-brand"
          />
          <button
            type="submit"
            :disabled="sending || !input.trim()"
            class="w-10 h-10 rounded-xl bg-brand text-white hover:bg-brand-dark disabled:opacity-50"
          >
            <i class="fas fa-paper-plane" />
          </button>
        </form>
      </div>
    </Transition>
  </div>
</template>

<style scoped>
.fade-enter-active, .fade-leave-active { transition: opacity 0.2s, transform 0.2s; }
.fade-enter-from, .fade-leave-to { opacity: 0; transform: translateY(8px); }
</style>
