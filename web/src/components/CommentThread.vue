<template>
  <div class="thread" :id="`comment-${comment.id}`">
    <div class="comment" :class="{ 'new-item': highlightId === comment.id }">
      <div class="comment-avatar-wrapper">
        <div class="comment-avatar avatar-clickable" @click="showUserCard">
          <UserAvatar :user-id="comment.user_id" :username="comment.username" :size="30" />
        </div>
        <UserCard ref="userCardRef" :user-id="comment.user_id ?? null" :username="comment.username" />
      </div>
      <div class="comment-main">
        <div class="comment-head">
          <span class="comment-author">{{ comment.username }}</span>
          <span class="comment-date">{{ formatDate(comment.created_at) }}</span>
        </div>
        <div class="comment-body">
          <span v-if="comment.reply_to_username" class="reply-to"
            >@{{ comment.reply_to_username }}</span
          >
          {{ comment.content }}
        </div>
        <div class="comment-actions">
          <button
            class="like-btn"
            :class="{ 'is-liked': isLiked }"
            :disabled="isLiking"
            @click="toggleLike(comment)"
          >
            <svg viewBox="0 0 24 24" class="like-icon">
              <path
                v-if="isLiked"
                d="M12 21.35l-1.45-1.32C5.4 15.36 2 12.28 2 8.5 2 5.42 4.42 3 7.5 3c1.74 0 3.41.81 4.5 2.09C13.09 3.81 14.76 3 16.5 3 19.58 3 22 5.42 22 8.5c0 3.78-3.4 6.86-8.55 11.54L12 21.35z"
              />
              <path
                v-else
                d="M16.5 3c-1.74 0-3.41.81-4.5 2.09C10.91 3.81 9.24 3 7.5 3 4.42 3 2 5.42 2 8.5c0 3.78 3.4 6.86 8.55 11.54L12 21.35l1.45-1.32C18.6 15.36 22 12.28 22 8.5 22 5.42 19.58 3 16.5 3zm-4.4 15.55l-.1.1-.1-.1C7.14 14.24 4 11.39 4 8.5 4 6.5 5.5 5 7.5 5c1.54 0 3.04.99 3.57 2.36h1.87C13.46 5.99 14.96 5 16.5 5c2 0 3.5 1.5 3.5 3.5 0 2.89-3.14 5.74-7.9 10.05z"
              />
            </svg>
            <span v-if="(comment.like_count || 0) > 0" class="like-count">{{
              comment.like_count
            }}</span>
          </button>
          <el-button v-if="isLoggedIn" text class="action-btn" @click="toggleReply(comment.id)">
            Reply
          </el-button>
          <el-button
            v-if="comment.username === currentUser"
            text
            class="action-btn"
            :loading="deletingId === comment.id"
            @click="deleteComment(comment.id)"
          >
            Delete
          </el-button>
        </div>
        <div v-if="replyingTo === comment.id" class="comment-box reply-box">
          <el-input
            ref="replyInput"
            v-model="replyContent"
            type="textarea"
            :rows="2"
            placeholder="Write a reply..."
            class="comment-input"
          />
          <div class="input-actions">
            <el-button
              class="bw-btn"
              size="small"
              :loading="replying"
              @click="submitReply(comment.id)"
              >Reply</el-button
            >
            <el-button class="bw-btn" size="small" @click="replyingTo = null">Cancel</el-button>
          </div>
        </div>
      </div>
    </div>

    <div class="reply-tree" v-if="replies.length > 0">
      <CommentThread v-for="reply in replies" :key="reply.id" :node="{ comment: reply, children: [] }" />
    </div>

    <div v-if="totalCount > 0 || loadingReplies" class="more-replies">
      <button class="more-btn" :disabled="loadingReplies" @click="toggleReplies">
        <span v-if="loadingReplies">Loading...</span>
        <span v-else-if="replies.length === 0">{{ totalCount === 1 ? 'Show 1 reply' : `Show ${totalCount} replies` }}</span>
        <span v-else>Hide replies</span>
      </button>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, inject, nextTick, onBeforeUnmount, onMounted, ref, watch } from 'vue'
import type { ComputedRef, Ref } from 'vue'
import type { ElInput } from 'element-plus'
import { apiFetch } from '@/api'
import UserAvatar from '@/components/UserAvatar.vue'
import UserCard from '@/components/UserCard.vue'

interface Comment {
  id: number
  user_id?: number
  username: string
  content: string
  created_at: string
  parent_id?: number | null
  reply_to_username?: string
  like_count?: number
}

export interface CommentNode {
  comment: Comment
  children: CommentNode[]
}

interface ThreadContext {
  formatDate: (value?: string) => string
  toggleReply: (commentId: number) => void
  deleteComment: (commentId: number) => void
  submitReply: (commentId: number) => void
  currentUser: Ref<string | null>
  isLoggedIn: ComputedRef<boolean>
  deletingId: Ref<number | null>
  replyingTo: Ref<number | null>
  replyContent: Ref<string>
  replying: Ref<boolean>
  toggleLike: (comment: Comment) => void
  likedComments: Ref<Set<number>>
  likingId: Ref<number | null>
  highlightId: Ref<number | null>
  registerReplyCallback: (commentId: number, cb: () => Promise<void>) => void
  unregisterReplyCallback: (commentId: number) => void
}

const props = defineProps<{
  node: CommentNode
}>()

const ctx = inject<ThreadContext>('commentThread')
if (!ctx) {
  throw new Error('CommentThread must be used inside Comments')
}

const {
  formatDate,
  toggleReply,
  deleteComment,
  submitReply,
  currentUser,
  isLoggedIn,
  deletingId,
  replyingTo,
  replyContent,
  replying,
  toggleLike,
  likedComments,
  likingId,
  highlightId,
  registerReplyCallback,
  unregisterReplyCallback,
} = ctx

const comment = computed(() => props.node.comment)

const userCardRef = ref<InstanceType<typeof UserCard> | null>(null)

function showUserCard(event: MouseEvent) {
  userCardRef.value?.show(event)
}

const replies = ref<Comment[]>([])
const totalCount = ref(0)
const loadingReplies = ref(false)
const repliesLoaded = ref(false)

async function fetchReplyCount() {
  try {
    const response = await apiFetch(`/comments/${comment.value.id}/replies?limit=1&offset=0`)
    if (!response.ok) return
    const count = Number(response.headers.get('X-Total-Count') ?? 0)
    totalCount.value = count
  } catch {
    // ignore
  }
}

async function toggleReplies() {
  if (repliesLoaded.value) {
    replies.value = []
    repliesLoaded.value = false
    return
  }
  loadingReplies.value = true
  try {
    const response = await apiFetch(`/comments/${comment.value.id}/replies?limit=100&offset=0&sort=asc`)
    if (!response.ok) throw new Error(`HTTP ${response.status}`)
    const json = await response.json()
    const data = json.data ?? json
    replies.value = Array.isArray(data) ? data : []
    repliesLoaded.value = true
  } catch {
    // ignore
  } finally {
    loadingReplies.value = false
  }
}

async function refreshReplies() {
  if (!repliesLoaded.value) {
    await fetchReplyCount()
    return
  }
  try {
    const response = await apiFetch(`/comments/${comment.value.id}/replies?limit=100&offset=0&sort=asc`)
    if (!response.ok) throw new Error(`HTTP ${response.status}`)
    const json = await response.json()
    const data = json.data ?? json
    replies.value = Array.isArray(data) ? data : []
  } catch {
    // ignore
  }
}

onMounted(() => {
  fetchReplyCount()
  registerReplyCallback(comment.value.id, refreshReplies)
})

onBeforeUnmount(() => {
  unregisterReplyCallback(comment.value.id)
})

const replyInput = ref<InstanceType<typeof ElInput> | null>(null)
watch(
  () => replyingTo.value,
  (id) => {
    if (id === comment.value.id) {
      nextTick(() => replyInput.value?.focus())
    }
  },
)

const isLiked = computed(() => likedComments.value.has(comment.value.id))
const isLiking = computed(() => likingId.value === comment.value.id)
</script>

<style scoped>
.thread {
  width: 100%;
}

.comment {
  display: flex;
  align-items: flex-start;
  gap: 10px;
  padding: 6px 2px;
}

.comment-avatar {
  flex-shrink: 0;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 30px;
  height: 30px;
  border-radius: 50%;
  background: #333;
  color: #e4e6e8;
  font-size: 14px;
  font-weight: 600;
}

.comment-avatar-wrapper {
  position: relative;
}

.avatar-clickable {
  cursor: pointer;
  transition: transform 0.2s ease;
}

.avatar-clickable:hover {
  transform: scale(1.1);
}

.comment-main {
  flex: 1;
  min-width: 0;
}

.comment-head {
  display: flex;
  align-items: baseline;
  gap: 8px;
  margin-bottom: 4px;
}

.comment-author {
  font-weight: 600;
  font-size: 13px;
  color: #6cbbf7;
}

.comment-date {
  font-size: 12px;
  color: #8c8c8c;
}

.comment-body {
  font-size: 14px;
  line-height: 1.45;
  color: #e4e6e8;
  white-space: pre-wrap;
  word-break: break-word;
  margin-bottom: 6px;
}

.reply-to {
  color: #8c8c8c;
}

.comment-actions {
  display: flex;
  justify-content: flex-end;
  gap: 14px;
}

.action-btn {
  padding: 0;
  font-size: 12px;
  color: #8c8c8c;
  background: transparent !important;
}

.action-btn:hover,
.action-btn:focus-visible {
  color: #6cbbf7;
  background: transparent !important;
  text-decoration: none;
}

.like-btn {
  display: inline-flex;
  align-items: center;
  gap: 4px;
  padding: 0;
  background: transparent;
  border: none;
  cursor: pointer;
  color: #8c8c8c;
  font-size: 12px;
  transition: color 0.2s;
}

.like-btn:hover:not(:disabled) {
  color: #e05c5c;
}

.like-btn:disabled {
  cursor: not-allowed;
  opacity: 0.6;
}

.like-btn.is-liked {
  color: #e05c5c;
}

.like-icon {
  width: 16px;
  height: 16px;
  fill: currentColor;
}

.like-count {
  line-height: 1;
}

.comment-box {
  border: 1px solid #262626;
  border-radius: 8px;
  padding: 12px;
  background: #141414;
}

.reply-box {
  margin-top: 8px;
}

.bw-btn,
.bw-btn + .bw-btn {
  margin-left: 0;
  background: #262626;
  color: #e4e6e8;
  border: 1px solid #333;
  font-weight: 600;
}

.bw-btn:hover {
  background: #333;
  color: #e4e6e8;
  border-color: #595959;
}

.comment-input :deep(.el-textarea__inner) {
  background: #141414;
  border: 1px solid #262626;
  border-radius: 6px;
  color: #e4e6e8;
}

.comment-input :deep(.el-textarea__inner::placeholder) {
  color: #8c8c8c;
}

.input-actions {
  display: flex;
  justify-content: flex-end;
  gap: 8px;
  margin-top: 10px;
}

.reply-tree {
  border-left: 2px solid #333;
  margin-left: 12px;
  padding-left: 16px;
}

.more-replies {
  padding: 8px 0;
}

.more-btn {
  background: transparent;
  border: none;
  padding: 0;
  font-size: 12px;
  color: #8c8c8c;
  cursor: pointer;
  text-decoration: underline;
  text-decoration-color: #8c8c8c;
  text-underline-offset: 3px;
}

.more-btn:hover:not(:disabled) {
  color: #6cbbf7;
  text-decoration-color: #6cbbf7;
}

.more-btn:disabled {
  cursor: wait;
}

.thread.highlight {
  background: rgba(37, 99, 235, 0.05);
  border-radius: 8px;
  transition: background 0.3s ease;
}

.comment.new-item {
  animation: highlight-flash 2.5s ease-out;
}

@keyframes highlight-flash {
  0% {
    background-color: rgba(64, 158, 255, 0.15);
  }
  100% {
    background-color: transparent;
  }
}
</style>
