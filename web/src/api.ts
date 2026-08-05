import { ref } from 'vue'
import { API_URL } from '@/main'

const TOKEN_KEY = 'auth_token'

const token = ref<string | null>(localStorage.getItem(TOKEN_KEY))

export function getToken(): string | null {
  return token.value
}

export function setToken(value: string) {
  token.value = value
  localStorage.setItem(TOKEN_KEY, value)
}

export function clearToken() {
  token.value = null
  localStorage.removeItem(TOKEN_KEY)
}

export async function apiFetch(path: string, options: RequestInit = {}): Promise<Response> {
  const headers = new Headers(options.headers)
  headers.set('Content-Type', 'application/json')
  const token = getToken()
  if (token) {
    headers.set('Authorization', `Bearer ${token}`)
  }
  return fetch(`${API_URL}${path}`, { ...options, headers })
}
