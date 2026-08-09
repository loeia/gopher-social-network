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

export interface SearchParams {
  search?: string
  author?: string
  tags?: string[]
  since?: string
  until?: string
  page?: number
}

export const SEARCH_PAGE_SIZE = 20

function localDateToUtc(date: string, endOfDay: boolean): string {
  const [y, m, d] = date.split('-')
  const dt = new Date(
    Number(y),
    Number(m) - 1,
    Number(d),
    endOfDay ? 23 : 0,
    endOfDay ? 59 : 0,
    0,
    0,
  )
  return dt.toISOString().replace(/\.\d{3}Z$/, 'Z')
}

type RawPost = {
  id: number
  title?: string
  content?: string
  created_at?: string
  user?: { username?: string | null } | null
  author?: string | null
}

function toFeedPost(p: RawPost): FeedPost {
  return {
    id: Number(p.id),
    title: p.title ?? '',
    content: p.content ?? '',
    created_at: p.created_at ?? '',
    user: {
      username: p.user?.username ?? p.author ?? '',
    },
  }
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
      const response = await apiFetch('/posts/free')
      if (!response.ok) throw new Error(`HTTP ${response.status}`)
      const json = await response.json()
      this.posts = Array.isArray(json) ? json : (json.data ?? [])
      this.postsLoaded = true
    },
    async refreshPosts() {
      const response = await apiFetch('/posts/free')
      if (!response.ok) throw new Error(`HTTP ${response.status}`)
      const json = await response.json()
      this.posts = Array.isArray(json) ? json : (json.data ?? [])
      this.postsLoaded = true
      this.feedScrollTop = 0
    },
    async fetchSearch(params: SearchParams): Promise<FeedPost[]> {
      const query = new URLSearchParams()
      if (params.search?.trim()) query.set('search', params.search.trim())
      if (params.author?.trim()) query.set('author', params.author.trim())
      if (params.tags && params.tags.length) query.set('tags', params.tags.join(','))
      if (params.since) query.set('since', localDateToUtc(params.since, false))
      if (params.until) query.set('until', localDateToUtc(params.until, true))
      const page = params.page && params.page > 1 ? params.page : 1
      query.set('limit', String(SEARCH_PAGE_SIZE))
      query.set('offset', String((page - 1) * SEARCH_PAGE_SIZE))
      const qs = query.toString()
      const response = await apiFetch(`/posts/search${qs ? `?${qs}` : ''}`)
      if (!response.ok) throw new Error(`HTTP ${response.status}`)
      const json = await response.json()
      const raw = Array.isArray(json) ? json : (json.data ?? [])
      return raw.map(toFeedPost)
    },
    async fetchLikedPosts() {
      const response = await apiFetch('/users/likes')
      if (!response.ok) throw new Error(`HTTP ${response.status}`)
      const json = await response.json()
      this.likedPosts = Array.isArray(json) ? json : (json.data ?? [])
      this.likedPostsLoaded = true
    },
  },
})
