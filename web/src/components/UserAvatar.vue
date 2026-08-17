<template>
  <img
    v-if="showImage && src"
    :src="src"
    :width="size"
    :height="size"
    class="avatar-img"
    :style="{ width: `${size}px`, height: `${size}px` }"
    :alt="username"
    loading="lazy"
    @error="onError"
  />
  <span
    v-else
    class="avatar-initials"
    :style="{
      width: `${size}px`,
      height: `${size}px`,
      fontSize: `${fontSize}px`,
      background: initialsColor,
    }"
  >
    {{ initials }}
  </span>
</template>

<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import { API_URL } from '@/main'

const AVATAR_COLORS = ['#5b8def', '#e07b5a', '#57c08a', '#c07bd8', '#d8a24a', '#6ec7c0']

const props = withDefaults(
  defineProps<{
    userId?: number | null
    username?: string
    size?: number
  }>(),
  {
    userId: null,
    username: '',
    size: 28,
  },
)

const showImage = ref(true)

watch(
  () => props.userId,
  () => {
    showImage.value = true
  },
)

const src = computed(() => {
  if (!props.userId) return ''
  return `${API_URL}/users/${props.userId}/avatar`
})

const initials = computed(() => {
  const name = props.username.trim()
  return name ? name[0]!.toUpperCase() : '?'
})

const initialsColor = computed(() => {
  const name = props.username
  let hash = 0
  for (let i = 0; i < name.length; i++) hash = (hash * 31 + name.charCodeAt(i)) >>> 0
  return AVATAR_COLORS[hash % AVATAR_COLORS.length]
})

const fontSize = computed(() => Math.max(12, Math.round(props.size * 0.4)))

function onError() {
  showImage.value = false
}
</script>

<style scoped>
.avatar-img,
.avatar-initials {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
  border-radius: 50%;
  border: 1px solid #262626;
}

.avatar-img {
  object-fit: cover;
}

.avatar-initials {
  color: #ffffff;
  font-weight: 700;
  line-height: 1;
}
</style>
