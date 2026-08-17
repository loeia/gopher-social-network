<template>
  <div class="feed" v-loading="loading">
    <div v-for="post in posts" :key="post.id" class="card" @click="openPost(post.id)">
      <div class="card-header">
        <h2 class="card-title">{{ post.title }}</h2>
        <div class="card-meta">
          <span class="stat" :title="`${post.comment_count} comments`">
            <svg class="stat-icon comment-icon" viewBox="0 0 24 24" aria-hidden="true">
              <path d="M20 2H4a2 2 0 0 0-2 2v18l4-4h14a2 2 0 0 0 2-2V4a2 2 0 0 0-2-2z" />
            </svg>
            <span>{{ post.comment_count }}</span>
          </span>
          <span class="stat" :title="`${post.like_count} likes`">
            <svg
              class="stat-icon like-icon"
              :class="{ liked: isLiked(post) }"
              viewBox="0 0 24 24"
              aria-hidden="true"
            >
              <path
                d="M12 21.35l-1.45-1.32C5.4 15.36 2 12.28 2 8.5 2 5.42 4.42 3 7.5 3c1.74 0 3.41.81 4.5 2.09C13.09 3.81 14.76 3 16.5 3 19.58 3 22 5.42 22 8.5c0 3.78-3.4 6.86-8.55 11.54L12 21.35z"
              />
            </svg>
            <span>{{ post.like_count }}</span>
          </span>
          <span class="card-date">{{ formatDate(post.created_at) }}</span>
        </div>
      </div>
      <div v-if="post.tags && post.tags.length" class="card-tags">
        <span v-for="tag in post.tags" :key="tag" class="tag-pill">{{ tag }}</span>
      </div>
      <div class="card-author">
        <UserAvatar :user-id="post.user_id" :username="post.user.username" :size="28" />
        <span>{{ post.user.username }}</span>
        <div v-if="editable" class="card-actions">
          <button class="back-btn" @click.stop="onEdit(post)">Edit</button>
          <button class="back-btn" @click.stop="onDelete(post)">Delete</button>
        </div>
      </div>
    </div>
    <button v-if="!isStandalone" class="new-btn" :disabled="loading" @click="loadNewPosts">
      New
    </button>
  </div>
</template>

<script setup lang="ts">
import {
  computed,
  nextTick,
  onActivated,
  onBeforeUnmount,
  onDeactivated,
  onMounted,
  ref,
} from 'vue'
import { useRouter } from 'vue-router'
import { storeToRefs } from 'pinia'
import { useFeedStore, type FeedPost } from '@/stores/feed'
import { getToken } from '@/api'
import { notify } from '@/utils/message'
import UserAvatar from '@/components/UserAvatar.vue'

const props = withDefaults(
  defineProps<{
    posts?: FeedPost[]
    loading?: boolean
    editable?: boolean
  }>(),
  {
    posts: undefined,
    loading: false,
    editable: false,
  },
)

const emit = defineEmits<{ edit: [post: FeedPost]; delete: [post: FeedPost] }>()

const store = useFeedStore()
const { posts: storePosts } = storeToRefs(store)
const fetching = ref(false)
const isStandalone = computed(() => !!props.posts)
const posts = computed(() => props.posts ?? storePosts.value)
const loading = computed(() => props.loading || fetching.value)

const likedIds = computed(() => new Set(store.likedPosts.map((p) => p.post_id)))

function isLiked(post: FeedPost): boolean {
  return likedIds.value.has(post.id)
}

const router = useRouter()

function openPost(id: number) {
  router.push(`/posts/${id}`)
}

function onEdit(post: FeedPost) {
  emit('edit', post)
}

function onDelete(post: FeedPost) {
  emit('delete', post)
}

function formatDate(value?: string) {
  if (!value) return ''
  const date = new Date(value)
  if (Number.isNaN(date.getTime())) return value
  return date.toLocaleString()
}

function saveScroll() {
  store.feedScrollTop = window.scrollY
}

function restoreScroll() {
  if (store.feedScrollTop > 0) {
    nextTick(() => window.scrollTo({ top: store.feedScrollTop }))
  }
}

async function loadPosts() {
  if (store.postsLoaded || isStandalone.value) return
  fetching.value = true
  try {
    await store.fetchPosts()
  } catch (error) {
    console.error('Load posts error:', error)
    notify('error', 'Failed to load posts')
  } finally {
    fetching.value = false
  }
}

async function loadNewPosts() {
  fetching.value = true
  try {
    await store.refreshPosts()
    window.scrollTo({ top: 0 })
  } catch (error) {
    console.error('Refresh posts error:', error)
    notify('error', 'Failed to refresh posts')
  } finally {
    fetching.value = false
  }
}

onMounted(() => {
  loadPosts()
  loadLikedPosts()
})

async function loadLikedPosts() {
  if (!getToken() || store.likedPostsLoaded) return
  try {
    await store.fetchLikedPosts()
  } catch (error) {
    console.error('Load liked posts error:', error)
  }
}
onActivated(() => {
  if (!isStandalone.value) restoreScroll()
})
onDeactivated(saveScroll)
onBeforeUnmount(saveScroll)
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
  align-items: flex-start;
  justify-content: space-between;
  gap: 16px;
}

.card-title {
  flex: 1;
  min-width: 0;
  margin: 0;
  font-size: 20px;
  font-weight: 600;
  line-height: 1.4;
  color: #ffffff;
  display: -webkit-box;
  -webkit-line-clamp: 3;
  -webkit-box-orient: vertical;
  overflow: hidden;
  word-break: break-word;
}

.card-meta {
  flex-shrink: 0;
  display: flex;
  align-items: center;
  gap: 16px;
}

.stat {
  display: inline-flex;
  align-items: center;
  gap: 5px;
  font-size: 13px;
  color: #8c8c8c;
}

.stat-icon {
  width: 16px;
  height: 16px;
}

.comment-icon path {
  fill: none;
  stroke: currentColor;
  stroke-width: 2;
}

.like-icon path {
  fill: none;
  stroke: #ffffff;
  stroke-width: 2;
}

.like-icon.liked path {
  fill: #e05c5c;
  stroke: #e05c5c;
}

.card-date {
  flex-shrink: 0;
  font-size: 13px;
  color: #8c8c8c;
}

.card-tags {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
  padding-top: 12px;
}

.tag-pill {
  padding: 2px 10px;
  border: 1px solid #3d444d;
  border-radius: 999px;
  font-size: 12px;
  color: #bfbfbf;
  background: #1f1f1f;
  white-space: nowrap;
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

.card-actions {
  margin-left: auto;
  display: flex;
  align-items: center;
  gap: 16px;
}

.back-btn {
  padding: 0;
  background: transparent;
  border: none;
  color: #6a737c;
  font-size: 14px;
  font-weight: 500;
  cursor: pointer;
}

.back-btn:hover,
.back-btn:focus,
.back-btn:focus-visible {
  color: #6a737c;
  background: transparent;
  text-decoration: underline;
  text-decoration-color: #6a737c;
  text-underline-offset: 4px;
}

.new-btn {
  flex-shrink: 0;
  align-self: center;
  width: 12%;
  padding: 10px 0;
  border: 1px solid #ffffff;
  border-radius: 8px;
  background: #ffffff;
  color: #141414;
  font-size: 15px;
  font-weight: 600;
  text-align: center;
  white-space: nowrap;
  cursor: pointer;
  transition:
    background 0.2s ease,
    color 0.2s ease;
}

.new-btn:hover {
  background: #141414;
  color: #ffffff;
}

.new-btn:disabled {
  opacity: 0.6;
  cursor: not-allowed;
}
</style>
