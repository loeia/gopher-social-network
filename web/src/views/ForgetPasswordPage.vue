<template>
  <div class="forgot-page">
    <div class="forgot-wrap">
      <div class="back-nav">
        <el-button text @click="goBack">← Back</el-button>
      </div>

      <div class="forgot-card">
        <template v-if="!sent">
          <h1 class="forgot-title">Reset your password</h1>

          <p class="forgot-desc">Enter your email and we'll send you a reset link.</p>

          <div class="field-group">
            <label class="field-label" for="forgot-email">Email</label>
            <el-input
              id="forgot-email"
              v-model="email"
              size="large"
              placeholder="you@example.com"
              class="field"
              @keyup.enter="handleSubmit"
            />
          </div>

          <el-button size="large" class="submit-btn" :loading="loading" @click="handleSubmit">
            Send reset link
          </el-button>
        </template>

        <div v-else class="success">
          <div class="success-icon">✓</div>
          <p class="success-title">Check your email</p>
          <p class="success-text">{{ successMessage }}</p>
          <el-button size="large" class="submit-btn" @click="router.push('/login')">
            Go to Sign in
          </el-button>
        </div>

        <div class="forgot-footer">
          Remember your password? <router-link to="/login" class="footer-link">Sign in</router-link>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import { apiFetch } from '@/api'
import { notify } from '@/utils/message'

const router = useRouter()

const email = ref('')
const loading = ref(false)
const sent = ref(false)
const successMessage = ref('')

function goBack() {
  router.push('/')
}

async function handleSubmit() {
  const mail = email.value.trim()
  if (!mail) {
    notify('warning', 'Please enter your email')
    return
  }

  loading.value = true
  try {
    const response = await apiFetch('/users/forgot-password', {
      method: 'POST',
      body: JSON.stringify({ email: mail }),
    })
    if (!response.ok) throw new Error(`HTTP ${response.status}`)

    const json = await response.json()
    const data = json.data ?? json
    successMessage.value = data?.message || 'If that email exists, a reset link has been sent.'
    sent.value = true
  } catch (error) {
    console.error('Forgot password error:', error)
    notify('error', 'Failed to send reset link, please try again')
  } finally {
    loading.value = false
  }
}
</script>

<style scoped>
.forgot-page {
  display: flex;
  align-items: center;
  justify-content: center;
  min-height: calc(100vh - 61px);
  padding: 48px 16px;
}

.forgot-wrap {
  width: 100%;
  max-width: 348px;
}

.back-nav {
  margin-bottom: 12px;
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

.back-nav :deep(.el-button.is-disabled) {
  color: #3d4043;
  background: transparent;
  text-decoration: none;
  cursor: not-allowed;
}

.forgot-card {
  background: #ffffff;
  border: 1px solid #d1d5db;
  border-radius: 8px;
  padding: 24px;
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.forgot-title {
  margin: 0 0 8px;
  font-size: 20px;
  font-weight: 500;
  color: #1f2328;
  text-align: center;
}

.forgot-desc {
  margin: 0;
  font-size: 14px;
  color: #57606a;
  text-align: center;
}

.field-group {
  display: flex;
  flex-direction: column;
  gap: 6px;
}

.field-label {
  font-size: 14px;
  font-weight: 600;
  color: #1f2328;
}

.field :deep(.el-input__wrapper) {
  background: #ffffff;
  box-shadow: 0 0 0 1px #d1d5db inset;
  border-radius: 6px;
  transition: box-shadow 0.2s ease;
}

.field :deep(.el-input__wrapper.is-focus) {
  box-shadow:
    0 0 0 1px #0969da inset,
    0 0 0 3px rgba(9, 105, 218, 0.3);
}

.field :deep(.el-input__inner) {
  color: #1f2328;
}

.field :deep(.el-input__inner::placeholder) {
  color: #6e7781;
}

.field :deep(.el-input__suffix) {
  color: #57606a;
}

.submit-btn {
  width: 100%;
  background: #1f2328;
  color: #ffffff;
  border: 1px solid #1f2328;
  border-radius: 6px;
  font-weight: 500;
}

.submit-btn:hover {
  background: #32383f;
  color: #ffffff;
  border-color: #32383f;
}

.submit-btn.is-loading {
  background: #6e7781;
  border-color: #6e7781;
  color: #ffffff;
}

.forgot-footer {
  padding-top: 16px;
  border-top: 1px solid #d1d5db;
  text-align: center;
  font-size: 14px;
  color: #1f2328;
}

.success {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 16px;
  padding: 12px 0;
  text-align: center;
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

.success-title {
  margin: 0;
  font-size: 18px;
  font-weight: 600;
  color: #1f2328;
}

.success-text {
  margin: 0;
  font-size: 14px;
  color: #57606a;
}

.footer-link {
  color: #0969da;
  text-decoration: none;
}

.footer-link:hover {
  text-decoration: underline;
}
</style>
