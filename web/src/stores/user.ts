import { defineStore } from 'pinia'
import { apiFetch, getCurrentUserId } from '@/api'

export const useUserStore = defineStore('user', {
  state: () => ({
    id: null as number | null,
    username: '',
    avatarUrl: '',
    loaded: false,
  }),
  actions: {
    async fetchCurrentUser(force = false) {
      const id = getCurrentUserId()
      if (!id) {
        this.reset()
        return
      }
      this.id = id
      if (this.loaded && !force) return
      try {
        const response = await apiFetch(`/users/${id}`)
        if (!response.ok) throw new Error(`HTTP ${response.status}`)
        const json = await response.json()
        const user = json.data ?? json
        this.username = user?.username ?? ''
        this.avatarUrl = user?.avatar_url ?? `/users/${id}/avatar`
        this.loaded = true
      } catch (error) {
        console.error('Load current user error:', error)
      }
    },
    reset() {
      this.id = null
      this.username = ''
      this.avatarUrl = ''
      this.loaded = false
    },
  },
})
