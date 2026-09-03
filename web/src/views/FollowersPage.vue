<template>
  <div class="followers-page">
    <FollowersList :user-id="userId" />
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useRoute } from 'vue-router'
import FollowersList from '@/components/FollowersList.vue'

defineOptions({ name: 'FollowersPage' })

const route = useRoute()
const userId = computed(() => Number(route.params.userId))
</script>

<style scoped>
.followers-page {
  min-height: 100vh;
  padding: 32px 0 80px;
}
</style>
