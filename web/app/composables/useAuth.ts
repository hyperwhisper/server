import type { User, AuthResponse, SignUpPayload, SignInPayload, TokenRefreshResponse, ApiError } from '~/types/auth'

// Token state (in-memory for access token)
const accessToken = ref<string | null>(null)
const tokenExpiresAt = ref<number | null>(null)
const user = ref<User | null>(null)
const isInitialized = ref(false)

// Refresh timer
let refreshTimer: ReturnType<typeof setTimeout> | null = null

export function useAuth() {
  const isAuthenticated = computed(() => !!accessToken.value && !!user.value)

  // Set token and schedule refresh
  const setToken = (token: string, expiresIn: number) => {
    accessToken.value = token
    tokenExpiresAt.value = Date.now() + expiresIn * 1000

    // Clear existing timer
    if (refreshTimer) {
      clearTimeout(refreshTimer)
    }

    // Schedule refresh 30 seconds before expiry
    const refreshIn = (expiresIn - 30) * 1000
    if (refreshIn > 0) {
      refreshTimer = setTimeout(() => {
        refreshAccessToken()
      }, refreshIn)
    }
  }

  // Clear auth state
  const clearAuth = () => {
    accessToken.value = null
    tokenExpiresAt.value = null
    user.value = null

    if (refreshTimer) {
      clearTimeout(refreshTimer)
      refreshTimer = null
    }
  }

  // Sign up
  const signUp = async (payload: SignUpPayload): Promise<{ success: boolean; error?: ApiError }> => {
    try {
      const response = await $fetch<AuthResponse>('/api/v1/signup', {
        method: 'POST',
        body: payload,
      })

      if (response) {
        setToken(response.access_token, response.expires_in)
        user.value = response.user
        return { success: true }
      }

      return { success: false, error: { error: 'Unknown error' } }
    } catch (e: any) {
      const apiError = e.data as ApiError
      return { success: false, error: apiError || { error: 'Network error' } }
    }
  }

  // Sign in
  const signIn = async (payload: SignInPayload): Promise<{ success: boolean; error?: ApiError }> => {
    try {
      const response = await $fetch<AuthResponse>('/api/v1/signin', {
        method: 'POST',
        body: payload,
      })

      if (response) {
        setToken(response.access_token, response.expires_in)
        user.value = response.user
        return { success: true }
      }

      return { success: false, error: { error: 'Unknown error' } }
    } catch (e: any) {
      const apiError = e.data as ApiError
      return { success: false, error: apiError || { error: 'Network error' } }
    }
  }

  // Sign out
  const signOut = async (): Promise<void> => {
    try {
      await $fetch('/api/v1/signout', {
        method: 'POST',
      })
    } catch (e) {
      // Ignore errors, clear local state anyway
    }
    clearAuth()
    await navigateTo('/')
  }

  // Refresh access token
  const refreshAccessToken = async (): Promise<boolean> => {
    try {
      const response = await $fetch<TokenRefreshResponse>('/api/v1/token_refresh', {
        method: 'POST',
      })

      if (response) {
        setToken(response.access_token, response.expires_in)
        return true
      }

      clearAuth()
      return false
    } catch (e) {
      clearAuth()
      return false
    }
  }

  // Fetch current user
  const fetchUser = async (): Promise<boolean> => {
    if (!accessToken.value) {
      return false
    }

    try {
      const response = await $fetch<User>('/api/v1/me', {
        headers: { Authorization: `Bearer ${accessToken.value}` },
      })

      if (response) {
        user.value = response
        return true
      }

      return false
    } catch (e) {
      return false
    }
  }

  // Initialize auth state (try to restore from refresh token)
  const initAuth = async (): Promise<void> => {
    if (isInitialized.value) return

    // Try to refresh token (cookie will be sent automatically)
    const refreshed = await refreshAccessToken()
    if (refreshed) {
      await fetchUser()
    }

    isInitialized.value = true
  }

  // Get authorization headers for API calls
  const getAuthHeaders = (): Record<string, string> => {
    if (accessToken.value) {
      return { Authorization: `Bearer ${accessToken.value}` }
    }
    return {}
  }

  return {
    // State
    user: readonly(user),
    isAuthenticated,
    isInitialized: readonly(isInitialized),
    accessToken: readonly(accessToken),

    // Actions
    signUp,
    signIn,
    signOut,
    refreshAccessToken,
    fetchUser,
    initAuth,
    getAuthHeaders,
  }
}
