<template>
  <div class="detail-page">
    <div class="detail-container" v-loading="loading">
      <template v-if="post">
        <h1 class="detail-title">{{ post.title }}</h1>

        <div class="detail-tags" v-if="post.tags && post.tags.length">
          <span v-for="tag in post.tags" :key="tag" class="tag">{{ tag }}</span>
        </div>

        <div class="meta">
          <UserAvatar :user-id="post.author_id" :username="post.author" :size="20" />
          <span class="meta-item">{{ post.author }}</span>
          <button
            v-if="canFollowAuthor"
            class="follow-btn"
            :disabled="followBusy"
            @click="toggleFollowAuthor"
          >
            {{ isFollowingAuthor ? 'Unfollow' : 'Follow' }}
          </button>
          <span class="meta-sep">·</span>
          <span class="meta-item">{{ formatDate(post.created_at) }}</span>
          <template v-if="post.updated_at && post.updated_at !== post.created_at">
            <span class="meta-sep">·</span>
            <span class="meta-item">Edited {{ formatDate(post.updated_at) }}</span>
          </template>
          <span class="meta-sep">·</span>
          <button
            class="like-btn"
            :class="likeBtnClass"
            :disabled="!isLoggedIn || liking"
            :title="likeTitle"
            @click="toggleLike"
          >
            <svg class="like-icon" viewBox="0 0 24 24" aria-hidden="true">
              <path
                d="M12 21.35l-1.45-1.32C5.4 15.36 2 12.28 2 8.5 2 5.42 4.42 3 7.5 3c1.74 0 3.41.81 4.5 2.09C13.09 3.81 14.76 3 16.5 3 19.58 3 22 5.42 22 8.5c0 3.78-3.4 6.86-8.55 11.54L12 21.35z"
              />
            </svg>
            <span class="meta-count">{{ likesCount }}</span>
          </button>
          <span class="meta-sep">·</span>
          <span class="comment-item">
            <svg class="comment-icon" viewBox="0 0 24 24" aria-hidden="true">
              <path d="M20 2H4a2 2 0 0 0-2 2v18l4-4h14a2 2 0 0 0 2-2V4a2 2 0 0 0-2-2z" />
            </svg>
            <span class="meta-count">{{ post.comment_count }}</span>
          </span>
          <span class="meta-sep">·</span>
          <span class="view-item">
            <svg class="view-icon" viewBox="0 0 24 24" aria-hidden="true">
              <path d="M1 12s4-8 11-8 11 8 11 8-4 8-11 8-11-8-11-8z" />
              <circle cx="12" cy="12" r="3" />
            </svg>
            <span class="meta-count">{{ viewsCount }}</span>
          </span>
        </div>

        <div class="markdown-body" v-html="renderedContent" />

        <Comments :post-id="post.id" @comment-count-changed="onCommentCountChanged" />
      </template>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, ref, watch } from 'vue'
import { useRoute } from 'vue-router'
import { storeToRefs } from 'pinia'
import { apiFetch, getCurrentUserId, getToken, handleApiError } from '@/api'
import { useFeedStore } from '@/stores/feed'
import { renderMarkdown } from '@/utils/markdown'
import { notify } from '@/utils/message'
import Comments from '@/components/Comments.vue'
import UserAvatar from '@/components/UserAvatar.vue'

interface PostDetail {
  id: number
  author_id?: number
  author: string
  title: string
  content: string
  tags: string[]
  like_count?: number
  comment_count?: number
  view_count?: number
  created_at: string
  updated_at: string
}

const route = useRoute()

const store = useFeedStore()
const { following } = storeToRefs(store)

const post = ref<PostDetail | null>(null)
const loading = ref(false)

const isLoggedIn = computed(() => !!getToken())
const liked = ref(false)
const liking = ref(false)
const likesCount = ref(0)
const viewsCount = ref(0)

const likeTitle = computed(() => {
  if (!isLoggedIn.value) return 'Log in to like'
  return liked.value ? 'Unlike' : 'Like'
})
const likeBtnClass = computed(() => ({ liked: isLoggedIn.value && liked.value }))

const followBusy = ref(false)
const canFollowAuthor = computed(() => {
  if (!isLoggedIn.value || !post.value?.author_id) return false
  return getCurrentUserId() !== post.value.author_id
})
const isFollowingAuthor = computed(() => {
  const authorId = post.value?.author_id
  return !!authorId && following.value.some((f) => f.following_id === authorId)
})

async function toggleFollowAuthor() {
  const current = post.value
  const authorId = current?.author_id
  if (!current || !authorId || followBusy.value) return
  followBusy.value = true
  try {
    if (isFollowingAuthor.value) {
      await store.unfollowUser(authorId)
    } else {
      await store.followUser(authorId, current.author)
    }
  } catch (error) {
    handleApiError(error, 'Failed to update follow')
  } finally {
    followBusy.value = false
  }
}

async function loadFollowState() {
  const currentUserId = getCurrentUserId()
  if (!currentUserId) return
  try {
    await store.fetchFollowing(currentUserId)
  } catch (error) {
    console.error('Load following error:', error)
  }
}

const renderedContent = computed(() => {
  if (!post.value) return ''
  return renderMarkdown(post.value.content)
})

function formatDate(value?: string) {
  if (!value) return ''
  const date = new Date(value)
  if (Number.isNaN(date.getTime())) return value
  return date.toLocaleString()
}

function loadPost() {
  const id = Number(route.params.postId)
  store.visitPost(id)
  loading.value = true
  post.value = null
  startLoad(id)
}

function onCommentCountChanged(delta: number) {
  if (post.value) {
    post.value.comment_count = Math.max(0, (post.value.comment_count ?? 0) + delta)
  }
}

async function toggleLike() {
  if (!isLoggedIn.value || !post.value) return
  liking.value = true
  try {
    const method = liked.value ? 'DELETE' : 'PUT'
    const response = await apiFetch(`/posts/${post.value.id}/like`, { method })
    if (!response.ok) throw new Error(`HTTP ${response.status}`)
    liked.value = !liked.value
    likesCount.value += liked.value ? 1 : -1
    store.togglePostLike(post.value.id, liked.value, likesCount.value)
  } catch (error) {
    handleApiError(error, 'Failed to update like')
  } finally {
    liking.value = false
  }
}

async function loadLikedState(id: number) {
  if (!getToken()) {
    liked.value = false
    return
  }
  const currentUserId = getCurrentUserId()
  if (!currentUserId) {
    liked.value = false
    return
  }
  try {
    const response = await apiFetch(`/users/${currentUserId}/post-likes`)
    if (!response.ok) throw new Error(`HTTP ${response.status}`)
    const json = await response.json()
    const list = json.data ?? json
    liked.value = Array.isArray(list) && list.some((p) => p?.post_id === id)
  } catch (error) {
    console.error('Load liked state error:', error)
    liked.value = false
  }
}

async function startLoad(id: number) {
  try {
    const response = await apiFetch(`/posts/${id}`)
    if (!response.ok) throw new Error(`HTTP ${response.status}`)
    const json = await response.json()
    const data = json.data ?? json
    post.value = data
    likesCount.value = Number(data.like_count ?? 0)
    viewsCount.value = Number(data.view_count ?? 0)
    loadLikedState(id)
  } catch (error) {
    handleApiError(error, 'Failed to load post')
  } finally {
    loading.value = false
  }
}

onMounted(() => {
  loadPost()
  loadFollowState()
})
watch(() => route.params.postId, loadPost)
</script>

<style scoped>
.detail-page {
  min-height: 100vh;
  padding: 32px 0 80px;
  font-family:
    -apple-system, BlinkMacSystemFont, 'Segoe UI', 'Liberation Sans', 'Helvetica Neue', Arial,
    sans-serif;
}

.detail-container {
  width: 75%;
  margin: 0 auto;
  padding: 0 24px;
}

.detail-title {
  margin: 0 0 16px;
  font-size: 27px;
  font-weight: 600;
  line-height: 1.2;
  color: #e4e6e8;
  word-break: break-word;
  overflow-wrap: anywhere;
}

.meta {
  display: flex;
  align-items: center;
  flex-wrap: wrap;
  gap: 8px;
  padding-bottom: 16px;
  border-bottom: 1px solid #262626;
  font-size: 13px;
  color: #8c8c8c;
}

.meta-sep {
  color: #595959;
}

.like-btn {
  display: inline-flex;
  align-items: center;
  gap: 5px;
  padding: 0;
  background: transparent;
  border: none;
  cursor: pointer;
  font-size: 13px;
  color: #8c8c8c;
}

.like-btn:disabled {
  cursor: not-allowed;
  color: #595959;
}

.like-btn:not(:disabled) {
  color: #e4e6e8;
}

.like-btn .like-icon {
  width: 16px;
  height: 16px;
}

.like-btn .like-icon path {
  fill: none;
  stroke: currentColor;
  stroke-width: 2;
}

.like-btn:not(:disabled):hover:not(.liked) .like-icon path {
  fill: currentColor;
}

.like-btn.liked {
  color: #e05c5c;
}

.like-btn.liked .like-icon path {
  fill: currentColor;
  stroke: currentColor;
}

.meta-count {
  font-size: 13px;
}

.comment-item {
  display: inline-flex;
  align-items: center;
  gap: 5px;
  color: #8c8c8c;
}

.comment-item .comment-icon {
  width: 16px;
  height: 16px;
}

.comment-item .comment-icon path {
  fill: none;
  stroke: currentColor;
  stroke-width: 2;
}

.view-item {
  display: inline-flex;
  align-items: center;
  gap: 5px;
  color: #8c8c8c;
}

.view-item .view-icon {
  width: 16px;
  height: 16px;
}

.view-item .view-icon path,
.view-item .view-icon circle {
  fill: none;
  stroke: currentColor;
  stroke-width: 2;
}

.follow-btn {
  display: inline-flex;
  align-items: center;
  padding: 2px 10px;
  font-size: 12px;
  font-weight: 600;
  color: #e4e6e8;
  background: #262626;
  border: 1px solid #333;
  border-radius: 6px;
  cursor: pointer;
}

.follow-btn:hover:not(:disabled) {
  background: #333;
  color: #e4e6e8;
  border-color: #595959;
}

.follow-btn:disabled {
  cursor: not-allowed;
  opacity: 0.6;
}

/* Markdown-rendered content, StackOverflow-style */
.markdown-body {
  margin: 24px 0;
  font-size: 15px;
  line-height: 1.5;
  color: #e4e6e8;
  word-break: break-word;
}

.markdown-body :deep(p) {
  margin: 0 0 16px;
}

.markdown-body :deep(h1),
.markdown-body :deep(h2),
.markdown-body :deep(h3),
.markdown-body :deep(h4),
.markdown-body :deep(h5),
.markdown-body :deep(h6) {
  margin: 28px 0 12px;
  font-weight: 600;
  line-height: 1.3;
  color: #e4e6e8;
}

.markdown-body :deep(h1) {
  font-size: 24px;
}

.markdown-body :deep(h2) {
  font-size: 21px;
}

.markdown-body :deep(h3) {
  font-size: 18px;
}

.markdown-body :deep(h4) {
  font-size: 16px;
}

.markdown-body :deep(h5) {
  font-size: 15px;
}

.markdown-body :deep(h6) {
  font-size: 14px;
}

.markdown-body :deep(a) {
  color: #6cbbf7;
  text-decoration: none;
}

.markdown-body :deep(a:hover) {
  text-decoration: underline;
}

.markdown-body :deep(ul),
.markdown-body :deep(ol) {
  margin: 0 0 16px;
  padding-left: 28px;
}

.markdown-body :deep(li) {
  margin: 4px 0;
}

.markdown-body :deep(pre) {
  margin: 0 0 16px;
  padding: 12px 16px;
  overflow-x: auto;
  background: #232629;
  border-radius: 4px;
  font-size: 13px;
  line-height: 1.5;
}

.markdown-body :deep(code) {
  font-family: ui-monospace, 'SF Mono', Menlo, Consolas, 'Liberation Mono', monospace;
  font-size: 13px;
}

.markdown-body :deep(:not(pre) > code) {
  padding: 2px 6px;
  background: #232629;
  border-radius: 4px;
  color: #e4e6e8;
}

.markdown-body :deep(pre code) {
  padding: 0;
  background: none;
  color: #e4e6e8;
}

.markdown-body :deep(blockquote) {
  margin: 0 0 16px;
  padding: 4px 16px;
  border-left: 4px solid #333;
  color: #8c8c8c;
}

.markdown-body :deep(img) {
  max-width: 100%;
  border-radius: 4px;
}

.markdown-body :deep(table) {
  margin: 0 0 16px;
  border-collapse: collapse;
  width: 100%;
}

.markdown-body :deep(th),
.markdown-body :deep(td) {
  padding: 8px 12px;
  border: 1px solid #333;
}

.markdown-body :deep(th) {
  background: #232629;
  color: #e4e6e8;
}

.markdown-body :deep(hr) {
  margin: 24px 0;
  border: none;
  border-top: 1px solid #262626;
}

.detail-tags {
  display: flex;
  flex-wrap: wrap;
  gap: 6px;
  margin-bottom: 16px;
}

.tag {
  padding: 4px 6px;
  border-radius: 4px;
  background: #1f1f1f;
  color: #8c8c8c;
  font-size: 12px;
}
</style>
