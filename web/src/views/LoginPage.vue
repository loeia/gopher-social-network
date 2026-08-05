<template>
  <div class="login-page">
    <div class="login-card">
      <h1 class="login-title">Log In</h1>

      <el-input
        v-model="email"
        size="large"
        placeholder="Email"
        class="field"
      />

      <el-input
        v-model="password"
        type="password"
        size="large"
        placeholder="Password"
        show-password
        class="field"
        @keyup.enter="handleLogin"
      />

      <el-button
        type="primary"
        size="large"
        class="submit-btn"
        :loading="loading"
        @click="handleLogin"
      >
        Log In
      </el-button>

      <div class="links">
        <a href="#" class="link">Sign up</a>
        <a href="#" class="link">Forgot password</a>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import { apiFetch, setToken } from '@/api'

const route = useRoute()
const router = useRouter()

const email = ref('')
const password = ref('')
const loading = ref(false)

async function handleLogin() {
  if (!email.value.trim() || !password.value) {
    ElMessage.warning('Please enter email and password')
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
    ElMessage.success('Logged in')
    const redirect = route.query.redirect
    if (
      typeof redirect === 'string' &&
      redirect.startsWith('/') &&
      !redirect.startsWith('//')
    ) {
      router.push(redirect)
    } else {
      router.push('/')
    }
  } catch (error) {
    console.error('Login error:', error)
    ElMessage.error('Login failed')
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
  min-height: 100vh;
  padding: 24px;
}

.login-card {
  width: 100%;
  max-width: 400px;
  background: #141414;
  border: 1px solid #262626;
  border-radius: 12px;
  padding: 32px;
  display: flex;
  flex-direction: column;
  gap: 20px;
}

.login-title {
  margin: 0;
  text-align: center;
  font-size: 24px;
  font-weight: 600;
  color: #ffffff;
}

.field :deep(.el-input__wrapper) {
  background: #0a0a0a;
  box-shadow: 0 0 0 1px #262626 inset;
}

.field :deep(.el-input__inner) {
  color: #ffffff;
}

.submit-btn {
  width: 100%;
}

.links {
  display: flex;
  justify-content: center;
  gap: 24px;
}

.link {
  font-size: 14px;
  color: #8c8c8c;
  text-decoration: none;
  transition: color 0.2s ease;
}

.link:hover {
  color: #ffffff;
}
</style>
