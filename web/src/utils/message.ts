import { ElMessage } from 'element-plus'

const MAX_VISIBLE = 3
const DURATION = 1000
const INTERVAL = Math.floor(DURATION / MAX_VISIBLE)

type MessageType = 'success' | 'warning' | 'error' | 'info'

interface PendingMessage {
  type: MessageType
  message: string
}

let visible = 0
let lastShownAt = 0
let nextTimer: number | null = null
const pending: PendingMessage[] = []

function canShowNow() {
  return pending.length > 0 && visible < MAX_VISIBLE && Date.now() - lastShownAt >= INTERVAL
}

function showOne() {
  const item = pending.shift()!
  lastShownAt = Date.now()
  ElMessage({
    type: item.type,
    message: item.message,
    duration: DURATION,
    onClose: () => {
      visible -= 1
      flush()
    },
  })
  visible += 1
}

function scheduleNext() {
  if (nextTimer !== null || pending.length === 0 || visible >= MAX_VISIBLE) return
  const wait = Math.max(0, lastShownAt + INTERVAL - Date.now())
  nextTimer = window.setTimeout(() => {
    nextTimer = null
    flush()
  }, wait)
}

function flush() {
  while (canShowNow()) showOne()
  scheduleNext()
}

export function notify(type: MessageType, message: string) {
  pending.push({ type, message })
  flush()
}
