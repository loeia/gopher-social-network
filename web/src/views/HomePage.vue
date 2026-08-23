<template>
  <div class="home">
    <PostsList v-if="!isLoggedIn || view === 'all'" />
    <LikedPosts v-else-if="view === 'liked'" />
    <LikedComments v-else-if="view === 'likedcomments'" />
    <FollowingList v-else-if="view === 'following'" />
    <FollowersList v-else />
  </div>
</template>

<script setup lang="ts">
import { computed, onActivated, ref } from 'vue'
import { storeToRefs } from 'pinia'
import PostsList from '@/components/PostsList.vue'
import LikedPosts from '@/components/LikedPosts.vue'
import LikedComments from '@/components/LikedComments.vue'
import FollowingList from '@/components/FollowingList.vue'
import FollowersList from '@/components/FollowersList.vue'
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
</style>
