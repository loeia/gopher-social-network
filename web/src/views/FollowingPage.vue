<template>
  <div class="following-page">
    <FollowingList />
  </div>
</template>

<script setup lang="ts">
import FollowingList from '@/components/FollowingList.vue'

defineOptions({ name: 'FollowingPage' })
</script>

<style scoped>
.following-page {
  min-height: 100vh;
  padding: 32px 0 80px;
}
</style>
