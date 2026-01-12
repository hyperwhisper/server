export default defineNuxtRouteMiddleware(async () => {
  const { isAuthenticated, isInitialized, initAuth } = useAuth()

  // Initialize auth if not done
  if (!isInitialized.value) {
    await initAuth()
  }

  // Redirect to signin if not authenticated
  if (!isAuthenticated.value) {
    return navigateTo('/signin')
  }
})
