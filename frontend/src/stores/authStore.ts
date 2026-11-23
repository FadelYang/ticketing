import { computed, ref } from 'vue'
import { defineStore } from 'pinia'
import { AUTH_TOKEN_KEY, AUTH_USER_KEY, ApiError, getStoredToken } from '@/services/api'
import { authService } from '@/services/authService'
import type { AuthUser, LoginPayload, RegisterPayload, TokenResponse } from '@/types/auth'

function readStoredUser(): AuthUser | null {
  const raw = localStorage.getItem(AUTH_USER_KEY) || sessionStorage.getItem(AUTH_USER_KEY)
  if (!raw) return null
  try {
    return JSON.parse(raw) as AuthUser
  } catch {
    return null
  }
}

export const useAuthStore = defineStore('auth', () => {
  const token = ref<string | null>(getStoredToken())
  const user = ref<AuthUser | null>(readStoredUser())
  const loading = ref(false)
  const error = ref<string | null>(null)

  const isAuthenticated = computed(() => Boolean(token.value && user.value))

  function persistSession(data: TokenResponse, remember: boolean) {
    token.value = data.access_token
    user.value = data.user

    localStorage.removeItem(AUTH_TOKEN_KEY)
    localStorage.removeItem(AUTH_USER_KEY)
    sessionStorage.removeItem(AUTH_TOKEN_KEY)
    sessionStorage.removeItem(AUTH_USER_KEY)

    const storage = remember ? localStorage : sessionStorage
    storage.setItem(AUTH_TOKEN_KEY, data.access_token)
    storage.setItem(AUTH_USER_KEY, JSON.stringify(data.user))
  }

  function clearSession() {
    token.value = null
    user.value = null
    localStorage.removeItem(AUTH_TOKEN_KEY)
    localStorage.removeItem(AUTH_USER_KEY)
    sessionStorage.removeItem(AUTH_TOKEN_KEY)
    sessionStorage.removeItem(AUTH_USER_KEY)
  }

  function clearError() {
    error.value = null
  }

  async function login(payload: LoginPayload, remember = false): Promise<boolean> {
    loading.value = true
    error.value = null
    try {
      const res = await authService.login(payload)
      persistSession(res, remember)
      return true
    } catch (err) {
      error.value = err instanceof ApiError ? err.message : 'Unable to sign in. Please try again.'
      return false
    } finally {
      loading.value = false
    }
  }

  async function register(payload: RegisterPayload): Promise<boolean> {
    loading.value = true
    error.value = null
    try {
      const res = await authService.register(payload)
      persistSession(res, true)
      return true
    } catch (err) {
      error.value =
        err instanceof ApiError ? err.message : 'Unable to create account. Please try again.'
      return false
    } finally {
      loading.value = false
    }
  }

  async function logout(): Promise<void> {
    try {
      if (token.value) {
        await authService.logout()
      }
    } catch {
      // Clear local session even if token already expired
    } finally {
      clearSession()
    }
  }

  return {
    token,
    user,
    loading,
    error,
    isAuthenticated,
    clearError,
    login,
    register,
    logout,
  }
})
