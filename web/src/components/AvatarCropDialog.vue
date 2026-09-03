<template>
  <el-dialog
    :model-value="visible"
    title="Crop avatar"
    width="540px"
    :close-on-click-modal="false"
    append-to-body
    class="avatar-crop-dialog"
    @update:model-value="onUpdateVisible"
    @opened="initCropper"
    @close="destroyCropper"
  >
    <div class="crop-body">
      <img ref="imgRef" :src="src" alt="avatar to crop" />
    </div>
    <template #footer>
      <el-button @click="cancel">Cancel</el-button>
      <el-button class="confirm-btn" :loading="cropping" :disabled="!ready" @click="confirmCrop">
        Confirm
      </el-button>
    </template>
  </el-dialog>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import Cropper from 'cropperjs'
import 'cropperjs/dist/cropper.css'
import { notify } from '@/utils/message'

const OUTPUT_SIZE = 2000
const MAX_AVATAR_SIZE = 2 * 1024 * 1024

const props = defineProps<{
  visible: boolean
  src: string
}>()

const emit = defineEmits<{ confirm: [blob: Blob]; close: [] }>()

const imgRef = ref<HTMLImageElement | null>(null)
const ready = ref(false)
const cropping = ref(false)

let cropper: Cropper | null = null

function onUpdateVisible(value: boolean) {
  if (!value) emit('close')
}

function initCropper() {
  destroyCropper()
  ready.value = false
  const img = imgRef.value
  if (!img) return
  cropper = new Cropper(img, {
    aspectRatio: 1,
    viewMode: 1,
    dragMode: 'move',
    autoCropArea: 1,
    background: false,
    responsive: true,
    ready() {
      ready.value = true
    },
  })
}

function destroyCropper() {
  if (cropper) {
    cropper.destroy()
    cropper = null
  }
  ready.value = false
}

function cancel() {
  emit('close')
}

function canvasToBlob(canvas: HTMLCanvasElement, quality: number): Promise<Blob> {
  return new Promise((resolve, reject) => {
    canvas.toBlob(
      (blob) => {
        if (blob) resolve(blob)
        else reject(new Error('Image encoding failed'))
      },
      'image/jpeg',
      quality,
    )
  })
}

async function confirmCrop() {
  if (!cropper || !ready.value) return
  cropping.value = true
  try {
    const canvas = cropper.getCroppedCanvas({
      width: OUTPUT_SIZE,
      height: OUTPUT_SIZE,
      imageSmoothingQuality: 'high',
    })
    let blob = await canvasToBlob(canvas, 0.92)
    for (const quality of [0.85, 0.75, 0.6]) {
      if (blob.size <= MAX_AVATAR_SIZE) break
      blob = await canvasToBlob(canvas, quality)
    }
    if (blob.size > MAX_AVATAR_SIZE) {
      notify('error', 'Image is too large, please choose a simpler image')
      return
    }
    emit('confirm', blob)
  } catch (error) {
    console.error('Crop error:', error)
    notify('error', 'Failed to crop image')
  } finally {
    cropping.value = false
  }
}
</script>

<style scoped>
.crop-body {
  width: 100%;
  height: 320px;
  overflow: hidden;
  background: #0a0a0a;
  border-radius: 6px;
}

.crop-body :deep(img) {
  display: block;
  max-width: 100%;
}

.confirm-btn {
  background: #141414;
  color: #ffffff;
  border: 1px solid #141414;
  font-weight: 600;
}

.confirm-btn:hover:not(.is-disabled) {
  background: #32383f;
  color: #ffffff;
  border-color: #32383f;
}

.confirm-btn.is-disabled {
  background: #6e7781;
  border-color: #6e7781;
  color: #ffffff;
}
</style>
