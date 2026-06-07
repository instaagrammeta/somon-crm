/** Lightweight toast notifications via a global Pinia-like ref. */
interface Toast {
  id: number;
  type: "success" | "error" | "info";
  text: string;
}

const toasts = ref<Toast[]>([]);
let nextId = 1;

export function useToast() {
  const push = (type: Toast["type"], text: string) => {
    const id = nextId++;
    toasts.value.push({ id, type, text });
    setTimeout(() => {
      toasts.value = toasts.value.filter((t) => t.id !== id);
    }, 3500);
  };
  return {
    toasts,
    success: (t: string) => push("success", t),
    error: (t: string) => push("error", t),
    info: (t: string) => push("info", t),
  };
}
