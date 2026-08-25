<template>
  <div class="login-page">
    <div class="login-wrap">
      <div class="back-nav">
        <el-button text @click="goBack">← Back</el-button>
      </div>

      <div class="login-card">
        <h1 class="login-title">Sign in to Gopher</h1>

        <div class="field-group">
          <label class="field-label" for="login-email">Email</label>
          <el-input
            id="login-email"
            v-model="email"
            size="large"
            placeholder="you@example.com"
            class="field"
          />
        </div>

        <div class="field-group">
          <div class="label-row">
            <label class="field-label" for="login-password">Password</label>
            <router-link to="/forgot-password" class="forgot-link">Forgot password?</router-link>
          </div>
          <el-input
            id="login-password"
            v-model="password"
            type="password"
            size="large"
            placeholder="Password"
            show-password
            class="field"
            @keyup.enter="handleLogin"
          />
        </div>

        <el-button size="large" class="submit-btn" :loading="loading" @click="handleLogin">
          Sign in
        </el-button>

        <div class="login-footer">
          New to Gopher?
          <router-link to="/signup" class="footer-link">Create an account</router-link>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { apiFetch, setToken, handleApiError } from '@/api'
import { notify } from '@/utils/message'

const route = useRoute()
const router = useRouter()

const email = ref('')
const password = ref('')
const loading = ref(false)

function goBack() {
  router.push('/')
}

async function handleLogin() {
  if (!email.value.trim() || !password.value) {
    notify('warning', 'Please enter email and password')
    return
  }

  loading.value = true
  try {
    const response = await apiFetch('/authentication/token', {
      method: 'POST',
      body: JSON.stringify({
        email: email.value.trim(),
        password: password.value,
      }),
    })
    if (!response.ok) throw new Error(`HTTP ${response.status}`)

    const json = await response.json()
    setToken(json.data)
    notify('success', 'Logged in')
    const redirect = route.query.redirect
    if (typeof redirect === 'string' && redirect.startsWith('/') && !redirect.startsWith('//')) {
      router.push(redirect)
    } else {
      router.push('/')
    }
  } catch (error) {
    handleApiError(error, 'Login failed')
  } finally {
    loading.value = false
  }
}
</script>

<style scoped>
.login-page {
  display: flex;
  align-items: center;
  justify-content: center;
  min-height: calc(100vh - 61px);
  padding: 48px 16px;
}

.login-wrap {
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

.login-card {
  background: #ffffff;
  border: 1px solid #d1d5db;
  border-radius: 8px;
  padding: 24px;
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.login-title {
  margin: 0 0 8px;
  font-size: 20px;
  font-weight: 500;
  color: #1f2328;
  text-align: center;
}

.field-group {
  display: flex;
  flex-direction: column;
  gap: 6px;
}

.label-row {
  display: flex;
  align-items: center;
  justify-content: space-between;
}

.field-label {
  font-size: 14px;
  font-weight: 600;
  color: #1f2328;
}

.forgot-link {
  font-size: 12px;
  color: #0969da;
  text-decoration: none;
}

.forgot-link:hover {
  text-decoration: underline;
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

.login-footer {
  padding-top: 16px;
  border-top: 1px solid #d1d5db;
  text-align: center;
  font-size: 14px;
  color: #1f2328;
}

.footer-link {
  color: #0969da;
  text-decoration: none;
}

.footer-link:hover {
  text-decoration: underline;
}
</style>
