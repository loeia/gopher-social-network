<template>
  <aside class="sidebar">
    <div class="user-box">
      <button
        type="button"
        class="avatar-btn"
        :title="'Change avatar'"
        :aria-label="'Change avatar'"
        @click="pickFile"
      >
        <img
          v-if="showAvatar && avatarSrc"
          :src="avatarSrc"
          class="avatar-img"
          alt="avatar"
          @error="onAvatarError"
        />
        <span v-else class="avatar-initials" :style="{ background: initialsColor }">
          {{ initials }}
        </span>
        <span class="avatar-overlay">
          <svg
            width="18"
            height="18"
            viewBox="0 0 24 24"
            fill="none"
            stroke="currentColor"
            stroke-width="2"
            stroke-linecap="round"
            stroke-linejoin="round"
          >
            <path d="M23 19a2 2 0 0 1-2 2H3a2 2 0 0 1-2-2V8a2 2 0 0 1 2-2h4l2-3h6l2 3h4a2 2 0 0 1 2 2z" />
            <circle cx="12" cy="13" r="4" />
          </svg>
        </span>
      </button>
      <span class="user-name">{{ userStore.username }}</span>
      <el-button
      class="btn profile-btn"
      :class="{ active: activeView === 'profile' }"
      @click="goToProfile"
    >
      Profile
    </el-button>
      <input
        ref="fileInput"
        type="file"
        accept="image/jpeg,image/png"
        hidden
        @change="onFileChange"
      />
    </div>

    <AvatarCropDialog :visible="cropVisible" :src="cropSrc" @close="cropVisible = false" @confirm="handleCropConfirm" />

    <el-button class="btn home-btn" :class="{ active: activeView === 'all' }" @click="goHome"
      >Home</el-button
    >
    <el-button
      class="btn create-btn"
      :class="{ active: activeView === 'create' }"
      @click="goToCreate"
    >
      Create
    </el-button>
    <el-button
      class="btn my-posts-btn"
      :class="{ active: activeView === 'myposts' }"
      @click="goToMyPosts"
    >
      My Posts
    </el-button>
    <el-button
      class="btn likes-btn"
      :class="{ active: activeView === 'liked' }"
      @click="toggleLikes"
    >
      Post Likes
    </el-button>
    <el-button
      class="btn comment-likes-btn"
      :class="{ active: activeView === 'likedcomments' }"
      @click="toggleLikedComments"
    >
      Comment Likes
    </el-button>
    <el-button
      class="btn following-btn"
      :class="{ active: activeView === 'following' }"
      @click="toggleFollowing"
    >
      Following
    </el-button>
    <el-button
      class="btn followers-btn"
      :class="{ active: activeView === 'followers' }"
      @click="toggleFollowers"
    >
      Followers
    </el-button>
    <el-button class="btn settings-btn" :class="{ active: activeView === 'settings' }" @click="goToSettings">
      Settings
    </el-button>
  </aside>
</template>

<script setup lang="ts">
import { computed, onMounted, ref, watchEffect } from 'vue'
import { useRouter } from 'vue-router'
import { apiFetch, getApiError } from '@/api'
import { API_URL } from '@/main'
import { notify } from '@/utils/message'
import { useUserStore } from '@/stores/user'
import AvatarCropDialog from '@/components/AvatarCropDialog.vue'
import type { ViewType } from '@/stores/feed'

const MAX_AVATAR_SIZE = 2 * 1024 * 1024

const AVATAR_COLORS = ['#5b8def', '#e07b5a', '#57c08a', '#c07bd8', '#d8a24a', '#6ec7c0']

const props = defineProps<{ activeView: ViewType | null }>()

const emit = defineEmits<{ view: [view: ViewType] }>()

const router = useRouter()
const userStore = useUserStore()

const fileInput = ref<HTMLInputElement | null>(null)
const uploading = ref(false)
const showAvatar = ref(false)
const avatarVersion = ref(0)
const cropVisible = ref(false)
const cropSrc = ref('')

const avatarSrc = computed(() => {
  if (!userStore.avatarUrl) return ''
  return `${API_URL}${userStore.avatarUrl}?v=${avatarVersion.value}`
})

const initials = computed(() => {
  const name = userStore.username.trim()
  return name ? name[0]!.toUpperCase() : '?'
})

const initialsColor = computed(() => {
  const name = userStore.username
  let hash = 0
  for (let i = 0; i < name.length; i++) hash = (hash * 31 + name.charCodeAt(i)) >>> 0
  return AVATAR_COLORS[hash % AVATAR_COLORS.length]
})

watchEffect(() => {
  showAvatar.value = !!userStore.avatarUrl
})

onMounted(() => {
  userStore.fetchCurrentUser()
})

function pickFile() {
  fileInput.value?.click()
}

function onAvatarError() {
  showAvatar.value = false
}

function readFileAsDataURL(file: File): Promise<string> {
  return new Promise((resolve, reject) => {
    const reader = new FileReader()
    reader.onload = () => resolve(reader.result as string)
    reader.onerror = () => reject(reader.error)
    reader.readAsDataURL(file)
  })
}

async function onFileChange(event: Event) {
  const input = event.target as HTMLInputElement
  const file = input.files?.[0]
  input.value = ''
  if (!file) return

  if (!/^image\/(jpeg|png)$/.test(file.type)) {
    notify('error', 'Only JPEG or PNG images are allowed')
    return
  }
  if (file.size > MAX_AVATAR_SIZE) {
    notify('error', 'Image must be 2 MB or smaller')
    return
  }

  try {
    cropSrc.value = await readFileAsDataURL(file)
    cropVisible.value = true
  } catch (error) {
    console.error('Read image error:', error)
    notify('error', 'Failed to read image')
  }
}

async function handleCropConfirm(blob: Blob) {
  cropVisible.value = false
  await uploadAvatar(blob)
}

async function uploadAvatar(blob: Blob) {
  uploading.value = true
  try {
    const form = new FormData()
    form.append('avatar', blob, 'avatar.jpg')
    const response = await apiFetch('/users/me/avatar', { method: 'PUT', body: form })
    if (!response.ok) {
      if (response.status === 401) return
      const message = (await getApiError(response)) ?? `Failed to upload avatar (HTTP ${response.status})`
      notify('error', message)
      return
    }
    showAvatar.value = true
    avatarVersion.value += 1
    await userStore.fetchCurrentUser(true)
    notify('success', 'Avatar updated')
  } catch (error) {
    console.error('Upload avatar error:', error)
    notify('error', 'Failed to upload avatar')
  } finally {
    uploading.value = false
  }
}

function goToCreate() {
  router.push('/posts/new')
}

function goToProfile() {
  if (userStore.id) {
    router.push(`/users/${userStore.id}`)
    return
  }
  userStore.fetchCurrentUser().then(() => {
    if (userStore.id) router.push(`/users/${userStore.id}`)
  })
}

function goToMyPosts() {
  router.push('/my-posts')
}

function goToSettings() {
  router.push('/settings')
}

function goHome() {
  emit('view', 'all')
  router.push('/')
}

function toggleLikes() {
  emit('view', props.activeView === 'liked' ? 'all' : 'liked')
  router.push('/')
}

function toggleLikedComments() {
  emit('view', props.activeView === 'likedcomments' ? 'all' : 'likedcomments')
  router.push('/')
}

function toggleFollowing() {
  emit('view', props.activeView === 'following' ? 'all' : 'following')
  router.push('/')
}

function toggleFollowers() {
  emit('view', props.activeView === 'followers' ? 'all' : 'followers')
  router.push('/')
}
</script>

<style scoped>
.sidebar {
  position: fixed;
  top: 0;
  left: 0;
  bottom: 0;
  width: 180px;
  padding: 80px 16px;
  background: #141414;
  border-right: 1px solid #262626;
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.user-box {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 8px;
  padding-bottom: 16px;
  margin-bottom: 4px;
  border-bottom: 1px solid #262626;
}

.avatar-btn {
  position: relative;
  width: 56px;
  height: 56px;
  padding: 0;
  border: none;
  border-radius: 50%;
  background: transparent;
  cursor: pointer;
  overflow: hidden;
}

.avatar-img,
.avatar-initials {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 100%;
  height: 100%;
  border-radius: 50%;
  border: 1px solid #262626;
}

.avatar-img {
  object-fit: cover;
  display: block;
}

.avatar-initials {
  font-size: 22px;
  font-weight: 700;
  color: #ffffff;
}

.avatar-overlay {
  position: absolute;
  inset: 0;
  display: flex;
  align-items: center;
  justify-content: center;
  border-radius: 50%;
  background: rgba(0, 0, 0, 0.55);
  color: #ffffff;
  opacity: 0;
  transition: opacity 0.2s ease;
}

.avatar-btn:hover .avatar-overlay,
.avatar-btn:focus-visible .avatar-overlay {
  opacity: 1;
}

.avatar-btn:focus-visible {
  outline: 2px solid rgba(255, 255, 255, 0.4);
  outline-offset: 2px;
}

.user-name {
  max-width: 100%;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  font-size: 13px;
  font-weight: 600;
  color: #e4e6e8;
}

.btn,
.btn + .btn {
  width: 100%;
  margin-left: 0;
  font-weight: 600;
  background: #ffffff;
  color: #141414;
  border: 1px solid #ffffff;
}

.btn:hover {
  background: #e4e6e8;
  color: #141414;
}

.btn.active {
  background: #141414;
  color: #ffffff;
  border: 1px solid #ffffff;
}

</style>
