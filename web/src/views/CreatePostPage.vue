<template>
  <div class="create-page">
    <div class="create-container">
      <h1 class="page-title">Create Post</h1>

      <div class="form">
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
            <span class="char-count" :class="{ over: isContentOver }"
              >{{ content.length }}/5000</span
            >
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
          <p v-if="isContentOver" class="form-error">Content cannot exceed 5000 characters</p>
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
          <el-button size="large" class="publish-btn" :loading="submitting" @click="handleSubmit">
            Publish
          </el-button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, ref } from 'vue'
import { useRouter } from 'vue-router'
import { apiFetch, handleApiError } from '@/api'
import { notify } from '@/utils/message'
import { renderMarkdown } from '@/utils/markdown'

const router = useRouter()

const title = ref('')
const content = ref('')
const contentTab = ref('edit')
const tagInput = ref('')
const tags = ref<string[]>([])
const submitting = ref(false)

const MAX_TITLE = 100
const MAX_CONTENT = 5000
const MAX_TAGS = 5
const MAX_TAG_LENGTH = 10

const isTitleOver = computed(() => title.value.length > MAX_TITLE)
const isContentOver = computed(() => content.value.length > MAX_CONTENT)
const isTagsOver = computed(() => tags.value.length > MAX_TAGS)
const isTagInputOver = computed(() => tagInput.value.trim().length > MAX_TAG_LENGTH)
const canAddTag = computed(
  () => !!tagInput.value.trim() && tags.value.length < MAX_TAGS && !isTagInputOver.value,
)
const renderedContent = computed(() => renderMarkdown(content.value))

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

async function handleSubmit() {
  if (isTitleOver.value || isContentOver.value || isTagsOver.value) {
    notify('warning', 'Please fix the fields exceeding the limit')
    return
  }
  if (!title.value.trim()) {
    notify('warning', 'Please enter a title')
    return
  }
  if (!content.value.trim()) {
    notify('warning', 'Please enter content')
    return
  }

  submitting.value = true
  try {
    const response = await apiFetch('/posts', {
      method: 'POST',
      body: JSON.stringify({
        title: title.value.trim(),
        content: content.value.trim(),
        tags: tags.value,
      }),
    })
    if (!response.ok) throw new Error(`HTTP ${response.status}`)
    const json = await response.json()
    const data = json.data ?? json
    const postId = Number(data?.id)
    notify('success', 'Post created')
    if (postId > 0) {
      router.push(`/posts/${postId}`)
    } else {
      router.push('/')
    }
  } catch (error) {
    handleApiError(error, 'Failed to create post')
  } finally {
    submitting.value = false
  }
}
</script>

<style scoped>
.create-page {
  min-height: 100vh;
  padding: 48px 0 96px;
}

.create-container {
  max-width: 860px;
  margin: 0 auto;
  padding: 0 32px;
}

.page-title {
  margin: 0 0 40px;
  font-size: 30px;
  font-weight: 600;
  color: #e4e6e8;
}

.form {
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
  color: #e4e6e8;
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
  box-shadow: 0 0 0 1px #333 inset;
}

.field :deep(.el-input__inner) {
  color: #e4e6e8;
}

.field :deep(.el-input__inner::placeholder) {
  color: #8c8c8c;
}

.field.over-limit :deep(.el-input__wrapper) {
  box-shadow: 0 0 0 1px #f56c6c inset;
}

.body-field :deep(.el-textarea__inner) {
  background: transparent;
  color: #e4e6e8;
  box-shadow: 0 0 0 1px #333 inset;
  border-radius: 4px;
  font-size: 15px;
  line-height: 1.7;
}

.body-field :deep(.el-textarea__inner::placeholder) {
  color: #8c8c8c;
}

.body-field.over-limit :deep(.el-textarea__inner) {
  box-shadow: 0 0 0 1px #f56c6c inset;
}

.content-tabs :deep(.el-tabs__item) {
  color: #8c8c8c;
}

.content-tabs :deep(.el-tabs__item.is-active) {
  color: #e4e6e8;
}

.content-tabs :deep(.el-tabs__active-bar) {
  background-color: #e4e6e8;
}

.content-tabs :deep(.el-tabs__nav-wrap::after) {
  background-color: #333;
}

.preview {
  min-height: 360px;
  padding: 16px;
  border: 1px solid #333;
  border-radius: 4px;
  background: #141414;
  font-size: 15px;
  line-height: 1.7;
  color: #8c8c8c;
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
  color: #e4e6e8;
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

.preview :deep(h4) {
  font-size: 16px;
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
  border-left: 4px solid #333;
  color: #8c8c8c;
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
  border: 1px solid #333;
}

.preview :deep(th) {
  background: #232629;
  color: #e4e6e8;
}

.preview :deep(hr) {
  margin: 24px 0;
  border: none;
  border-top: 1px solid #333;
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
  background: #1f1f1f;
  color: #e4e6e8;
  border: 1px solid #333;
}

.tags-list :deep(.el-tag .el-tag__close) {
  color: #8c8c8c;
}

.tags-list :deep(.el-tag .el-tag__close:hover) {
  background: #333;
  color: #ffffff;
}

.form-error {
  margin: 0;
  font-size: 13px;
  color: #f56c6c;
}

.add-btn {
  min-width: 88px;
  background: #262626;
  color: #e4e6e8;
  border: 1px solid #333;
}

.add-btn:hover {
  background: #333;
  color: #e4e6e8;
}

.add-btn.is-disabled,
.add-btn.is-disabled:hover {
  background: #1f1f1f;
  border-color: #262626;
  color: #595959;
}

.form-actions {
  display: flex;
  justify-content: flex-end;
}

.publish-btn {
  background: #e4e6e8;
  color: #1a1a1a;
  border: 1px solid #e4e6e8;
  font-weight: 600;
}

.publish-btn:hover {
  background: #ffffff;
  color: #1a1a1a;
}
</style>
