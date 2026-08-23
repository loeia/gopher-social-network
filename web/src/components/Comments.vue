<template>
  <div class="comments" v-loading="loading">
    <h3 class="comments-title" v-if="comments.length">Comments ({{ comments.length }})</h3>

    <CommentThread
      v-for="root in commentTree"
      :key="root.comment.id"
      :node="root"
    />

    <div class="add-comment" v-if="isLoggedIn">
      <el-button v-if="!showCommentBox" text class="add-comment-link" @click="showCommentBox = true">
        Add a comment
      </el-button>
      <div v-else class="comment-box">
        <el-input
          ref="commentInputRef"
          v-model="newComment"
          type="textarea"
          :rows="3"
          placeholder="Write a comment..."
          class="comment-input"
        />
        <div class="comment-actions">
          <el-button class="bw-btn" size="small" :loading="commenting" @click="submitComment">
            Add Comment
          </el-button>
          <el-button class="bw-btn" size="small" @click="cancelComment">Cancel</el-button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, nextTick, onBeforeUnmount, onMounted, provide, ref, watch } from 'vue'
import { apiFetch, getToken } from '@/api'
import { notify } from '@/utils/message'
import type { ElInput } from 'element-plus'
import CommentThread, { type CommentNode } from '@/components/CommentThread.vue'
import { useFeedStore } from '@/stores/feed'

const store = useFeedStore()

interface Comment {
  id: number
  user_id?: number
  username: string
  content: string
  created_at: string
  parent_id?: number | null
  reply_to_user_id?: number | null
  reply_to_username?: string
  like_count?: number
}

const props = defineProps<{
  postId: number
}>()

const emit = defineEmits<{
  count: [value: number]
}>()

const comments = ref<Comment[]>([])
const loading = ref(false)

const isLoggedIn = computed(() => !!getToken())
const showCommentBox = ref(false)
const newComment = ref('')
const commenting = ref(false)
const commentInputRef = ref<InstanceType<typeof ElInput> | null>(null)
const replyingTo = ref<number | null>(null)
const replyContent = ref('')
const replying = ref(false)
const deletingId = ref<number | null>(null)
const currentUser = ref<string | null>(null)
const likedComments = ref<Set<number>>(new Set())
const likingId = ref<number | null>(null)

const replyShown = ref<Record<number, number>>({})
function shownFor(parentId: number): number {
  return replyShown.value[parentId] ?? 0
}
function moreReplies(parentId: number, total: number) {
  const cur = shownFor(parentId)
  if (cur <= 0) {
    replyShown.value = { ...replyShown.value, [parentId]: Math.min(3, total) }
  } else if (cur >= total) {
    replyShown.value = { ...replyShown.value, [parentId]: 0 }
  } else {
    replyShown.value = { ...replyShown.value, [parentId]: Math.min(cur + 5, total) }
  }
}

function buildTree(list: Comment[]): CommentNode[] {
  const map = new Map<number, CommentNode>()
  for (const c of list) map.set(c.id, { comment: c, children: [] })
  const roots: CommentNode[] = []
  for (const c of list) {
    const node = map.get(c.id)!
    if (c.parent_id != null && map.has(c.parent_id)) {
      map.get(c.parent_id)!.children.push(node)
    } else {
      roots.push(node)
    }
  }
  return roots
}

const commentTree = computed(() => buildTree(comments.value))

provide('commentThread', {
  formatDate,
  toggleReply,
  deleteComment,
  submitReply,
  shownFor,
  moreReplies,
  currentUser,
  isLoggedIn,
  deletingId,
  replyingTo,
  replyContent,
  replying,
  toggleLike,
  likedComments,
  likingId,
})

function decodeJwtPayload(token: string): Record<string, unknown> | null {
  try {
    const base64 = token.split('.')[1]
    if (!base64) return null
    const json = base64.replace(/-/g, '+').replace(/_/g, '/')
    return JSON.parse(atob(json))
  } catch {
    return null
  }
}

async function loadCurrentUser() {
  const token = getToken()
  if (!token) {
    currentUser.value = null
    return
  }
  const payload = decodeJwtPayload(token)
  const userId = payload?.sub as number | undefined
  if (!userId) {
    currentUser.value = null
    return
  }
  try {
    const response = await apiFetch(`/users/${userId}`)
    if (!response.ok) throw new Error(`HTTP ${response.status}`)
    const json = await response.json()
    const user = json.data ?? json
    currentUser.value = user?.username ?? null
  } catch (error) {
    console.error('Load current user error:', error)
    currentUser.value = null
  }
}

async function fetchLikedComments() {
  if (!isLoggedIn.value) {
    likedComments.value = new Set()
    return
  }
  try {
    const response = await apiFetch('/users/comment-likes')
    if (!response.ok) throw new Error(`HTTP ${response.status}`)
    const json = await response.json()
    const data = json.data ?? json
    const ids = Array.isArray(data) ? data.map((item: { comment_id: number }) => item.comment_id) : []
    likedComments.value = new Set(ids)
  } catch (error) {
    console.error('Fetch liked comments error:', error)
    likedComments.value = new Set()
  }
}

async function deleteComment(commentId: number) {
  deletingId.value = commentId
  try {
    const response = await apiFetch(`/comments/${commentId}`, { method: 'DELETE' })
    if (!response.ok) throw new Error(`HTTP ${response.status}`)
    notify('success', 'Comment deleted')
    await loadComments()
    store.updatePostCommentCount(props.postId, comments.value.length)
  } catch (error) {
    console.error('Delete comment error:', error)
    notify('error', 'Failed to delete comment')
  } finally {
    deletingId.value = null
  }
}

function cancelComment() {
  showCommentBox.value = false
  newComment.value = ''
}

watch(showCommentBox, (open) => {
  if (open) {
    nextTick(() => commentInputRef.value?.focus())
  }
})

function toggleReply(commentId: number) {
  replyingTo.value = replyingTo.value === commentId ? null : commentId
  replyContent.value = ''
}

async function submitReply(commentId: number) {
  const content = replyContent.value.trim()
  if (!content) {
    notify('warning', 'Please enter a reply')
    return
  }
  replying.value = true
  try {
    const response = await apiFetch(`/posts/${props.postId}/comments/${commentId}/reply`, {
      method: 'POST',
      body: JSON.stringify({ content }),
    })
    if (!response.ok) throw new Error(`HTTP ${response.status}`)
    replyContent.value = ''
    replyingTo.value = null
    notify('success', 'Reply added')
    await loadComments()
    store.updatePostCommentCount(props.postId, comments.value.length)
  } catch (error) {
    console.error('Reply comment error:', error)
    notify('error', 'Failed to add reply')
  } finally {
    replying.value = false
  }
}

async function submitComment() {
  const content = newComment.value.trim()
  if (!content) {
    notify('warning', 'Please enter a comment')
    return
  }
  commenting.value = true
  try {
    const response = await apiFetch(`/posts/${props.postId}/comments`, {
      method: 'POST',
      body: JSON.stringify({ content }),
    })
    if (!response.ok) throw new Error(`HTTP ${response.status}`)
    newComment.value = ''
    showCommentBox.value = false
    notify('success', 'Comment added')
    await loadComments()
    store.updatePostCommentCount(props.postId, comments.value.length)
  } catch (error) {
    console.error('Add comment error:', error)
    notify('error', 'Failed to add comment')
  } finally {
    commenting.value = false
  }
}

function formatDate(value?: string) {
  if (!value) return ''
  const date = new Date(value)
  if (Number.isNaN(date.getTime())) return value
  return date.toLocaleString()
}

async function toggleLike(comment: Comment) {
  if (!isLoggedIn.value) {
    notify('warning', 'Please sign in to like comments')
    return
  }
  if (likingId.value === comment.id) return

  likingId.value = comment.id
  const isLiked = likedComments.value.has(comment.id)
  try {
    const endpoint = isLiked ? `/comments/${comment.id}/dislike` : `/comments/${comment.id}/like`
    const response = await apiFetch(endpoint, { method: 'PUT' })
    if (!response.ok) throw new Error(`HTTP ${response.status}`)
    if (isLiked) {
      likedComments.value.delete(comment.id)
      comment.like_count = Math.max(0, (comment.like_count || 1) - 1)
    } else {
      likedComments.value.add(comment.id)
      comment.like_count = (comment.like_count || 0) + 1
    }
    likedComments.value = new Set(likedComments.value)
  } catch (error) {
    console.error('Toggle like error:', error)
    notify('error', 'Failed to like comment')
  } finally {
    likingId.value = null
  }
}

async function loadComments(silent = false) {
  if (!silent) loading.value = true
  try {
    const response = await apiFetch(`/posts/${props.postId}/comments`)
    if (!response.ok) throw new Error(`HTTP ${response.status}`)
    const json = await response.json()
    comments.value = json.data ?? []
    emit('count', comments.value.length)
    if (!silent) nextTick(scrollToComment)
  } catch (error) {
    console.error('Load comments error:', error)
    if (!silent) notify('error', 'Failed to load comments')
  } finally {
    if (!silent) loading.value = false
  }
}

function scrollToComment() {
  const hash = window.location.hash
  if (!hash || !hash.startsWith('#comment-')) return
  const id = hash.slice(9)
  const el = document.getElementById(`comment-${id}`)
  if (el) {
    el.scrollIntoView({ behavior: 'smooth', block: 'center' })
    el.classList.add('highlight')
    setTimeout(() => el.classList.remove('highlight'), 2000)
  }
}

let pollTimer: number | null = null

function startPolling() {
  stopPolling()
  pollTimer = window.setInterval(() => {
    if (!loading.value) loadComments(true)
  }, 5000)
}

function stopPolling() {
  if (pollTimer !== null) {
    clearInterval(pollTimer)
    pollTimer = null
  }
}

onMounted(() => {
  loadCurrentUser()
  loadComments()
  fetchLikedComments()
  startPolling()
})
watch(() => props.postId, () => {
  loadCurrentUser()
  loadComments()
  fetchLikedComments()
  startPolling()
})
onBeforeUnmount(stopPolling)
</script>

<style scoped>
.comments {
  margin-top: 32px;
  padding-top: 16px;
  border-top: 1px solid #3d4043;
}

.comments-title {
  margin: 0 0 12px;
  font-size: 16px;
  font-weight: 600;
  color: #ffffff;
}

.add-comment {
  margin-top: 16px;
}

.add-comment-link {
  font-size: 13px;
  color: #8c8c8c;
  padding: 0;
  background: transparent !important;
  text-decoration: underline;
  text-decoration-color: #595959;
  text-underline-offset: 3px;
}

.add-comment-link:hover,
.add-comment-link:active {
  color: #8c8c8c;
  background: transparent !important;
  text-decoration: underline;
  text-decoration-color: #e4e6e8;
  text-underline-offset: 3px;
}

.comment-box {
  border: 1px solid #3d4043;
  border-radius: 8px;
  padding: 12px;
  background: #0f1010;
}

.bw-btn,
.bw-btn + .bw-btn {
  margin-left: 0;
  background: #ffffff;
  color: #141414;
  border: 1px solid #ffffff;
  font-weight: 600;
}

.bw-btn:hover {
  background: #e4e6e8;
  color: #141414;
  border-color: #e4e6e8;
}

.comment-input :deep(.el-textarea__inner) {
  background: #0a0a0a;
  border: 1px solid #262626;
  border-radius: 6px;
  color: #e4e6e8;
}

.comment-input :deep(.el-textarea__inner::placeholder) {
  color: #595959;
}

.comment-actions {
  display: flex;
  justify-content: flex-end;
  gap: 8px;
  margin-top: 10px;
}
</style>