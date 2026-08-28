<template>
  <div class="reset-container">
    <div class="reset-card">
      <template v-if="!done">
        <h1 class="reset-title">Reset your password</h1>
        <p class="reset-desc">Enter your new password below.</p>

        <el-input
          v-model="newPassword"
          type="password"
          size="large"
          placeholder="New password"
          show-password
          class="reset-field"
          @keyup.enter="handleReset"
        />

        <el-button size="large" class="reset-btn" :loading="loading" @click="handleReset">
          Reset password
        </el-button>
      </template>

      <template v-else>
        <div class="success-icon">✓</div>
        <h1 class="reset-title">Password has been reset</h1>
        <p class="reset-desc">You can now sign in with your new password.</p>
        <el-button size="large" class="reset-btn" @click="router.push('/login')">
          Go to Sign in
        </el-button>
      </template>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { apiFetch, handleApiError } from '@/api'
import { notify } from '@/utils/message'

const route = useRoute()
const router = useRouter()

const token = route.params.token as string
const newPassword = ref('')
const loading = ref(false)
const done = ref(false)

async function handleReset() {
  if (!newPassword.value) {
    notify('warning', 'Please enter a new password')
    return
  }

  loading.value = true
  try {
    const response = await apiFetch('/users/reset-password', {
      method: 'POST',
      body: JSON.stringify({
        token,
        new_password: newPassword.value,
      }),
    })
    if (!response.ok) {
      const json = await response.json().catch(() => null)
      const msg = json?.error || json?.data?.message || 'Invalid or expired token'
      notify('error', msg)
      return
    }

    done.value = true
  } catch (error) {
    handleApiError(error, 'Failed to reset password, please try again')
  } finally {
    loading.value = false
  }
}
</script>

<style scoped>
.reset-container {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 24px;
  min-height: 100vh;
}

.reset-card {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 16px;
  width: 100%;
  max-width: 360px;
  padding: 32px;
}

.reset-title {
  margin: 0;
  font-size: 22px;
  font-weight: 600;
  color: #e4e6e8;
  text-align: center;
}

.reset-desc {
  margin: 0;
  font-size: 14px;
  color: #8c8c8c;
  text-align: center;
}

.reset-field {
  width: 100%;
}

.reset-field :deep(.el-input__wrapper) {
  background: transparent;
  box-shadow: 0 0 0 1px #333 inset;
}

.reset-field :deep(.el-input__wrapper.is-focus) {
  box-shadow:
    0 0 0 1px #e4e6e8 inset,
    0 0 0 3px rgba(228, 230, 232, 0.1);
}

.reset-field :deep(.el-input__inner) {
  color: #e4e6e8;
}

.reset-field :deep(.el-input__inner::placeholder) {
  color: #8c8c8c;
}

.reset-btn {
  width: 100%;
  background: #e4e6e8;
  color: #1a1a1a;
  border: 1px solid #e4e6e8;
  font-weight: 600;
}

.reset-btn:hover {
  background: #ffffff;
  color: #1a1a1a;
  border-color: #ffffff;
}

.reset-btn.is-loading {
  background: #e4e6e8;
  border-color: #e4e6e8;
  color: #ffffff;
}

.success-icon {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 48px;
  height: 48px;
  border-radius: 50%;
  background: #2da44e;
  color: #ffffff;
  font-size: 26px;
  font-weight: 600;
}
</style>
