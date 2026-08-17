<template>
  <div class="thread">
    <div class="comment">
      <div class="comment-avatar">
        <UserAvatar :user-id="comment.user_id" :username="comment.username" :size="30" />
      </div>
      <div class="comment-main">
        <div class="comment-head">
          <span class="comment-author">{{ comment.username }}</span>
          <span class="comment-date">{{ formatDate(comment.created_at) }}</span>
        </div>
        <div class="comment-body">
          <span v-if="comment.reply_to_username" class="reply-to">@{{ comment.reply_to_username }}</span>
          {{ comment.content }}
        </div>
        <div class="comment-actions">
          <el-button
            v-if="isLoggedIn"
            text
            class="action-btn"
            @click="toggleReply(comment.id)"
          >
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
            <el-button class="bw-btn" size="small" :loading="replying" @click="submitReply(comment.id)">Reply</el-button>
            <el-button class="bw-btn" size="small" @click="replyingTo = null">Cancel</el-button>
          </div>
        </div>
      </div>
    </div>

    <div class="reply-tree" v-if="node.children.length">
      <CommentThread
        v-for="child in visibleChildren"
        :key="child.comment.id"
        :node="child"
      />
      <div class="more-replies">
        <button class="more-btn" @click="moreReplies(comment.id, node.children.length)">
          {{ replyLabel }}
        </button>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, inject, nextTick, ref, watch } from 'vue'
import type { ComputedRef, Ref } from 'vue'
import type { ElInput } from 'element-plus'
import UserAvatar from '@/components/UserAvatar.vue'

interface Comment {
  id: number
  user_id?: number
  username: string
  content: string
  created_at: string
  parent_id?: number | null
  reply_to_username?: string
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
  shownFor: (parentId: number) => number
  moreReplies: (parentId: number, total: number) => void
  currentUser: Ref<string | null>
  isLoggedIn: ComputedRef<boolean>
  deletingId: Ref<number | null>
  replyingTo: Ref<number | null>
  replyContent: Ref<string>
  replying: Ref<boolean>
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
  shownFor,
  moreReplies,
  currentUser,
  isLoggedIn,
  deletingId,
  replyingTo,
  replyContent,
  replying,
} = ctx

const comment = computed(() => props.node.comment)
const visibleChildren = computed(() =>
  props.node.children.slice(0, shownFor(comment.value.id)),
)
const replyLabel = computed(() => {
  const total = props.node.children.length
  const cur = shownFor(comment.value.id)
  if (cur <= 0) return total === 1 ? 'Show 1 reply' : `Show ${total} replies`
  if (cur >= total) return 'Hide replies'
  const left = total - cur
  return left === 1 ? 'Show 1 more reply' : `Show ${left} more replies`
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
  background: #343536;
  color: #ffffff;
  font-size: 14px;
  font-weight: 600;
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
  color: #818384;
}

.comment-body {
  font-size: 14px;
  line-height: 1.45;
  color: #d7dadc;
  white-space: pre-wrap;
  word-break: break-word;
  margin-bottom: 6px;
}

.reply-to {
  color: #818384;
}

.comment-actions {
  display: flex;
  justify-content: flex-end;
  gap: 14px;
}

.action-btn {
  padding: 0;
  font-size: 12px;
  color: #818384;
  background: transparent !important;
}

.action-btn:hover,
.action-btn:focus-visible {
  color: #6cbbf7;
  background: transparent !important;
  text-decoration: none;
}

.comment-box {
  border: 1px solid #3d4043;
  border-radius: 8px;
  padding: 12px;
  background: #0a0a0a;
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

.input-actions {
  display: flex;
  justify-content: flex-end;
  gap: 8px;
  margin-top: 10px;
}

.reply-tree {
  border-left: 2px solid #343536;
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
  color: #818384;
  cursor: pointer;
  text-decoration: underline;
  text-decoration-color: #595959;
  text-underline-offset: 3px;
}

.more-btn:hover {
  color: #6cbbf7;
  text-decoration-color: #6cbbf7;
}
</style>