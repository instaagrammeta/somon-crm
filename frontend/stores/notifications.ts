import { defineStore } from "pinia";
import type { Notification } from "~/types";

export const useNotificationsStore = defineStore("notifications", {
  state: () => ({
    items: [] as Notification[],
    unread: 0,
    loading: false,
  }),
  actions: {
    /** Load the full feed for the current user. */
    async fetch() {
      this.loading = true;
      try {
        const api = useApi();
        const rows = await api.get<Notification[]>("/api/notifications");
        this.items = rows || [];
        this.unread = this.items.filter((n) => !n.is_read).length;
      } catch {
        /* keep previous state on failure */
      } finally {
        this.loading = false;
      }
    },
    /** Cheap poll for the topbar badge. */
    async fetchUnread() {
      try {
        const api = useApi();
        const res = await api.get<{ count: number }>(
          "/api/notifications/unread-count"
        );
        this.unread = res?.count || 0;
      } catch {
        /* ignore */
      }
    },
    async markRead(id: number) {
      const n = this.items.find((x) => x.id === id);
      if (n && !n.is_read) {
        n.is_read = true;
        this.unread = Math.max(0, this.unread - 1);
      }
      try {
        await useApi().post(`/api/notifications/mark-read/${id}`);
      } catch {
        /* optimistic update already applied */
      }
    },
    async markAll() {
      this.items.forEach((n) => (n.is_read = true));
      this.unread = 0;
      try {
        await useApi().post("/api/notifications/read-all");
      } catch {
        /* ignore */
      }
    },
    async remove(id: number) {
      this.items = this.items.filter((x) => x.id !== id);
      this.unread = this.items.filter((n) => !n.is_read).length;
      try {
        await useApi().del(`/api/notifications/${id}`);
      } catch {
        /* ignore */
      }
    },
  },
});
