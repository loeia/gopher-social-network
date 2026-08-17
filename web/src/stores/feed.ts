import { defineStore } from 'pinia'
import { apiFetch } from '@/api'

export const MAX_POST_HISTORY = 50

export interface FeedPost {
  id: number
  title: string
  content: string
  tags: string[]
  user: { username: string }
  user_id?: number
  comment_count: number
  like_count: number
  created_at: string
}

export interface LikedPost {
  post_id: number
  author: string
  title: string
  tags: string[]
  comment_count?: number
  like_count?: number
  created_at: string
}

export interface FollowingUser {
  following_id: number
  username: string
  created_at: string
}

export interface FollowerUser {
  follower_id: number
  username: string
  created_at: string
}

export type ViewType = 'all' | 'liked' | 'following' | 'followers' | 'myposts'

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
  id?: number
  post_id?: number
  user_id?: number
  author_id?: number
  title?: string
  content?: string
  tags?: string[] | null
  created_at?: string
  user?: { username?: string | null } | null
  author?: string | null
  comment_count?: number
  like_count?: number
}

export function toFeedPost(p: RawPost): FeedPost {
  const userId = Number(p.user_id ?? p.author_id)
  return {
    id: Number(p.id ?? p.post_id ?? 0),
    title: p.title ?? '',
    content: p.content ?? '',
    tags: p.tags ?? [],
    user_id: Number.isFinite(userId) && userId > 0 ? userId : undefined,
    comment_count: Number(p.comment_count ?? 0),
    like_count: Number(p.like_count ?? 0),
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
    activeNav: 'all' as ViewType,
    likedPosts: [] as LikedPost[],
    likedPostsLoaded: false,
    following: [] as FollowingUser[],
    followingLoaded: false,
    followers: [] as FollowerUser[],
    followersLoaded: false,
    view: 'all' as ViewType,
    feedScrollTop: 0,
    postHistory: [] as number[],
    postHistoryIndex: -1,
  }),
  actions: {
    setView(view: ViewType) {
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
      const raw = Array.isArray(json) ? json : (json.data ?? [])
      this.posts = raw.map(toFeedPost)
      this.postsLoaded = true
    },
    async refreshPosts() {
      const response = await apiFetch('/posts/free')
      if (!response.ok) throw new Error(`HTTP ${response.status}`)
      const json = await response.json()
      const raw = Array.isArray(json) ? json : (json.data ?? [])
      this.posts = raw.map(toFeedPost)
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
    async fetchFollowing() {
      const response = await apiFetch('/users/following')
      if (!response.ok) throw new Error(`HTTP ${response.status}`)
      const json = await response.json()
      this.following = Array.isArray(json) ? json : (json.data ?? [])
      this.followingLoaded = true
    },
    async fetchFollowers() {
      const response = await apiFetch('/users/followers')
      if (!response.ok) throw new Error(`HTTP ${response.status}`)
      const json = await response.json()
      this.followers = Array.isArray(json) ? json : (json.data ?? [])
      this.followersLoaded = true
    },
    async unfollowUser(userId: number) {
      const response = await apiFetch(`/users/${userId}/unfollow`, { method: 'PUT' })
      if (!response.ok) throw new Error(`HTTP ${response.status}`)
      this.following = this.following.filter((user) => user.following_id !== userId)
    },
    async followUser(userId: number, username: string) {
      const response = await apiFetch(`/users/${userId}/follow`, { method: 'PUT' })
      if (!response.ok) throw new Error(`HTTP ${response.status}`)
      this.following.push({ following_id: userId, username, created_at: '' })
    },
  },
})
