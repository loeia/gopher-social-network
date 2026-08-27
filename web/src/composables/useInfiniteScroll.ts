import { onActivated, onBeforeUnmount, onDeactivated, onMounted, ref, type Ref } from 'vue'

export function useInfiniteScroll(callback: () => Promise<void>, enabled: Ref<boolean>) {
  const loadingMore = ref(false)
  const active = ref(true)

  async function handleScroll() {
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

  function addListener() {
    window.addEventListener('scroll', handleScroll)
  }

  function removeListener() {
    window.removeEventListener('scroll', handleScroll)
  }

  onMounted(addListener)
  onBeforeUnmount(removeListener)

  onActivated(() => {
    active.value = true
    addListener()
  })
  onDeactivated(() => {
    active.value = false
    removeListener()
  })

  return { loadingMore }
}
