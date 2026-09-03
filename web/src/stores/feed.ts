import { defineStore } from 'pinia'
import { apiFetch, getCurrentUserId } from '@/api'

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
  view_count: number
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
  user_id?: number
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

export interface LikedComment {
  comment_id: number
  post_id: number
  content: string
  username: string
  user_id: number
  like_count?: number
  reply_count?: number
  created_at: string
}

export type ViewType =
  | 'all'
  | 'liked'
  | 'likedcomments'
  | 'following'
  | 'followers'
  | 'myposts'
  | 'create'
  | 'profile'
  | 'settings'

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
  view_count?: number
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
    view_count: Number(p.view_count ?? 0),
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
    feedPosts: [] as FeedPost[],
    feedPostsLoaded: false,
    feedOffset: 0,
    feedHasMore: true,
    activeNav: 'all' as ViewType,
    likedPosts: [] as LikedPost[],
    likedPostsLoaded: false,
    following: [] as FollowingUser[],
    followingLoaded: false,
    followingUserId: 0,
    followers: [] as FollowerUser[],
    followersLoaded: false,
    followersUserId: 0,
    likedComments: [] as LikedComment[],
    likedCommentsLoaded: false,
    view: 'all' as ViewType,
    postHistory: [] as number[],
    postHistoryIndex: -1,
    newPostIds: new Set<number>(),
  }),
  actions: {
    setView(view: ViewType) {
      this.view = view
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
    updatePostLike(postId: number, liked: boolean, likeCount: number) {
      const post = this.posts.find((p) => p.id === postId)
      if (post) {
        post.like_count = likeCount
      }
      const likedPost = this.likedPosts.find((p) => p.post_id === postId)
      if (likedPost) {
        likedPost.like_count = likeCount
      }
      if (liked) {
        if (!this.likedPosts.some((p) => p.post_id === postId)) {
          const post = this.posts.find((p) => p.id === postId)
          if (post) {
            this.likedPosts.push({
              post_id: post.id,
              author: post.user.username,
              title: post.title,
              tags: post.tags,
              comment_count: post.comment_count,
              like_count: post.like_count,
              created_at: post.created_at,
              user_id: post.user_id,
            })
          }
        }
      } else {
        this.likedPosts = this.likedPosts.filter((p) => p.post_id !== postId)
      }
    },
    togglePostLike(postId: number, liked: boolean, likeCount: number) {
      const post = this.posts.find((p) => p.id === postId)
      if (post) {
        post.like_count = likeCount
      }
      if (liked) {
        if (!this.likedPosts.some((p) => p.post_id === postId)) {
          if (post) {
            this.likedPosts.push({
              post_id: post.id,
              author: post.user.username,
              title: post.title,
              tags: post.tags,
              comment_count: post.comment_count,
              like_count: post.like_count,
              created_at: post.created_at,
              user_id: post.user_id,
            })
          } else {
            this.likedPosts.push({
              post_id: postId,
              author: '',
              title: '',
              tags: [],
              comment_count: 0,
              like_count: likeCount,
              created_at: '',
              user_id: undefined,
            })
          }
        }
      } else {
        this.likedPosts = this.likedPosts.filter((p) => p.post_id !== postId)
      }
    },
    updatePostCommentCount(postId: number, commentCount: number) {
      const post = this.posts.find((p) => p.id === postId)
      if (post) {
        post.comment_count = commentCount
      }
      const likedPost = this.likedPosts.find((p) => p.post_id === postId)
      if (likedPost) {
        likedPost.comment_count = commentCount
      }
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
    async loadMorePosts() {
      const response = await apiFetch('/posts/free')
      if (!response.ok) throw new Error(`HTTP ${response.status}`)
      const json = await response.json()
      const raw = Array.isArray(json) ? json : (json.data ?? [])
      const newPosts = raw.map(toFeedPost)
      const existingIds = new Set(this.posts.map((p) => p.id))
      let firstNewId: number | null = null
      for (const post of newPosts) {
        if (!existingIds.has(post.id)) {
          this.posts.push(post)
          if (firstNewId === null) {
            firstNewId = post.id
          }
        }
      }
      if (firstNewId !== null) {
        this.newPostIds.add(firstNewId)
        setTimeout(() => {
          this.newPostIds.delete(firstNewId!)
        }, 2500)
      }
    },
    async fetchFeedPosts() {
      if (this.feedPostsLoaded) return
      const response = await apiFetch('/users/feed?limit=20&offset=0&sort=desc')
      if (!response.ok) throw new Error(`HTTP ${response.status}`)
      const json = await response.json()
      const raw = Array.isArray(json) ? json : (json.data ?? [])
      this.feedPosts = raw.map(toFeedPost)
      this.feedOffset = this.feedPosts.length
      this.feedHasMore = this.feedPosts.length >= 20
      this.feedPostsLoaded = true
    },
    async loadMoreFeedPosts() {
      const response = await apiFetch(`/users/feed?limit=20&offset=${this.feedOffset}&sort=desc`)
      if (!response.ok) throw new Error(`HTTP ${response.status}`)
      const json = await response.json()
      const raw = Array.isArray(json) ? json : (json.data ?? [])
      const newPosts = raw.map(toFeedPost)
      const existingIds = new Set(this.feedPosts.map((p) => p.id))
      for (const post of newPosts) {
        if (!existingIds.has(post.id)) {
          this.feedPosts.push(post)
        }
      }
      this.feedOffset = this.feedPosts.length
      this.feedHasMore = newPosts.length >= 20
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
      const currentUserId = getCurrentUserId()
      if (!currentUserId) return
      const response = await apiFetch(`/users/${currentUserId}/post-likes`)
      if (!response.ok) throw new Error(`HTTP ${response.status}`)
      const json = await response.json()
      const raw = Array.isArray(json) ? json : (json.data ?? [])
      this.likedPosts = raw.map((p: RawPost) => {
        const userId = Number(p.user_id ?? p.author_id)
        return {
          post_id: Number(p.post_id ?? p.id ?? 0),
          author: p.user?.username ?? p.author ?? '',
          title: p.title ?? '',
          tags: p.tags ?? [],
          comment_count: Number(p.comment_count ?? 0),
          like_count: Number(p.like_count ?? 0),
          created_at: p.created_at ?? '',
          user_id: Number.isFinite(userId) && userId > 0 ? userId : undefined,
        }
      })
      this.likedPostsLoaded = true
    },
    async fetchLikedComments() {
      const response = await apiFetch('/users/me/comment-likes')
      if (!response.ok) throw new Error(`HTTP ${response.status}`)
      const json = await response.json()
      const raw = Array.isArray(json) ? json : (json.data ?? [])
      this.likedComments = raw.map((c: Record<string, unknown>) => ({
        comment_id: Number(c.comment_id ?? c.id ?? 0),
        post_id: Number(c.post_id ?? 0),
        content: String(c.content ?? ''),
        username: String(c.username ?? ''),
        user_id: Number(c.user_id ?? 0),
        like_count: Number(c.like_count ?? 0),
        reply_count: Number(c.reply_count ?? 0),
        created_at: String(c.created_at ?? ''),
      }))
      this.likedCommentsLoaded = true
    },
    async fetchFollowing(userId: number) {
      if (!userId || (this.followingLoaded && this.followingUserId === userId)) return
      const response = await apiFetch(`/users/${userId}/following`)
      if (!response.ok) throw new Error(`HTTP ${response.status}`)
      const json = await response.json()
      this.following = Array.isArray(json) ? json : (json.data ?? [])
      this.followingLoaded = true
      this.followingUserId = userId
    },
    async fetchFollowers(userId: number) {
      if (!userId || (this.followersLoaded && this.followersUserId === userId)) return
      const response = await apiFetch(`/users/${userId}/followers`)
      if (!response.ok) throw new Error(`HTTP ${response.status}`)
      const json = await response.json()
      this.followers = Array.isArray(json) ? json : (json.data ?? [])
      this.followersLoaded = true
      this.followersUserId = userId
    },
    async unfollowUser(userId: number) {
      const response = await apiFetch(`/users/${userId}/follow`, {
        method: 'DELETE',
      })
      if (!response.ok) throw new Error(`HTTP ${response.status}`)
      this.following = this.following.filter((user) => user.following_id !== userId)
    },
    async followUser(userId: number, username: string) {
      const response = await apiFetch(`/users/${userId}/follow`, {
        method: 'PUT',
      })
      if (!response.ok) throw new Error(`HTTP ${response.status}`)
      this.following.push({
        following_id: userId,
        username,
        created_at: '',
      })
    },
  },
})
