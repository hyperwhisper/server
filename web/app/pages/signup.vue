<script setup lang="ts">
import { Loader2 } from 'lucide-vue-next'
import type { SignUpPayload } from '~/types/auth'

definePageMeta({
  middleware: 'guest'
})

useHead({
  title: 'Sign Up - HyperWhisper'
})

const { signUp } = useAuth()

const form = reactive<SignUpPayload>({
  username: '',
  email: '',
  password: '',
  first_name: '',
  last_name: '',
})

const confirmPassword = ref('')
const isLoading = ref(false)
const errorMessage = ref('')
const fieldErrors = ref<Record<string, string>>({})

const passwordsMatch = computed(() => form.password === confirmPassword.value)

const handleSubmit = async () => {
  errorMessage.value = ''
  fieldErrors.value = {}

  if (!passwordsMatch.value) {
    fieldErrors.value.confirmPassword = 'Passwords do not match'
    return
  }

  isLoading.value = true

  const { success, error } = await signUp(form)

  isLoading.value = false

  if (success) {
    await navigateTo('/dashboard')
  } else if (error) {
    errorMessage.value = error.error
    if (error.details) {
      fieldErrors.value = error.details
    }
  }
}
</script>

<template>
  <div class="min-h-screen bg-white dark:bg-black">
    <AppNavbar />

    <div class="min-h-screen flex items-center justify-center px-4 pt-16">
      <Card class="w-full max-w-md">
        <CardHeader class="text-center">
          <CardTitle class="text-2xl">Create an account</CardTitle>
          <CardDescription>Enter your details to get started</CardDescription>
        </CardHeader>
        <CardContent>
          <form @submit.prevent="handleSubmit" class="space-y-4">
            <Alert v-if="errorMessage" variant="destructive">
              <AlertDescription>{{ errorMessage }}</AlertDescription>
            </Alert>

            <div class="grid grid-cols-2 gap-4">
              <div class="space-y-2">
                <Label for="first_name">First Name</Label>
                <Input
                  id="first_name"
                  v-model="form.first_name"
                  placeholder="John"
                />
              </div>
              <div class="space-y-2">
                <Label for="last_name">Last Name</Label>
                <Input
                  id="last_name"
                  v-model="form.last_name"
                  placeholder="Doe"
                />
              </div>
            </div>

            <div class="space-y-2">
              <Label for="username">Username</Label>
              <Input
                id="username"
                v-model="form.username"
                placeholder="johndoe"
                required
                :class="{ 'border-destructive': fieldErrors.username }"
              />
              <p v-if="fieldErrors.username" class="text-sm text-destructive">
                {{ fieldErrors.username }}
              </p>
            </div>

            <div class="space-y-2">
              <Label for="email">Email</Label>
              <Input
                id="email"
                type="email"
                v-model="form.email"
                placeholder="john@example.com"
                required
                :class="{ 'border-destructive': fieldErrors.email }"
              />
              <p v-if="fieldErrors.email" class="text-sm text-destructive">
                {{ fieldErrors.email }}
              </p>
            </div>

            <div class="space-y-2">
              <Label for="password">Password</Label>
              <Input
                id="password"
                type="password"
                v-model="form.password"
                placeholder="********"
                required
                :class="{ 'border-destructive': fieldErrors.password }"
              />
              <p v-if="fieldErrors.password" class="text-sm text-destructive">
                {{ fieldErrors.password }}
              </p>
            </div>

            <div class="space-y-2">
              <Label for="confirmPassword">Confirm Password</Label>
              <Input
                id="confirmPassword"
                type="password"
                v-model="confirmPassword"
                placeholder="********"
                required
                :class="{ 'border-destructive': !passwordsMatch && confirmPassword.length > 0 }"
              />
              <p v-if="!passwordsMatch && confirmPassword.length > 0" class="text-sm text-destructive">
                Passwords do not match
              </p>
            </div>

            <button
              type="submit"
              :disabled="isLoading"
              class="inline-flex items-center justify-center gap-2 whitespace-nowrap rounded-md text-sm font-medium ring-offset-background transition-colors focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2 disabled:pointer-events-none disabled:opacity-50 bg-primary text-primary-foreground hover:bg-primary/90 h-10 px-4 py-2 w-full"
            >
              <Loader2 v-if="isLoading" class="mr-2 h-4 w-4 animate-spin" />
              {{ isLoading ? 'Creating account...' : 'Sign Up' }}
            </button>
          </form>

          <div class="mt-6 text-center text-sm text-muted-foreground">
            Already have an account?
            <NuxtLink to="/signin" class="text-primary hover:underline">
              Sign in
            </NuxtLink>
          </div>
        </CardContent>
      </Card>
    </div>
  </div>
</template>
