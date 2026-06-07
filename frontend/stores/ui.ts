import { defineStore } from "pinia";

export const useUiStore = defineStore("ui", {
  state: () => ({
    sidebarCollapsed: false,
    mobileSidebarOpen: false,
  }),
  actions: {
    toggleSidebar() {
      this.sidebarCollapsed = !this.sidebarCollapsed;
    },
    openMobileSidebar() {
      this.mobileSidebarOpen = true;
    },
    closeMobileSidebar() {
      this.mobileSidebarOpen = false;
    },
  },
});
