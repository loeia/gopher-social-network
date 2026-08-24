<template>
  <div class="feed" v-loading="loading">
    <div
      v-for="comment in comments"
      :key="comment.comment_id"
      class="topic-row"
      @click="openPost(comment.post_id, comment.comment_id)"
    >
      <div class="topic-top">
        <p class="topic-content">{{ truncateContent(comment.content) }}</p>
        <div class="topic-stats">
          <span class="topic-stat">
            <svg class="stat-icon like-icon" viewBox="0 0 24 24" aria-hidden="true">
              <path
                d="M12 21.35l-1.45-1.32C5.4 15.36 2 12.28 2 8.5 2 5.42 4.42 3 7.5 3c1.74 0 3.41.81 4.5 2.09C13.09 3.81 14.76 3 16.5 3 19.58 3 22 5.42 22 8.5c0 3.78-3.4 6.86-8.55 11.54L12 21.35z"
              />
            </svg>
            {{ comment.like_count ?? 0 }}
          </span>
          <span class="topic-stat">
            <svg class="stat-icon" viewBox="0 0 24 24" aria-hidden="true">
              <path d="M20 2H4a2 2 0 0 0-2 2v18l4-4h14a2 2 0 0 0 2-2V4a2 2 0 0 0-2-2z" />
            </svg>
            {{ comment.reply_count ?? 0 }}
          </span>
        </div>
      </div>
      <div class="topic-bottom">
        <UserAvatar :user-id="comment.user_id" :username="comment.username" :size="20" />
        <span class="topic-author">{{ comment.username }}</span>
        <span class="meta-dot">&middot;</span>
        <span class="topic-time">{{ formatDate(comment.created_at) }}</span>
      </div>
    </div>
    <div v-if="!loading && comments.length === 0" class="empty">No liked comments yet</div>
  </div>
</template>

<script setup lang="ts">
import { nextTick, onActivated, onBeforeUnmount, onDeactivated, onMounted, ref } from 'vue'
import { useRouter } from 'vue-router'
import { storeToRefs } from 'pinia'
import { useFeedStore } from '@/stores/feed'
import { notify } from '@/utils/message'
import UserAvatar from '@/components/UserAvatar.vue'

const store = useFeedStore()
const { likedComments: comments } = storeToRefs(store)
const loading = ref(false)

const router = useRouter()

function openPost(postId: number, commentId: number) {
  router.push(`/posts/${postId}#comment-${commentId}`)
}

function truncateContent(content: string): string {
  if (!content) return ''
  return content.length > 80 ? content.slice(0, 80) + '...' : content
}

function formatDate(value?: string) {
  if (!value) return ''
  const date = new Date(value)
  if (Number.isNaN(date.getTime())) return value
  return date.toLocaleString()
}

function saveScroll() {
  store.likedCommentsScrollTop = window.scrollY
}

function restoreScroll() {
  nextTick(() => window.scrollTo({ top: store.likedCommentsScrollTop }))
}

async function loadLikedComments() {
  loading.value = true
  try {
    await store.fetchLikedComments()
  } catch (error) {
    console.error('Load liked comments error:', error)
    notify('error', 'Failed to load liked comments')
  } finally {
    loading.value = false
  }
}

onMounted(() => {
  restoreScroll()
  loadLikedComments()
})
onActivated(() => {
  restoreScroll()
  loadLikedComments()
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

.topic-content {
  flex: 1;
  min-width: 0;
  margin: 0;
  font-size: 15px;
  line-height: 1.5;
  color: #e4e6e8;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  word-break: break-word;
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

.like-icon {
  fill: #e05c5c;
  stroke: #e05c5c;
}

.topic-bottom {
  display: flex;
  align-items: center;
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

.empty {
  text-align: center;
  padding: 40px;
  color: #8c8c8c;
  font-size: 14px;
}
</style>
