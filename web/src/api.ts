import { ref } from 'vue'
import { API_URL } from '@/main'
import { notify } from '@/utils/message'

const TOKEN_KEY = 'auth_token'

const token = ref<string | null>(localStorage.getItem(TOKEN_KEY))

let sessionExpiredHandled = false

export function getToken(): string | null {
  return token.value
}

export function setToken(value: string) {
  token.value = value
  localStorage.setItem(TOKEN_KEY, value)
  sessionExpiredHandled = false
}

export function clearToken() {
  token.value = null
  localStorage.removeItem(TOKEN_KEY)
}

function decodeJwtPayload(value: string): Record<string, unknown> | null {
  try {
    const payloadPart = value.split('.')[1]
    if (!payloadPart) return null
    return JSON.parse(atob(payloadPart.replace(/-/g, '+').replace(/_/g, '/')))
  } catch {
    return null
  }
}

export function isTokenExpired(value?: string | null): boolean {
  const t = value ?? getToken()
  if (!t) return false
  const exp = decodeJwtPayload(t)?.exp
  if (typeof exp !== 'number' || !Number.isFinite(exp)) return false
  return exp * 1000 <= Date.now()
}

export function getCurrentUserId(): number | null {
  const t = getToken()
  if (!t) return null
  const sub = decodeJwtPayload(t)?.sub
  if (typeof sub === 'number' && Number.isFinite(sub)) return sub
  return null
}

export async function getApiError(response: Response): Promise<string | null> {
  try {
    const json = await response.json()
    if (json && typeof json.error === 'string' && json.error) return json.error
    return null
  } catch {
    return null
  }
}

export async function handleSessionExpired() {
  if (!getToken() || sessionExpiredHandled) return
  sessionExpiredHandled = true
  clearToken()
  notify('warning', 'Session expired, please sign in again')
  const { default: router } = await import('./router')
  if (router.currentRoute.value.name !== 'Login') {
    router.push({ name: 'Login' })
  }
}

export async function apiFetch(path: string, options: RequestInit = {}): Promise<Response> {
  const headers = new Headers(options.headers)
  const isFormData = options.body instanceof FormData
  if (!isFormData) {
    headers.set('Content-Type', 'application/json')
  }
  const currentToken = getToken()
  if (currentToken) {
    headers.set('Authorization', `Bearer ${currentToken}`)
  }
  const response = await fetch(`${API_URL}${path}`, { ...options, headers })

  if (response.status === 401 && currentToken && path !== '/authentication/token') {
    await handleSessionExpired()
  }

  return response
}
