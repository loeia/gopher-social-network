<template>
  <aside class="sidebar">
    <el-button class="btn home-btn" :class="{ active: activeView === 'all' }" @click="goHome"
      >Home</el-button
    >
    <el-button class="btn create-btn" @click="goToCreate">Create</el-button>
    <el-button
      class="btn my-posts-btn"
      :class="{ active: activeView === 'myposts' }"
      @click="goToMyPosts"
    >
      My Posts
    </el-button>
    <el-button
      class="btn likes-btn"
      :class="{ active: activeView === 'liked' }"
      @click="toggleLikes"
    >
      Likes
    </el-button>
    <el-button
      class="btn following-btn"
      :class="{ active: activeView === 'following' }"
      @click="toggleFollowing"
    >
      Following
    </el-button>
    <el-button
      class="btn followers-btn"
      :class="{ active: activeView === 'followers' }"
      @click="toggleFollowers"
    >
      Followers
    </el-button>
  </aside>
</template>

<script setup lang="ts">
import { useRouter } from 'vue-router'
import type { ViewType } from '@/stores/feed'

const props = defineProps<{ activeView: ViewType }>()

const emit = defineEmits<{ view: [view: ViewType] }>()

const router = useRouter()

function goToCreate() {
  router.push('/posts/new')
}

function goToMyPosts() {
  router.push('/my-posts')
}

function goHome() {
  emit('view', 'all')
  router.push('/')
}

function toggleLikes() {
  emit('view', props.activeView === 'liked' ? 'all' : 'liked')
  router.push('/')
}

function toggleFollowing() {
  emit('view', props.activeView === 'following' ? 'all' : 'following')
  router.push('/')
}

function toggleFollowers() {
  emit('view', props.activeView === 'followers' ? 'all' : 'followers')
  router.push('/')
}
</script>

<style scoped>
.sidebar {
  position: fixed;
  top: 0;
  left: 0;
  bottom: 0;
  width: 180px;
  padding: 80px 16px;
  background: #141414;
  border-right: 1px solid #262626;
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.btn,
.btn + .btn {
  width: 100%;
  margin-left: 0;
  font-weight: 600;
  background: #ffffff;
  color: #141414;
  border: 1px solid #ffffff;
}

.btn:hover {
  background: #e4e6e8;
  color: #141414;
}

.btn.active {
  background: #141414;
  color: #ffffff;
  border: 1px solid #ffffff;
}
</style>
