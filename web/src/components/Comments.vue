<template>
  <div class="comments" v-loading="loading">
    <h3 class="comments-title" v-if="comments.length">Comments ({{ comments.length }})</h3>

    <CommentThread v-for="root in commentTree" :key="root.comment.id" :node="root" />

    <div v-if="loadingMore" class="loading-more">Loading...</div>

    <div class="add-comment" v-if="isLoggedIn">
      <el-button
        v-if="!showCommentBox"
        text
        class="add-comment-link"
        @click="showCommentBox = true"
      >
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
import { computed, nextTick, onMounted, provide, ref, watch } from 'vue'
import { apiFetch, getToken, handleApiError } from '@/api'
import { notify } from '@/utils/message'
import type { ElInput } from 'element-plus'
import CommentThread, { type CommentNode } from '@/components/CommentThread.vue'
import { useFeedStore } from '@/stores/feed'
import { useInfiniteScroll } from '@/composables/useInfiniteScroll'

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
const commentsOffset = ref(0)
const hasMore = ref(true)
const highlightId = ref<number | null>(null)
const hasScrolledToComment = ref(false)

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
  highlightId,
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
    const response = await apiFetch('/users/comment-likes?limit=20&offset=0&sort=desc')
    if (!response.ok) throw new Error(`HTTP ${response.status}`)
    const json = await response.json()
    const data = json.data ?? json
    const ids = Array.isArray(data)
      ? data.map((item: { comment_id: number }) => item.comment_id)
      : []
    likedComments.value = new Set(ids)
  } catch (error) {
    console.error('Fetch liked comments error:', error)
    likedComments.value = new Set()
  }
}

const canLoadMore = computed(() => hasMore.value && !loading.value)
const { loadingMore } = useInfiniteScroll(loadMoreComments, canLoadMore)

async function deleteComment(commentId: number) {
  deletingId.value = commentId
  try {
    const response = await apiFetch(`/comments/${commentId}`, { method: 'DELETE' })
    if (!response.ok) throw new Error(`HTTP ${response.status}`)
    notify('success', 'Comment deleted')
    commentsOffset.value = 0
    hasMore.value = true
    await loadComments()
    store.updatePostCommentCount(props.postId, comments.value.length)
  } catch (error) {
    handleApiError(error, 'Failed to delete comment')
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
    commentsOffset.value = 0
    hasMore.value = true
    await loadComments()
    store.updatePostCommentCount(props.postId, comments.value.length)
  } catch (error) {
    handleApiError(error, 'Failed to add reply')
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
    commentsOffset.value = 0
    hasMore.value = true
    await loadComments()
    store.updatePostCommentCount(props.postId, comments.value.length)
  } catch (error) {
    handleApiError(error, 'Failed to add comment')
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
    handleApiError(error, 'Failed to like comment')
  } finally {
    likingId.value = null
  }
}

async function loadComments(silent = false) {
  if (!silent) loading.value = true
  try {
    const response = await apiFetch(
      `/posts/${props.postId}/comments?limit=20&offset=0&sort=desc`,
    )
    if (!response.ok) throw new Error(`HTTP ${response.status}`)
    const json = await response.json()
    const data = json.data ?? json
    if (silent) {
      const existingIds = new Set(comments.value.map((c) => c.id))
      const newComments = Array.isArray(data) ? data.filter((c: Comment) => !existingIds.has(c.id)) : []
      if (newComments.length > 0) {
        comments.value = [...newComments, ...comments.value]
      }
      hasMore.value = comments.value.length >= 20
    } else {
      comments.value = Array.isArray(data) ? data : []
      commentsOffset.value = 0
      hasMore.value = comments.value.length >= 20
    }
    emit('count', comments.value.length)
    if (!silent) nextTick(scrollToComment)
  } catch (error) {
    if (!silent) handleApiError(error, 'Failed to load comments')
  } finally {
    if (!silent) loading.value = false
  }
}

async function loadMoreComments() {
  if (!hasMore.value) return
  try {
    const response = await apiFetch(
      `/posts/${props.postId}/comments?limit=20&offset=${commentsOffset.value}&sort=desc`,
    )
    if (!response.ok) throw new Error(`HTTP ${response.status}`)
    const json = await response.json()
    const data = json.data ?? json
    const newComments = Array.isArray(data) ? data : []
    if (newComments.length > 0) {
      comments.value = [...comments.value, ...newComments]
      commentsOffset.value += newComments.length
      highlightId.value = newComments[0].id
      setTimeout(() => { highlightId.value = null }, 2500)
    }
    if (newComments.length < 20) {
      hasMore.value = false
    }
  } catch (error) {
    handleApiError(error, 'Failed to load more comments')
  }
}

function expandParentComments(targetId: number) {
  const commentMap = new Map<number, Comment>()
  for (const c of comments.value) commentMap.set(c.id, c)

  let current = commentMap.get(targetId)
  while (current?.parent_id) {
    const parentId = current.parent_id
    const parentChildren = comments.value.filter((c) => c.parent_id === parentId)
    const curShown = shownFor(parentId)
    if (curShown < parentChildren.length) {
      replyShown.value = { ...replyShown.value, [parentId]: parentChildren.length }
    }
    current = commentMap.get(parentId)
  }
}

function scrollToComment() {
  if (hasScrolledToComment.value) return
  const hash = window.location.hash
  if (!hash || !hash.startsWith('#comment-')) return
  const id = parseInt(hash.slice(9), 10)
  if (isNaN(id)) return

  hasScrolledToComment.value = true
  expandParentComments(id)

  nextTick(() => {
    nextTick(() => {
      const el = document.getElementById(`comment-${id}`)
      if (el) {
        el.scrollIntoView({ behavior: 'smooth', block: 'center' })
        el.classList.add('highlight')
        setTimeout(() => el.classList.remove('highlight'), 2000)
      }
    })
  })
}

onMounted(() => {
  hasScrolledToComment.value = false
  loadCurrentUser()
  loadComments()
  fetchLikedComments()
})
watch(
  () => props.postId,
  () => {
    loadCurrentUser()
    loadComments()
    fetchLikedComments()
  },
)
</script>

<style scoped>
.comments {
  margin-top: 32px;
  padding-top: 16px;
  border-top: 1px solid #262626;
}

.comments-title {
  margin: 0 0 12px;
  font-size: 16px;
  font-weight: 600;
  color: #e4e6e8;
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
  text-decoration-color: #8c8c8c;
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
  border: 1px solid #262626;
  border-radius: 8px;
  padding: 12px;
  background: #141414;
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

.comment-actions {
  display: flex;
  justify-content: flex-end;
  gap: 8px;
  margin-top: 10px;
}

.loading-more {
  text-align: center;
  padding: 16px;
  color: #8c8c8c;
  font-size: 14px;
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
