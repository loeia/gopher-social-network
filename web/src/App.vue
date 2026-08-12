<template>
  <el-config-provider :message="{ placement: 'top-right', offset: 64 }">
    <div class="layout">
      <NavBar v-if="!route.meta.hideNavBar" />
      <SideBar v-if="showSidebar" :active-view="activeView" @view="onSidebarView" />
      <div class="main" :class="{ 'with-sidebar': showSidebar }">
        <keep-alive include="HomePage">
          <router-view />
        </keep-alive>
      </div>
    </div>
  </el-config-provider>
</template>

<script setup lang="ts">
import { computed, watch } from 'vue'
import { useRoute } from 'vue-router'
import { storeToRefs } from 'pinia'
import { ElConfigProvider } from 'element-plus'
import NavBar from '@/components/NavBar.vue'
import SideBar from '@/components/SideBar.vue'
import { getToken } from '@/api'
import { useFeedStore } from '@/stores/feed'
import { useUIStore } from '@/stores/ui'
import type { ViewType } from '@/stores/feed'

const route = useRoute()
const feedStore = useFeedStore()
const uiStore = useUIStore()
const { view, activeNav } = storeToRefs(feedStore)
const { sidebarOpen } = storeToRefs(uiStore)
const isLoggedIn = computed(() => !!getToken())
const showSidebar = computed(() => isLoggedIn.value && sidebarOpen.value && !route.meta.hideNavBar)
const activeView = computed<ViewType>(() => {
  if (route.name === 'Home') return view.value
  if (route.name === 'MyPosts') return 'myposts'
  return activeNav.value
})

function onSidebarView(next: ViewType) {
  view.value = next
  activeNav.value = next
}

watch(
  () => route.name,
  (name) => {
    if (name === 'MyPosts') activeNav.value = 'myposts'
  },
  { immediate: true },
)
</script>

<style scoped>
.main {
  min-height: 100vh;
}

.main.with-sidebar {
  padding-left: 216px;
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

.el-message :is(.el-message-icon--success, .el-message-icon--warning, .el-message-icon--error, .el-message-icon--info, .el-message-icon--primary) {
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
