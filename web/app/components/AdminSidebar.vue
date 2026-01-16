<script setup lang="ts">
import { Users, Key, LayoutDashboard, BarChart3, FlaskConical } from 'lucide-vue-next'

const route = useRoute()
const { user } = useAuth()

const isAdmin = computed(() => user.value?.user_type === 'admin')

const adminLinks = [
  { to: '/dashboard', label: 'Dashboard', icon: LayoutDashboard },
  { to: '/admin/model/users', label: 'Users', icon: Users },
  { to: '/admin/model/tokens', label: 'Tokens', icon: Key },
  { to: '/admin/model/usage', label: 'Usage', icon: BarChart3 },
  { to: '/admin/model/trials', label: 'Trials', icon: FlaskConical },
]

function isActive(path: string) {
  return route.path === path
}
</script>

<template>
  <aside
    v-if="isAdmin"
    class="hidden md:flex fixed right-0 top-0 h-screen w-16 pt-20 border-l border-neutral-200 dark:border-white/10 bg-white dark:bg-black z-40 flex-col items-center py-4"
  >
    <nav class="flex flex-col gap-2 mt-4">
      <NuxtLink
        v-for="link in adminLinks"
        :key="link.to"
        :to="link.to"
        :title="link.label"
        :class="[
          'p-3 rounded-lg transition-colors',
          isActive(link.to)
            ? 'bg-neutral-100 dark:bg-white/10 text-neutral-900 dark:text-white'
            : 'text-neutral-500 dark:text-neutral-400 hover:bg-neutral-50 dark:hover:bg-white/5 hover:text-neutral-900 dark:hover:text-white'
        ]"
      >
        <component :is="link.icon" class="size-5" />
      </NuxtLink>
    </nav>
  </aside>
</template>
