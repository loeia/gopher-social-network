<template>
  <div class="followers-page">
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
import { computed, onMounted, ref, watch } from 'vue'
import { storeToRefs } from 'pinia'
import { useFeedStore } from '@/stores/feed'
import { getCurrentUserId, handleApiError } from '@/api'
import UserAvatar from '@/components/UserAvatar.vue'

const props = withDefaults(defineProps<{ userId?: number }>(), { userId: 0 })

const store = useFeedStore()
const { followers: users } = storeToRefs(store)
const loading = ref(false)

// Target user: the profile being viewed, falling back to the current user
// (e.g. when shown from the home page views).
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
    await store.fetchFollowers(targetUserId.value)
  } catch (error) {
    handleApiError(error, 'Failed to load followers')
  } finally {
    loading.value = false
  }
}

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
