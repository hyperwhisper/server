<script setup lang="ts">
import { Github, Activity, LogOut } from 'lucide-vue-next'

const { isAuthenticated, isInitialized, signOut, initAuth } = useAuth()

// Initialize auth on component mount
onMounted(async () => {
  if (!isInitialized.value) {
    await initAuth()
  }
})

const handleSignOut = async () => {
  await signOut()
}
</script>

<template>
  <header class="fixed top-0 left-0 right-0 z-50 border-b border-neutral-200 dark:border-white/5 bg-white/80 dark:bg-black/80 backdrop-blur-sm">
    <div class="container mx-auto px-4 py-4 flex items-center justify-between">
      <NuxtLink to="/" class="text-lg font-medium tracking-tight">
        hyperwhisper
      </NuxtLink>
      <nav class="flex items-center gap-2 sm:gap-4">
        <a
          href="https://github.com/hyperwhisper"
          target="_blank"
          rel="noopener"
          class="p-2 text-neutral-500 dark:text-neutral-400 hover:text-neutral-900 dark:hover:text-white transition-colors"
          title="GitHub"
        >
          <Github class="size-5" />
        </a>
        <NuxtLink
          to="/ht"
          class="p-2 text-neutral-500 dark:text-neutral-400 hover:text-neutral-900 dark:hover:text-white transition-colors"
          title="Status"
        >
          <Activity class="size-5" />
        </NuxtLink>
        <ThemeToggle />

        <!-- Auth button -->
        <template v-if="isInitialized">
          <template v-if="isAuthenticated">
            <Button size="sm" variant="ghost" @click="handleSignOut">
              <LogOut class="size-4 mr-2" />
              Sign Out
            </Button>
          </template>
          <template v-else>
            <NuxtLink to="/signin">
              <Button size="sm">
                Sign In
              </Button>
            </NuxtLink>
          </template>
        </template>
      </nav>
    </div>
  </header>
</template>
