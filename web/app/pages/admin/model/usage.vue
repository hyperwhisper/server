<script setup lang="ts">
import { RefreshCw, ChevronLeft, ChevronRight, BarChart3, Key, Activity } from 'lucide-vue-next'
import type {
  SystemUsageSummary,
  AdminTranscriptionLog,
  AdminAPIKey,
  PaginatedResponse
} from '~/types/deepgram'
import type { ApiError } from '~/types/auth'

definePageMeta({
  middleware: 'admin'
})

useHead({
  title: 'Usage - Admin - HyperWhisper'
})

const { getAuthHeaders } = useAuth()

// Usage summary state
const usage = ref<SystemUsageSummary | null>(null)
const isLoadingUsage = ref(false)

// Logs state
const logs = ref<AdminTranscriptionLog[]>([])
const logsTotal = ref(0)
const logsPage = ref(1)
const logsPerPage = ref(10)
const logsTotalPages = ref(0)
const isLoadingLogs = ref(false)
const logsError = ref<string | null>(null)

// API Keys state
const apiKeys = ref<AdminAPIKey[]>([])
const keysTotal = ref(0)
const keysPage = ref(1)
const keysPerPage = ref(10)
const keysTotalPages = ref(0)
const isLoadingKeys = ref(false)
const keysError = ref<string | null>(null)

// Active tab
const activeTab = ref<'logs' | 'keys'>('logs')

// Fetch usage summary
async function fetchUsage() {
  isLoadingUsage.value = true

  try {
    const response = await $fetch<SystemUsageSummary>('/api/v1/admin/deepgram/usage', {
      headers: getAuthHeaders(),
      credentials: 'include'
    })
    usage.value = response
  } catch (e: any) {
    console.error('Failed to fetch usage:', e)
  } finally {
    isLoadingUsage.value = false
  }
}

// Fetch transcription logs
async function fetchLogs() {
  isLoadingLogs.value = true
  logsError.value = null

  try {
    const response = await $fetch<PaginatedResponse<AdminTranscriptionLog>>('/api/v1/admin/deepgram/logs', {
      headers: getAuthHeaders(),
      query: { page: logsPage.value, per_page: logsPerPage.value },
      credentials: 'include'
    })

    logs.value = response.data
    logsTotal.value = response.total
    logsTotalPages.value = response.total_pages
  } catch (e: any) {
    const apiError = e.data as ApiError
    logsError.value = apiError?.error || 'Failed to fetch logs'
  } finally {
    isLoadingLogs.value = false
  }
}

// Fetch API keys
async function fetchKeys() {
  isLoadingKeys.value = true
  keysError.value = null

  try {
    const response = await $fetch<PaginatedResponse<AdminAPIKey>>('/api/v1/admin/deepgram/keys', {
      headers: getAuthHeaders(),
      query: { page: keysPage.value, per_page: keysPerPage.value },
      credentials: 'include'
    })

    apiKeys.value = response.data
    keysTotal.value = response.total
    keysTotalPages.value = response.total_pages
  } catch (e: any) {
    const apiError = e.data as ApiError
    keysError.value = apiError?.error || 'Failed to fetch API keys'
  } finally {
    isLoadingKeys.value = false
  }
}

// Pagination for logs
function nextLogsPage() {
  if (logsPage.value < logsTotalPages.value) {
    logsPage.value++
    fetchLogs()
  }
}

function prevLogsPage() {
  if (logsPage.value > 1) {
    logsPage.value--
    fetchLogs()
  }
}

// Pagination for keys
function nextKeysPage() {
  if (keysPage.value < keysTotalPages.value) {
    keysPage.value++
    fetchKeys()
  }
}

function prevKeysPage() {
  if (keysPage.value > 1) {
    keysPage.value--
    fetchKeys()
  }
}

// Format bytes
function formatBytes(bytes: number): string {
  if (bytes === 0) return '0 B'
  const k = 1024
  const sizes = ['B', 'KB', 'MB', 'GB']
  const i = Math.floor(Math.log(bytes) / Math.log(k))
  return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i]
}

// Format duration
function formatDuration(seconds: number): string {
  if (seconds < 60) return `${seconds.toFixed(1)}s`
  const minutes = Math.floor(seconds / 60)
  const remainingSeconds = seconds % 60
  if (minutes < 60) return `${minutes}m ${remainingSeconds.toFixed(0)}s`
  const hours = Math.floor(minutes / 60)
  const remainingMinutes = minutes % 60
  return `${hours}h ${remainingMinutes}m`
}

// Status badge variant
function getStatusVariant(status: string): 'default' | 'secondary' | 'destructive' | 'outline' {
  switch (status) {
    case 'completed':
      return 'default'
    case 'active':
      return 'secondary'
    case 'error':
    case 'timeout':
      return 'destructive'
    default:
      return 'outline'
  }
}

// Refresh all data
function refreshAll() {
  fetchUsage()
  fetchLogs()
  fetchKeys()
}

onMounted(() => {
  fetchUsage()
  fetchLogs()
  fetchKeys()
})
</script>

<template>
  <div class="min-h-screen bg-white dark:bg-black">
    <AppNavbar />
    <AdminSidebar />

    <main class="container mx-auto px-4 py-12 pt-24 pr-20">
      <div class="max-w-6xl mx-auto">
        <!-- Header -->
        <div class="flex items-center justify-between mb-8">
          <div>
            <h1 class="text-3xl font-bold mb-2">Usage Statistics</h1>
            <p class="text-neutral-600 dark:text-neutral-400">
              System-wide usage and API key management
            </p>
          </div>
          <Button variant="outline" size="sm" @click="refreshAll">
            <RefreshCw class="size-4 mr-2" />
            Refresh All
          </Button>
        </div>

        <!-- Usage Summary Cards -->
        <div class="grid grid-cols-4 gap-4 mb-8">
          <Card>
            <CardContent class="p-6 text-center">
              <div v-if="isLoadingUsage" class="text-neutral-500">Loading...</div>
              <template v-else-if="usage">
                <div class="text-3xl font-bold">{{ usage.unique_users }}</div>
                <div class="text-sm text-neutral-500 dark:text-neutral-400 mt-1">Active Users</div>
              </template>
            </CardContent>
          </Card>
          <Card>
            <CardContent class="p-6 text-center">
              <div v-if="isLoadingUsage" class="text-neutral-500">Loading...</div>
              <template v-else-if="usage">
                <div class="text-3xl font-bold">{{ usage.total_sessions }}</div>
                <div class="text-sm text-neutral-500 dark:text-neutral-400 mt-1">Total Sessions</div>
              </template>
            </CardContent>
          </Card>
          <Card>
            <CardContent class="p-6 text-center">
              <div v-if="isLoadingUsage" class="text-neutral-500">Loading...</div>
              <template v-else-if="usage">
                <div class="text-3xl font-bold">{{ formatDuration(usage.total_duration_seconds) }}</div>
                <div class="text-sm text-neutral-500 dark:text-neutral-400 mt-1">Audio Duration</div>
              </template>
            </CardContent>
          </Card>
          <Card>
            <CardContent class="p-6 text-center">
              <div v-if="isLoadingUsage" class="text-neutral-500">Loading...</div>
              <template v-else-if="usage">
                <div class="text-3xl font-bold">{{ formatBytes(usage.total_bytes_sent) }}</div>
                <div class="text-sm text-neutral-500 dark:text-neutral-400 mt-1">Data Transferred</div>
              </template>
            </CardContent>
          </Card>
        </div>

        <!-- Tabs -->
        <div class="flex gap-2 mb-4">
          <Button
            :variant="activeTab === 'logs' ? 'default' : 'outline'"
            size="sm"
            @click="activeTab = 'logs'"
          >
            <Activity class="size-4 mr-2" />
            Transcription Logs
          </Button>
          <Button
            :variant="activeTab === 'keys' ? 'default' : 'outline'"
            size="sm"
            @click="activeTab = 'keys'"
          >
            <Key class="size-4 mr-2" />
            API Keys
          </Button>
        </div>

        <!-- Transcription Logs Table -->
        <Card v-if="activeTab === 'logs'">
          <CardHeader>
            <div class="flex items-center justify-between">
              <div class="flex items-center gap-2">
                <Activity class="size-5" />
                <CardTitle>All Transcription Logs</CardTitle>
              </div>
              <Button variant="outline" size="sm" @click="fetchLogs" :disabled="isLoadingLogs">
                <RefreshCw :class="['size-4 mr-2', isLoadingLogs && 'animate-spin']" />
                Refresh
              </Button>
            </div>
          </CardHeader>
          <CardContent class="p-0">
            <Alert v-if="logsError" variant="destructive" class="m-4">
              <AlertDescription>{{ logsError }}</AlertDescription>
            </Alert>

            <div class="overflow-x-auto">
              <table class="w-full">
                <thead class="border-b border-neutral-200 dark:border-white/10">
                  <tr class="text-left text-sm text-neutral-500 dark:text-neutral-400">
                    <th class="p-4 font-medium">User</th>
                    <th class="p-4 font-medium">API Key</th>
                    <th class="p-4 font-medium">Started</th>
                    <th class="p-4 font-medium">Duration</th>
                    <th class="p-4 font-medium">Status</th>
                    <th class="p-4 font-medium">Data</th>
                  </tr>
                </thead>
                <tbody class="divide-y divide-neutral-200 dark:divide-white/10">
                  <tr v-if="isLoadingLogs">
                    <td colspan="6" class="p-8 text-center text-neutral-500">
                      Loading...
                    </td>
                  </tr>
                  <tr v-else-if="logs.length === 0">
                    <td colspan="6" class="p-8 text-center text-neutral-500">
                      No transcription logs found
                    </td>
                  </tr>
                  <tr v-else v-for="log in logs" :key="log.id" class="hover:bg-neutral-50 dark:hover:bg-white/5">
                    <td class="p-4">
                      <div class="font-medium">{{ log.username }}</div>
                      <div class="text-sm text-neutral-500">{{ log.email }}</div>
                    </td>
                    <td class="p-4 text-sm">{{ log.api_key_name }}</td>
                    <td class="p-4 text-sm text-neutral-500">
                      {{ new Date(log.started_at).toLocaleString() }}
                    </td>
                    <td class="p-4 text-sm">
                      {{ log.duration_seconds ? formatDuration(log.duration_seconds) : '-' }}
                    </td>
                    <td class="p-4">
                      <Badge :variant="getStatusVariant(log.status)">
                        {{ log.status }}
                      </Badge>
                    </td>
                    <td class="p-4 text-sm">{{ formatBytes(log.bytes_sent) }}</td>
                  </tr>
                </tbody>
              </table>
            </div>

            <!-- Pagination -->
            <div v-if="logsTotalPages > 1" class="flex items-center justify-between p-4 border-t border-neutral-200 dark:border-white/10">
              <span class="text-sm text-neutral-500">
                Showing {{ (logsPage - 1) * logsPerPage + 1 }} to {{ Math.min(logsPage * logsPerPage, logsTotal) }} of {{ logsTotal }} logs
              </span>
              <div class="flex gap-2">
                <Button variant="outline" size="sm" @click="prevLogsPage" :disabled="logsPage <= 1">
                  <ChevronLeft class="size-4" />
                </Button>
                <Button variant="outline" size="sm" @click="nextLogsPage" :disabled="logsPage >= logsTotalPages">
                  <ChevronRight class="size-4" />
                </Button>
              </div>
            </div>
          </CardContent>
        </Card>

        <!-- API Keys Table -->
        <Card v-if="activeTab === 'keys'">
          <CardHeader>
            <div class="flex items-center justify-between">
              <div class="flex items-center gap-2">
                <Key class="size-5" />
                <CardTitle>All API Keys</CardTitle>
              </div>
              <Button variant="outline" size="sm" @click="fetchKeys" :disabled="isLoadingKeys">
                <RefreshCw :class="['size-4 mr-2', isLoadingKeys && 'animate-spin']" />
                Refresh
              </Button>
            </div>
          </CardHeader>
          <CardContent class="p-0">
            <Alert v-if="keysError" variant="destructive" class="m-4">
              <AlertDescription>{{ keysError }}</AlertDescription>
            </Alert>

            <div class="overflow-x-auto">
              <table class="w-full">
                <thead class="border-b border-neutral-200 dark:border-white/10">
                  <tr class="text-left text-sm text-neutral-500 dark:text-neutral-400">
                    <th class="p-4 font-medium">User</th>
                    <th class="p-4 font-medium">Key Name</th>
                    <th class="p-4 font-medium">Key Prefix</th>
                    <th class="p-4 font-medium">Created</th>
                    <th class="p-4 font-medium">Last Used</th>
                    <th class="p-4 font-medium">Status</th>
                  </tr>
                </thead>
                <tbody class="divide-y divide-neutral-200 dark:divide-white/10">
                  <tr v-if="isLoadingKeys">
                    <td colspan="6" class="p-8 text-center text-neutral-500">
                      Loading...
                    </td>
                  </tr>
                  <tr v-else-if="apiKeys.length === 0">
                    <td colspan="6" class="p-8 text-center text-neutral-500">
                      No API keys found
                    </td>
                  </tr>
                  <tr v-else v-for="key in apiKeys" :key="key.id" class="hover:bg-neutral-50 dark:hover:bg-white/5">
                    <td class="p-4">
                      <div class="font-medium">{{ key.username }}</div>
                      <div class="text-sm text-neutral-500">{{ key.email }}</div>
                    </td>
                    <td class="p-4 font-medium">{{ key.name }}</td>
                    <td class="p-4">
                      <code class="text-sm bg-neutral-100 dark:bg-white/10 px-2 py-1 rounded">
                        {{ key.key_prefix }}...
                      </code>
                    </td>
                    <td class="p-4 text-sm text-neutral-500">
                      {{ new Date(key.created_at).toLocaleDateString() }}
                    </td>
                    <td class="p-4 text-sm text-neutral-500">
                      {{ key.last_used_at ? new Date(key.last_used_at).toLocaleDateString() : 'Never' }}
                    </td>
                    <td class="p-4">
                      <Badge :variant="key.revoked_at ? 'destructive' : 'default'">
                        {{ key.revoked_at ? 'Revoked' : 'Active' }}
                      </Badge>
                    </td>
                  </tr>
                </tbody>
              </table>
            </div>

            <!-- Pagination -->
            <div v-if="keysTotalPages > 1" class="flex items-center justify-between p-4 border-t border-neutral-200 dark:border-white/10">
              <span class="text-sm text-neutral-500">
                Showing {{ (keysPage - 1) * keysPerPage + 1 }} to {{ Math.min(keysPage * keysPerPage, keysTotal) }} of {{ keysTotal }} keys
              </span>
              <div class="flex gap-2">
                <Button variant="outline" size="sm" @click="prevKeysPage" :disabled="keysPage <= 1">
                  <ChevronLeft class="size-4" />
                </Button>
                <Button variant="outline" size="sm" @click="nextKeysPage" :disabled="keysPage >= keysTotalPages">
                  <ChevronRight class="size-4" />
                </Button>
              </div>
            </div>
          </CardContent>
        </Card>
      </div>
    </main>
  </div>
</template>
