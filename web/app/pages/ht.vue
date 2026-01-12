<script setup lang="ts">
import { RefreshCw, ArrowLeft, CheckCircle, XCircle } from 'lucide-vue-next'

interface HealthCheck {
  all: boolean
  db: boolean
  api: boolean
}

const { data: health, status, refresh } = useFetch<HealthCheck>('/api/v1/ht', {
  server: false, // Only fetch on client-side, not during SSR
})
const isRefreshing = ref(false)

async function handleRefresh() {
  isRefreshing.value = true
  await refresh()
  setTimeout(() => {
    isRefreshing.value = false
  }, 500)
}

useHead({
  title: 'Status - HyperWhisper'
})
</script>

<template>
  <div class="min-h-screen bg-white dark:bg-black text-neutral-900 dark:text-white transition-colors duration-300">
    <!-- Subtle grain texture overlay -->
    <div class="fixed inset-0 opacity-[0.02] dark:opacity-[0.015] pointer-events-none bg-[url('data:image/svg+xml;base64,PHN2ZyB4bWxucz0iaHR0cDovL3d3dy53My5vcmcvMjAwMC9zdmciIHdpZHRoPSIzMDAiIGhlaWdodD0iMzAwIj48ZmlsdGVyIGlkPSJhIiB4PSIwIiB5PSIwIj48ZmVUdXJidWxlbmNlIGJhc2VGcmVxdWVuY3k9Ii43NSIgc3RpdGNoVGlsZXM9InN0aXRjaCIgdHlwZT0iZnJhY3RhbE5vaXNlIi8+PC9maWx0ZXI+PHJlY3Qgd2lkdGg9IjMwMCIgaGVpZ2h0PSIzMDAiIGZpbHRlcj0idXJsKCNhKSIgb3BhY2l0eT0iMSIvPjwvc3ZnPg==')]" />

    <AppNavbar />

    <main class="min-h-screen flex items-center justify-center px-4 pt-16">
      <div class="w-full max-w-md">
        <div class="text-center mb-8">
          <h1 class="text-2xl sm:text-3xl font-bold mb-2">System Status</h1>
          <p class="text-sm text-neutral-500">Real-time health overview</p>
        </div>

        <div class="rounded-xl border border-neutral-200 dark:border-white/10 bg-neutral-50 dark:bg-white/[0.02] backdrop-blur-sm overflow-hidden">
          <!-- Loading State (idle = before client fetch starts, pending = during fetch) -->
          <div v-if="status === 'pending' || status === 'idle'" class="divide-y divide-neutral-200 dark:divide-white/5">
            <div v-for="i in 2" :key="i" class="flex items-center justify-between p-4">
              <Skeleton class="h-4 w-20" />
              <Skeleton class="h-5 w-24" />
            </div>
            <div class="flex items-center justify-between p-4 bg-neutral-100 dark:bg-white/[0.02]">
              <Skeleton class="h-4 w-24" />
              <Skeleton class="h-5 w-28" />
            </div>
          </div>

          <!-- Error State -->
          <div v-else-if="status === 'error'" class="p-6">
            <div class="text-center py-8">
              <XCircle class="size-12 text-red-500 mx-auto mb-4" />
              <p class="text-neutral-500 dark:text-neutral-400">Failed to fetch status</p>
            </div>
          </div>

          <!-- Success State -->
          <div v-else-if="health" class="divide-y divide-neutral-200 dark:divide-white/5">
            <div class="flex items-center justify-between p-4">
              <span class="text-sm font-medium">API Server</span>
              <div class="flex items-center gap-2">
                <span :class="health.api ? 'text-emerald-600 dark:text-emerald-400' : 'text-red-600 dark:text-red-400'" class="text-sm">
                  {{ health.api ? 'Operational' : 'Down' }}
                </span>
                <CheckCircle v-if="health.api" class="size-4 text-emerald-600 dark:text-emerald-400" />
                <XCircle v-else class="size-4 text-red-600 dark:text-red-400" />
              </div>
            </div>

            <div class="flex items-center justify-between p-4">
              <span class="text-sm font-medium">Database</span>
              <div class="flex items-center gap-2">
                <span :class="health.db ? 'text-emerald-600 dark:text-emerald-400' : 'text-red-600 dark:text-red-400'" class="text-sm">
                  {{ health.db ? 'Operational' : 'Down' }}
                </span>
                <CheckCircle v-if="health.db" class="size-4 text-emerald-600 dark:text-emerald-400" />
                <XCircle v-else class="size-4 text-red-600 dark:text-red-400" />
              </div>
            </div>

            <div class="flex items-center justify-between p-4 bg-neutral-100 dark:bg-white/[0.02]">
              <span class="text-sm font-medium">Overall Status</span>
              <div class="flex items-center gap-2">
                <span :class="health.all ? 'text-emerald-600 dark:text-emerald-400' : 'text-red-600 dark:text-red-400'" class="text-sm font-medium">
                  {{ health.all ? 'All Systems Go' : 'Issues Detected' }}
                </span>
                <CheckCircle v-if="health.all" class="size-4 text-emerald-600 dark:text-emerald-400" />
                <XCircle v-else class="size-4 text-red-600 dark:text-red-400" />
              </div>
            </div>
          </div>
        </div>

        <!-- Actions -->
        <div class="flex flex-col sm:flex-row gap-3 mt-6">
          <Button
            variant="outline"
            class="flex-1 border-neutral-300 dark:border-white/10 hover:bg-neutral-100 dark:hover:bg-white/5 gap-2"
            @click="handleRefresh"
            :disabled="isRefreshing"
          >
            <RefreshCw :class="['size-4', isRefreshing && 'animate-spin']" />
            Refresh
          </Button>
          <Button
            variant="ghost"
            class="flex-1 text-neutral-500 dark:text-neutral-400 hover:text-neutral-900 dark:hover:text-white hover:bg-neutral-100 dark:hover:bg-white/5 gap-2"
            as-child
          >
            <NuxtLink to="/">
              <ArrowLeft class="size-4" />
              Back to Home
            </NuxtLink>
          </Button>
        </div>
      </div>
    </main>
  </div>
</template>
