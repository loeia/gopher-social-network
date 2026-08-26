<template>
  <el-config-provider :message="{ placement: 'top-right', offset: 64 }">
    <div class="layout">
      <NavBar v-if="!route.meta.hideNavBar" />
      <div class="main">
        <router-view v-slot="{ Component }">
          <keep-alive include="HomePage,MyPostsPage,SearchResults">
            <component :is="Component" />
          </keep-alive>
        </router-view>
      </div>
    </div>
  </el-config-provider>
</template>

<script setup lang="ts">
import { computed, onBeforeUnmount, onMounted, watch } from 'vue'
import { useRoute } from 'vue-router'
import { ElConfigProvider } from 'element-plus'
import NavBar from '@/components/NavBar.vue'
import { getToken, handleSessionExpired, isTokenExpired } from '@/api'
import { useUserStore } from '@/stores/user'

const route = useRoute()
const userStore = useUserStore()
const isLoggedIn = computed(() => !!getToken())

let sessionTimer: number | null = null

function checkSession() {
  if (getToken() && isTokenExpired()) {
    handleSessionExpired()
  }
}

onMounted(() => {
  sessionTimer = window.setInterval(checkSession, 30_000)
})

onBeforeUnmount(() => {
  if (sessionTimer !== null) window.clearInterval(sessionTimer)
})

watch(isLoggedIn, (loggedIn) => {
  if (!loggedIn) userStore.reset()
})
</script>

<style scoped>
.main {
  min-height: 100vh;
}
</style>

<style>
body {
  margin: 0;
  background: #0a0a0a;
  color: #ffffff;
}

.el-message {
  --el-message-bg-color: #ffffff;
  --el-message-border-color: #ffffff;
  --el-message-close-icon-color: #595959;
  --el-message-close-hover-color: #141414;
  background-color: #ffffff;
  border-color: #ffffff;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.15);
}

.el-message .el-message__content,
.el-message--success .el-message__content,
.el-message--warning .el-message__content,
.el-message--error .el-message__content,
.el-message--info .el-message__content {
  color: #141414;
}

.el-message
  :is(
    .el-message-icon--success,
    .el-message-icon--warning,
    .el-message-icon--error,
    .el-message-icon--info,
    .el-message-icon--primary
  ) {
  color: #141414;
}

.el-message-fade-enter-from.is-right {
  transform: translateX(calc(100% + 32px));
  opacity: 0;
}

.el-message-fade-leave-to.is-right {
  transform: translateX(calc(100% + 32px));
  opacity: 0;
}

.el-message-fade-enter-active.is-right,
.el-message-fade-leave-active.is-right {
  transition:
    opacity 0.35s ease,
    transform 0.35s ease;
}
</style>
