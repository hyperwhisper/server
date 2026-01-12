export default defineNuxtRouteMiddleware(async () => {
  const { isAuthenticated, isInitialized, initAuth, user } = useAuth()

  // Initialize auth if not done
  if (!isInitialized.value) {
    await initAuth()
  }

  // Redirect to signin if not authenticated
  if (!isAuthenticated.value) {
    return navigateTo('/signin')
  }

  // Return 401 if not admin
  if (user.value?.user_type !== 'admin') {
    throw createError({
      statusCode: 401,
      statusMessage: 'Unauthorized',
      message: 'Admin access required'
    })
  }
})
