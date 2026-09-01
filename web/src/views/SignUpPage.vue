<template>
  <div class="signup-page">
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
              :class="{ 'is-error': usernameError }"
              @input="usernameError = ''"
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
              :class="{ 'is-error': emailError }"
              @input="emailError = ''"
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

          <el-button size="large" class="submit-btn" :loading="loading" @click="handleRegister">
            Create account
          </el-button>

          <div class="signup-footer">
            Already have an account?
            <router-link to="/login" class="footer-link">Sign in</router-link>
          </div>
        </template>

        <div v-else class="success">
          <div class="success-icon">✓</div>
          <p class="success-title">Registration successful</p>
          <p class="success-text">
            An activation link has been sent to your email. Please check your inbox to confirm your
            account.
          </p>
          <el-button size="large" class="submit-btn" @click="router.push('/login')">
            Go to Sign in
          </el-button>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import { apiFetch, handleApiError } from '@/api'
import { notify } from '@/utils/message'

const router = useRouter()

const username = ref('')
const email = ref('')
const password = ref('')
const loading = ref(false)
const registered = ref(false)
const usernameError = ref('')
const emailError = ref('')

async function handleRegister() {
  const name = username.value.trim()
  const mail = email.value.trim()

  usernameError.value = ''
  emailError.value = ''

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

    if (response.status === 409) {
      const data = await response.json()
      const message: string = data?.error || 'Conflict detected, please try again'
      if (message.includes('username')) {
        usernameError.value = message
      } else if (message.includes('email')) {
        emailError.value = message
      }
      notify('error', message)
      return
    }

    if (!response.ok) throw new Error(`HTTP ${response.status}`)

    registered.value = true
    password.value = ''
  } catch (error) {
    handleApiError(error, 'Email sending failed, please try again')
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
  height: 100vh;
  padding: 32px;
  box-sizing: border-box;
}

.signup-card {
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

.signup-title {
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

.field-label {
  font-size: 14px;
  font-weight: 600;
  color: #e4e6e8;
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

.field.is-error :deep(.el-input__wrapper),
.field.is-error :deep(.el-input__wrapper.is-focus) {
  box-shadow:
    0 0 0 1px #f85149 inset,
    0 0 0 3px rgba(248, 81, 73, 0.3);
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

.signup-footer {
  padding-top: 16px;
  border-top: 1px solid #333;
  text-align: center;
  font-size: 14px;
  color: #e4e6e8;
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
  background: #238636;
  color: #ffffff;
  font-size: 26px;
  font-weight: 600;
}

.success-title {
  margin: 0;
  font-size: 18px;
  font-weight: 600;
  color: #e4e6e8;
}

.success-text {
  margin: 0;
  font-size: 14px;
  color: #8c8c8c;
}

.footer-link {
  color: #58a6ff;
  text-decoration: none;
}

.footer-link:hover {
  text-decoration: underline;
}
</style>
