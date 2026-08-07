import { defineStore } from 'pinia'
import { apiFetch } from '@/api'

export const MAX_POST_HISTORY = 50

export interface FeedPost {
  id: number
  title: string
  content: string
  user: { username: string }
  created_at: string
}

export interface LikedPost {
  post_id: number
  author: string
  title: string
  created_at: string
}

export const useFeedStore = defineStore('feed', {
  state: () => ({
    posts: [] as FeedPost[],
    postsLoaded: false,
    likedPosts: [] as LikedPost[],
    likedPostsLoaded: false,
    view: 'all' as 'all' | 'liked',
    feedScrollTop: 0,
    postHistory: [] as number[],
    postHistoryIndex: -1,
  }),
  actions: {
    setView(view: 'all' | 'liked') {
      this.view = view
    },
    setFeedScrollTop(top: number) {
      this.feedScrollTop = top
    },
    visitPost(id: number) {
      if (this.postHistory[this.postHistoryIndex] === id) return
      this.postHistory = this.postHistory.filter((pid) => pid !== id)
      this.postHistory.push(id)
      if (this.postHistory.length > MAX_POST_HISTORY) {
        this.postHistory.shift()
      }
      this.postHistoryIndex = this.postHistory.length - 1
    },
    goBackPost(): number | null {
      if (this.postHistoryIndex <= 0) return null
      this.postHistoryIndex -= 1
      return this.postHistory[this.postHistoryIndex] ?? null
    },
    goForwardPost(): number | null {
      if (this.postHistoryIndex >= this.postHistory.length - 1) return null
      this.postHistoryIndex += 1
      return this.postHistory[this.postHistoryIndex] ?? null
    },
    clearPostHistory() {
      this.postHistory = []
      this.postHistoryIndex = -1
    },
    async fetchPosts() {
      if (this.postsLoaded) return
      const response = await apiFetch('/free')
      if (!response.ok) throw new Error(`HTTP ${response.status}`)
      const json = await response.json()
      this.posts = Array.isArray(json) ? json : json.data ?? []
      this.postsLoaded = true
    },
    async refreshPosts() {
      const response = await apiFetch('/free')
      if (!response.ok) throw new Error(`HTTP ${response.status}`)
      const json = await response.json()
      this.posts = Array.isArray(json) ? json : json.data ?? []
      this.postsLoaded = true
      this.feedScrollTop = 0
    },
    async fetchLikedPosts() {
      if (this.likedPostsLoaded) return
      const response = await apiFetch('/users/likes')
      if (!response.ok) throw new Error(`HTTP ${response.status}`)
      const json = await response.json()
      this.likedPosts = Array.isArray(json) ? json : json.data ?? []
      this.likedPostsLoaded = true
    },
  },
})
