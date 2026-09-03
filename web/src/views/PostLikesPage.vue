<template>
  <div class="likes-page">
    <LikedPosts />
  </div>
</template>

<script setup lang="ts">
import LikedPosts from '@/components/LikedPosts.vue'

defineOptions({ name: 'PostLikesPage' })
</script>

<style scoped>
.likes-page {
  min-height: 100vh;
  padding: 32px 0 80px;
}
</style>
