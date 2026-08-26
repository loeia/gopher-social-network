<template>
  <div class="user-profile-page">
    <div class="profile-card" v-loading="loading">
      <template v-if="notFound">
        <h1 class="not-found-title">User not found</h1>
        <p class="not-found-hint">This user may have been removed or the link is incorrect.</p>
      </template>
      <template v-else>
        <UserAvatar :user-id="userId" :username="user.username" :size="96" />

        <h1 class="username">{{ user.username }}</h1>

        <div class="follow-stats">
          <span
            class="follow-stat"
            :class="{ clickable: isOwnProfile }"
            @click="isOwnProfile && goToFollowers()"
            ><strong>{{ user.followers_count }}</strong> followers</span
          >
          <span class="follow-dot">·</span>
          <span
            class="follow-stat"
            :class="{ clickable: isOwnProfile }"
            @click="isOwnProfile && goToFollowing()"
            ><strong>{{ user.following_count }}</strong> following</span
          >
        </div>

        <p v-if="user.bio" class="bio">{{ user.bio }}</p>

        <div class="info-row">
          <svg
            class="info-icon"
            viewBox="0 0 24 24"
            fill="none"
            stroke="currentColor"
            stroke-width="2"
            stroke-linecap="round"
            stroke-linejoin="round"
          >
            <rect x="3" y="4" width="18" height="18" rx="2" ry="2" />
            <line x1="16" y1="2" x2="16" y2="6" />
            <line x1="8" y1="2" x2="8" y2="6" />
            <line x1="3" y1="10" x2="21" y2="10" />
          </svg>
          <span>Joined {{ formatDate(user.created_at) }}</span>
        </div>

        <div v-for="(link, index) in user.links" :key="index" class="info-row">
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
        </div>

        <el-button v-if="isOwnProfile" class="edit-btn" @click="openEdit"> Edit profile </el-button>
      </template>
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
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, ref, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { apiFetch, getApiError, getCurrentUserId, handleApiError } from '@/api'
import { notify } from '@/utils/message'
import UserAvatar from '@/components/UserAvatar.vue'
import { useFeedStore } from '@/stores/feed'

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

const route = useRoute()
const router = useRouter()
const feedStore = useFeedStore()

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
  feedStore.setView('followers')
  router.push({ name: 'Home' })
}

function goToFollowing() {
  feedStore.setView('following')
  router.push({ name: 'Home' })
}

async function load() {
  const id = userId.value
  if (!id) return
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
  } catch (error) {
    handleApiError(error, 'Failed to load profile')
  } finally {
    loading.value = false
  }
}

onMounted(load)
watch(() => route.params.userId, load)
</script>

<style scoped>
.user-profile-page {
  min-height: 100vh;
  padding: 32px 0 80px;
}

.profile-card {
  width: 75%;
  margin: 0 auto;
  padding: 40px 24px;
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 12px;
  text-align: center;
  background: #141414;
  border: 1px solid #262626;
  border-radius: 12px;
}

.username {
  margin: 0;
  font-size: 24px;
  font-weight: 700;
  color: #ffffff;
}

.follow-stats {
  display: flex;
  align-items: center;
  gap: 6px;
  font-size: 14px;
  color: #8c8c8c;
}

.follow-stat strong {
  font-weight: 700;
  color: #e4e6e8;
}

.follow-stat.clickable {
  cursor: pointer;
}

.follow-stat.clickable:hover strong {
  color: #6cbbf7;
}

.follow-dot {
  color: #8c8c8c;
}

.bio {
  margin: 0;
  font-size: 14px;
  line-height: 1.6;
  color: #e4e6e8;
  word-break: break-word;
}

.info-row {
  display: flex;
  align-items: center;
  gap: 6px;
  font-size: 13px;
  color: #8c8c8c;
}

.info-icon {
  width: 16px;
  height: 16px;
  flex-shrink: 0;
}

.website {
  color: #6cbbf7;
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

.edit-btn {
  margin-top: 12px;
  padding: 4px 16px;
  font-size: 13px;
  font-weight: 600;
  color: #141414;
  background: #ffffff;
  border: 1px solid #ffffff;
  border-radius: 6px;
  cursor: pointer;
}

.edit-btn:hover:not(:disabled) {
  background: #e4e6e8;
  color: #141414;
  border-color: #e4e6e8;
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
</style>
