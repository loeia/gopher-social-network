<template>
  <aside class="sidebar">
    <el-button class="btn home-btn" :class="{ active: activeView === 'all' }" @click="goHome">Home</el-button>
    <el-button class="btn create-btn" @click="goToCreate">Create</el-button>
    <el-button class="btn likes-btn" :class="{ active: activeView === 'liked' }" @click="toggleLikes">
      Likes
    </el-button>
  </aside>
</template>

<script setup lang="ts">
import { useRouter } from 'vue-router'

const props = defineProps<{ activeView: 'all' | 'liked' }>()

const emit = defineEmits<{ view: [view: 'all' | 'liked'] }>()

const router = useRouter()

function goToCreate() {
  router.push('/posts/new')
}

function goHome() {
  emit('view', 'all')
}

function toggleLikes() {
  emit('view', props.activeView === 'liked' ? 'all' : 'liked')
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