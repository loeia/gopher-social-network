<template>
  <div class="following-page">
    <FollowingList :user-id="userId" />
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useRoute } from 'vue-router'
import FollowingList from '@/components/FollowingList.vue'

defineOptions({ name: 'FollowingPage' })

const route = useRoute()
const userId = computed(() => Number(route.params.userId))
</script>

<style scoped>
.following-page {
  min-height: 100vh;
  padding: 32px 0 80px;
}
</style>
