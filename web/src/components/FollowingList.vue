<template>
  <div class="list" v-loading="loading">
    <div v-for="user in users" :key="user.following_id" class="topic-row">
      <UserAvatar :user-id="user.following_id" :username="user.username" :size="36" />
      <div class="topic-main">
        <span class="topic-username">{{ user.username }}</span>
        <span class="topic-time">{{ formatDate(user.created_at) }}</span>
      </div>
      <el-button
        class="unfollow-btn"
        :disabled="unfollowingId === user.following_id"
        @click="unfollow(user.following_id)"
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
import UserAvatar from '@/components/UserAvatar.vue'

const store = useFeedStore()
const { following: users } = storeToRefs(store)
const loading = ref(false)
const unfollowingId = ref<number | null>(null)

function formatDate(value?: string) {
  if (!value) return ''
  const date = new Date(value)
  if (Number.isNaN(date.getTime())) return value
  return date.toLocaleString()
}

function saveScroll() {
  store.followingScrollTop = window.scrollY
}

function restoreScroll() {
  nextTick(() => window.scrollTo({ top: store.followingScrollTop }))
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
  border-bottom: 1px solid #1f1f1f;
  transition: background 0.15s ease;
}

.topic-row:first-child {
  border-top: 1px solid #1f1f1f;
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
  padding: 60px 24px;
  text-align: center;
  font-size: 14px;
  color: #8c8c8c;
}
</style>
