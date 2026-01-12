<script setup lang="ts">
import { Loader2 } from 'lucide-vue-next'
import type { SignInPayload } from '~/types/auth'

definePageMeta({
  middleware: 'guest'
})

useHead({
  title: 'Sign In - HyperWhisper'
})

const route = useRoute()
const { signIn } = useAuth()

const form = reactive<SignInPayload>({
  identifier: '',
  password: '',
})

const isLoading = ref(false)
const errorMessage = ref('')

const handleSubmit = async () => {
  errorMessage.value = ''
  isLoading.value = true

  const { success, error } = await signIn(form)

  isLoading.value = false

  if (success) {
    // Redirect to intended destination or default to dashboard
    const redirect = route.query.redirect as string
    await navigateTo(redirect || '/dashboard')
  } else if (error) {
    errorMessage.value = error.error
  }
}
</script>

<template>
  <div class="min-h-screen bg-white dark:bg-black">
    <AppNavbar />

    <div class="min-h-screen flex items-center justify-center px-4 pt-16">
      <Card class="w-full max-w-md">
        <CardHeader class="text-center">
          <CardTitle class="text-2xl">Welcome back</CardTitle>
          <CardDescription>Sign in to your account</CardDescription>
        </CardHeader>
        <CardContent>
          <form @submit.prevent="handleSubmit" class="space-y-4">
            <Alert v-if="errorMessage" variant="destructive">
              <AlertDescription>{{ errorMessage }}</AlertDescription>
            </Alert>

            <div class="space-y-2">
              <Label for="identifier">Email or Username</Label>
              <Input
                id="identifier"
                v-model="form.identifier"
                placeholder="john@example.com or johndoe"
                required
              />
            </div>

            <div class="space-y-2">
              <Label for="password">Password</Label>
              <Input
                id="password"
                type="password"
                v-model="form.password"
                placeholder="********"
                required
              />
            </div>

            <button
              type="submit"
              :disabled="isLoading"
              class="inline-flex items-center justify-center gap-2 whitespace-nowrap rounded-md text-sm font-medium ring-offset-background transition-colors focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2 disabled:pointer-events-none disabled:opacity-50 bg-primary text-primary-foreground hover:bg-primary/90 h-10 px-4 py-2 w-full"
            >
              <Loader2 v-if="isLoading" class="mr-2 h-4 w-4 animate-spin" />
              {{ isLoading ? 'Signing in...' : 'Sign In' }}
            </button>
          </form>

          <div class="mt-6 text-center text-sm text-muted-foreground">
            Don't have an account?
            <NuxtLink :to="{ path: '/signup', query: route.query }" class="text-primary hover:underline">
              Sign up
            </NuxtLink>
          </div>
        </CardContent>
      </Card>
    </div>
  </div>
</template>
