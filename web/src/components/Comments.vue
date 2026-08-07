<template>
  <div class="comments" v-loading="loading">
    <h3 class="comments-title" v-if="comments.length">Comments ({{ comments.length }})</h3>

    <div v-for="comment in comments" :key="comment.id" class="comment">
      <div class="comment-avatar">{{ (comment.username || 'G').charAt(0).toUpperCase() }}</div>
      <div class="comment-body">
        <div class="comment-meta">
          <span class="comment-username">{{ comment.username }}</span>
          <span class="comment-date">{{ formatDate(comment.created_at) }}</span>
        </div>
        <div class="comment-content">{{ comment.content }}</div>
        <div class="comment-footer">
          <el-button
            v-if="comment.username === currentUser"
            text
            class="reply-btn"
            :loading="deletingId === comment.id"
            @click="deleteComment(comment.id)"
          >
            Delete
          </el-button>
          <el-button
            v-if="isLoggedIn"
            text
            class="reply-btn"
            @click="toggleReply(comment.id)"
          >
            Reply
          </el-button>
        </div>
        <div v-if="replyingTo === comment.id" class="comment-box reply-box">
          <el-input
            v-model="replyContent"
            type="textarea"
            :rows="2"
            placeholder="Write a reply..."
            class="comment-input"
          />
          <div class="comment-actions">
            <el-button class="bw-btn" size="small" @click="submitReply">Reply</el-button>
            <el-button class="bw-btn" size="small" @click="replyingTo = null">Cancel</el-button>
          </div>
        </div>
      </div>
    </div>

    <div class="add-comment" v-if="isLoggedIn">
      <el-button v-if="!showCommentBox" text class="add-comment-link" @click="showCommentBox = true">
        Add a comment
      </el-button>
      <div v-else class="comment-box">
        <el-input
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
import { computed, onMounted, ref, watch } from 'vue'
import { apiFetch, getToken } from '@/api'
import { notify } from '@/utils/message'

interface Comment {
  id: number
  username: string
  content: string
  created_at: string
}

const props = defineProps<{
  postId: number
}>()

const comments = ref<Comment[]>([])
const loading = ref(false)

const isLoggedIn = computed(() => !!getToken())
const showCommentBox = ref(false)
const newComment = ref('')
const commenting = ref(false)
const replyingTo = ref<number | null>(null)
const replyContent = ref('')
const deletingId = ref<number | null>(null)
const currentUser = ref<string | null>(null)

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

async function deleteComment(commentId: number) {
  deletingId.value = commentId
  try {
    const response = await apiFetch(`/comments/${commentId}`, { method: 'DELETE' })
    if (!response.ok) throw new Error(`HTTP ${response.status}`)
    notify('success', 'Comment deleted')
    comments.value = comments.value.filter((c) => c.id !== commentId)
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

function toggleReply(commentId: number) {
  replyingTo.value = replyingTo.value === commentId ? null : commentId
  replyContent.value = ''
}

function submitReply() {
  replyingTo.value = null
  replyContent.value = ''
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

async function loadComments() {
  loading.value = true
  try {
    const response = await apiFetch(`/posts/${props.postId}/comments`)
    if (!response.ok) throw new Error(`HTTP ${response.status}`)
    const json = await response.json()
    comments.value = json.data ?? []
  } catch (error) {
    console.error('Load comments error:', error)
    notify('error', 'Failed to load comments')
  } finally {
    loading.value = false
  }
}

onMounted(() => {
  loadCurrentUser()
  loadComments()
})
watch(() => props.postId, () => {
  loadCurrentUser()
  loadComments()
})
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

.comment {
  display: flex;
  align-items: flex-start;
  gap: 10px;
  padding: 10px 0;
  border-bottom: 1px solid #313335;
}

.comment-avatar {
  flex-shrink: 0;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 28px;
  height: 28px;
  border-radius: 50%;
  background: #3d4043;
  color: #ffffff;
  font-size: 13px;
  font-weight: 600;
}

.comment-body {
  flex: 1;
  min-width: 0;
}

.comment-meta {
  display: flex;
  align-items: baseline;
  gap: 8px;
  margin-bottom: 2px;
}

.comment-username {
  font-weight: 600;
  font-size: 13px;
  color: #6cbbf7;
}

.comment-date {
  font-size: 12px;
  color: #6a737c;
}

.comment-content {
  font-size: 13px;
  line-height: 1.5;
  color: #c4c9cc;
  white-space: pre-wrap;
  word-break: break-word;
}

.comment-footer {
  display: flex;
  justify-content: flex-end;
  margin-top: 4px;
}

.reply-btn {
  padding: 0;
  font-size: 12px;
  color: #6a737c;
  background: transparent !important;
}

.reply-btn:hover,
.reply-btn:focus-visible {
  color: #6a737c;
  background: transparent !important;
  text-decoration: underline;
  text-decoration-color: #6a737c;
  text-underline-offset: 4px;
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

.reply-box {
  margin-top: 8px;
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