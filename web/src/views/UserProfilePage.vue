<template>
  <div class="user-profile-page">
    <div class="profile-card" v-loading="loading">
      <template v-if="notFound">
        <h1 class="not-found-title">User not found</h1>
        <p class="not-found-hint">This user may have been removed or the link is incorrect.</p>
      </template>
      <template v-else>
        <div class="profile-header">
          <div class="avatar-wrapper" :class="{ own: isOwnProfile }">
            <UserAvatar :user-id="userId" :username="user.username" :size="112" />
            <label v-if="isOwnProfile" class="avatar-overlay">
              <input
                ref="avatarInputRef"
                type="file"
                accept="image/*"
                class="hidden-input"
                @change="onAvatarFileChange"
              />
              <svg
                class="camera-icon"
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
            </label>
          </div>

          <h1 class="username">{{ user.username }}</h1>
          <span class="handle">Joined {{ formatDate(user.created_at) }}</span>

          <p v-if="user.bio" class="bio">{{ user.bio }}</p>

          <div class="meta-row">
            <span v-for="(link, index) in user.links" :key="index" class="meta-item">
              <svg
                class="info-icon"
                viewBox="0 0 24 24"
                fill="none"
                stroke="currentColor"
                stroke-width="2"
                stroke-linecap="round"
                stroke-linejoin="round"
              >
                <path d="M10 13a5 5 0 0 0 7.54.54l3-3a5 5 0 0 0-7.07-7.07l-1.72 1.71" />
                <path d="M14 11a5 5 0 0 0-7.54-.54l-3 3a5 5 0 0 0 7.07 7.07l1.71-1.71" />
              </svg>
              <a :href="link" class="website" target="_blank" rel="noopener noreferrer">
                {{ link }}
              </a>
            </span>
          </div>

          <div class="stats-row">
            <span class="stat">
              <strong>{{ userPosts.length }}</strong>
              <span class="stat-label">Posts</span>
            </span>
            <span class="stat">
              <strong>{{ userReplies.length }}</strong>
              <span class="stat-label">Replies</span>
            </span>
            <span class="stat">
              <strong>{{ userLikedPosts.length }}</strong>
              <span class="stat-label">Likes</span>
            </span>
            <span
              class="stat"
              :class="{ clickable: isOwnProfile }"
              @click="isOwnProfile && goToFollowers()"
            >
              <strong>{{ user.followers_count }}</strong>
              <span class="stat-label">Followers</span>
            </span>
            <span
              class="stat"
              :class="{ clickable: isOwnProfile }"
              @click="isOwnProfile && goToFollowing()"
            >
              <strong>{{ user.following_count }}</strong>
              <span class="stat-label">Following</span>
            </span>
          </div>

          <el-button v-if="isOwnProfile" class="edit-btn" @click="openEdit">
            Edit profile
          </el-button>
        </div>
      </template>
    </div>

    <div v-if="!notFound && !loading" class="profile-tabs">
      <div class="tabs-bar">
        <button
          v-for="tab in tabs"
          :key="tab.key"
          class="tab-btn"
          :class="{ active: activeTab === tab.key }"
          @click="activeTab = tab.key"
        >
          {{ tab.label }}
        </button>
      </div>

      <div class="tabs-content">
        <div v-show="activeTab === 'posts'" class="tab-panel">
          <div v-loading="postsLoading" class="feed">
            <div
              v-for="post in userPosts"
              :key="post.id"
              class="topic-row"
              @click="openPost(post.id)"
            >
              <div class="topic-top">
                <h2 class="topic-title">{{ post.title }}</h2>
                <div class="topic-stats">
                  <span class="topic-stat" :title="`${post.comment_count} comments`">
                    <svg class="stat-icon" viewBox="0 0 24 24" aria-hidden="true">
                      <path d="M20 2H4a2 2 0 0 0-2 2v18l4-4h14a2 2 0 0 0 2-2V4a2 2 0 0 0-2-2z" />
                    </svg>
                    {{ post.comment_count }}
                  </span>
                  <span class="topic-stat" :title="`${post.like_count} likes`">
                    <svg class="stat-icon like-icon" viewBox="0 0 24 24" aria-hidden="true">
                      <path
                        d="M12 21.35l-1.45-1.32C5.4 15.36 2 12.28 2 8.5 2 5.42 4.42 3 7.5 3c1.74 0 3.41.81 4.5 2.09C13.09 3.81 14.76 3 16.5 3 19.58 3 22 5.42 22 8.5c0 3.78-3.4 6.86-8.55 11.54L12 21.35z"
                      />
                    </svg>
                    {{ post.like_count }}
                  </span>
                </div>
              </div>
              <div class="topic-meta">
                <span class="topic-date">{{ formatDate(post.created_at) }}</span>
                <template v-if="post.tags && post.tags.length">
                  <span class="meta-dot">&middot;</span>
                  <span v-for="tag in post.tags" :key="tag" class="topic-tag">{{ tag }}</span>
                </template>
              </div>
            </div>
            <div v-if="!postsLoading && userPosts.length === 0" class="empty-hint">
              No posts yet.
            </div>
          </div>
        </div>

        <div v-show="activeTab === 'replies'" class="tab-panel">
          <div v-loading="repliesLoading" class="feed">
            <div
              v-for="reply in userReplies"
              :key="reply.id"
              class="topic-row reply-row"
              @click="openPost(reply.post_id)"
            >
              <div class="reply-context">
                <svg class="reply-icon" viewBox="0 0 24 24" aria-hidden="true">
                  <path d="M9 17H7v-7h2v7zm4 0h-2V7h2v10zm4 0h-2v-4h2v4z" />
                </svg>
                <span class="reply-hint">replied to a post</span>
              </div>
              <p class="reply-content">{{ reply.content }}</p>
              <div class="topic-meta">
                <span class="topic-date">{{ formatDate(reply.created_at) }}</span>
              </div>
            </div>
            <div v-if="!repliesLoading && userReplies.length === 0" class="empty-hint">
              No replies yet.
            </div>
          </div>
        </div>

        <div v-show="activeTab === 'likes'" class="tab-panel">
          <div v-loading="likesLoading" class="feed">
            <div
              v-for="post in userLikedPosts"
              :key="post.post_id"
              class="topic-row"
              @click="openPost(post.post_id)"
            >
              <div class="topic-top">
                <h2 class="topic-title">{{ post.title }}</h2>
                <div class="topic-stats">
                  <span class="topic-stat" :title="`${post.comment_count} comments`">
                    <svg class="stat-icon" viewBox="0 0 24 24" aria-hidden="true">
                      <path d="M20 2H4a2 2 0 0 0-2 2v18l4-4h14a2 2 0 0 0 2-2V4a2 2 0 0 0-2-2z" />
                    </svg>
                    {{ post.comment_count ?? 0 }}
                  </span>
                  <span class="topic-stat" :title="`${post.like_count} likes`">
                    <svg class="stat-icon like-icon" viewBox="0 0 24 24" aria-hidden="true">
                      <path
                        d="M12 21.35l-1.45-1.32C5.4 15.36 2 12.28 2 8.5 2 5.42 4.42 3 7.5 3c1.74 0 3.41.81 4.5 2.09C13.09 3.81 14.76 3 16.5 3 19.58 3 22 5.42 22 8.5c0 3.78-3.4 6.86-8.55 11.54L12 21.35z"
                      />
                    </svg>
                    {{ post.like_count ?? 0 }}
                  </span>
                </div>
              </div>
              <div class="topic-bottom">
                <span class="topic-author">{{ post.author }}</span>
                <template v-if="post.tags && post.tags.length">
                  <span class="meta-dot">&middot;</span>
                  <span v-for="tag in post.tags" :key="tag" class="topic-tag">{{ tag }}</span>
                </template>
              </div>
            </div>
            <div v-if="!likesLoading && userLikedPosts.length === 0" class="empty-hint">
              No liked posts yet.
            </div>
          </div>
        </div>
      </div>
    </div>

    <el-dialog
      :model-value="editVisible"
      title="Edit profile"
      width="520px"
      append-to-body
      @update:model-value="editVisible = $event"
    >
      <div class="field-group">
        <label class="field-label" for="profile-links">Links</label>
        <div class="links-list">
          <div
            v-for="index in 5"
            :key="index"
            class="link-field"
            :class="{ 'has-error': linkInvalid(index - 1) }"
          >
            <el-input
              :model-value="links[index - 1]"
              :placeholder="`https://link-${index}.com`"
              class="field"
              @update:model-value="onLinkInput(index - 1, $event)"
            />
            <span v-if="linkInvalid(index - 1)" class="link-hint">
              Must start with http(s)://
            </span>
          </div>
        </div>
      </div>
      <div class="field-group">
        <div class="field-label-row">
          <label class="field-label" for="profile-bio">Bio</label>
          <span class="char-count" :class="{ over: bioOverLimit }">{{ bio.length }}/500</span>
        </div>
        <el-input
          id="profile-bio"
          v-model="bio"
          type="textarea"
          :rows="4"
          placeholder="Tell us a bit about yourself"
          class="field"
          :class="{ 'is-over': bioOverLimit }"
        />
      </div>
      <template #footer>
        <div class="dialog-footer">
          <el-button @click="editVisible = false">Cancel</el-button>
          <el-button
            class="save-btn"
            :loading="saving"
            :disabled="bioOverLimit || hasInvalidLink"
            @click="saveProfile"
          >
            Save
          </el-button>
        </div>
      </template>
    </el-dialog>

    <AvatarCropDialog
      :visible="cropVisible"
      :src="cropSrc"
      @confirm="uploadAvatar"
      @close="cropVisible = false"
    />
  </div>
</template>

<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { apiFetch, getApiError, getCurrentUserId, handleApiError } from '@/api'
import { notify } from '@/utils/message'
import UserAvatar from '@/components/UserAvatar.vue'
import AvatarCropDialog from '@/components/AvatarCropDialog.vue'
import { useUserStore } from '@/stores/user'

defineOptions({ name: 'UserProfilePage' })

interface UserProfile {
  id: number
  username: string
  created_at: string
  bio: string
  links: string[]
  followers_count: number
  following_count: number
}

interface UserPost {
  id: number
  title: string
  tags: string[]
  comment_count: number
  like_count: number
  created_at: string
}

interface UserReply {
  id: number
  post_id: number
  content: string
  created_at: string
}

interface LikedPost {
  post_id: number
  author: string
  title: string
  tags: string[]
  comment_count?: number
  like_count?: number
  created_at: string
}

const route = useRoute()
const router = useRouter()
const userStore = useUserStore()

// ID of the user whose data is currently loaded. Used to avoid re-fetching when
// returning to the same profile (keep-alive restores the cached DOM, so the
// router's scrollBehavior can restore the exact scroll position, same as Home).
let loadedUserId = 0

const loading = ref(false)
const notFound = ref(false)
const user = ref<UserProfile>({
  id: 0,
  username: '',
  created_at: '',
  bio: '',
  links: [],
  followers_count: 0,
  following_count: 0,
})

const editVisible = ref(false)
const saving = ref(false)
const links = ref<string[]>(Array(5).fill(''))
const bio = ref('')
const avatarInputRef = ref<HTMLInputElement | null>(null)
const cropVisible = ref(false)
const cropSrc = ref('')

const activeTab = ref<'posts' | 'replies' | 'likes'>('posts')
const tabs = [
  { key: 'posts' as const, label: 'Posts' },
  { key: 'replies' as const, label: 'Replies' },
  { key: 'likes' as const, label: 'Likes' },
]

const userPosts = ref<UserPost[]>([])
const postsLoading = ref(false)
const userReplies = ref<UserReply[]>([])
const repliesLoading = ref(false)
const userLikedPosts = ref<LikedPost[]>([])
const likesLoading = ref(false)

const userId = computed(() => Number(route.params.userId))
const isOwnProfile = computed(() => getCurrentUserId() === userId.value)
const bioOverLimit = computed(() => bio.value.length > 500)

function isValidLink(value: string) {
  const trimmed = value.trim()
  return trimmed === '' || /^https?:\/\//i.test(trimmed)
}

function linkInvalid(index: number) {
  return !isValidLink(links.value[index] ?? '')
}

const hasInvalidLink = computed(() => links.value.some((link) => !isValidLink(link)))

function formatDate(value?: string) {
  if (!value) return ''
  const date = new Date(value)
  if (Number.isNaN(date.getTime())) return value
  const y = date.getFullYear()
  const m = String(date.getMonth() + 1).padStart(2, '0')
  const d = String(date.getDate()).padStart(2, '0')
  return `${y}/${m}/${d}`
}

function onLinkInput(index: number, value: string) {
  links.value[index] = value
}

function openEdit() {
  bio.value = user.value.bio
  links.value = Array(5).fill('')
  ;(user.value.links ?? []).slice(0, 5).forEach((link, index) => {
    links.value[index] = link
  })
  editVisible.value = true
}

async function saveProfile() {
  if (bioOverLimit.value) {
    notify('error', 'Bio must be 500 characters or fewer')
    return
  }
  if (hasInvalidLink.value) {
    notify('error', 'Links must start with http(s)://')
    return
  }
  const filledLinks = links.value.map((link) => link.trim()).filter(Boolean)
  saving.value = true
  try {
    const response = await apiFetch('/users/me/profile', {
      method: 'PUT',
      body: JSON.stringify({
        bio: bio.value.trim(),
        links: filledLinks,
      }),
    })
    if (!response.ok) {
      const message =
        (await getApiError(response)) ?? `Failed to save profile (HTTP ${response.status})`
      notify('error', message)
      return
    }
    editVisible.value = false
    await load()
    notify('success', 'Profile updated')
  } catch (error) {
    handleApiError(error, 'Failed to save profile')
  } finally {
    saving.value = false
  }
}

function goToFollowers() {
  router.push({ name: 'Followers' })
}

function goToFollowing() {
  router.push({ name: 'Following' })
}

function onAvatarFileChange(e: Event) {
  const file = (e.target as HTMLInputElement).files?.[0]
  if (!file) return
  cropSrc.value = URL.createObjectURL(file)
  cropVisible.value = true
  if (avatarInputRef.value) avatarInputRef.value.value = ''
}

async function uploadAvatar(blob: Blob) {
  try {
    const formData = new FormData()
    formData.append('avatar', blob, 'avatar.jpg')
    const response = await apiFetch('/users/me/avatar', {
      method: 'PUT',
      body: formData,
    })
    if (!response.ok) {
      const message = (await getApiError(response)) ?? `Failed to upload avatar (HTTP ${response.status})`
      notify('error', message)
      return
    }
    userStore.bumpAvatarVersion()
    cropVisible.value = false
    notify('success', 'Avatar updated')
  } catch (error) {
    handleApiError(error, 'Failed to upload avatar')
  }
}

async function load() {
  const id = userId.value
  if (!id) return
  loadedUserId = id
  loading.value = true
  notFound.value = false
  user.value = {
    id,
    username: '',
    created_at: '',
    bio: '',
    links: [],
    followers_count: 0,
    following_count: 0,
  }
  try {
    const response = await apiFetch(`/users/${id}`)
    if (!response.ok) {
      if (response.status === 404) {
        notFound.value = true
        return
      }
      throw new Error(`HTTP ${response.status}`)
    }
    const json = await response.json()
    const data = json.data ?? json
    user.value = {
      id: Number(data.id),
      username: data.username ?? '',
      created_at: data.created_at ?? '',
      bio: data.bio ?? '',
      links: Array.isArray(data.links) ? data.links : [],
      followers_count: Number(data.followers_count ?? 0),
      following_count: Number(data.following_count ?? 0),
    }
    // Preload all tab data up front so switching tabs never has to
    // mount an empty panel (which collapses the page height and makes
    // the scroll position jump/jitter when scrolled down).
    await Promise.all([loadUserPosts(), loadUserReplies(), loadUserLikedPosts()])
  } catch (error) {
    handleApiError(error, 'Failed to load profile')
  } finally {
    loading.value = false
  }
}

async function loadUserPosts() {
  if (postsLoading.value) return
  postsLoading.value = true
  try {
    const response = await apiFetch(`/users/${userId.value}/posts`)
    if (!response.ok) throw new Error(`HTTP ${response.status}`)
    const json = await response.json()
    const raw = Array.isArray(json) ? json : (json.data ?? [])
    userPosts.value = raw.map((p: Record<string, unknown>) => ({
      id: Number(p.post_id ?? p.id ?? 0),
      title: String(p.title ?? ''),
      tags: Array.isArray(p.tags) ? p.tags : [],
      comment_count: Number(p.comment_count ?? 0),
      like_count: Number(p.like_count ?? 0),
      created_at: String(p.created_at ?? ''),
    }))
  } catch (error) {
    handleApiError(error, 'Failed to load posts')
  } finally {
    postsLoading.value = false
  }
}

async function loadUserReplies() {
  if (repliesLoading.value) return
  repliesLoading.value = true
  try {
    const response = await apiFetch(`/users/${userId.value}/comments`)
    if (!response.ok) throw new Error(`HTTP ${response.status}`)
    const json = await response.json()
    const raw = Array.isArray(json) ? json : (json.data ?? [])
    userReplies.value = raw.map((c: Record<string, unknown>) => ({
      id: Number(c.comment_id ?? c.id ?? 0),
      post_id: Number(c.post_id ?? 0),
      content: String(c.content ?? ''),
      created_at: String(c.created_at ?? ''),
    }))
  } catch (error) {
    handleApiError(error, 'Failed to load replies')
  } finally {
    repliesLoading.value = false
  }
}

async function loadUserLikedPosts() {
  if (likesLoading.value) return
  likesLoading.value = true
  try {
    const response = await apiFetch(`/users/${userId.value}/post-likes`)
    if (!response.ok) throw new Error(`HTTP ${response.status}`)
    const json = await response.json()
    const raw = Array.isArray(json) ? json : (json.data ?? [])
    userLikedPosts.value = raw.map((p: Record<string, unknown>) => ({
      post_id: Number(p.post_id ?? p.id ?? 0),
      author: String(p.author ?? ''),
      title: String(p.title ?? ''),
      tags: Array.isArray(p.tags) ? p.tags : [],
      comment_count: Number(p.comment_count ?? 0),
      like_count: Number(p.like_count ?? 0),
      created_at: String(p.created_at ?? ''),
    }))
  } catch (error) {
    handleApiError(error, 'Failed to load liked posts')
  } finally {
    likesLoading.value = false
  }
}

function openPost(id: number) {
  router.push(`/posts/${id}`)
}

// Load on first mount, and only reload when the profile user actually changes.
// Returning to the same profile (e.g. back from a post) must NOT reload, otherwise
// the loading collapse breaks scroll restoration.
watch(
  () => route.params.userId,
  (id) => {
    const newId = Number(id)
    if (newId && newId !== loadedUserId) load()
  },
  { immediate: true },
)
</script>

<style scoped>
.user-profile-page {
  min-height: 100vh;
  padding: 32px 0 80px;
}

.profile-card {
  width: 75%;
  margin: 0 auto;
  padding: 48px 24px 8px;
  display: flex;
  flex-direction: column;
  align-items: center;
  text-align: center;
}

.profile-header {
  width: 100%;
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 6px;
}

.avatar-wrapper {
  position: relative;
  display: inline-block;
}

.avatar-wrapper.own {
  cursor: pointer;
}

.avatar-overlay {
  position: absolute;
  inset: 0;
  display: flex;
  align-items: center;
  justify-content: center;
  border-radius: 50%;
  background: rgba(0, 0, 0, 0.5);
  opacity: 0;
  transition: opacity 0.2s ease;
  cursor: pointer;
}

.avatar-wrapper.own:hover .avatar-overlay {
  opacity: 1;
}

.hidden-input {
  display: none;
}

.camera-icon {
  width: 28px;
  height: 28px;
  color: #ffffff;
}

.username {
  margin: 18px 0 0;
  font-size: 28px;
  font-weight: 700;
  line-height: 1.2;
  color: #ffffff;
}

.handle {
  font-size: 15px;
  color: #8b949e;
}

.bio {
  margin: 8px 0 0;
  max-width: 560px;
  font-size: 14px;
  line-height: 1.6;
  color: #c9d1d9;
  word-break: break-word;
}

.meta-row {
  display: flex;
  align-items: center;
  justify-content: center;
  flex-wrap: wrap;
  gap: 6px 22px;
  margin-top: 10px;
  font-size: 13px;
  color: #8b949e;
}

.meta-item {
  display: inline-flex;
  align-items: center;
  gap: 6px;
}

.info-icon {
  width: 14px;
  height: 14px;
  flex-shrink: 0;
  color: #8b949e;
}

.website {
  color: #58a6ff;
  text-decoration: none;
  word-break: break-all;
}

.website:hover {
  text-decoration: underline;
}

.not-found-title {
  margin: 0;
  font-size: 22px;
  font-weight: 700;
  color: #ffffff;
}

.not-found-hint {
  margin: 0;
  font-size: 13px;
  color: #8c8c8c;
}

.stats-row {
  display: flex;
  align-items: center;
  justify-content: center;
  flex-wrap: wrap;
  gap: 8px 36px;
  width: 100%;
  max-width: 660px;
  margin-top: 20px;
  padding: 18px 0 2px;
  border-top: 1px solid #21262d;
}

.stat {
  display: inline-flex;
  flex-direction: column;
  align-items: center;
  gap: 3px;
  min-width: 62px;
}

.stat strong {
  font-size: 20px;
  font-weight: 700;
  color: #ffffff;
  font-variant-numeric: tabular-nums;
}

.stat-label {
  font-size: 12px;
  color: #8b949e;
}

.stat.clickable {
  cursor: pointer;
}

.stat.clickable:hover strong {
  color: #58a6ff;
}

.edit-btn {
  margin-top: 16px;
  padding: 6px 22px;
  font-size: 14px;
  font-weight: 600;
  color: #141414;
  background: #ffffff;
  border: 1px solid #ffffff;
  border-radius: 8px;
  cursor: pointer;
  transition:
    background 0.15s ease,
    border-color 0.15s ease;
}

.edit-btn:hover:not(:disabled) {
  background: #e4e6e8;
  border-color: #e4e6e8;
  color: #141414;
}

.field-group {
  display: flex;
  flex-direction: column;
  gap: 6px;
  margin-bottom: 16px;
}

.links-list {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.link-field {
  display: flex;
  flex-direction: column;
}

.link-field.has-error :deep(.el-input__wrapper) {
  box-shadow: 0 0 0 1px #cf222e inset;
}

.link-field.has-error :deep(.el-input__wrapper.is-focus) {
  box-shadow:
    0 0 0 1px #cf222e inset,
    0 0 0 3px rgba(207, 34, 46, 0.2);
}

.link-hint {
  align-self: flex-end;
  margin-top: 2px;
  font-size: 12px;
  line-height: 1.4;
  color: #cf222e;
}

.field-label {
  font-size: 14px;
  font-weight: 600;
  color: #1f2328;
}

.field-label-row {
  display: flex;
  align-items: center;
  justify-content: space-between;
}

.char-count {
  font-size: 12px;
  color: #6e7781;
}

.char-count.over {
  color: #cf222e;
  font-weight: 600;
}

.field.is-over :deep(.el-textarea__inner) {
  border-color: #cf222e;
  box-shadow: 0 0 0 1px #cf222e inset;
}

.field.is-over :deep(.el-textarea__inner:focus) {
  border-color: #cf222e;
  box-shadow:
    0 0 0 1px #cf222e inset,
    0 0 0 3px rgba(207, 34, 46, 0.2);
}

.dialog-footer {
  display: flex;
  align-items: center;
  justify-content: flex-end;
  gap: 8px;
}

.save-btn {
  margin-left: 0;
  font-size: 13px;
  font-weight: 600;
  color: #141414;
  background: #ffffff;
  border: 1px solid #ffffff;
  border-radius: 6px;
}

.save-btn:hover:not(:disabled) {
  background: #e4e6e8;
  color: #141414;
  border-color: #e4e6e8;
}

.profile-tabs {
  width: 75%;
  margin: 24px auto 0;
}

.tabs-bar {
  display: flex;
  justify-content: center;
  gap: 0;
  border-bottom: 1px solid #21262d;
}

.tab-btn {
  padding: 12px 28px;
  background: transparent;
  border: none;
  border-bottom: 2px solid transparent;
  color: #8b949e;
  font-size: 14px;
  font-weight: 500;
  cursor: pointer;
  transition: all 0.15s ease;
}

.tab-btn:hover {
  color: #e6edf3;
}

.tab-btn.active {
  color: #ffffff;
  border-bottom-color: #ffffff;
}

.tabs-content {
  margin-top: 8px;
}

.tab-panel {
  min-height: 120px;
}

.feed {
  display: flex;
  flex-direction: column;
}

.topic-row {
  padding: 14px 0;
  border-bottom: 1px solid #1f1f1f;
  cursor: pointer;
  transition: background 0.15s ease;
}

.topic-row:first-child {
  border-top: 1px solid #1f1f1f;
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
  fill: none;
  stroke: #8c8c8c;
  stroke-width: 2;
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
  color: #bfbfbf;
  font-weight: 500;
}

.topic-meta {
  display: flex;
  align-items: center;
  gap: 6px;
  margin-top: 6px;
  font-size: 13px;
  color: #8c8c8c;
}

.topic-date {
  color: #8c8c8c;
}

.meta-dot {
  color: #555;
}

.topic-tag {
  padding: 1px 8px;
  background: #1f1f1f;
  border: 1px solid #333;
  border-radius: 4px;
  font-size: 12px;
  color: #bfbfbf;
  white-space: nowrap;
}

.reply-row {
  display: block;
}

.reply-context {
  display: flex;
  align-items: center;
  gap: 6px;
  font-size: 13px;
  color: #8c8c8c;
}

.reply-icon {
  width: 16px;
  height: 16px;
  fill: currentColor;
  flex-shrink: 0;
}

.reply-hint {
  color: #8c8c8c;
}

.reply-content {
  margin: 8px 0 0;
  font-size: 14px;
  line-height: 1.5;
  color: #e4e6e8;
  overflow: hidden;
  text-overflow: ellipsis;
  display: -webkit-box;
  -webkit-line-clamp: 2;
  -webkit-box-orient: vertical;
}

.empty-hint {
  text-align: center;
  padding: 32px 16px;
  color: #8c8c8c;
  font-size: 14px;
}
</style>
