<template>
  <div class="list" v-loading="loading">
    <div v-for="user in users" :key="user.follower_id" class="card">
      <UserAvatar :user-id="user.follower_id" :username="user.username" :size="48" />
      <div class="user-info">
        <span class="username">{{ user.username }}</span>
      </div>
      <span class="date">{{ formatDate(user.created_at) }}</span>
    </div>
    <div v-if="!loading && users.length === 0" class="empty">
      <p>No followers yet.</p>
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
const { followers: users } = storeToRefs(store)
const loading = ref(false)

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

.empty {
  padding: 80px 24px;
  border: 1px dashed #262626;
  border-radius: 12px;
  text-align: center;
  font-size: 16px;
  color: #8c8c8c;
}
</style>
