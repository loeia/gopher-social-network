<template>
  <el-config-provider :message="{ placement: 'top-right', offset: 64 }">
    <div class="layout">
      <NavBar />
      <keep-alive include="HomePage">
        <router-view />
      </keep-alive>
    </div>
  </el-config-provider>
</template>

<script setup lang="ts">
import { ElConfigProvider } from 'element-plus'
import NavBar from '@/components/NavBar.vue'
</script>

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
