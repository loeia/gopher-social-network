<template>
  <div class="feed" v-loading="loading">
    <div
      v-for="post in posts"
      :key="post.post_id"
      class="card"
      @click="openPost(post.post_id)"
    >
      <div class="card-header">
        <h2 class="card-title">{{ post.title }}</h2>
        <span class="card-date">{{ formatDate(post.created_at) }}</span>
      </div>
      <div class="card-author">
        <span class="avatar">G</span>
        <span>{{ post.author }}</span>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { useRouter } from 'vue-router'
import { storeToRefs } from 'pinia'
import { ElMessage } from 'element-plus'
import { useFeedStore } from '@/stores/feed'

const store = useFeedStore()
const { likedPosts: posts } = storeToRefs(store)
const loading = ref(false)

const router = useRouter()

function openPost(id: number) {
  router.push(`/posts/${id}`)
}

function formatDate(value?: string) {
  if (!value) return ''
  const date = new Date(value)
  if (Number.isNaN(date.getTime())) return value
  return date.toLocaleString()
}

async function loadLikedPosts() {
  if (store.likedPostsLoaded) return
  loading.value = true
  try {
    await store.fetchLikedPosts()
  } catch (error) {
    console.error('Load liked posts error:', error)
    ElMessage.error('Failed to load liked posts')
  } finally {
    loading.value = false
  }
}

onMounted(loadLikedPosts)
</script>

<style scoped>
.feed {
  margin: 0 20%;
  display: flex;
  flex-direction: column;
  gap: 20px;
}

.card {
  background: #141414;
  border: 1px solid #262626;
  border-radius: 12px;
  padding: 24px;
  cursor: pointer;
  transition:
    border-color 0.2s ease,
    transform 0.2s ease;
}

.card:hover {
  border-color: #ffffff;
  transform: translateY(-2px);
}

.card-header {
  display: flex;
  align-items: baseline;
  justify-content: space-between;
  gap: 16px;
}

.card-title {
  margin: 0;
  font-size: 20px;
  font-weight: 600;
  color: #ffffff;
}

.card-date {
  flex-shrink: 0;
  font-size: 13px;
  color: #8c8c8c;
}

.card-author {
  display: flex;
  align-items: center;
  gap: 10px;
  padding-top: 16px;
  border-top: 1px solid #262626;
  font-size: 14px;
  color: #8c8c8c;
}

.avatar {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 28px;
  height: 28px;
  border-radius: 50%;
  background: #ffffff;
  color: #141414;
  font-size: 13px;
  font-weight: 600;
}
</style>