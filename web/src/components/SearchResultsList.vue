<template>
  <div class="results" v-loading="loading"   element-loading-background="rgba(20, 20, 20, 0.8)">
    <PostsList
      v-if="results.length"
      :posts="results"
      :loading="loading"
      :highlight-first="highlightFirst"
    />
    <div v-if="loadingMore" class="loading-more">Loading...</div>

    <div v-if="!loading && !hasSearchParams" class="empty">
      <p>Enter a search keyword, author, or tag to find posts.</p>
    </div>

    <div v-else-if="!loading && hasSearchParams && results.length === 0" class="empty">
      <p>No posts match your search.</p>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, onActivated, onBeforeUnmount, onDeactivated, onMounted, ref, watch } from 'vue'
import { useFeedStore, type FeedPost, type SearchParams, SEARCH_PAGE_SIZE } from '@/stores/feed'
import { handleApiError } from '@/api'
import PostsList from '@/components/PostsList.vue'

defineOptions({ name: 'SearchResultsList' })

const props = defineProps<{
  params: SearchParams
}>()

const store = useFeedStore()
const results = ref<FeedPost[]>([])
const loading = ref(false)
const loadingMore = ref(false)
const currentPage = ref(1)
const hasMore = ref(true)
const highlightFirst = ref(false)

let requestId = 0
let rafId = 0
let deactivated = false
let lastFetchedKey = ''

async function fetchPage(page: number, append: boolean) {
  const id = ++requestId
  if (append) {
    loadingMore.value = true
  } else {
    loading.value = results.value.length === 0
  }
  try {
    const posts = await store.fetchSearch({ ...props.params, page })
    if (id !== requestId) return
    if (append) {
      results.value = [...results.value, ...posts]
    } else {
      results.value = posts
      highlightFirst.value = true
      setTimeout(() => {
        highlightFirst.value = false
      }, 2500)
    }
    hasMore.value = posts.length === SEARCH_PAGE_SIZE
  } catch (error) {
    if (id !== requestId) return
    handleApiError(error, 'Search failed')
  } finally {
    if (id === requestId) {
      loading.value = false
      loadingMore.value = false
    }
  }
}

function loadMore() {
  if (loadingMore.value || !hasMore.value) return
  currentPage.value++
  fetchPage(currentPage.value, true)
}

function handleScroll() {
  if (rafId) return
  rafId = requestAnimationFrame(() => {
    rafId = 0
    const scrollHeight = document.documentElement.scrollHeight
    const scrollTop = window.scrollY
    const clientHeight = window.innerHeight
    if (scrollTop + clientHeight >= scrollHeight - 200) {
      loadMore()
    }
  })
}

const paramsKey = computed(() => JSON.stringify(props.params))

const hasSearchParams = computed(() => {
  const p = props.params
  return !!(p.search?.trim() || p.author?.trim() || (p.tags && p.tags.length) || p.since || p.until)
})

watch(
  paramsKey,
  (newKey) => {
    if (deactivated) return
    if (!hasSearchParams.value) return
    if (newKey === lastFetchedKey) return
    lastFetchedKey = newKey
    currentPage.value = 1
    hasMore.value = true
    requestId++
    fetchPage(1, false)
  },
  { immediate: true },
)

onMounted(() => {
  window.addEventListener('scroll', handleScroll)
})

onDeactivated(() => {
  deactivated = true
})

onActivated(() => {
  deactivated = false
  lastFetchedKey = ''
})

onBeforeUnmount(() => {
  if (rafId) cancelAnimationFrame(rafId)
  window.removeEventListener('scroll', handleScroll)
})
</script>

<style scoped>
.results {
  min-height: 200px;
}

.loading-more {
  text-align: center;
  padding: 16px;
  color: #8c8c8c;
  font-size: 14px;
}

.empty {
  padding: 80px 24px;
  border: 1px dashed #333;
  border-radius: 12px;
  text-align: center;
  font-size: 16px;
  color: #8c8c8c;
}
</style>
