<template>
  <div class="results" v-loading="loading" element-loading-background="rgba(20, 20, 20, 0.6)">
    <PostsList v-if="results.length" :posts="results" :loading="loading" />

    <div v-if="results.length" class="pagination">
      <el-button
        class="page-btn"
        :disabled="!hasPrev"
        @click="goToPage(current - 1)"
      >
        Previous
      </el-button>
      <span class="page-indicator">Page {{ current }}</span>
      <el-button
        class="page-btn"
        :disabled="!hasNext"
        @click="goToPage(current + 1)"
      >
        Next
      </el-button>
    </div>

    <div v-else-if="!loading" class="empty">
      <p>No posts match your search.</p>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, nextTick, onActivated, onBeforeUnmount, ref, watch } from 'vue'
import { onBeforeRouteLeave, useRoute, useRouter } from 'vue-router'
import {
  useFeedStore,
  type FeedPost,
  type SearchParams,
  SEARCH_PAGE_SIZE,
} from '@/stores/feed'
import { notify } from '@/utils/message'
import PostsList from '@/components/PostsList.vue'

defineOptions({ name: 'SearchResultsList' })

const props = defineProps<{
  params: SearchParams
}>()

const store = useFeedStore()
const route = useRoute()
const router = useRouter()
const results = ref<FeedPost[]>([])
const loading = ref(false)

const current = computed(() => props.params.page ?? 1)
const hasPrev = computed(() => current.value > 1)
const hasNext = computed(() => results.value.length === SEARCH_PAGE_SIZE)

function goToPage(page: number) {
  if (page < 1) return
  const query = { ...route.query }
  if (page === 1) delete query.page
  else query.page = String(page)
  router.push({ path: '/search', query })
}

function saveScroll() {
  store.searchScrollTop = window.scrollY
}

function restoreScroll() {
  const top = store.searchScrollTop
  if (top > 0) {
    nextTick(() => window.scrollTo({ top }))
  }
}

let requestId = 0

watch(
  () => props.params,
  async (params) => {
    const id = ++requestId
    loading.value = true
    try {
      const posts = await store.fetchSearch(params)
      if (id !== requestId) return
      results.value = posts
    } catch {
      if (id !== requestId) return
      notify('error', 'Search failed')
    } finally {
      if (id === requestId) loading.value = false
    }
  },
  { immediate: true },
)

onBeforeRouteLeave(saveScroll)
onBeforeUnmount(saveScroll)
onActivated(restoreScroll)
</script>

<style scoped>
.results {
  min-height: 200px;
}

.pagination {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 16px;
  margin-top: 24px;
}

.page-btn {
  min-width: 110px;
  background: #ffffff;
  color: #141414;
  border: 1px solid #ffffff;
  font-weight: 600;
}

.page-btn:hover,
.page-btn:focus,
.page-btn:focus-visible {
  background: #e4e6e8;
  color: #141414;
  border-color: #e4e6e8;
}

.page-btn.is-disabled,
.page-btn.is-disabled:hover,
.page-btn.is-disabled:focus {
  background: #1f1f1f;
  color: #8c8c8c;
  border-color: #262626;
}

.page-indicator {
  font-size: 14px;
  color: #8c8c8c;
}

.empty {
  padding: 80px 24px;
  border: 1px dashed #262626;
  border-radius: 12px;
  text-align: center;
  font-size: 16px;
  color: #8c8c8c;
}
</style>