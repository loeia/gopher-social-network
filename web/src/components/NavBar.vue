<template>
  <header class="navbar">
    <router-link to="/" class="brand">Gopher</router-link>

    <el-input
      v-model="searchId"
      class="navbar-search"
      placeholder="Search post by ID"
      clearable
      @keyup.enter="searchPost"
    />

    <div class="navbar-right">
      <template v-if="isLoggedIn">
        <el-button @click="handleLogout">Logout</el-button>
      </template>
      <template v-else>
        <el-button @click="goToLogin">Login</el-button>
        <el-button type="primary" @click="goToSignUp">Sign Up</el-button>
      </template>
    </div>
  </header>
</template>

<script setup lang="ts">
import { ref, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import { getToken, clearToken } from '@/api'

const router = useRouter()
const route = useRoute()

const searchId = ref('')
const isLoggedIn = ref(!!getToken())

watch(
  () => route.fullPath,
  () => {
    isLoggedIn.value = !!getToken()
  },
)

function searchPost() {
  const id = Number(searchId.value.trim())
  if (!searchId.value.trim() || !Number.isInteger(id) || id <= 0) {
    ElMessage.warning('Please enter a valid post ID')
    return
  }
  router.push(`/posts/${id}`)
  searchId.value = ''
}

function goToLogin() {
  router.push({ path: '/login', query: { redirect: route.fullPath } })
}

function goToSignUp() {
  // Register page not built yet.
  ElMessage.info('Sign up page is not ready yet')
}

async function handleLogout() {
  // TODO: call the logout API to invalidate the server-side token/session.
  // The backend endpoint is not implemented yet, so for now we only clear the local token.
  clearToken()
  isLoggedIn.value = false
  ElMessage.success('Logged out')
  router.push('/')
}
</script>

<style scoped>
.navbar {
  position: sticky;
  top: 0;
  z-index: 100;
  display: flex;
  align-items: center;
  gap: 16px;
  padding: 10px 24px;
  background: #141414;
  border-bottom: 1px solid #262626;
}

.brand {
  margin-right: 16px;
  font-size: 22px;
  font-weight: 700;
  color: #ffffff;
  text-decoration: none;
  white-space: nowrap;
}

.navbar-left {
  display: flex;
  align-items: center;
}

.navbar-search {
  flex: 1;
  max-width: 460px;
  margin: 0 auto;
}

.navbar-search :deep(.el-input__wrapper) {
  background: #0a0a0a;
  box-shadow: 0 0 0 1px #262626 inset;
}

.navbar-search :deep(.el-input__inner) {
  color: #ffffff;
}

.navbar-search :deep(.el-input__inner::placeholder) {
  color: #595959;
}

.navbar-right {
  display: flex;
  align-items: center;
  gap: 12px;
}
</style>