<script setup lang="ts">
import { Github, Activity, LogOut, Menu, X, LayoutDashboard, FileText, Users, Key, BarChart3 } from 'lucide-vue-next'

const route = useRoute()
const { isAuthenticated, isInitialized, signOut, initAuth, user } = useAuth()

// Mobile menu state
const mobileMenuOpen = ref(false)

// Close mobile menu on route change
watch(() => route.path, () => {
  mobileMenuOpen.value = false
})

// Initialize auth on component mount
onMounted(async () => {
  if (!isInitialized.value) {
    await initAuth()
  }
})

const handleSignOut = async () => {
  mobileMenuOpen.value = false
  await signOut()
}

const isOnDashboard = computed(() => route.path === '/dashboard')
const isAdmin = computed(() => user.value?.user_type === 'admin')

const userLinks = [
  { to: '/dashboard', label: 'Dashboard', icon: LayoutDashboard },
  { to: '/docs', label: 'Docs', icon: FileText },
]

const adminLinks = [
  { to: '/admin/model/users', label: 'Users', icon: Users },
  { to: '/admin/model/tokens', label: 'Tokens', icon: Key },
  { to: '/admin/model/usage', label: 'Usage', icon: BarChart3 },
]
</script>

<template>
  <header class="fixed top-0 left-0 right-0 z-50 border-b border-neutral-200 dark:border-white/5 bg-white/80 dark:bg-black/80 backdrop-blur-sm">
    <div class="container mx-auto px-4 py-4 flex items-center justify-between">
      <NuxtLink to="/" class="text-lg font-medium tracking-tight">
        hyperwhisper
      </NuxtLink>
      <nav class="flex items-center gap-2 sm:gap-4">
        <!-- Desktop nav items -->
        <a
          href="https://github.com/hyperwhisper"
          target="_blank"
          rel="noopener"
          class="hidden sm:block p-2 text-neutral-500 dark:text-neutral-400 hover:text-neutral-900 dark:hover:text-white transition-colors"
          title="GitHub"
        >
          <Github class="size-5" />
        </a>
        <NuxtLink
          to="/health"
          class="hidden sm:block p-2 text-neutral-500 dark:text-neutral-400 hover:text-neutral-900 dark:hover:text-white transition-colors"
          title="Status"
        >
          <Activity class="size-5" />
        </NuxtLink>
        <ThemeToggle />

        <!-- Desktop Auth button -->
        <template v-if="isInitialized">
          <template v-if="isAuthenticated">
            <template v-if="isOnDashboard">
              <Button size="sm" variant="ghost" @click="handleSignOut" class="hidden sm:flex">
                <LogOut class="size-4 mr-2" />
                Sign Out
              </Button>
            </template>
            <template v-else>
              <NuxtLink to="/dashboard" class="hidden sm:block">
                <Button size="sm">
                  Dashboard
                </Button>
              </NuxtLink>
            </template>
          </template>
          <template v-else>
            <NuxtLink to="/signin" class="hidden sm:block">
              <Button size="sm">
                Sign In
              </Button>
            </NuxtLink>
          </template>
        </template>

        <!-- Mobile menu button -->
        <button
          @click="mobileMenuOpen = !mobileMenuOpen"
          class="md:hidden p-2 text-neutral-500 dark:text-neutral-400 hover:text-neutral-900 dark:hover:text-white transition-colors"
        >
          <X v-if="mobileMenuOpen" class="size-5" />
          <Menu v-else class="size-5" />
        </button>
      </nav>
    </div>

    <!-- Mobile menu -->
    <div
      v-if="mobileMenuOpen"
      class="md:hidden border-t border-neutral-200 dark:border-white/10 bg-white dark:bg-black"
    >
      <div class="container mx-auto px-4 py-4 space-y-4">
        <!-- User links (when authenticated) -->
        <template v-if="isAuthenticated">
          <div class="space-y-1">
            <p class="text-xs font-medium text-neutral-500 dark:text-neutral-400 uppercase tracking-wider px-3 py-1">
              Navigation
            </p>
            <NuxtLink
              v-for="link in userLinks"
              :key="link.to"
              :to="link.to"
              class="flex items-center gap-3 px-3 py-2 rounded-lg text-neutral-700 dark:text-neutral-300 hover:bg-neutral-100 dark:hover:bg-white/10 transition-colors"
            >
              <component :is="link.icon" class="size-5" />
              {{ link.label }}
            </NuxtLink>
          </div>

          <!-- Admin links -->
          <div v-if="isAdmin" class="space-y-1 pt-2 border-t border-neutral-200 dark:border-white/10">
            <p class="text-xs font-medium text-neutral-500 dark:text-neutral-400 uppercase tracking-wider px-3 py-1">
              Admin
            </p>
            <NuxtLink
              v-for="link in adminLinks"
              :key="link.to"
              :to="link.to"
              class="flex items-center gap-3 px-3 py-2 rounded-lg text-neutral-700 dark:text-neutral-300 hover:bg-neutral-100 dark:hover:bg-white/10 transition-colors"
            >
              <component :is="link.icon" class="size-5" />
              {{ link.label }}
            </NuxtLink>
          </div>
        </template>

        <!-- General links -->
        <div class="space-y-1 pt-2 border-t border-neutral-200 dark:border-white/10">
          <a
            href="https://github.com/hyperwhisper"
            target="_blank"
            rel="noopener"
            class="flex items-center gap-3 px-3 py-2 rounded-lg text-neutral-700 dark:text-neutral-300 hover:bg-neutral-100 dark:hover:bg-white/10 transition-colors"
          >
            <Github class="size-5" />
            GitHub
          </a>
          <NuxtLink
            to="/health"
            class="flex items-center gap-3 px-3 py-2 rounded-lg text-neutral-700 dark:text-neutral-300 hover:bg-neutral-100 dark:hover:bg-white/10 transition-colors"
          >
            <Activity class="size-5" />
            Status
          </NuxtLink>
        </div>

        <!-- Auth actions -->
        <div class="pt-2 border-t border-neutral-200 dark:border-white/10">
          <template v-if="isAuthenticated">
            <button
              @click="handleSignOut"
              class="flex items-center gap-3 px-3 py-2 rounded-lg text-red-600 dark:text-red-400 hover:bg-red-50 dark:hover:bg-red-950/50 transition-colors w-full"
            >
              <LogOut class="size-5" />
              Sign Out
            </button>
          </template>
          <template v-else>
            <NuxtLink
              to="/signin"
              class="flex items-center gap-3 px-3 py-2 rounded-lg text-neutral-700 dark:text-neutral-300 hover:bg-neutral-100 dark:hover:bg-white/10 transition-colors"
            >
              Sign In
            </NuxtLink>
          </template>
        </div>
      </div>
    </div>
  </header>
</template>
