<template>
  <div class="search-page">
    <div class="back-nav">
      <el-button text @click="goHome">← Back</el-button>
    </div>

    <div class="search-header">
      <h1 class="page-title">Search Results</h1>
      <p class="summary">{{ summary }}</p>
    </div>

    <SearchResultsList :params="params" />
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import SearchResultsList from '@/components/SearchResultsList.vue'
import { useFeedStore, type SearchParams } from '@/stores/feed'

defineOptions({ name: 'SearchResults' })

const store = useFeedStore()
const route = useRoute()
const router = useRouter()

function q(value: unknown): string {
  if (typeof value === 'string') return value
  if (Array.isArray(value)) return value[0] ?? ''
  return ''
}

const params = computed<SearchParams>(() => ({
  search: q(route.query.search).trim() || undefined,
  author: q(route.query.author).trim() || undefined,
  tags: q(route.query.tags)
    .split(',')
    .map((tag) => tag.trim())
    .filter(Boolean),
  since: q(route.query.since) || undefined,
  until: q(route.query.until) || undefined,
  page: Number(q(route.query.page)) > 1 ? Number(q(route.query.page)) : undefined,
}))

const summary = computed(() => {
  const p = params.value
  const parts: string[] = []
  if (p.search) parts.push(`keyword "${p.search}"`)
  if (p.author) parts.push(`author "${p.author}"`)
  if (p.tags && p.tags.length) parts.push(`tags ${p.tags.join(', ')}`)
  if (p.since || p.until) parts.push(`${p.since ?? '…'} ~ ${p.until ?? '…'}`)
  return parts.length ? `Results for ${parts.join(' · ')}` : ''
})

function goHome() {
  store.clearPostHistory()
  router.push('/')
}
</script>

<style scoped>
.search-page {
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

.search-header {
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

.summary {
  margin: 0;
  font-size: 14px;
  color: #8c8c8c;
}
</style>
