<template>
  <header class="navbar">
    <div class="navbar-left">
      <el-button v-if="isLoggedIn" class="menu-btn" @click="toggleSidebar">
        <svg
          class="menu-icon"
          width="18"
          height="18"
          viewBox="0 0 24 24"
          fill="none"
          stroke="currentColor"
          stroke-width="2"
          stroke-linecap="round"
        >
          <line x1="3" y1="6" x2="21" y2="6" />
          <line x1="3" y1="12" x2="21" y2="12" />
          <line x1="3" y1="18" x2="21" y2="18" />
        </svg>
      </el-button>
      <router-link to="/" class="brand">Gopher</router-link>
    </div>

    <div class="navbar-center">
      <el-popover
        v-model:visible="filterOpen"
        placement="bottom-start"
        :width="300"
        trigger="click"
        popper-class="search-filters-popper"
      >
        <template #reference>
          <el-button class="filters-btn" aria-label="Filters">
            <svg
              class="filter-icon"
              width="18"
              height="18"
              viewBox="0 0 24 24"
              fill="none"
              stroke="currentColor"
              stroke-width="2"
              stroke-linecap="round"
              stroke-linejoin="round"
            >
              <line x1="3" y1="6" x2="21" y2="6" />
              <line x1="3" y1="12" x2="21" y2="12" />
              <line x1="3" y1="18" x2="21" y2="18" />
              <circle cx="8" cy="6" r="2.2" fill="currentColor" stroke="none" />
              <circle cx="16" cy="12" r="2.2" fill="currentColor" stroke="none" />
              <circle cx="12" cy="18" r="2.2" fill="currentColor" stroke="none" />
            </svg>
          </el-button>
        </template>

        <div class="filters-panel">
          <label class="filter-label" for="filters-since">Start date</label>
          <div class="filter-control">
            <el-date-picker
              id="filters-since"
              v-model="startDate"
              class="date-field"
              type="date"
              placeholder="Start date"
              value-format="YYYY-MM-DD"
              :teleported="false"
              popper-class="date-filter-popper"
              clearable
            />
          </div>
          <label class="filter-label" for="filters-until">End date</label>
          <div class="filter-control">
            <el-date-picker
              id="filters-until"
              v-model="endDate"
              class="date-field"
              type="date"
              placeholder="End date"
              value-format="YYYY-MM-DD"
              :teleported="false"
              popper-class="date-filter-popper"
              clearable
            />
          </div>
          <label class="filter-label" for="filters-author">Author</label>
          <el-input
            id="filters-author"
            v-model="author"
            class="filter-control"
            placeholder="Author"
            clearable
            @keyup.enter="handleSearch"
          />
          <label class="filter-label" for="filters-tags">Tags</label>
          <el-input
            id="filters-tags"
            v-model="tags"
            class="filter-control"
            placeholder="Tags (comma separated)"
            clearable
            @keyup.enter="handleSearch"
          />
        </div>
      </el-popover>

      <el-input
        v-model="searchQuery"
        class="navbar-search"
        placeholder="Search posts"
        clearable
        @keyup.enter="handleSearch"
      />

      <el-button class="navbar-search-btn" aria-label="Search" @click="handleSearch">
        <svg
          class="search-icon"
          width="18"
          height="18"
          viewBox="0 0 24 24"
          fill="none"
          stroke="currentColor"
          stroke-width="2"
          stroke-linecap="round"
        >
          <circle cx="11" cy="11" r="8" />
          <line x1="21" y1="21" x2="16.65" y2="16.65" />
        </svg>
      </el-button>
    </div>

    <div class="navbar-right">
      <template v-if="isLoggedIn">
        <el-button @click="handleLogout">Logout</el-button>
      </template>
      <template v-else>
        <el-button @click="goToLogin">Login</el-button>
        <el-button type="primary" @click="goToSignUp">Sign Up</el-button>
      </template>
    </div>
  </header>
</template>

<script setup lang="ts">
import { ref, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { getToken, clearToken } from '@/api'
import { notify } from '@/utils/message'
import { useUIStore } from '@/stores/ui'

const router = useRouter()
const route = useRoute()
const uiStore = useUIStore()

const searchQuery = ref('')
const author = ref('')
const tags = ref('')
const startDate = ref<string | null>(null)
const endDate = ref<string | null>(null)
const filterOpen = ref(false)
const isLoggedIn = ref(!!getToken())

function toggleSidebar() {
  uiStore.toggleSidebar()
}

watch(
  () => route.fullPath,
  () => {
    isLoggedIn.value = !!getToken()
  },
)

function handleSearch() {
  const query: Record<string, string> = {}
  const searchText = searchQuery.value.trim()
  const authorText = author.value.trim()
  const tagsList = tags.value
    .split(',')
    .map((tag) => tag.trim())
    .filter(Boolean)
  if (searchText) query.search = searchText
  if (authorText) query.author = authorText
  if (tagsList.length) query.tags = tagsList.join(',')
  if (startDate.value) query.since = startDate.value
  if (endDate.value) query.until = endDate.value

  if (Object.keys(query).length === 0) {
    notify('warning', 'Please enter search criteria')
    return
  }

  filterOpen.value = false
  router.push({ path: '/search', query })
}

function goToLogin() {
  if (route.path === '/login') return
  router.push({ path: '/login', query: { redirect: route.fullPath } })
}

function goToSignUp() {
  if (route.path === '/signup') return
  router.push({ path: '/signup', query: { redirect: route.fullPath } })
}

async function handleLogout() {
  // TODO: call the logout API to invalidate the server-side token/session.
  // The backend endpoint is not implemented yet, so for now we only clear the local token.
  clearToken()
  isLoggedIn.value = false
  searchQuery.value = ''
  author.value = ''
  tags.value = ''
  startDate.value = null
  endDate.value = null
  filterOpen.value = false
  notify('success', 'Logged out')
  router.push('/')
}
</script>

<style scoped>
.navbar {
  position: sticky;
  top: 0;
  z-index: 100;
  display: grid;
  grid-template-columns: 1fr auto 1fr;
  align-items: center;
  gap: 16px;
  padding: 10px 24px;
  background: #141414;
  border-bottom: 1px solid #262626;
}

.navbar-left {
  justify-self: start;
  display: flex;
  align-items: center;
  gap: 8px;
}

.menu-btn {
  padding: 8px;
  border: 1px solid transparent;
  border-radius: 8px;
  color: #ffffff;
  transition: background 0.2s ease;
  --el-button-bg-color: transparent;
  --el-button-border-color: transparent;
  --el-button-text-color: #ffffff;
  --el-button-hover-bg-color: transparent;
  --el-button-hover-border-color: transparent;
  --el-button-hover-text-color: #ffffff;
  --el-button-active-bg-color: transparent;
  --el-button-active-border-color: transparent;
  --el-button-active-text-color: #ffffff;
}

.menu-btn:hover,
.menu-btn:active,
.menu-btn.is-active {
  background: rgba(255, 255, 255, 0.1);
  color: #ffffff;
  border-color: transparent;
}

.menu-btn:focus-visible {
  outline: 2px solid rgba(255, 255, 255, 0.4);
  outline-offset: 1px;
}

.menu-icon {
  display: block;
}

.brand {
  font-size: 22px;
  font-weight: 700;
  color: #ffffff;
  text-decoration: none;
  white-space: nowrap;
}

.navbar-center {
  justify-self: center;
  display: flex;
  align-items: center;
  gap: 8px;
}

.navbar-search {
  width: 420px;
  max-width: 60vw;
}

.filters-btn {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 38px;
  height: 32px;
  padding: 0;
  background: #0a0a0a;
  color: #ffffff;
  border: 1px solid #262626;
  box-shadow: 0 0 0 1px #262626 inset;
  --el-button-bg-color: #0a0a0a;
  --el-button-border-color: #262626;
  --el-button-text-color: #ffffff;
  --el-button-hover-bg-color: #1f1f1f;
  --el-button-hover-border-color: #ffffff;
  --el-button-hover-text-color: #ffffff;
  --el-button-active-bg-color: #1f1f1f;
  --el-button-active-border-color: #ffffff;
  --el-button-active-text-color: #ffffff;
}

.filters-btn:hover,
.filters-btn:active,
.filters-btn.is-active {
  background: #1f1f1f;
  color: #ffffff;
  border-color: #ffffff;
}

.filter-icon {
  display: block;
}

.navbar-search :deep(.el-input__wrapper) {
  background: #0a0a0a;
  box-shadow: 0 0 0 1px #262626 inset;
}

.navbar-search :deep(.el-input__wrapper.is-focus) {
  box-shadow: 0 0 0 1px #ffffff inset;
}

.navbar-search :deep(.el-input__inner) {
  color: #ffffff;
}

.navbar-search :deep(.el-input__inner::placeholder) {
  color: #595959;
}

:global(.search-filters-popper.el-popper) {
  background: #141414;
  border: 1px solid #262626;
  box-shadow: 0 8px 24px rgba(0, 0, 0, 0.5);
}

:global(.search-filters-popper .el-popper__arrow) {
  display: none;
}

.filters-panel {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.filter-label {
  font-size: 13px;
  font-weight: 600;
  color: #8c8c8c;
}

.filter-control {
  width: 100%;
  --el-input-bg-color: #0a0a0a;
  --el-input-border-color: #262626;
  --el-date-editor-width: 100%;
}

.filter-control :deep(.el-date-editor) {
  width: 100%;
  background: #0a0a0a;
}

.filter-control :deep(.el-date-editor .el-input__wrapper) {
  background: #0a0a0a;
  box-shadow: 0 0 0 1px #262626 inset;
}

.filter-control :deep(.el-date-editor .el-input__wrapper.is-focus) {
  box-shadow: 0 0 0 1px #ffffff inset;
}

.filter-control :deep(.el-date-editor .el-input__inner) {
  color: #ffffff;
}

.filter-control :deep(.el-date-editor .el-input__inner::placeholder) {
  color: #595959;
}

.filter-control :deep(.el-input__wrapper) {
  background: #0a0a0a;
  box-shadow: 0 0 0 1px #262626 inset;
}

.filter-control :deep(.el-input__wrapper.is-focus) {
  box-shadow: 0 0 0 1px #ffffff inset;
}

.filter-control :deep(.el-input__inner) {
  color: #ffffff;
}

.filter-control :deep(.el-input__inner::placeholder) {
  color: #595959;
}

:global(.date-filter-popper.el-popper) {
  padding: 0;
  background: #ffffff;
  border: 1px solid #d0d0d0;
  border-radius: 6px;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.12);
  color: #141414;
}

:global(.date-filter-popper .el-popper__arrow::before) {
  background: #ffffff;
  border-color: #d0d0d0;
}

:global(.date-filter-popper .el-date-picker__header-label) {
  color: #141414;
  font-weight: 600;
}

:global(.date-filter-popper .el-date-picker__prev-btn button),
:global(.date-filter-popper .el-date-picker__next-btn button) {
  color: #777;
}

:global(.date-filter-popper .el-date-picker__prev-btn button:hover),
:global(.date-filter-popper .el-date-picker__next-btn button:hover) {
  background: #f0f0f0;
  color: #141414;
}

:global(.date-filter-popper .el-date-picker__content) {
  background: #ffffff;
}

:global(.date-filter-popper th.el-date-table) {
  color: #141414;
}

:global(.date-filter-popper table th) {
  color: #777;
  font-weight: 600;
}

:global(.date-filter-popper .el-date-table td .el-date-table-cell__text) {
  color: #141414;
}

:global(.date-filter-popper .el-date-table td:not(.disabled):hover .el-date-table-cell__text) {
  background: #e8e8e8;
  color: #141414;
}

:global(.date-filter-popper .el-date-table td.current .el-date-table-cell__text) {
  background: #141414;
  color: #ffffff;
}

:global(.date-filter-popper .el-date-table td.today .el-date-table-cell__text) {
  border: 1px solid #141414;
  color: #141414;
}

:global(.date-filter-popper .el-date-table td.available:hover .el-date-table-cell__text) {
  background: #e8e8e8;
  color: #141414;
}

:global(.date-filter-popper .el-picker-panel__footer) {
  background: #ffffff;
  border-top: 1px solid #e6e6e6;
}

:global(.date-filter-popper .el-picker-panel__footer button.el-button) {
  color: #141414;
  border: 1px solid #d0d0d0;
}

:global(.date-filter-popper .el-picker-panel__footer button.el-button:hover) {
  background: #f0f0f0;
  color: #141414;
}

.navbar-search-btn {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 38px;
  height: 32px;
  padding: 0;
  background: #0a0a0a;
  color: #ffffff;
  border: 1px solid #262626;
  box-shadow: 0 0 0 1px #262626 inset;
  --el-button-bg-color: #0a0a0a;
  --el-button-border-color: #262626;
  --el-button-text-color: #ffffff;
  --el-button-hover-bg-color: #1f1f1f;
  --el-button-hover-border-color: #ffffff;
  --el-button-hover-text-color: #ffffff;
  --el-button-active-bg-color: #1f1f1f;
  --el-button-active-border-color: #ffffff;
  --el-button-active-text-color: #ffffff;
}

.navbar-search-btn:hover,
.navbar-search-btn:active {
  background: #1f1f1f;
  color: #ffffff;
  border-color: #ffffff;
}

.search-icon {
  display: block;
}

.navbar-right {
  justify-self: end;
  display: flex;
  align-items: center;
  gap: 12px;
}

.navbar-right :deep(.el-button) {
  background: #ffffff;
  color: #141414;
  border: 1px solid #ffffff;
  font-weight: 600;
}

.navbar-right :deep(.el-button:hover) {
  background: #e4e6e8;
  color: #141414;
  border-color: #e4e6e8;
}
</style>
