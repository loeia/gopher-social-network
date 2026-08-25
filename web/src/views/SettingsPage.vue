<template>
  <div class="settings-page">
    <div class="settings-container">
      <div class="settings-sidebar">
        <h1 class="settings-title">Settings</h1>
        <el-button
          class="menu-btn"
          :class="{ active: activeTab === 'password' }"
          @click="activeTab = 'password'"
        >
          Password
        </el-button>
        <el-button
          class="menu-btn"
          :class="{ active: activeTab === 'rename' }"
          @click="activeTab = 'rename'"
        >
          Rename
        </el-button>
      </div>

      <div class="settings-main">
        <div v-if="activeTab === 'password'" class="reset-section">
          <h2 class="section-title">Reset Password</h2>

          <div class="field-group">
            <label class="field-label" for="old-password">Old Password</label>
            <el-input
              id="old-password"
              v-model="oldPassword"
              type="password"
              size="large"
              placeholder="Enter old password"
              show-password
              class="field"
              :class="{ 'is-error': oldPasswordError }"
              @input="oldPasswordError = ''"
              @keyup.enter="handleSubmit"
            />
            <p v-if="oldPasswordError" class="field-error">{{ oldPasswordError }}</p>
          </div>

          <div class="field-group">
            <label class="field-label" for="new-password">New Password</label>
            <el-input
              id="new-password"
              v-model="newPassword"
              type="password"
              size="large"
              placeholder="Enter new password"
              show-password
              class="field"
              @keyup.enter="handleSubmit"
            />
          </div>

          <div class="field-group">
            <label class="field-label" for="confirm-password">Confirm New Password</label>
            <el-input
              id="confirm-password"
              v-model="confirmPassword"
              type="password"
              size="large"
              placeholder="Confirm new password"
              show-password
              class="field"
              @keyup.enter="handleSubmit"
            />
          </div>

          <el-button size="large" class="submit-btn" :loading="loading" @click="handleSubmit">
            Reset
          </el-button>
        </div>

        <div v-if="activeTab === 'rename'" class="rename-section">
          <h2 class="section-title">Rename</h2>

          <div class="field-group">
            <label class="field-label" for="new-username">New Username</label>
            <el-input
              id="new-username"
              v-model="newUsername"
              type="text"
              size="large"
              placeholder="Enter new username"
              maxlength="25"
              class="field"
              :class="{ 'is-error': renameError || isUsernameTooShort }"
              @input="renameError = ''"
              @keyup.enter="handleRename"
            />
            <p class="field-hint">{{ newUsername.length }}/25</p>
            <p v-if="renameError" class="field-error">{{ renameError }}</p>
            <p v-else-if="isUsernameTooShort" class="field-error">
              Username must be at least 4 characters
            </p>
          </div>

          <el-button
            size="large"
            class="submit-btn rename-submit"
            :loading="renameLoading"
            @click="handleRename"
          >
            Save
          </el-button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed } from 'vue'
import { useRouter } from 'vue-router'
import { apiFetch, clearToken, getApiError, handleApiError } from '@/api'
import { notify } from '@/utils/message'
import { useUserStore } from '@/stores/user'

const router = useRouter()
const userStore = useUserStore()

const activeTab = ref<'password' | 'rename'>('password')
const oldPassword = ref('')
const newPassword = ref('')
const confirmPassword = ref('')
const loading = ref(false)
const oldPasswordError = ref('')

const newUsername = ref('')
const renameLoading = ref(false)
const renameError = ref('')

const isUsernameTooShort = computed(
  () => newUsername.value.length > 0 && newUsername.value.length < 4,
)

async function handleSubmit() {
  oldPasswordError.value = ''
  if (!oldPassword.value) {
    notify('warning', 'Please enter your old password')
    return
  }
  if (!newPassword.value) {
    notify('warning', 'Please enter a new password')
    return
  }
  if (newPassword.value !== confirmPassword.value) {
    notify('warning', 'New passwords do not match')
    return
  }

  loading.value = true
  try {
    const response = await apiFetch('/users/reset', {
      method: 'PATCH',
      body: JSON.stringify({
        old_password: oldPassword.value,
        new_password: newPassword.value,
      }),
    })
    if (!response.ok) {
      if (response.status === 401) {
        oldPasswordError.value = 'Incorrect old password'
        return
      }
      const json = await response.json().catch(() => null)
      const msg = json?.error || 'Failed to update password'
      notify('error', msg)
      return
    }

    notify('success', 'Password updated')
    clearToken()
    router.push('/login')
  } catch (error) {
    handleApiError(error, 'Failed to update password')
  } finally {
    loading.value = false
  }
}

async function handleRename() {
  const name = newUsername.value.trim()
  if (!name) {
    notify('warning', 'Please enter a new username')
    return
  }
  if (name.length < 4) {
    renameError.value = 'Username must be at least 4 characters'
    return
  }
  if (name === userStore.username) {
    notify('warning', 'New username is the same as the current username')
    return
  }
  renameLoading.value = true
  renameError.value = ''
  try {
    const response = await apiFetch('/users/rename', {
      method: 'PATCH',
      body: JSON.stringify({ new_name: name }),
    })
    if (!response.ok) {
      if (response.status === 401) return
      const message = (await getApiError(response)) ?? `Failed to rename (HTTP ${response.status})`
      notify('error', message)
      return
    }
    await userStore.fetchCurrentUser(true)
    newUsername.value = ''
    notify('success', 'Username updated')
  } catch (error) {
    handleApiError(error, 'Failed to rename')
  } finally {
    renameLoading.value = false
  }
}
</script>

<style scoped>
.settings-page {
  min-height: 100vh;
  padding: 32px 0 80px;
}

.settings-container {
  width: 75%;
  margin: 0 auto;
  display: flex;
  gap: 32px;
  padding: 0 24px;
}

.settings-sidebar {
  width: 180px;
  flex-shrink: 0;
}

.settings-title {
  margin: 0 0 24px;
  font-size: 22px;
  font-weight: 600;
  color: #ffffff;
}

.menu-btn {
  width: 100%;
  margin-left: 0 !important;
  text-align: left;
  justify-content: flex-start;
  background: transparent;
  color: #8c8c8c;
  border: 1px solid transparent;
  font-weight: 500;
}

.menu-btn:hover {
  color: #ffffff;
  background: transparent;
  border-color: transparent;
}

.menu-btn.active {
  color: #ffffff;
  background: #262626;
  border-color: #3d444d;
}

.settings-main {
  flex: 1;
  min-width: 0;
}

.reset-section,
.rename-section {
  display: flex;
  flex-direction: column;
  gap: 24px;
  max-width: 300px;
}

.rename-submit {
  margin-top: 8px;
}

.section-title {
  margin: 0;
  font-size: 20px;
  font-weight: 600;
  color: #ffffff;
}

.field-group {
  display: flex;
  flex-direction: column;
  gap: 6px;
}

.field-label {
  font-size: 14px;
  font-weight: 600;
  color: #e4e6e8;
}

.field :deep(.el-input__wrapper) {
  background: transparent;
  box-shadow: 0 0 0 1px #262626 inset;
  border-radius: 6px;
}

.field :deep(.el-input__wrapper.is-focus) {
  box-shadow:
    0 0 0 1px #ffffff inset,
    0 0 0 3px rgba(255, 255, 255, 0.15);
}

.field :deep(.el-input__inner) {
  color: #ffffff;
}

.field :deep(.el-input__inner::placeholder) {
  color: #595959;
}

.field.is-error :deep(.el-input__wrapper),
.field.is-error :deep(.el-input__wrapper.is-focus) {
  box-shadow:
    0 0 0 1px #cf222e inset,
    0 0 0 3px rgba(207, 34, 46, 0.2);
}

.field-error {
  margin: 0;
  font-size: 13px;
  color: #cf222e;
}

.field-hint {
  margin: 0;
  font-size: 12px;
  color: #595959;
  text-align: right;
}

.submit-btn {
  width: 100%;
  background: #ffffff;
  color: #141414;
  border: 1px solid #ffffff;
  font-weight: 600;
}

.submit-btn:hover {
  background: #e4e6e8;
  color: #141414;
  border-color: #e4e6e8;
}

.submit-btn.is-loading {
  background: #ffffff;
  border-color: #ffffff;
  color: #141414;
}
</style>
