<template>
  <div class="detail-page">
    <div class="back-nav">
      <el-button text @click="backHome">← Back</el-button>
      <div class="nav-arrows">
        <el-button text :disabled="!hasPrevPost" @click="goToPrevious">&lt;</el-button>
        <el-button text :disabled="!hasNextPost" @click="goToNext">&gt;</el-button>
      </div>
    </div>

    <div class="detail-container" v-loading="loading">
      <template v-if="post">
        <h1 class="detail-title">{{ post.title }}</h1>

        <div class="detail-tags" v-if="post.tags && post.tags.length">
          <span v-for="tag in post.tags" :key="tag" class="tag">{{ tag }}</span>
        </div>

        <div class="meta">
          <span class="meta-item">{{ post.author }}</span>
          <span class="meta-sep">·</span>
          <span class="meta-item">{{ formatDate(post.created_at) }}</span>
          <span class="meta-sep">·</span>
          <span class="meta-item">{{ post.likes }} likes</span>
        </div>

        <div class="markdown-body" v-html="renderedContent" />

        <div class="comments" v-if="post.comments && post.comments.length">
          <h3 class="comments-title">Comments ({{ post.comments.length }})</h3>
          <div v-for="comment in post.comments" :key="comment.id" class="comment">
            <div class="comment-avatar">{{ (comment.username || 'G').charAt(0).toUpperCase() }}</div>
            <div class="comment-body">
              <div class="comment-meta">
                <span class="comment-username">{{ comment.username }}</span>
                <span class="comment-date">{{ formatDate(comment.created_at) }}</span>
              </div>
              <div class="comment-content">{{ comment.content }}</div>
              <div class="comment-footer">
                <el-button v-if="isLoggedIn" text class="reply-btn" @click="toggleReply(comment.id)">
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
      </template>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, ref, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import { apiFetch, getToken } from '@/api'
import { useFeedStore } from '@/stores/feed'
import { renderMarkdown } from '@/utils/markdown'

interface Comment {
  id: number
  username: string
  content: string
  created_at: string
}

interface PostDetail {
  id: number
  author_id: number
  author: string
  title: string
  content: string
  tags: string[]
  comments: Comment[]
  likes: number
  created_at: string
  updated_at: string
}

const route = useRoute()
const router = useRouter()

const store = useFeedStore()

const post = ref<PostDetail | null>(null)
const loading = ref(false)

const isLoggedIn = computed(() => !!getToken())
const showCommentBox = ref(false)
const newComment = ref('')
const commenting = ref(false)
const replyingTo = ref<number | null>(null)
const replyContent = ref('')

const hasPrevPost = computed(() => store.postHistoryIndex > 0)
const hasNextPost = computed(() => store.postHistoryIndex < store.postHistory.length - 1)
const renderedContent = computed(() => {
  if (!post.value) return ''
  return renderMarkdown(post.value.content)
})

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
    ElMessage.warning('Please enter a comment')
    return
  }
  commenting.value = true
  try {
    const response = await apiFetch(`/posts/${route.params.postId}/comments`, {
      method: 'POST',
      body: JSON.stringify({ content }),
    })
    if (!response.ok) throw new Error(`HTTP ${response.status}`)
    newComment.value = ''
    showCommentBox.value = false
    ElMessage.success('Comment added')
    await startLoad(Number(route.params.postId))
  } catch (error) {
    console.error('Add comment error:', error)
    ElMessage.error('Failed to add comment')
  } finally {
    commenting.value = false
  }
}

function backHome() {
  store.clearPostHistory()
  router.push('/')
}

function goToPrevious() {
  const id = store.goBackPost()
  if (id !== null) router.push(`/posts/${id}`)
}

function goToNext() {
  const id = store.goForwardPost()
  if (id !== null) router.push(`/posts/${id}`)
}

function formatDate(value?: string) {
  if (!value) return ''
  const date = new Date(value)
  if (Number.isNaN(date.getTime())) return value
  return date.toLocaleString()
}

function loadPost() {
  const id = Number(route.params.postId)
  store.visitPost(id)
  loading.value = true
  post.value = null
  startLoad(id)
}

async function startLoad(id: number) {
  try {
    const response = await apiFetch(`/posts/${id}`)
    if (!response.ok) throw new Error(`HTTP ${response.status}`)
    const json = await response.json()
    post.value = json.data ?? json
  } catch (error) {
    console.error('Load post error:', error)
    ElMessage.error('Failed to load post')
  } finally {
    loading.value = false
  }
}

onMounted(loadPost)
watch(() => route.params.postId, loadPost)
</script>

<style scoped>
.detail-page {
  min-height: 100vh;
  padding: 32px 0 80px;
  font-family:
    -apple-system,
    BlinkMacSystemFont,
    'Segoe UI',
    'Liberation Sans',
    'Helvetica Neue',
    Arial,
    sans-serif;
}

.back-nav {
  width: 75%;
  margin: 0 auto 16px;
  padding: 0 24px;
  display: flex;
  align-items: center;
  gap: 12px;
}

.nav-arrows {
  margin-left: auto;
  display: flex;
  gap: 4px;
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

.nav-arrows :deep(.el-button:hover),
.nav-arrows :deep(.el-button:focus),
.nav-arrows :deep(.el-button:focus-visible) {
  color: #6a737c;
  background: transparent;
  font-weight: 600;
  text-decoration: none;
}

.back-nav :deep(.el-button.is-disabled) {
  color: #3d4043;
  background: transparent;
  text-decoration: none;
  cursor: not-allowed;
}

.detail-container {
  width: 75%;
  margin: 0 auto;
  padding: 0 24px;
}

.detail-title {
  margin: 0 0 16px;
  font-size: 27px;
  font-weight: 600;
  line-height: 1.2;
  color: #ffffff;
}

.meta {
  display: flex;
  align-items: center;
  flex-wrap: wrap;
  gap: 8px;
  padding-bottom: 16px;
  border-bottom: 1px solid #3d4043;
  font-size: 13px;
  color: #9fa6ad;
}

.meta-sep {
  color: #6a737c;
}

/* Markdown-rendered content, StackOverflow-style */
.markdown-body {
  margin: 24px 0;
  font-size: 15px;
  line-height: 1.5;
  color: #e4e6e8;
  word-break: break-word;
}

.markdown-body :deep(p) {
  margin: 0 0 16px;
}

.markdown-body :deep(h1),
.markdown-body :deep(h2),
.markdown-body :deep(h3),
.markdown-body :deep(h4),
.markdown-body :deep(h5),
.markdown-body :deep(h6) {
  margin: 28px 0 12px;
  font-weight: 600;
  line-height: 1.3;
  color: #ffffff;
}

.markdown-body :deep(h1) {
  font-size: 24px;
}

.markdown-body :deep(h2) {
  font-size: 21px;
}

.markdown-body :deep(h3) {
  font-size: 18px;
}

.markdown-body :deep(h4) {
  font-size: 16px;
}

.markdown-body :deep(h5) {
  font-size: 15px;
}

.markdown-body :deep(h6) {
  font-size: 14px;
}

.markdown-body :deep(a) {
  color: #6cbbf7;
  text-decoration: none;
}

.markdown-body :deep(a:hover) {
  text-decoration: underline;
}

.markdown-body :deep(ul),
.markdown-body :deep(ol) {
  margin: 0 0 16px;
  padding-left: 28px;
}

.markdown-body :deep(li) {
  margin: 4px 0;
}

.markdown-body :deep(pre) {
  margin: 0 0 16px;
  padding: 12px 16px;
  overflow-x: auto;
  background: #232629;
  border-radius: 4px;
  font-size: 13px;
  line-height: 1.5;
}

.markdown-body :deep(code) {
  font-family: ui-monospace, 'SF Mono', Menlo, Consolas, 'Liberation Mono', monospace;
  font-size: 13px;
}

.markdown-body :deep(:not(pre) > code) {
  padding: 2px 6px;
  background: #232629;
  border-radius: 4px;
  color: #e4e6e8;
}

.markdown-body :deep(pre code) {
  padding: 0;
  background: none;
  color: #e4e6e8;
}

.markdown-body :deep(blockquote) {
  margin: 0 0 16px;
  padding: 4px 16px;
  border-left: 4px solid #3d4043;
  color: #b2b6b9;
}

.markdown-body :deep(img) {
  max-width: 100%;
  border-radius: 4px;
}

.markdown-body :deep(table) {
  margin: 0 0 16px;
  border-collapse: collapse;
  width: 100%;
}

.markdown-body :deep(th),
.markdown-body :deep(td) {
  padding: 8px 12px;
  border: 1px solid #3d4043;
}

.markdown-body :deep(th) {
  background: #232629;
  color: #ffffff;
}

.markdown-body :deep(hr) {
  margin: 24px 0;
  border: none;
  border-top: 1px solid #3d4043;
}

.detail-tags {
  display: flex;
  flex-wrap: wrap;
  gap: 6px;
  margin-bottom: 16px;
}

.tag {
  padding: 4px 6px;
  border-radius: 4px;
  background: #26324a;
  color: #9faccc;
  font-size: 12px;
}

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