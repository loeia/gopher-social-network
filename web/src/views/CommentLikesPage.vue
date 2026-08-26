<template>
  <div class="likes-page">
    <LikedComments />
  </div>
</template>

<script setup lang="ts">
import LikedComments from '@/components/LikedComments.vue'

defineOptions({ name: 'CommentLikesPage' })
</script>

<style scoped>
.likes-page {
  min-height: 100vh;
  padding: 32px 0 80px;
}
</style>
