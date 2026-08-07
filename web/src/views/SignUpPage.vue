<template>
  <div class="signup-page">
    <div class="signup-wrap">
      <div class="back-nav">
        <el-button text @click="goBack">← Back</el-button>
      </div>

      <div class="signup-card">
        <h1 class="signup-title">Create your Gopher account</h1>

        <template v-if="!registered">
          <div class="field-group">
            <label class="field-label" for="signup-username">Username</label>
            <el-input
              id="signup-username"
              v-model="username"
              size="large"
              placeholder="Choose a username"
              autocomplete="username"
              class="field"
            />
          </div>

          <div class="field-group">
            <label class="field-label" for="signup-email">Email</label>
            <el-input
              id="signup-email"
              v-model="email"
              size="large"
              placeholder="you@example.com"
              autocomplete="new-password"
              class="field"
            />
          </div>

          <div class="field-group">
            <label class="field-label" for="signup-password">Password</label>
            <el-input
              id="signup-password"
              v-model="password"
              type="password"
              size="large"
              placeholder="Password"
              show-password
              autocomplete="new-password"
              class="field"
              @keyup.enter="handleRegister"
            />
          </div>

          <el-button
            size="large"
            class="submit-btn"
            :loading="loading"
            @click="handleRegister"
          >
            Create account
          </el-button>

          <div class="signup-footer">
            Already have an account? <router-link to="/login" class="footer-link">Sign in</router-link>
          </div>
        </template>

        <div v-else class="success">
          <div class="success-icon">✓</div>
          <p class="success-title">Registration successful</p>
          <p class="success-text">An activation link has been sent to your email. Please check your inbox to confirm your account.</p>
          <el-button size="large" class="submit-btn" @click="router.push('/login')">
            Go to Sign in
          </el-button>
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

const username = ref('')
const email = ref('')
const password = ref('')
const loading = ref(false)
const registered = ref(false)

function goBack() {
  router.push('/')
}

async function handleRegister() {
  const name = username.value.trim()
  const mail = email.value.trim()

  if (!name) {
    notify('warning', 'Please enter a username')
    return
  }
  if (!mail) {
    notify('warning', 'Please enter your email')
    return
  }
  if (!password.value) {
    notify('warning', 'Please enter a password')
    return
  }

  loading.value = true
  try {
    const response = await apiFetch('/authentication/users', {
      method: 'POST',
      body: JSON.stringify({
        username: name,
        email: mail,
        password: password.value,
      }),
    })
    if (!response.ok) throw new Error(`HTTP ${response.status}`)

    registered.value = true
    password.value = ''
  } catch (error) {
    console.error('Register error:', error)
    notify('error', 'Email sending failed, please try again')
  } finally {
    loading.value = false
  }
}
</script>

<style scoped>
.signup-page {
  display: flex;
  align-items: center;
  justify-content: center;
  min-height: calc(100vh - 61px);
  padding: 48px 16px;
}

.signup-wrap {
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

.signup-card {
  background: #ffffff;
  border: 1px solid #d1d5db;
  border-radius: 8px;
  padding: 24px;
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.signup-title {
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

.signup-footer {
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