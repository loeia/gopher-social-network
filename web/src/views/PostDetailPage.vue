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
          <span class="meta-item">{{ post.likes_count }} likes</span>
          <span class="meta-sep">·</span>
          <span class="meta-item">{{ post.comments_count }} comments</span>
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
import { apiFetch } from '@/api'
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
  likes_count: number
  comments_count: number
  created_at: string
  updated_at: string
}

const route = useRoute()
const router = useRouter()

const store = useFeedStore()

const post = ref<PostDetail | null>(null)
const loading = ref(false)

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

async function startLoad(id: number) {
  try {
    const response = await apiFetch(`/posts/${id}`)
    if (!response.ok) throw new Error(`HTTP ${response.status}`)
    const json = await response.json()
    post.value = json.data ?? json
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