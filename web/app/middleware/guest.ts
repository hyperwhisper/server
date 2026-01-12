export default defineNuxtRouteMiddleware(async (to) => {
  const { isAuthenticated, isInitialized, initAuth } = useAuth()

  // Initialize auth if not done
  if (!isInitialized.value) {
    await initAuth()
  }

  // Redirect to intended destination or dashboard if already authenticated
  if (isAuthenticated.value) {
    const redirect = to.query.redirect as string
    return navigateTo(redirect || '/dashboard')
  }
})
