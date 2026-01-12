export default defineNuxtRouteMiddleware(async () => {
  const { isAuthenticated, isInitialized, initAuth } = useAuth()

  // Initialize auth if not done
  if (!isInitialized.value) {
    await initAuth()
  }

  // Redirect to dashboard if already authenticated
  if (isAuthenticated.value) {
    return navigateTo('/dashboard')
  }
})
