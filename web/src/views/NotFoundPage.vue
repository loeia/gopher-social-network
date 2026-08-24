<template>
  <div class="not-found">
    <h1 class="nf-title">404</h1>
    <p class="nf-text">Page not found</p>
    <p class="nf-hint">
      The page you are looking for doesn't exist or has been moved.<br />
      Please check the URL and try again.
    </p>
    <el-button class="nf-btn" @click="router.push('/')">Go to Home</el-button>
  </div>
</template>

<script setup lang="ts">
import { useRouter } from 'vue-router'

const router = useRouter()
</script>

<style scoped>
.not-found {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 16px;
  min-height: 100vh;
  text-align: center;
}

.nf-title {
  margin: 0;
  font-size: 64px;
  font-weight: 700;
  color: #ffffff;
}

.nf-text {
  margin: 0;
  font-size: 16px;
  font-weight: 600;
  color: #ffffff;
}

.nf-hint {
  margin: 0;
  font-size: 14px;
  line-height: 1.6;
  color: #8c8c8c;
}

.nf-btn {
  background: #ffffff;
  color: #141414;
  border: 1px solid #ffffff;
  font-weight: 600;
}

.nf-btn:hover {
  background: #e4e6e8;
  color: #141414;
  border-color: #e4e6e8;
}
</style>
