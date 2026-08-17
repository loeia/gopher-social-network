<template>
  <div class="my-posts-page">
    <div class="back-nav">
      <el-button text @click="router.back()">← Back</el-button>
    </div>

    <div class="page-header">
      <h1 class="page-title">My Posts</h1>
    </div>

    <PostsList :posts="posts" :loading="loading" editable />
  </div>
</template>

<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { useRouter } from 'vue-router'
import PostsList from '@/components/PostsList.vue'
import { apiFetch, getCurrentUserId } from '@/api'
import { toFeedPost, type FeedPost } from '@/stores/feed'
import { notify } from '@/utils/message'

defineOptions({ name: 'MyPostsPage' })

const router = useRouter()
const posts = ref<FeedPost[]>([])
const loading = ref(false)
const currentUserId = getCurrentUserId()

async function loadMyPosts() {
  loading.value = true
  try {
    const response = await apiFetch('/users/posts')
    if (!response.ok) throw new Error(`HTTP ${response.status}`)
    const json = await response.json()
    const raw = Array.isArray(json) ? json : (json.data ?? [])
    posts.value = raw.map((p: any) => ({
      ...toFeedPost(p),
      user_id: currentUserId ?? undefined,
    }))
  } catch (error) {
    console.error('Load my posts error:', error)
    notify('error', 'Failed to load my posts')
  } finally {
    loading.value = false
  }
}

onMounted(loadMyPosts)
</script>

<style scoped>
.my-posts-page {
  min-height: 100vh;
  padding: 32px 0 80px;
}

.back-nav {
  width: 75%;
  margin: 0 auto 16px;
  padding: 0 24px;
}

.back-nav :deep(.el-button) {
  color: #6a737c;
  background: transparent;
}

.back-nav :deep(.el-button:hover),
.back-nav :deep(.el-button:focus),
.back-nav :deep(.el-button:focus-visible) {
  color: #6a737c;
  background: transparent;
  text-decoration: underline;
  text-decoration-color: #6a737c;
  text-underline-offset: 4px;
}

.page-header {
  margin: 0 20% 24px;
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.page-title {
  margin: 0;
  font-size: 28px;
  font-weight: 600;
  color: #ffffff;
}
</style>
