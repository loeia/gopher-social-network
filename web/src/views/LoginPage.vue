<template>
  <div class="login-page">
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
  height: 100vh;
  padding: 32px;
  box-sizing: border-box;
}

.login-card {
  width: 100%;
  max-width: 348px;
  
  background: #1a1a1a;
  border: 1px solid #333;
  border-radius: 8px;
  padding: 32px 24px;
  display: flex;
  flex-direction: column;
  justify-content: center;
  gap: 16px;
}

.login-title {
  margin: 0 0 8px;
  font-size: 20px;
  font-weight: 500;
  color: #e4e6e8;
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
  color: #e4e6e8;
}

.forgot-link {
  font-size: 12px;
  color: #58a6ff;
  text-decoration: none;
}

.forgot-link:hover {
  text-decoration: underline;
}

.field :deep(.el-input__wrapper) {
  background: transparent;
  box-shadow: 0 0 0 1px #333 inset;
  border-radius: 6px;
  transition: box-shadow 0.2s ease;
}

.field :deep(.el-input__wrapper.is-focus) {
  box-shadow:
    0 0 0 1px #e4e6e8 inset,
    0 0 0 3px rgba(228, 230, 232, 0.1);
}

.field :deep(.el-input__inner) {
  color: #e4e6e8;
}

.field :deep(.el-input__inner::placeholder) {
  color: #8c8c8c;
}

.field :deep(.el-input__inner:-webkit-autofill) {
  -webkit-box-shadow: 0 0 0 1000px #1a1a1a inset !important;
  -webkit-text-fill-color: #e4e6e8 !important;
  caret-color: #e4e6e8;
}

.submit-btn {
  width: 100%;
  background: #e4e6e8;
  color: #1a1a1a;
  border: 1px solid #e4e6e8;
  border-radius: 6px;
  font-weight: 500;
}

.submit-btn:hover {
  background: #ffffff;
  color: #1a1a1a;
  border-color: #ffffff;
}

.submit-btn.is-loading {
  background: #6e7781;
  border-color: #6e7781;
  color: #ffffff;
}

.login-footer {
  padding-top: 16px;
  border-top: 1px solid #333;
  text-align: center;
  font-size: 14px;
  color: #e4e6e8;
}

.footer-link {
  color: #58a6ff;
  text-decoration: none;
}

.footer-link:hover {
  text-decoration: underline;
}
</style>
