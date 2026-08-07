<template>
  <div class="confirm-container">
    <h1>Confirmation</h1>
    <el-button type="primary" size="large" :loading="loading" @click="handleConfirm">
      Click to confirm
    </el-button>
  </div>
</template>

<script setup>
import { ref } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { apiFetch } from '@/api'
import { notify } from '@/utils/message'

const route = useRoute()
const router = useRouter()

const token = route.params.token || ''
const loading = ref(false)

const handleConfirm = async () => {
  loading.value = true
  try {
    const response = await apiFetch(`/users/activate/${token}`, {
      method: 'PUT',
    })

    if (response.ok) {
      notify('success', 'Confirmed successfully')
      router.push('/')
    } else {
      notify('error', 'Failed to confirm token')
    }
  } catch (error) {
    console.error('Confirm error:', error)
    notify('error', 'Network error, please try again')
  } finally {
    loading.value = false
  }
}
</script>

<style scoped>
.confirm-container {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 24px;
  min-height: 100vh;
}
</style>