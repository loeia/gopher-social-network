import { nextTick, onActivated, onBeforeUnmount, onDeactivated, onMounted, ref, watch, type Ref } from 'vue'

export function useInfiniteScroll(callback: () => Promise<void>, enabled: Ref<boolean>) {
  const loadingMore = ref(false)
  const active = ref(true)
  let lastScrollTime = 0
  const scrollDebounceMs = 300

  async function tryLoad() {
    if (!active.value || loadingMore.value || !enabled.value) return
    const scrollHeight = document.documentElement.scrollHeight
    const scrollTop = window.scrollY
    const clientHeight = window.innerHeight
    if (scrollTop + clientHeight >= scrollHeight - 200) {
      loadingMore.value = true
      try {
        await callback()
      } finally {
        loadingMore.value = false
      }
    }
  }

  async function handleScroll() {
    if (!active.value || loadingMore.value || !enabled.value) return

    const now = Date.now()
    if (now - lastScrollTime < scrollDebounceMs) return
    lastScrollTime = now

    await tryLoad()
  }

  function addListener() {
    window.addEventListener('scroll', handleScroll)
  }

  function removeListener() {
    window.removeEventListener('scroll', handleScroll)
  }

  watch(enabled, (val) => {
    if (val) nextTick(tryLoad)
  })

  onMounted(() => {
    addListener()
    nextTick(tryLoad)
  })
  onBeforeUnmount(removeListener)

  onActivated(() => {
    active.value = true
    addListener()
    nextTick(tryLoad)
  })
  onDeactivated(() => {
    active.value = false
    removeListener()
  })

  return { loadingMore }
}
