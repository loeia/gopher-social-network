<template>
  <div class="followers-page">
    <div class="back-nav">
      <el-button text @click="goBack">← Back</el-button>
    </div>
    <div class="list" v-loading="loading">
      <div v-for="user in users" :key="user.follower_id" class="topic-row">
        <UserAvatar :user-id="user.follower_id" :username="user.username" :size="36" />
        <div class="topic-main">
          <span class="topic-username">{{ user.username }}</span>
          <span class="topic-time">{{ formatDate(user.created_at) }}</span>
        </div>
      </div>
      <div v-if="!loading && users.length === 0" class="empty">
        <p>No followers yet.</p>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { nextTick, onActivated, onBeforeUnmount, onDeactivated, onMounted, ref } from 'vue'
import { useRouter } from 'vue-router'
import { storeToRefs } from 'pinia'
import { useFeedStore } from '@/stores/feed'
import { notify } from '@/utils/message'
import UserAvatar from '@/components/UserAvatar.vue'

const router = useRouter()
const store = useFeedStore()
const { followers: users } = storeToRefs(store)
const loading = ref(false)

function goBack() {
  const state = window.history.state as { back?: unknown } | null
  if (state && state.back != null) {
    router.back()
  } else {
    router.push('/')
  }
}

function formatDate(value?: string) {
  if (!value) return ''
  const date = new Date(value)
  if (Number.isNaN(date.getTime())) return value
  return date.toLocaleString()
}

function saveScroll() {
  store.followersScrollTop = window.scrollY
}

function restoreScroll() {
  nextTick(() => window.scrollTo({ top: store.followersScrollTop }))
}

async function loadFollowers() {
  loading.value = true
  try {
    await store.fetchFollowers()
  } catch (error) {
    console.error('Load followers error:', error)
    notify('error', 'Failed to load followers')
  } finally {
    loading.value = false
  }
}

onMounted(() => {
  restoreScroll()
  loadFollowers()
})
onActivated(() => {
  restoreScroll()
  loadFollowers()
})
onDeactivated(saveScroll)
onBeforeUnmount(saveScroll)
</script>

<style scoped>
.followers-page {
  min-height: 100vh;
  padding: 32px 0 80px;
}

.back-nav {
  margin: 0 0 16px;
  padding: 0 24px;
  display: flex;
  align-items: center;
  gap: 12px;
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

.empty {
  padding: 60px 24px;
  text-align: center;
  font-size: 14px;
  color: #8c8c8c;
}
</style>
