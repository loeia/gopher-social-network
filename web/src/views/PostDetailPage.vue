<template>
  <div class="detail-page">
    <div class="back-nav">
      <el-button text @click="backHome">← Back</el-button>
      <div class="nav-arrows">
        <el-button text :disabled="!hasPrevPost" @click="goToPrevious">&lt;</el-button>
        <el-button text :disabled="!hasNextPost" @click="goToNext">&gt;</el-button>
      </div>
    </div>

    <div class="detail-container" v-loading="loading">
      <template v-if="post">
        <h1 class="detail-title">{{ post.title }}</h1>

        <div class="detail-tags" v-if="post.tags && post.tags.length">
          <span v-for="tag in post.tags" :key="tag" class="tag">{{ tag }}</span>
        </div>

        <div class="meta">
          <span class="meta-item">{{ post.author }}</span>
          <span class="meta-sep">·</span>
          <span class="meta-item">{{ formatDate(post.created_at) }}</span>
          <span class="meta-sep">·</span>
          <span class="meta-item">{{ likesCount }} likes</span>
          <span class="meta-sep">·</span>
          <span class="meta-item">{{ commentsCount }} comments</span>
          <span class="meta-sep">·</span>
          <button
            class="like-btn"
            :class="likeBtnClass"
            :disabled="!isLoggedIn || liking"
            :title="likeTitle"
            @click="toggleLike"
          >
            <svg class="like-icon" viewBox="0 0 24 24" aria-hidden="true">
              <path d="M12 21.35l-1.45-1.32C5.4 15.36 2 12.28 2 8.5 2 5.42 4.42 3 7.5 3c1.74 0 3.41.81 4.5 2.09C13.09 3.81 14.76 3 16.5 3 19.58 3 22 5.42 22 8.5c0 3.78-3.4 6.86-8.55 11.54L12 21.35z" />
            </svg>
          </button>
        </div>

        <div class="markdown-body" v-html="renderedContent" />

        <Comments :post-id="post.id" />
      </template>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, ref, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { apiFetch, getToken } from '@/api'
import { useFeedStore } from '@/stores/feed'
import { renderMarkdown } from '@/utils/markdown'
import { notify } from '@/utils/message'
import Comments from '@/components/Comments.vue'

interface PostDetail {
  id: number
  author: string
  title: string
  content: string
  tags: string[]
  likes_count?: number
  likes?: number
  comments_count?: number
  comment_count?: number
  created_at: string
  updated_at: string
}

const route = useRoute()
const router = useRouter()

const store = useFeedStore()

const post = ref<PostDetail | null>(null)
const loading = ref(false)

const isLoggedIn = computed(() => !!getToken())
const liked = ref(false)
const liking = ref(false)
const likesCount = ref(0)
const commentsCount = ref(0)

const likeTitle = computed(() => {
  if (!isLoggedIn.value) return 'Log in to like'
  return liked.value ? 'Unlike' : 'Like'
})
const likeBtnClass = computed(() => ({ liked: isLoggedIn.value && liked.value }))

const hasPrevPost = computed(() => store.postHistoryIndex > 0)
const hasNextPost = computed(() => store.postHistoryIndex < store.postHistory.length - 1)
const renderedContent = computed(() => {
  if (!post.value) return ''
  return renderMarkdown(post.value.content)
})

function backHome() {
  store.clearPostHistory()
  router.push('/')
}

function goToPrevious() {
  const id = store.goBackPost()
  if (id !== null) router.push(`/posts/${id}`)
}

function goToNext() {
  const id = store.goForwardPost()
  if (id !== null) router.push(`/posts/${id}`)
}

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

async function toggleLike() {
  if (!isLoggedIn.value || !post.value) return
  liking.value = true
  try {
    const action = liked.value ? 'dislike' : 'like'
    const response = await apiFetch(`/posts/${post.value.id}/${action}`, { method: 'PUT' })
    if (!response.ok) throw new Error(`HTTP ${response.status}`)
    liked.value = !liked.value
    likesCount.value += liked.value ? 1 : -1
  } catch (error) {
    console.error('Toggle like error:', error)
    notify('error', 'Failed to update like')
  } finally {
    liking.value = false
  }
}

async function loadLikedState(id: number) {
  if (!getToken()) {
    liked.value = false
    return
  }
  try {
    const response = await apiFetch('/users/likes')
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
    likesCount.value = Number(data.likes_count ?? data.likes ?? 0)
    commentsCount.value = Number(data.comments_count ?? data.comment_count ?? 0)
    loadLikedState(id)
  } catch (error) {
    console.error('Load post error:', error)
    notify('error', 'Failed to load post')
  } finally {
    loading.value = false
  }
}

onMounted(loadPost)
watch(() => route.params.postId, loadPost)
</script>

<style scoped>
.detail-page {
  min-height: 100vh;
  padding: 32px 0 80px;
  font-family:
    -apple-system,
    BlinkMacSystemFont,
    'Segoe UI',
    'Liberation Sans',
    'Helvetica Neue',
    Arial,
    sans-serif;
}

.back-nav {
  width: 75%;
  margin: 0 auto 16px;
  padding: 0 24px;
  display: flex;
  align-items: center;
  gap: 12px;
}

.nav-arrows {
  margin-left: auto;
  display: flex;
  gap: 4px;
}

.back-nav :deep(.el-button) {
  color: #6a737c;
  background: transparent;
}

.back-nav :deep(.el-button:hover),
.back-nav :deep(.el-button:focus),
.back-nav :deep(.el-button:focus-visible) {
  color: #6a737c;
  background: transparent;
  text-decoration: underline;
  text-decoration-color: #6a737c;
  text-underline-offset: 4px;
}

.nav-arrows :deep(.el-button:hover),
.nav-arrows :deep(.el-button:focus),
.nav-arrows :deep(.el-button:focus-visible) {
  color: #6a737c;
  background: transparent;
  font-weight: 600;
  text-decoration: none;
}

.back-nav :deep(.el-button.is-disabled) {
  color: #3d4043;
  background: transparent;
  text-decoration: none;
  cursor: not-allowed;
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
  color: #ffffff;
}

.meta {
  display: flex;
  align-items: center;
  flex-wrap: wrap;
  gap: 8px;
  padding-bottom: 16px;
  border-bottom: 1px solid #3d4043;
  font-size: 13px;
  color: #9fa6ad;
}

.meta-sep {
  color: #6a737c;
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
  color: #6a737c;
}

.like-btn:disabled {
  cursor: not-allowed;
  color: #6a737c;
}

.like-btn:not(:disabled) {
  color: #ffffff;
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
  color: #ffffff;
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
  border-left: 4px solid #3d4043;
  color: #b2b6b9;
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
  border: 1px solid #3d4043;
}

.markdown-body :deep(th) {
  background: #232629;
  color: #ffffff;
}

.markdown-body :deep(hr) {
  margin: 24px 0;
  border: none;
  border-top: 1px solid #3d4043;
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
  background: #26324a;
  color: #9faccc;
  font-size: 12px;
}


</style>