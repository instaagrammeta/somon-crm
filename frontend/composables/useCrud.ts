/** Boilerplate-free CRUD composable. */
import type { Ref } from "vue";

export function useCrud<T extends { id: number }>(baseUrl: string, key: string) {
  const api = useApi();
  const toast = useToast();
  const { t } = useI18n();

  const { data, pending, refresh } = useAsyncData<T[]>(key, () =>
    api.get<T[]>(baseUrl)
  );

  const open = ref(false);
  const editing = ref<Partial<T>>({});
  const confirmState = ref<{ open: boolean; id?: number }>({ open: false });

  const openCreate = (defaults: Partial<T> = {}) => {
    editing.value = { ...defaults } as Partial<T>;
    open.value = true;
  };
  const openEdit = (row: T) => {
    editing.value = { ...row };
    open.value = true;
  };

  const save = async () => {
    try {
      const e = editing.value as any;
      if (e.id) {
        await api.put(`${baseUrl}/${e.id}`, e);
        toast.success(t("notify.updated"));
      } else {
        await api.post(baseUrl, e);
        toast.success(t("notify.created"));
      }
      open.value = false;
      await refresh();
    } catch (err: any) {
      toast.error(err?.data?.error || t("notify.error"));
    }
  };

  const askDelete = (row: T) => (confirmState.value = { open: true, id: row.id });
  const doDelete = async () => {
    if (!confirmState.value.id) return;
    try {
      await api.del(`${baseUrl}/${confirmState.value.id}`);
      toast.success(t("notify.deleted"));
      await refresh();
    } catch (err: any) {
      toast.error(err?.data?.error || t("notify.error"));
    }
  };

  return {
    rows: data as Ref<T[] | null>,
    pending,
    refresh,
    open,
    editing,
    confirmState,
    openCreate,
    openEdit,
    save,
    askDelete,
    doDelete,
  };
}
