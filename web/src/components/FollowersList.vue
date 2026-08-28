<template>
  <div class="followers-page">
    <div class="list" v-loading="loading">
      <div v-for="user in users" :key="user.follower_id" class="topic-row" :class="{ 'new-item': highlightId === user.follower_id }">
        <UserAvatar :user-id="user.follower_id" :username="user.username" :size="36" />
        <div class="topic-main">
          <span class="topic-username">{{ user.username }}</span>
          <span class="topic-time">{{ formatDate(user.created_at) }}</span>
        </div>
      </div>
      <div v-if="loadingMore" class="loading-more">Loading...</div>
      <div v-if="!loading && users.length === 0" class="empty">
        <p>No followers yet.</p>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, ref, watch } from 'vue'
import { apiFetch, getCurrentUserId, handleApiError } from '@/api'
import UserAvatar from '@/components/UserAvatar.vue'
import { useInfiniteScroll } from '@/composables/useInfiniteScroll'

interface FollowerUser {
  follower_id: number
  username: string
  created_at: string
}

const props = withDefaults(defineProps<{ userId?: number }>(), { userId: 0 })

const users = ref<FollowerUser[]>([])
const loading = ref(false)
const followersOffset = ref(0)
const hasMore = ref(true)
const highlightId = ref<number | null>(null)

const targetUserId = computed(() => {
  if (props.userId && Number.isFinite(props.userId)) return props.userId
  return getCurrentUserId() ?? 0
})

function formatDate(value?: string) {
  if (!value) return ''
  const date = new Date(value)
  if (Number.isNaN(date.getTime())) return value
  return date.toLocaleString()
}

async function loadFollowers() {
  if (!targetUserId.value) return
  loading.value = true
  try {
    const response = await apiFetch(
      `/users/${targetUserId.value}/followers?limit=20&offset=0&sort=desc`,
    )
    if (!response.ok) throw new Error(`HTTP ${response.status}`)
    const json = await response.json()
    const data = json.data ?? json
    users.value = Array.isArray(data) ? data : []
    followersOffset.value = users.value.length
    hasMore.value = users.value.length >= 20
  } catch (error) {
    handleApiError(error, 'Failed to load followers')
  } finally {
    loading.value = false
  }
}

async function loadMoreFollowers() {
  if (!targetUserId.value || !hasMore.value) return
  try {
    const response = await apiFetch(
      `/users/${targetUserId.value}/followers?limit=20&offset=${followersOffset.value}&sort=desc`,
    )
    if (!response.ok) throw new Error(`HTTP ${response.status}`)
    const json = await response.json()
    const data = json.data ?? json
    const newUsers = Array.isArray(data) ? data : []
    if (newUsers.length > 0) {
      users.value = [...users.value, ...newUsers]
      followersOffset.value += newUsers.length
      highlightId.value = newUsers[0].follower_id
      setTimeout(() => { highlightId.value = null }, 2500)
    }
    if (newUsers.length < 20) {
      hasMore.value = false
    }
  } catch (error) {
    handleApiError(error, 'Failed to load more followers')
  }
}

const canLoadMore = computed(() => hasMore.value && !loading.value)
const { loadingMore } = useInfiniteScroll(loadMoreFollowers, canLoadMore)

onMounted(loadFollowers)
watch(targetUserId, loadFollowers)
</script>

<style scoped>
.followers-page {
  min-height: 100vh;
  padding: 32px 0 80px;
}

.list {
  margin: 0 auto;
  max-width: 1100px;
  padding: 0 20px;
  display: flex;
  flex-direction: column;
}

.topic-row {
  display: flex;
  align-items: center;
  gap: 14px;
  padding: 14px 0;
  border-bottom: 1px solid #262626;
  transition: background 0.15s ease;
}

.topic-row:first-child {
  border-top: 1px solid #262626;
}

.topic-row:hover {
  background: rgba(255, 255, 255, 0.03);
}

.topic-main {
  flex: 1;
  min-width: 0;
  display: flex;
  align-items: center;
  gap: 10px;
}

.topic-username {
  font-size: 15px;
  font-weight: 600;
  color: #e4e6e8;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.topic-time {
  font-size: 13px;
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

.empty {
  padding: 60px 24px;
  text-align: center;
  font-size: 14px;
  color: #8c8c8c;
}
</style>
