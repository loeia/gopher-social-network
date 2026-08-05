<template>
  <div class="feed" v-loading="loading">
    <div v-for="post in posts" :key="post.id" class="card" @click="openPost(post.id)">
      <div class="card-header">
        <h2 class="card-title">{{ post.title }}</h2>
        <span class="card-date">{{ formatDate(post.created_at) }}</span>
      </div>
      <div class="card-author">
        <span class="avatar">G</span>
        <span>{{ post.user.username }}</span>
      </div>
    </div>
    <button class="new-btn" :disabled="loading" @click="loadNewPosts">New</button>
  </div>
</template>

<script setup lang="ts">
import { nextTick, onActivated, onBeforeUnmount, onDeactivated, onMounted, ref } from 'vue'
import { useRouter } from 'vue-router'
import { storeToRefs } from 'pinia'
import { ElMessage } from 'element-plus'
import { useFeedStore } from '@/stores/feed'

const store = useFeedStore()
const { posts } = storeToRefs(store)
const loading = ref(false)

const router = useRouter()

function openPost(id: number) {
  router.push(`/posts/${id}`)
}

function formatDate(value?: string) {
  if (!value) return ''
  const date = new Date(value)
  if (Number.isNaN(date.getTime())) return value
  return date.toLocaleString()
}

function saveScroll() {
  store.feedScrollTop = window.scrollY
}

function restoreScroll() {
  if (store.feedScrollTop > 0) {
    nextTick(() => window.scrollTo({ top: store.feedScrollTop }))
  }
}

async function loadPosts() {
  if (store.postsLoaded) return
  loading.value = true
  try {
    await store.fetchPosts()
  } catch (error) {
    console.error('Load posts error:', error)
    ElMessage.error('Failed to load posts')
  } finally {
    loading.value = false
  }
}

async function loadNewPosts() {
  loading.value = true
  try {
    await store.refreshPosts()
    window.scrollTo({ top: 0 })
  } catch (error) {
    console.error('Refresh posts error:', error)
    ElMessage.error('Failed to refresh posts')
  } finally {
    loading.value = false
  }
}

onMounted(loadPosts)
onMounted(restoreScroll)
onActivated(restoreScroll)
onDeactivated(saveScroll)
onBeforeUnmount(saveScroll)
</script>

<style scoped>
.feed {
  margin: 0 20%;
  display: flex;
  flex-direction: column;
  gap: 20px;
}

.card {
  background: #141414;
  border: 1px solid #262626;
  border-radius: 12px;
  padding: 24px;
  cursor: pointer;
  transition:
    border-color 0.2s ease,
    transform 0.2s ease;
}

.card:hover {
  border-color: #ffffff;
  transform: translateY(-2px);
}

.card-header {
  display: flex;
  align-items: baseline;
  justify-content: space-between;
  gap: 16px;
}

.card-title {
  margin: 0;
  font-size: 20px;
  font-weight: 600;
  color: #ffffff;
}

.card-date {
  flex-shrink: 0;
  font-size: 13px;
  color: #8c8c8c;
}

.card-content {
  display: -webkit-box;
  -webkit-line-clamp: 3;
  -webkit-box-orient: vertical;
  overflow: hidden;
  margin: 16px 0;
  line-height: 1.7;
  color: #bfbfbf;
  word-break: break-word;
}

.card-content :deep(p),
.card-content :deep(ul),
.card-content :deep(ol),
.card-content :deep(pre),
.card-content :deep(blockquote),
.card-content :deep(h1),
.card-content :deep(h2),
.card-content :deep(h3),
.card-content :deep(h4),
.card-content :deep(h5),
.card-content :deep(h6) {
  margin: 0;
}

.card-content :deep(pre) {
  overflow: hidden;
}

.card-author {
  display: flex;
  align-items: center;
  gap: 10px;
  padding-top: 16px;
  border-top: 1px solid #262626;
  font-size: 14px;
  color: #8c8c8c;
}

.avatar {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 28px;
  height: 28px;
  border-radius: 50%;
  background: #ffffff;
  color: #141414;
  font-size: 13px;
  font-weight: 600;
}

.new-btn {
  flex-shrink: 0;
  align-self: center;
  width: 12%;
  padding: 10px 0;
  border: 1px solid #ffffff;
  border-radius: 8px;
  background: #ffffff;
  color: #141414;
  font-size: 15px;
  font-weight: 600;
  text-align: center;
  white-space: nowrap;
  cursor: pointer;
  transition:
    background 0.2s ease,
    color 0.2s ease;
}

.new-btn:hover {
  background: #141414;
  color: #ffffff;
}

.new-btn:disabled {
  opacity: 0.6;
  cursor: not-allowed;
}
</style>