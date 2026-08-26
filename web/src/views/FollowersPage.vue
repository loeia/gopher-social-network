<template>
  <div class="followers-page">
    <FollowersList />
  </div>
</template>

<script setup lang="ts">
import FollowersList from '@/components/FollowersList.vue'

defineOptions({ name: 'FollowersPage' })
</script>

<style scoped>
.followers-page {
  min-height: 100vh;
  padding: 32px 0 80px;
}
</style>
