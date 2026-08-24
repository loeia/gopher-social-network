<template>
  <div class="my-posts-page">
    <div v-if="editingPost" class="back-nav">
      <el-button text @click="goBack">← Back</el-button>
    </div>

    <div v-if="editingPost" class="page-header">
      <h1 class="page-title">Edit Post</h1>
    </div>

    <div v-if="editingPost" class="form">
      <div class="form-group">
        <div class="label-row">
          <label class="form-label" for="title">Title</label>
          <span class="char-count" :class="{ over: isTitleOver }">{{ title.length }}/100</span>
        </div>
        <el-input
          id="title"
          v-model="title"
          size="large"
          class="field"
          :class="{ 'over-limit': isTitleOver }"
          placeholder="Be specific and imagine you're asking a question to another person"
        />
        <p v-if="isTitleOver" class="form-error">Title cannot exceed 100 characters</p>
      </div>

      <div class="form-group">
        <div class="label-row">
          <label class="form-label" for="body">Content</label>
          <span class="char-count" :class="{ over: isContentOver }">{{ content.length }}/1000</span>
        </div>
        <el-tabs v-model="contentTab" class="content-tabs">
          <el-tab-pane label="Edit" name="edit">
            <el-input
              id="body"
              v-model="content"
              type="textarea"
              :rows="14"
              resize="none"
              class="body-field"
              :class="{ 'over-limit': isContentOver }"
              placeholder="Include all the information someone would need to answer your question"
            />
          </el-tab-pane>
          <el-tab-pane label="Preview" name="preview">
            <div class="preview markdown-body" v-html="renderedContent"></div>
          </el-tab-pane>
        </el-tabs>
        <p v-if="isContentOver" class="form-error">Content cannot exceed 1000 characters</p>
      </div>

      <div class="form-group">
        <div class="label-row">
          <label class="form-label" for="tags">Tags</label>
          <span class="char-count" :class="{ over: isTagsOver }">{{ tags.length }}/5</span>
        </div>
        <div class="tags-input-row">
          <el-input
            id="tags"
            v-model="tagInput"
            size="large"
            class="field"
            :class="{ 'over-limit': isTagInputOver }"
            placeholder="e.g. golang, postgres"
            @keyup.enter.prevent="addTag"
          />
          <el-button size="large" class="add-btn" :disabled="!canAddTag" @click="addTag">
            Add
          </el-button>
        </div>
        <div class="tags-list" v-if="tags.length">
          <el-tag
            v-for="(tag, index) in tags"
            :key="`${tag}-${index}`"
            closable
            @close="removeTag(index)"
          >
            {{ tag }}
          </el-tag>
        </div>
        <p v-if="isTagsOver" class="form-error">You can add at most 5 tags</p>
        <p v-if="isTagInputOver" class="form-error">Tag cannot exceed 10 characters</p>
      </div>

      <div class="form-actions">
        <el-button size="large" class="cancel-btn" @click="cancelEdit">Cancel</el-button>
        <el-button
          size="large"
          class="publish-btn"
          :loading="submitting"
          :disabled="!hasChanges"
          @click="handleEdit"
        >
          Save Changes
        </el-button>
      </div>
    </div>

    <PostsList
      v-else
      :posts="posts"
      :loading="loading"
      editable
      @edit="startEdit"
      @delete="handleDelete"
    />
  </div>
</template>

<script setup lang="ts">
import { computed, nextTick, onActivated, onBeforeUnmount, onMounted, ref } from 'vue'
import { onBeforeRouteLeave, useRouter } from 'vue-router'
import { ElMessageBox } from 'element-plus'
import PostsList from '@/components/PostsList.vue'
import { apiFetch, getCurrentUserId } from '@/api'
import { useFeedStore, toFeedPost, type FeedPost } from '@/stores/feed'
import { renderMarkdown } from '@/utils/markdown'
import { notify } from '@/utils/message'

defineOptions({ name: 'MyPostsPage' })

const router = useRouter()
const feedStore = useFeedStore()
const posts = ref<FeedPost[]>([])
const loading = ref(false)
const currentUserId = getCurrentUserId()

const editingPost = ref<FeedPost | null>(null)
const title = ref('')
const content = ref('')
const contentTab = ref('edit')
const tagInput = ref('')
const tags = ref<string[]>([])
const submitting = ref(false)

const MAX_TITLE = 100
const MAX_CONTENT = 1000
const MAX_TAGS = 5
const MAX_TAG_LENGTH = 10

const isTitleOver = computed(() => title.value.length > MAX_TITLE)
const isTagsOver = computed(() => tags.value.length > MAX_TAGS)
const isTagInputOver = computed(() => tagInput.value.trim().length > MAX_TAG_LENGTH)
const canAddTag = computed(
  () => !!tagInput.value.trim() && tags.value.length < MAX_TAGS && !isTagInputOver.value,
)
const renderedContent = computed(() => renderMarkdown(content.value))

const originalTitle = ref('')
const originalContent = ref('')
const originalTags = ref<string[]>([])

const titleChanged = computed(() => title.value.trim() !== originalTitle.value.trim())
const contentChanged = computed(() => content.value !== originalContent.value)
const tagsChanged = computed(() => {
  if (tags.value.length !== originalTags.value.length) return true
  return tags.value.some((tag, index) => tag !== originalTags.value[index])
})
const hasChanges = computed(() => titleChanged.value || contentChanged.value || tagsChanged.value)
const isContentOver = computed(() => contentChanged.value && content.value.length > MAX_CONTENT)

function saveScroll() {
  feedStore.myPostsScrollTop = window.scrollY
}

function restoreScroll() {
  const top = feedStore.myPostsScrollTop
  nextTick(() => window.scrollTo({ top }))
}

onMounted(loadMyPosts)
onBeforeRouteLeave(saveScroll)
onBeforeUnmount(saveScroll)
onActivated(() => {
  restoreScroll()
  loadMyPosts()
})

function goBack() {
  if (editingPost.value) {
    cancelEdit()
    return
  }
  router.back()
}

async function loadMyPosts() {
  loading.value = true
  try {
    const response = await apiFetch('/users/posts')
    if (!response.ok) throw new Error(`HTTP ${response.status}`)
    const json = await response.json()
    const raw = Array.isArray(json) ? json : (json.data ?? [])
    posts.value = raw.map((p: any) => ({
      ...toFeedPost(p),
      comment_count: Number(p.comment_count ?? p.comments_count ?? 0),
      like_count: Number(p.like_count ?? p.likes_count ?? 0),
      user_id: currentUserId ?? undefined,
    }))
  } catch (error) {
    console.error('Load my posts error:', error)
    notify('error', 'Failed to load my posts')
  } finally {
    loading.value = false
  }
}

function startEdit(post: FeedPost) {
  editingPost.value = post
  title.value = post.title
  content.value = ''
  contentTab.value = 'edit'
  tagInput.value = ''
  tags.value = [...(post.tags ?? [])]
  originalTitle.value = post.title
  originalContent.value = ''
  originalTags.value = [...(post.tags ?? [])]
  loadPostForEdit(post.id)
}

async function loadPostForEdit(postId: number) {
  try {
    const response = await apiFetch(`/posts/${postId}`)
    if (!response.ok) throw new Error(`HTTP ${response.status}`)
    const json = await response.json()
    const data = json.data ?? json
    title.value = data.title ?? title.value
    content.value = data.content ?? ''
    tags.value = Array.isArray(data.tags) ? data.tags : tags.value
    originalTitle.value = title.value
    originalContent.value = content.value
    originalTags.value = [...tags.value]
  } catch (error) {
    console.error('Load post for edit error:', error)
    notify('error', 'Failed to load post content')
  }
}

function cancelEdit() {
  editingPost.value = null
}

function addTag() {
  const tag = tagInput.value.trim()
  if (!tag) return
  if (tag.length > MAX_TAG_LENGTH) return
  if (tags.value.includes(tag)) {
    tagInput.value = ''
    return
  }
  if (tags.value.length >= MAX_TAGS) {
    notify('warning', `You can add at most ${MAX_TAGS} tags`)
    return
  }
  tags.value.push(tag)
  tagInput.value = ''
}

function removeTag(index: number) {
  tags.value.splice(index, 1)
}

async function handleEdit() {
  const post = editingPost.value
  if (!post) return
  if (!hasChanges.value) {
    notify('info', 'No changes to save')
    return
  }
  if (isTitleOver.value || isContentOver.value || isTagsOver.value) {
    notify('warning', 'Please fix the fields exceeding the limit')
    return
  }
  if (titleChanged.value && !title.value.trim()) {
    notify('warning', 'Please enter a title')
    return
  }
  if (contentChanged.value && !content.value.trim()) {
    notify('warning', 'Please enter content')
    return
  }

  const payload: Record<string, unknown> = {}
  if (titleChanged.value) payload.title = title.value.trim()
  if (contentChanged.value) payload.content = content.value.trim()
  if (tagsChanged.value) payload.tags = tags.value

  submitting.value = true
  try {
    const response = await apiFetch(`/posts/${post.id}`, {
      method: 'PATCH',
      body: JSON.stringify(payload),
    })
    if (!response.ok) throw new Error(`HTTP ${response.status}`)
    notify('success', 'Post updated')
    cancelEdit()
    await loadMyPosts()
  } catch (error) {
    console.error('Update post error:', error)
    notify('error', 'Failed to update post')
  } finally {
    submitting.value = false
  }
}

async function handleDelete(post: FeedPost) {
  try {
    await ElMessageBox.confirm('This action cannot be undone. Delete this post?', 'Delete post', {
      confirmButtonText: 'Delete',
      cancelButtonText: 'Cancel',
      type: 'warning',
      customClass: 'bw-messagebox',
    })
  } catch {
    return
  }
  try {
    const response = await apiFetch(`/posts/${post.id}`, {
      method: 'DELETE',
    })
    if (!response.ok) throw new Error(`HTTP ${response.status}`)
    notify('success', 'Post deleted')
    if (editingPost.value?.id === post.id) cancelEdit()
    await loadMyPosts()
  } catch (error) {
    console.error('Delete post error:', error)
    notify('error', 'Failed to delete post')
  }
}
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

.form {
  margin: 0 20%;
  display: flex;
  flex-direction: column;
  gap: 40px;
}

.form-group {
  display: flex;
  flex-direction: column;
  gap: 10px;
}

.label-row {
  display: flex;
  align-items: baseline;
  justify-content: space-between;
  gap: 16px;
}

.form-label {
  font-size: 16px;
  font-weight: 600;
  color: #ffffff;
}

.char-count {
  font-size: 13px;
  color: #8c8c8c;
}

.char-count.over {
  color: #f56c6c;
  font-weight: 600;
}

.field :deep(.el-input__wrapper) {
  background: transparent;
  box-shadow: 0 0 0 1px #262626 inset;
}

.field :deep(.el-input__inner) {
  color: #ffffff;
}

.field :deep(.el-input__inner::placeholder) {
  color: #595959;
}

.field.over-limit :deep(.el-input__wrapper) {
  box-shadow: 0 0 0 1px #f56c6c inset;
}

.body-field :deep(.el-textarea__inner) {
  background: transparent;
  color: #ffffff;
  box-shadow: 0 0 0 1px #262626 inset;
  border-radius: 4px;
  font-size: 15px;
  line-height: 1.7;
}

.body-field :deep(.el-textarea__inner::placeholder) {
  color: #595959;
}

.body-field.over-limit :deep(.el-textarea__inner) {
  box-shadow: 0 0 0 1px #f56c6c inset;
}

.content-tabs :deep(.el-tabs__item) {
  color: #8c8c8c;
}

.content-tabs :deep(.el-tabs__item.is-active) {
  color: #ffffff;
}

.content-tabs :deep(.el-tabs__active-bar) {
  background-color: #ffffff;
}

.content-tabs :deep(.el-tabs__nav-wrap::after) {
  background-color: #262626;
}

.preview {
  min-height: 360px;
  padding: 16px;
  border: 1px solid #262626;
  border-radius: 4px;
  background: #141414;
  font-size: 15px;
  line-height: 1.7;
  color: #bfbfbf;
  word-break: break-word;
  overflow-wrap: anywhere;
}

.preview :deep(p) {
  margin: 0 0 16px;
}

.preview :deep(h1),
.preview :deep(h2),
.preview :deep(h3),
.preview :deep(h4),
.preview :deep(h5),
.preview :deep(h6) {
  margin: 24px 0 12px;
  font-weight: 600;
  line-height: 1.3;
  color: #ffffff;
}

.preview :deep(h1) {
  font-size: 24px;
}

.preview :deep(h2) {
  font-size: 21px;
}

.preview :deep(h3) {
  font-size: 18px;
}

.preview :deep(a) {
  color: #6cbbf7;
  text-decoration: none;
}

.preview :deep(a:hover) {
  text-decoration: underline;
}

.preview :deep(ul),
.preview :deep(ol) {
  margin: 0 0 16px;
  padding-left: 28px;
}

.preview :deep(li) {
  margin: 4px 0;
}

.preview :deep(pre) {
  margin: 0 0 16px;
  padding: 12px 16px;
  overflow-x: auto;
  background: #232629;
  border-radius: 4px;
  font-size: 13px;
  line-height: 1.5;
}

.preview :deep(code) {
  font-family: ui-monospace, 'SF Mono', Menlo, Consolas, 'Liberation Mono', monospace;
  font-size: 13px;
}

.preview :deep(:not(pre) > code) {
  padding: 2px 6px;
  background: #232629;
  border-radius: 4px;
  color: #e4e6e8;
}

.preview :deep(pre code) {
  padding: 0;
  background: none;
  color: #e4e6e8;
}

.preview :deep(blockquote) {
  margin: 0 0 16px;
  padding: 4px 16px;
  border-left: 4px solid #3d4043;
  color: #b2b6b9;
}

.preview :deep(img) {
  max-width: 100%;
  border-radius: 4px;
}

.preview :deep(table) {
  margin: 0 0 16px;
  border-collapse: collapse;
  width: 100%;
}

.preview :deep(th),
.preview :deep(td) {
  padding: 8px 12px;
  border: 1px solid #3d4043;
}

.preview :deep(th) {
  background: #232629;
  color: #ffffff;
}

.preview :deep(hr) {
  margin: 24px 0;
  border: none;
  border-top: 1px solid #3d4043;
}

.tags-input-row {
  display: flex;
  gap: 12px;
}

.tags-input-row .field {
  flex: 1;
}

.tags-list {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
}

.tags-list :deep(.el-tag) {
  background: #ffffff;
  color: #141414;
  border: 1px solid #ffffff;
}

.tags-list :deep(.el-tag .el-tag__close) {
  color: #595959;
}

.tags-list :deep(.el-tag .el-tag__close:hover) {
  background: #141414;
  color: #ffffff;
}

.form-error {
  margin: 0;
  font-size: 13px;
  color: #f56c6c;
}

.add-btn {
  min-width: 88px;
  background: #ffffff;
  color: #141414;
  border: 1px solid #ffffff;
}

.add-btn:hover {
  background: #e4e6e8;
  color: #141414;
}

.add-btn.is-disabled,
.add-btn.is-disabled:hover {
  background: #262626;
  border-color: #262626;
  color: #595959;
}

.form-actions {
  display: flex;
  justify-content: flex-end;
  gap: 12px;
}

.cancel-btn {
  background: transparent;
  color: #8c8c8c;
  border: 1px solid #262626;
}

.cancel-btn:hover {
  background: #262626;
  color: #ffffff;
  border-color: #262626;
}

.publish-btn {
  background: #ffffff;
  color: #141414;
  border: 1px solid #ffffff;
  font-weight: 600;
}

.publish-btn:hover {
  background: #e4e6e8;
  color: #141414;
}

:global(.bw-messagebox) {
  background: #141414;
  border: 1px solid #262626;
  border-radius: 8px;
  padding: 20px;
}

:global(.bw-messagebox .el-message-box__header) {
  padding: 0 0 12px;
}

:global(.bw-messagebox .el-message-box__title) {
  color: #ffffff;
  font-weight: 600;
}

:global(.bw-messagebox .el-message-box__content) {
  padding: 0 0 20px;
}

:global(.bw-messagebox .el-message-box__message p) {
  color: #8c8c8c;
}

:global(.bw-messagebox .el-message-box__btns) {
  display: flex;
  justify-content: flex-end;
  gap: 12px;
}

:global(.bw-messagebox .el-button--primary) {
  background: #ffffff;
  border-color: #ffffff;
  color: #141414;
  font-weight: 600;
}

:global(.bw-messagebox .el-button--primary:hover) {
  background: #e4e6e8;
  border-color: #e4e6e8;
  color: #141414;
}

:global(.bw-messagebox .el-button:not(.el-button--primary)) {
  background: transparent;
  border-color: #262626;
  color: #8c8c8c;
}

:global(.bw-messagebox .el-button:not(.el-button--primary):hover) {
  background: #262626;
  border-color: #262626;
  color: #ffffff;
}
</style>
