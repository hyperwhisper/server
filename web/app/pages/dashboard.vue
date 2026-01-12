<script setup lang="ts">
definePageMeta({
  middleware: 'auth'
})

useHead({
  title: 'Dashboard - HyperWhisper'
})

const { user } = useAuth()
</script>

<template>
  <div class="min-h-screen bg-white dark:bg-black">
    <AppNavbar />
    <AdminSidebar />

    <main class="container mx-auto px-4 py-12 pt-24 pr-20">
      <div class="max-w-2xl mx-auto text-center">
        <h1 class="text-4xl font-bold mb-4">Welcome to hyperwhisper</h1>
        <p class="text-neutral-600 dark:text-neutral-400 mb-8">
          Hello, {{ user?.first_name || user?.username }}! You're now signed in.
        </p>

        <Card>
          <CardHeader>
            <CardTitle>Your Profile</CardTitle>
          </CardHeader>
          <CardContent class="text-left space-y-2">
            <div class="flex justify-between">
              <span class="text-neutral-500 dark:text-neutral-400">Username:</span>
              <span>{{ user?.username }}</span>
            </div>
            <div class="flex justify-between">
              <span class="text-neutral-500 dark:text-neutral-400">Email:</span>
              <span>{{ user?.email }}</span>
            </div>
            <div class="flex justify-between">
              <span class="text-neutral-500 dark:text-neutral-400">Name:</span>
              <span>{{ user?.first_name }} {{ user?.last_name }}</span>
            </div>
            <div class="flex justify-between">
              <span class="text-neutral-500 dark:text-neutral-400">Account Type:</span>
              <Badge>{{ user?.user_type }}</Badge>
            </div>
            <div class="flex justify-between">
              <span class="text-neutral-500 dark:text-neutral-400">Member since:</span>
              <span>{{ new Date(user?.created_at || '').toLocaleDateString() }}</span>
            </div>
          </CardContent>
        </Card>
      </div>
    </main>
  </div>
</template>
