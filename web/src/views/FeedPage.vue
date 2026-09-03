<template>
  <div class="feed-page">
    <div class="page-header">
      <h1 class="page-title">Subscriptions</h1>
      <p class="page-desc">Posts from users you follow</p>
    </div>
    <PostsList
      :posts="posts"
      :loading="loading"
      :highlight-id="highlightId"
    />
    <div v-if="loadingMore" class="loading-more">Loading...</div>
    <div v-if="!loading && posts.length === 0" class="empty-state">
      <p>No posts yet.</p>
      <p class="empty-hint">Follow some users to see their posts here.</p>
    </div>
  </div>
</template>

<script setup lang="ts">
import { onActivated, onBeforeUnmount, onDeactivated, onMounted, ref } from 'vue'
import { useRouter } from 'vue-router'
import PostsList from '@/components/PostsList.vue'
import { apiFetch, handleApiError } from '@/api'
import { toFeedPost, type FeedPost } from '@/stores/feed'

defineOptions({ name: 'FeedPage' })

const router = useRouter()
const posts = ref<FeedPost[]>([])
const loading = ref(false)
const postsOffset = ref(0)
const hasMore = ref(true)
const highlightId = ref<number | null>(null)
const loadingMore = ref(false)

onMounted(loadFeed)

async function loadFeed() {
  loading.value = true
  try {
    const response = await apiFetch('/users/feed?limit=20&offset=0&sort=desc')
    if (!response.ok) {
      if (response.status === 401) {
        router.push('/login')
        return
      }
      throw new Error(`HTTP ${response.status}`)
    }
    const json = await response.json()
    const raw = Array.isArray(json) ? json : (json.data ?? [])
    posts.value = raw.map(toFeedPost)
    postsOffset.value = posts.value.length
    hasMore.value = posts.value.length >= 20
  } catch (error) {
    handleApiError(error, 'Failed to load feed')
  } finally {
    loading.value = false
  }
}

async function loadMoreFeed() {
  if (!hasMore.value || loadingMore.value) return
  loadingMore.value = true
  try {
    const response = await apiFetch(`/users/feed?limit=20&offset=${postsOffset.value}&sort=desc`)
    if (!response.ok) throw new Error(`HTTP ${response.status}`)
    const json = await response.json()
    const raw = Array.isArray(json) ? json : (json.data ?? [])
    const newPosts = raw.map(toFeedPost)
    if (newPosts.length > 0) {
      const existingIds = new Set(posts.value.map((p: FeedPost) => p.id))
      const uniqueNewPosts = newPosts.filter((p: FeedPost) => !existingIds.has(p.id))
      posts.value = [...posts.value, ...uniqueNewPosts]
      postsOffset.value += uniqueNewPosts.length
      highlightId.value = uniqueNewPosts[0]?.id ?? null
      setTimeout(() => { highlightId.value = null }, 2500)
    }
    if (newPosts.length < 20) {
      hasMore.value = false
    }
  } catch (error) {
    handleApiError(error, 'Failed to load more posts')
  } finally {
    loadingMore.value = false
  }
}

function nearBottom() {
  const scrollHeight = document.documentElement.scrollHeight
  const scrollTop = window.scrollY
  const clientHeight = window.innerHeight
  return scrollTop + clientHeight >= scrollHeight - 200
}

function handleScroll() {
  if (loadingMore.value || !hasMore.value || loading.value) return
  if (nearBottom()) {
    loadMoreFeed()
  }
}

onMounted(() => {
  window.addEventListener('scroll', handleScroll)
})
onBeforeUnmount(() => {
  window.removeEventListener('scroll', handleScroll)
})
onActivated(() => {
  window.addEventListener('scroll', handleScroll)
})
onDeactivated(() => {
  window.removeEventListener('scroll', handleScroll)
})
</script>

<style scoped>
.feed-page {
  min-height: 100vh;
  padding: 32px 0 80px;
}

.page-header {
  width: 75%;
  margin: 0 auto 24px;
  padding: 0 24px;
}

.page-title {
  margin: 0 0 4px;
  font-size: 28px;
  font-weight: 600;
  color: #e4e6e8;
}

.page-desc {
  margin: 0;
  font-size: 14px;
  color: #8c8c8c;
}

.loading-more {
  text-align: center;
  padding: 16px;
  color: #8c8c8c;
  font-size: 14px;
}

.empty-state {
  text-align: center;
  padding: 64px 16px;
  color: #8c8c8c;
}

.empty-state p {
  margin: 0 0 8px;
  font-size: 16px;
}

.empty-hint {
  font-size: 14px;
  color: #595959;
}
</style>
