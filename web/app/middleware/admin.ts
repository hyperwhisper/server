export default defineNuxtRouteMiddleware(async (to) => {
  const { isAuthenticated, isInitialized, initAuth, user } = useAuth()

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

  // Return 401 if not admin
  if (user.value?.user_type !== 'admin') {
    throw createError({
      statusCode: 401,
      statusMessage: 'Unauthorized',
      message: 'Admin access required'
    })
  }
})
