<template>
  <div class="feed" v-loading="loading">
    <div
      v-for="post in posts"
      :key="post.post_id"
      class="topic-row"
      :class="{ 'new-item': highlightId === post.post_id }"
      @click="openPost(post.post_id)"
    >
      <div class="topic-top">
        <h2 class="topic-title">{{ post.title }}</h2>
        <div class="topic-stats">
          <span class="topic-stat">
            <svg class="stat-icon like-icon" viewBox="0 0 24 24" aria-hidden="true">
              <path
                d="M12 21.35l-1.45-1.32C5.4 15.36 2 12.28 2 8.5 2 5.42 4.42 3 7.5 3c1.74 0 3.41.81 4.5 2.09C13.09 3.81 14.76 3 16.5 3 19.58 3 22 5.42 22 8.5c0 3.78-3.4 6.86-8.55 11.54L12 21.35z"
              />
            </svg>
            {{ post.like_count ?? 0 }}
          </span>
          <span class="topic-stat">
            <svg class="stat-icon" viewBox="0 0 24 24" aria-hidden="true">
              <path d="M20 2H4a2 2 0 0 0-2 2v18l4-4h14a2 2 0 0 0 2-2V4a2 2 0 0 0-2-2z" />
            </svg>
            {{ post.comment_count ?? 0 }}
          </span>
        </div>
      </div>
      <div class="topic-bottom">
        <UserAvatar :user-id="post.user_id" :username="post.author" :size="20" />
        <span class="topic-author">{{ post.author }}</span>
        <template v-if="post.tags && post.tags.length">
          <span class="meta-dot">&middot;</span>
          <span v-for="tag in post.tags" :key="tag" class="topic-tag">{{ tag }}</span>
        </template>
      </div>
    </div>
    <div v-if="loadingMore" class="loading-more">Loading...</div>
  </div>
</template>

<script setup lang="ts">
import { onActivated, onBeforeUnmount, onDeactivated, onMounted, ref } from 'vue'
import { useRouter } from 'vue-router'
import { getCurrentUserId, apiFetch, handleApiError } from '@/api'
import UserAvatar from '@/components/UserAvatar.vue'

interface LikedPost {
  post_id: number
  author: string
  title: string
  tags: string[]
  comment_count?: number
  like_count?: number
  created_at: string
  user_id?: number
}

const posts = ref<LikedPost[]>([])
const loading = ref(false)
const postsOffset = ref(0)
const hasMore = ref(true)
const highlightId = ref<number | null>(null)
const loadingMore = ref(false)

const router = useRouter()

function openPost(id: number) {
  router.push(`/posts/${id}`)
}

async function loadLikedPosts() {
  const userId = getCurrentUserId()
  if (!userId) return
  loading.value = true
  try {
    const response = await apiFetch(
      `/users/${userId}/post-likes?limit=20&offset=0&sort=desc`,
    )
    if (!response.ok) throw new Error(`HTTP ${response.status}`)
    const json = await response.json()
    const data = json.data ?? json
    posts.value = Array.isArray(data) ? data : []
    postsOffset.value = posts.value.length
    hasMore.value = posts.value.length >= 20
  } catch (error) {
    handleApiError(error, 'Failed to load liked posts')
  } finally {
    loading.value = false
  }
}

async function loadMoreLikedPosts() {
  const userId = getCurrentUserId()
  if (!userId || !hasMore.value || loadingMore.value) return
  loadingMore.value = true
  try {
    const response = await apiFetch(
      `/users/${userId}/post-likes?limit=20&offset=${postsOffset.value}&sort=desc`,
    )
    if (!response.ok) throw new Error(`HTTP ${response.status}`)
    const json = await response.json()
    const data = json.data ?? json
    const newPosts = Array.isArray(data) ? data : []
    if (newPosts.length > 0) {
      posts.value = [...posts.value, ...newPosts]
      postsOffset.value += newPosts.length
      highlightId.value = newPosts[0].post_id
      setTimeout(() => { highlightId.value = null }, 2500)
    }
    if (newPosts.length < 20) {
      hasMore.value = false
    }
  } catch (error) {
    handleApiError(error, 'Failed to load more liked posts')
  } finally {
    loadingMore.value = false
  }
}

// Home-style infinite scroll: fire as soon as the user scrolls within 200px of
// the bottom, with no debounce and no automatic load on mount/activation — only
// from scroll events, exactly like the Home page (PostsList.vue).
function nearBottom() {
  const scrollHeight = document.documentElement.scrollHeight
  const scrollTop = window.scrollY
  const clientHeight = window.innerHeight
  return scrollTop + clientHeight >= scrollHeight - 200
}

function handleScroll() {
  if (loadingMore.value || !hasMore.value || loading.value) return
  if (nearBottom()) {
    loadMoreLikedPosts()
  }
}

onMounted(() => {
  loadLikedPosts()
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
.feed {
  margin: 0 auto;
  max-width: 1100px;
  padding: 0 20px;
  display: flex;
  flex-direction: column;
}

.topic-row {
  padding: 14px 0;
  border-bottom: 1px solid #262626;
  cursor: pointer;
  transition: background 0.15s ease;
}

.topic-row:first-child {
  border-top: 1px solid #262626;
}

.topic-row:hover {
  background: rgba(255, 255, 255, 0.03);
}

.topic-top {
  display: flex;
  align-items: center;
  gap: 12px;
}

.topic-title {
  flex: 1;
  min-width: 0;
  margin: 0;
  font-size: 16px;
  font-weight: 600;
  line-height: 1.4;
  color: #e4e6e8;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.topic-stats {
  flex-shrink: 0;
  display: flex;
  align-items: center;
  gap: 14px;
}

.topic-stat {
  display: inline-flex;
  align-items: center;
  gap: 4px;
  font-size: 13px;
  color: #8c8c8c;
  white-space: nowrap;
}

.stat-icon {
  width: 15px;
  height: 15px;
  fill: none;
  stroke: currentColor;
  stroke-width: 2;
}

.like-icon {
  fill: #e05c5c;
  stroke: #e05c5c;
}

.topic-bottom {
  display: flex;
  align-items: center;
  flex-wrap: wrap;
  gap: 6px;
  margin-top: 6px;
  padding-left: 2px;
  font-size: 13px;
  color: #8c8c8c;
}

.topic-author {
  color: #8c8c8c;
  font-weight: 500;
}

.meta-dot {
  color: #595959;
}

.topic-tag {
  padding: 1px 8px;
  background: #1f1f1f;
  border: 1px solid #333;
  border-radius: 4px;
  font-size: 12px;
  color: #8c8c8c;
  white-space: nowrap;
}

.loading-more {
  text-align: center;
  padding: 16px;
  color: #8c8c8c;
  font-size: 14px;
}

.topic-row.new-item {
  animation: highlight-flash 2.5s ease-out;
}

@keyframes highlight-flash {
  0% {
    background-color: rgba(64, 158, 255, 0.15);
  }
  100% {
    background-color: transparent;
  }
}
</style>
