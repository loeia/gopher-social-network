<template>
  <div class="following-page">
    <div class="list" v-loading="loading">
      <div v-for="user in users" :key="user.following_id" class="topic-row" :class="{ 'new-item': highlightId === user.following_id }">
        <UserAvatar :user-id="user.following_id" :username="user.username" :size="36" />
        <div class="topic-main">
          <span class="topic-username">{{ user.username }}</span>
          <span class="topic-time">{{ formatDate(user.created_at) }}</span>
        </div>
        <el-button
          v-if="isOwn"
          class="unfollow-btn"
          :disabled="unfollowingId === user.following_id"
          @click="unfollow(user.following_id)"
        >
          Unfollow
        </el-button>
      </div>
      <div v-if="loadingMore" class="loading-more">Loading...</div>
      <div v-if="!loading && users.length === 0" class="empty">
        <p>No followings yet.</p>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, onActivated, onMounted, ref, watch } from 'vue'
import { apiFetch, getCurrentUserId, handleApiError } from '@/api'
import { notify } from '@/utils/message'
import UserAvatar from '@/components/UserAvatar.vue'
import { useInfiniteScroll } from '@/composables/useInfiniteScroll'

interface FollowingUser {
  following_id: number
  username: string
  created_at: string
}

const props = withDefaults(defineProps<{ userId?: number }>(), { userId: 0 })

const users = ref<FollowingUser[]>([])
const loading = ref(false)
const followingOffset = ref(0)
const hasMore = ref(true)
const highlightId = ref<number | null>(null)
const unfollowingId = ref<number | null>(null)

const targetUserId = computed(() => {
  if (props.userId && Number.isFinite(props.userId)) return props.userId
  return getCurrentUserId() ?? 0
})

const isOwn = computed(() => targetUserId.value === getCurrentUserId())

function formatDate(value?: string) {
  if (!value) return ''
  const date = new Date(value)
  if (Number.isNaN(date.getTime())) return value
  return date.toLocaleString()
}

async function loadFollowing() {
  if (!targetUserId.value) return
  loading.value = true
  try {
    const response = await apiFetch(
      `/users/${targetUserId.value}/following?limit=20&offset=0&sort=desc`,
    )
    if (!response.ok) throw new Error(`HTTP ${response.status}`)
    const json = await response.json()
    const data = json.data ?? json
    users.value = Array.isArray(data) ? data : []
    followingOffset.value = users.value.length
    hasMore.value = users.value.length >= 20
  } catch (error) {
    handleApiError(error, 'Failed to load followings')
  } finally {
    loading.value = false
  }
}

async function loadMoreFollowing() {
  if (!targetUserId.value || !hasMore.value) return
  try {
    const response = await apiFetch(
      `/users/${targetUserId.value}/following?limit=20&offset=${followingOffset.value}&sort=desc`,
    )
    if (!response.ok) throw new Error(`HTTP ${response.status}`)
    const json = await response.json()
    const data = json.data ?? json
    const newUsers = Array.isArray(data) ? data : []
    if (newUsers.length > 0) {
      users.value = [...users.value, ...newUsers]
      followingOffset.value += newUsers.length
      highlightId.value = newUsers[0].following_id
      setTimeout(() => { highlightId.value = null }, 2500)
    }
    if (newUsers.length < 20) {
      hasMore.value = false
    }
  } catch (error) {
    handleApiError(error, 'Failed to load more followings')
  }
}

const canLoadMore = computed(() => hasMore.value && !loading.value)
const { loadingMore } = useInfiniteScroll(loadMoreFollowing, canLoadMore)

async function unfollow(userId: number) {
  unfollowingId.value = userId
  try {
    const response = await apiFetch(`/users/${userId}/follow`, { method: 'DELETE' })
    if (!response.ok) throw new Error(`HTTP ${response.status}`)
    users.value = users.value.filter((u) => u.following_id !== userId)
  } catch (error) {
    handleApiError(error, 'Failed to unfollow')
  } finally {
    unfollowingId.value = null
  }
}

onMounted(loadFollowing)
onActivated(loadFollowing)
watch(targetUserId, loadFollowing)
</script>

<style scoped>
.following-page {
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

.unfollow-btn {
  flex-shrink: 0;
  margin-left: auto;
  font-weight: 500;
  --el-button-bg-color: transparent;
  --el-button-border-color: #333;
  --el-button-text-color: #e4e6e8;
  --el-button-hover-bg-color: transparent;
  --el-button-hover-border-color: #f85149;
  --el-button-hover-text-color: #f85149;
  --el-button-active-bg-color: transparent;
  --el-button-active-border-color: #f85149;
  --el-button-active-text-color: #f85149;
}

.unfollow-btn.is-disabled,
.unfollow-btn.is-disabled:hover {
  border-color: #333;
  color: #595959;
  background: transparent;
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
