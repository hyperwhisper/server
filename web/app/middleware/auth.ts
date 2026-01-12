export default defineNuxtRouteMiddleware(async (to) => {
  const { isAuthenticated, isInitialized, initAuth } = useAuth()

  // Initialize auth if not done
  if (!isInitialized.value) {
    await initAuth()
  }

  // Redirect to signin if not authenticated, preserving intended destination
  if (!isAuthenticated.value) {
    return navigateTo({
      path: '/signin',
      query: { redirect: to.fullPath }
    })
  }
})
