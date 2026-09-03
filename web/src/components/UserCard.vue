<template>
  <Transition name="card-fade">
    <div
      v-if="visible"
      ref="cardRef"
      class="user-card"
      @click.stop
    >
      <div class="card-content">
        <div class="card-avatar-section">
          <div class="card-avatar avatar-link" @click="goToProfile">
            <UserAvatar :user-id="userId" :username="username" :size="64" />
          </div>
          <button
            v-if="canFollow"
            class="follow-btn"
            :class="{ 'is-following': isFollowing }"
            :disabled="followLoading"
            @click="toggleFollow"
          >
            {{ isFollowing ? 'Unfollow' : 'Follow' }}
          </button>
        </div>
        <div class="card-info">
          <div class="card-username">{{ userData?.username || username }}</div>
          <div v-if="userData?.bio" class="card-bio">{{ userData.bio }}</div>
          <div class="card-joined">
            <svg class="joined-icon" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
              <rect x="3" y="4" width="18" height="18" rx="2" ry="2" />
              <line x1="16" y1="2" x2="16" y2="6" />
              <line x1="8" y1="2" x2="8" y2="6" />
              <line x1="3" y1="10" x2="21" y2="10" />
            </svg>
            <span>Joined {{ formatDate(userData?.created_at) }}</span>
          </div>
          <div class="card-stats">
            <span class="stat-item">
              <strong>{{ userData?.posts_count ?? 0 }}</strong>
              <span>Posts</span>
            </span>
            <span class="stat-item">
              <strong>{{ userData?.followers_count ?? 0 }}</strong>
              <span>Followers</span>
            </span>
            <span class="stat-item">
              <strong>{{ userData?.following_count ?? 0 }}</strong>
              <span>Following</span>
            </span>
          </div>
        </div>
      </div>
    </div>
  </Transition>
</template>

<script setup lang="ts">
import { computed, nextTick, onBeforeUnmount, ref } from 'vue'
import { useRouter } from 'vue-router'
import { apiFetch, getCurrentUserId, getToken } from '@/api'
import UserAvatar from '@/components/UserAvatar.vue'

const router = useRouter()

interface UserData {
  id: number
  username: string
  avatar_url: string
  bio: string
  created_at: string
  followers_count: number
  following_count: number
  posts_count: number
}

const props = defineProps<{
  userId: number | null
  username: string
}>()

const visible = ref(false)
const userData = ref<UserData | null>(null)
const loading = ref(false)
const cardRef = ref<HTMLElement | null>(null)
const triggerElement = ref<HTMLElement | null>(null)

const isFollowing = ref(false)
const followLoading = ref(false)

const isLoggedIn = computed(() => !!getToken())
const currentUserId = computed(() => getCurrentUserId())
const canFollow = computed(() => {
  if (!isLoggedIn.value || !props.userId) return false
  return currentUserId.value !== props.userId
})

async function checkFollowStatus() {
  if (!isLoggedIn.value || !currentUserId.value || !props.userId) {
    isFollowing.value = false
    return
  }
  try {
    const response = await apiFetch(`/users/${currentUserId.value}/following`)
    if (!response.ok) return
    const json = await response.json()
    const following = Array.isArray(json) ? json : (json.data ?? [])
    isFollowing.value = following.some((f: { following_id: number }) => f.following_id === props.userId)
  } catch {
    isFollowing.value = false
  }
}

async function toggleFollow() {
  if (!isLoggedIn.value || !props.userId || followLoading.value) return
  followLoading.value = true
  try {
    const method = isFollowing.value ? 'DELETE' : 'PUT'
    const response = await apiFetch(`/users/${props.userId}/follow`, { method })
    if (!response.ok) throw new Error(`HTTP ${response.status}`)
    isFollowing.value = !isFollowing.value
    if (userData.value) {
      userData.value.followers_count += isFollowing.value ? 1 : -1
    }
  } catch (error) {
    console.error('Follow error:', error)
  } finally {
    followLoading.value = false
  }
}

const cardPosition = computed(() => {
  if (!triggerElement.value) return { top: '0px', left: '0px', display: 'none' as const }
  
  const rect = triggerElement.value.getBoundingClientRect()
  const parent = triggerElement.value.offsetParent as HTMLElement | null
  
  if (!parent) return { top: '0px', left: '0px', display: 'none' as const }
  
  const parentRect = parent.getBoundingClientRect()
  const cardWidth = 400
  const gap = 12
  
  let left = rect.right - parentRect.left + gap
  let top = rect.top - parentRect.top
  
  if (left + cardWidth > parentRect.width) {
    left = rect.left - parentRect.left - cardWidth - gap
  }
  if (left < 0) left = 0
  
  return {
    position: 'absolute' as const,
    top: `${top}px`,
    left: `${left}px`,
  }
})

async function fetchUserData() {
  if (!props.userId) return
  loading.value = true
  try {
    const response = await apiFetch(`/users/${props.userId}`)
    if (!response.ok) throw new Error(`HTTP ${response.status}`)
    const json = await response.json()
    userData.value = json.data ?? json
  } catch (error) {
    console.error('Failed to load user data:', error)
  } finally {
    loading.value = false
  }
}

function show(event: MouseEvent) {
  event.stopPropagation()
  triggerElement.value = event.currentTarget as HTMLElement
  visible.value = true
  fetchUserData()
  checkFollowStatus()
  nextTick(() => {
    setTimeout(() => {
      document.addEventListener('click', handleOutsideClick)
    }, 0)
  })
}

function hide() {
  visible.value = false
  userData.value = null
  document.removeEventListener('click', handleOutsideClick)
}

function handleOutsideClick(e: MouseEvent) {
  if (cardRef.value && !cardRef.value.contains(e.target as Node)) {
    hide()
  }
}

function formatDate(value?: string) {
  if (!value) return ''
  const date = new Date(value)
  if (Number.isNaN(date.getTime())) return value
  const y = date.getFullYear()
  const m = String(date.getMonth() + 1).padStart(2, '0')
  const d = String(date.getDate()).padStart(2, '0')
  return `${y}/${m}/${d}`
}

function goToProfile() {
  const id = userData.value?.id || props.userId
  if (id) {
    hide()
    router.push(`/users/${id}`)
  }
}

onBeforeUnmount(() => {
  document.removeEventListener('click', handleOutsideClick)
})

defineExpose({ show, hide })
</script>

<style scoped>
.user-card {
  position: absolute;
  width: 400px;
  background: #1e1e1e;
  border: 1px solid #333;
  border-radius: 12px;
  box-shadow: 0 8px 24px rgba(0, 0, 0, 0.5);
  overflow: hidden;
  z-index: 999;
}

.card-content {
  display: flex;
  gap: 16px;
  padding: 20px;
}

.card-avatar-section {
  flex-shrink: 0;
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 8px;
}

.card-avatar {
  cursor: pointer;
  transition: opacity 0.2s ease;
}

.card-avatar:hover {
  opacity: 0.8;
}

.follow-btn {
  display: inline-flex;
  align-items: center;
  padding: 2px 10px;
  font-size: 12px;
  font-weight: 600;
  color: #e4e6e8;
  background: #262626;
  border: 1px solid #333;
  border-radius: 6px;
  cursor: pointer;
  transition: all 0.2s ease;
}

.follow-btn:hover:not(:disabled) {
  background: #333;
  color: #e4e6e8;
  border-color: #595959;
}

.follow-btn.is-following {
  background: transparent;
  border: 1px solid #333;
  color: #8c8c8c;
}

.follow-btn.is-following:hover:not(:disabled) {
  background: #262626;
  color: #e4e6e8;
  border-color: #595959;
}

.follow-btn:disabled {
  cursor: not-allowed;
  opacity: 0.6;
}

.card-info {
  flex: 1;
  min-width: 0;
}

.card-username {
  font-size: 16px;
  font-weight: 600;
  color: #e4e6e8;
  margin-bottom: 6px;
}

.card-bio {
  font-size: 13px;
  line-height: 1.5;
  color: #8c8c8c;
  margin-bottom: 8px;
  word-break: break-word;
}

.card-joined {
  display: flex;
  align-items: center;
  gap: 6px;
  font-size: 12px;
  color: #595959;
  margin-bottom: 12px;
}

.joined-icon {
  width: 14px;
  height: 14px;
}

.card-stats {
  display: flex;
  gap: 16px;
  padding-top: 12px;
  border-top: 1px solid #333;
}

.stat-item {
  display: flex;
  align-items: center;
  gap: 4px;
}

.stat-item strong {
  font-size: 14px;
  font-weight: 700;
  color: #e4e6e8;
}

.stat-item span {
  font-size: 12px;
  color: #8c8c8c;
}

.card-fade-enter-active,
.card-fade-leave-active {
  transition: opacity 0.15s ease;
}

.card-fade-enter-from,
.card-fade-leave-to {
  opacity: 0;
}
</style>
