<template>
  <div class="feed" v-loading="loading">
    <div v-for="post in posts" :key="post.id" class="topic-row" @click="openPost(post.id)">
      <div class="topic-top">
        <h2 class="topic-title">{{ post.title }}</h2>
        <div v-if="editable" class="topic-actions">
          <button class="action-btn" @click.stop="onEdit(post)">Edit</button>
          <button class="action-btn" @click.stop="onDelete(post)">Delete</button>
        </div>
        <div class="topic-stats">
          <span class="topic-stat" :title="`${post.comment_count} comments`">
            <svg class="stat-icon" viewBox="0 0 24 24" aria-hidden="true">
              <path d="M20 2H4a2 2 0 0 0-2 2v18l4-4h14a2 2 0 0 0 2-2V4a2 2 0 0 0-2-2z" />
            </svg>
            {{ post.comment_count }}
          </span>
          <span class="topic-stat" :title="`${post.like_count} likes`">
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
            {{ post.like_count }}
          </span>
        </div>
      </div>
      <div class="topic-bottom">
        <UserAvatar :user-id="post.user_id" :username="post.user.username" :size="20" />
        <span class="topic-author">{{ post.user.username }}</span>
        <span class="meta-dot">&middot;</span>
        <span class="topic-time">{{ formatDate(post.created_at) }}</span>
        <template v-if="post.tags && post.tags.length">
          <span class="meta-dot">&middot;</span>
          <span v-for="tag in post.tags" :key="tag" class="topic-tag">{{ tag }}</span>
        </template>
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

const emit = defineEmits<{
  edit: [post: FeedPost]
  delete: [post: FeedPost]
}>()

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
  nextTick(() => window.scrollTo({ top: store.feedScrollTop }))
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
  margin: 0 auto;
  max-width: 1100px;
  padding: 0 20px;
  display: flex;
  flex-direction: column;
}

.topic-row {
  padding: 14px 0;
  border-bottom: 1px solid #1f1f1f;
  cursor: pointer;
  transition: background 0.15s ease;
}

.topic-row:first-child {
  border-top: 1px solid #1f1f1f;
}

.topic-row:hover {
  background: rgba(255, 255, 255, 0.03);
}

.topic-top {
  display: flex;
  align-items: center;
  gap: 12px;
}

.topic-title {
  flex: 1;
  min-width: 0;
  margin: 0;
  font-size: 16px;
  font-weight: 600;
  line-height: 1.4;
  color: #e4e6e8;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.topic-actions {
  flex-shrink: 0;
  display: flex;
  gap: 6px;
}

.action-btn {
  padding: 2px 8px;
  background: transparent;
  border: 1px solid #333;
  border-radius: 4px;
  color: #8c8c8c;
  font-size: 12px;
  cursor: pointer;
  transition: all 0.15s ease;
}

.action-btn:hover {
  background: #262626;
  color: #e4e6e8;
  border-color: #555;
}

.topic-stats {
  flex-shrink: 0;
  display: flex;
  align-items: center;
  gap: 14px;
}

.topic-stat {
  display: inline-flex;
  align-items: center;
  gap: 4px;
  font-size: 13px;
  color: #8c8c8c;
  white-space: nowrap;
}

.stat-icon {
  width: 15px;
  height: 15px;
  fill: none;
  stroke: currentColor;
  stroke-width: 2;
}

.like-icon.liked {
  fill: #e05c5c;
  stroke: #e05c5c;
}

.topic-bottom {
  display: flex;
  align-items: center;
  flex-wrap: wrap;
  gap: 6px;
  margin-top: 6px;
  padding-left: 2px;
  font-size: 13px;
  color: #8c8c8c;
}

.topic-author {
  color: #bfbfbf;
  font-weight: 500;
}

.meta-dot {
  color: #555;
}

.topic-tag {
  padding: 1px 8px;
  background: #1f1f1f;
  border: 1px solid #333;
  border-radius: 4px;
  font-size: 12px;
  color: #bfbfbf;
  white-space: nowrap;
}

.new-btn {
  flex-shrink: 0;
  align-self: center;
  width: 12%;
  margin-top: 16px;
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
