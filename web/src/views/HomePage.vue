<template>
  <div class="home" :class="{ 'with-sidebar': isLoggedIn }">
    <SideBar v-if="isLoggedIn" :active-view="view" @view="view = $event" />
    <PostsList v-if="!isLoggedIn || view === 'all'" />
    <LikedPosts v-else />
  </div>
</template>

<script setup lang="ts">
import { computed, onActivated, ref } from 'vue'
import { storeToRefs } from 'pinia'
import SideBar from '@/components/SideBar.vue'
import PostsList from '@/components/PostsList.vue'
import LikedPosts from '@/components/LikedPosts.vue'
import { getToken } from '@/api'
import { useFeedStore } from '@/stores/feed'

defineOptions({ name: 'HomePage' })

const store = useFeedStore()
const { view } = storeToRefs(store)
const isLoggedIn = computed(() => !!getToken())

onActivated(() => store.clearPostHistory())
</script>

<style scoped>
.home {
  min-height: 100vh;
  padding: 32px 0 80px;
}

.home.with-sidebar {
  padding-left: 216px;
}
</style>
