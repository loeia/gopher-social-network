<template>
  <div class="search-page">
    <div class="search-header">
      <h1 class="page-title">Search Results</h1>
      <p class="summary">{{ summary }}</p>
    </div>

    <SearchResultsList :params="params" />
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useRoute } from 'vue-router'
import SearchResultsList from '@/components/SearchResultsList.vue'
import type { SearchParams } from '@/stores/feed'

defineOptions({ name: 'SearchResults' })

const route = useRoute()

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
</script>

<style scoped>
.search-page {
  min-height: 100vh;
  padding: 32px 0 80px;
}

.search-header {
  margin: 0 auto 24px;
  max-width: 1100px;
  padding: 0 20px;
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
