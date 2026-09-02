<template>
  <div class="confirm-container">
    <h1>Confirmation</h1>
    <el-button size="large" class="confirm-btn" :loading="loading" @click="handleConfirm">
      Click to confirm
    </el-button>
  </div>
</template>

<script setup>
import { ref } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { apiFetch, handleApiError } from '@/api'
import { notify } from '@/utils/message'

const route = useRoute()
const router = useRouter()

const token = route.params.token || ''
const loading = ref(false)

const handleConfirm = async () => {
  loading.value = true
  try {
    const response = await apiFetch(`/auth/activate/${token}`, {
      method: 'POST',
    })

    if (response.ok) {
      notify('success', 'Confirmed successfully')
      router.push('/')
    } else {
      notify('error', 'Failed to confirm token')
    }
  } catch (error) {
    handleApiError(error, 'Network error, please try again')
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

.confirm-btn {
  background: #ffffff;
  color: #141414;
  border: 1px solid #ffffff;
  font-weight: 600;
}

.confirm-btn:hover {
  background: #e4e6e8;
  color: #141414;
  border-color: #e4e6e8;
}

.confirm-btn.is-loading {
  background: #ffffff;
  border-color: #ffffff;
  color: #141414;
}
</style>
