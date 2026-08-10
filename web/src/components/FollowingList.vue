<template>
  <div class="list" v-loading="loading">
    <div v-for="user in users" :key="user.follower_id" class="card">
      <div class="avatar">{{ avatarLabel(user.username) }}</div>
      <div class="user-info">
        <span class="username">{{ user.username }}</span>
      </div>
      <span class="date">{{ formatDate(user.created_at) }}</span>
      <el-button
        class="unfollow-btn"
        :disabled="unfollowingId === user.follower_id"
        @click="unfollow(user.follower_id)"
      >
        Unfollow
      </el-button>
    </div>
    <div v-if="!loading && users.length === 0" class="empty">
      <p>No followings yet.</p>
    </div>
  </div>
</template>

<script setup lang="ts">
import { nextTick, onActivated, onBeforeUnmount, onDeactivated, onMounted, ref } from 'vue'
import { storeToRefs } from 'pinia'
import { useFeedStore } from '@/stores/feed'
import { notify } from '@/utils/message'

const store = useFeedStore()
const { following: users } = storeToRefs(store)
const loading = ref(false)
const unfollowingId = ref<number | null>(null)

function avatarLabel(username: string) {
  return username ? username.charAt(0).toUpperCase() : '?'
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

async function loadFollowing() {
  loading.value = true
  try {
    await store.fetchFollowing()
  } catch (error) {
    console.error('Load following error:', error)
    notify('error', 'Failed to load followings')
  } finally {
    loading.value = false
  }
}

async function unfollow(userId: number) {
  unfollowingId.value = userId
  try {
    await store.unfollowUser(userId)
    notify('success', 'Unfollowed')
  } catch (error) {
    console.error('Unfollow error:', error)
    notify('error', 'Failed to unfollow')
  } finally {
    unfollowingId.value = null
  }
}

onMounted(() => {
  restoreScroll()
  loadFollowing()
})
onActivated(() => {
  restoreScroll()
  loadFollowing()
})
onDeactivated(saveScroll)
onBeforeUnmount(saveScroll)
</script>

<style scoped>
.list {
  margin: 0 20%;
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.card {
  display: flex;
  align-items: center;
  gap: 16px;
  background: #141414;
  border: 1px solid #262626;
  border-radius: 12px;
  padding: 16px 24px;
  transition: border-color 0.2s ease;
}

.card:hover {
  border-color: #3d444d;
}

.avatar {
  flex-shrink: 0;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 48px;
  height: 48px;
  border-radius: 50%;
  background: #ffffff;
  color: #141414;
  font-size: 18px;
  font-weight: 600;
}

.user-info {
  min-width: 0;
  flex: 1;
  display: flex;
  flex-direction: column;
}

.username {
  font-size: 16px;
  font-weight: 600;
  color: #ffffff;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.date {
  flex-shrink: 0;
  font-size: 13px;
  color: #8c8c8c;
}

.unfollow-btn {
  flex-shrink: 0;
  margin-left: 0;
  font-weight: 500;
  --el-button-bg-color: transparent;
  --el-button-border-color: #3d444d;
  --el-button-text-color: #f0f6fc;
  --el-button-hover-bg-color: transparent;
  --el-button-hover-border-color: #f85149;
  --el-button-hover-text-color: #f85149;
  --el-button-active-bg-color: transparent;
  --el-button-active-border-color: #f85149;
  --el-button-active-text-color: #f85149;
}

.unfollow-btn.is-disabled,
.unfollow-btn.is-disabled:hover {
  border-color: #3d444d;
  color: #8c8c8c;
  background: transparent;
}

.empty {
  padding: 80px 24px;
  border: 1px dashed #262626;
  border-radius: 12px;
  text-align: center;
  font-size: 16px;
  color: #8c8c8c;
}
</style>
