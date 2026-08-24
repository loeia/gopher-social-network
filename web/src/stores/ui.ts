import { defineStore } from 'pinia'

export const useUIStore = defineStore('ui', {
  state: () => ({
    sidebarOpen: true,
  }),
  actions: {
    toggleSidebar() {
      this.sidebarOpen = !this.sidebarOpen
    },
  },
})
