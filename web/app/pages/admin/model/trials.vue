<script setup lang="ts">
import { RefreshCw, ChevronLeft, ChevronRight, Trash2, Settings, Sparkles } from 'lucide-vue-next'
import type { TrialAPIKey, TrialUsageSummary, TrialLimits, UpdateTrialLimitsRequest } from '~/types/trial'
import type { PaginatedResponse } from '~/types/deepgram'
import type { ApiError } from '~/types/auth'

definePageMeta({
  middleware: 'admin'
})

useHead({
  title: 'Trials - Admin - HyperWhisper'
})

const { getAuthHeaders } = useAuth()

// Usage summary state
const usage = ref<TrialUsageSummary | null>(null)
const isLoadingUsage = ref(false)

// Limits state
const limits = ref<TrialLimits | null>(null)
const isLoadingLimits = ref(false)

// Trial keys state
const keys = ref<TrialAPIKey[]>([])
const keysTotal = ref(0)
const keysPage = ref(1)
const keysPerPage = ref(10)
const keysTotalPages = ref(0)
const isLoadingKeys = ref(false)
const keysError = ref<string | null>(null)

// Edit limits dialog
const showEditLimitsDialog = ref(false)
const editLimitsForm = ref<UpdateTrialLimitsRequest>({
  max_duration_seconds: 0,
  max_sessions: 0,
  max_session_duration_seconds: 0,
  expiry_days: 0
})
const isUpdatingLimits = ref(false)
const limitsError = ref<string | null>(null)

// Revoke confirmation dialog
const showRevokeDialog = ref(false)
const keyToRevoke = ref<TrialAPIKey | null>(null)
const isRevoking = ref(false)

// Cleanup state
const isCleaningUp = ref(false)

// Fetch usage summary
async function fetchUsage() {
  isLoadingUsage.value = true

  try {
    const response = await $fetch<TrialUsageSummary>('/api/v1/admin/trial/usage', {
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

// Fetch limits
async function fetchLimits() {
  isLoadingLimits.value = true

  try {
    const response = await $fetch<TrialLimits>('/api/v1/admin/trial/limits', {
      headers: getAuthHeaders(),
      credentials: 'include'
    })
    limits.value = response
  } catch (e: any) {
    console.error('Failed to fetch limits:', e)
  } finally {
    isLoadingLimits.value = false
  }
}

// Fetch trial keys
async function fetchKeys() {
  isLoadingKeys.value = true
  keysError.value = null

  try {
    const response = await $fetch<PaginatedResponse<TrialAPIKey>>('/api/v1/admin/trial/keys', {
      headers: getAuthHeaders(),
      query: { page: keysPage.value, per_page: keysPerPage.value },
      credentials: 'include'
    })

    keys.value = response.data
    keysTotal.value = response.total
    keysTotalPages.value = response.total_pages
  } catch (e: any) {
    const apiError = e.data as ApiError
    keysError.value = apiError?.error || 'Failed to fetch trial keys'
  } finally {
    isLoadingKeys.value = false
  }
}

// Update limits
async function updateLimits() {
  isUpdatingLimits.value = true
  limitsError.value = null

  try {
    const response = await $fetch<TrialLimits>('/api/v1/admin/trial/limits', {
      method: 'PUT',
      headers: getAuthHeaders(),
      body: editLimitsForm.value,
      credentials: 'include'
    })

    limits.value = response
    showEditLimitsDialog.value = false
  } catch (e: any) {
    const apiError = e.data as ApiError
    limitsError.value = apiError?.error || 'Failed to update limits'
  } finally {
    isUpdatingLimits.value = false
  }
}

// Revoke key
async function revokeKey() {
  if (!keyToRevoke.value) return

  isRevoking.value = true

  try {
    await $fetch(`/api/v1/admin/trial/keys/${keyToRevoke.value.id}`, {
      method: 'DELETE',
      headers: getAuthHeaders(),
      credentials: 'include'
    })

    showRevokeDialog.value = false
    keyToRevoke.value = null
    await fetchKeys()
    await fetchUsage()
  } catch (e: any) {
    const apiError = e.data as ApiError
    keysError.value = apiError?.error || 'Failed to revoke key'
  } finally {
    isRevoking.value = false
  }
}

// Cleanup expired keys
async function cleanupExpired() {
  isCleaningUp.value = true

  try {
    await $fetch('/api/v1/admin/trial/cleanup', {
      method: 'POST',
      headers: getAuthHeaders(),
      credentials: 'include'
    })

    await fetchKeys()
    await fetchUsage()
  } catch (e: any) {
    const apiError = e.data as ApiError
    keysError.value = apiError?.error || 'Failed to cleanup expired keys'
  } finally {
    isCleaningUp.value = false
  }
}

function confirmRevoke(key: TrialAPIKey) {
  keyToRevoke.value = key
  showRevokeDialog.value = true
}

function openEditLimits() {
  if (limits.value) {
    editLimitsForm.value = {
      max_duration_seconds: limits.value.max_duration_seconds,
      max_sessions: limits.value.max_sessions,
      max_session_duration_seconds: limits.value.max_session_duration_seconds,
      expiry_days: limits.value.expiry_days
    }
  }
  limitsError.value = null
  showEditLimitsDialog.value = true
}

// Pagination
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
function formatDuration(seconds: number | null | undefined): string {
  if (seconds == null || seconds === 0) return '-'
  if (seconds < 60) return `${seconds.toFixed(1)}s`
  const minutes = Math.floor(seconds / 60)
  const remainingSeconds = seconds % 60
  if (minutes < 60) return `${minutes}m ${remainingSeconds.toFixed(0)}s`
  const hours = Math.floor(minutes / 60)
  const remainingMinutes = minutes % 60
  return `${hours}h ${remainingMinutes}m`
}

// Get key status
function getKeyStatus(key: TrialAPIKey): 'active' | 'expired' | 'revoked' {
  if (key.revoked_at) return 'revoked'
  if (new Date(key.expires_at) < new Date()) return 'expired'
  return 'active'
}

// Get badge variant
function getStatusVariant(status: 'active' | 'expired' | 'revoked'): 'default' | 'secondary' | 'destructive' {
  switch (status) {
    case 'active':
      return 'default'
    case 'expired':
      return 'secondary'
    case 'revoked':
      return 'destructive'
  }
}

// Refresh all data
function refreshAll() {
  fetchUsage()
  fetchLimits()
  fetchKeys()
}

onMounted(() => {
  fetchUsage()
  fetchLimits()
  fetchKeys()
})
</script>

<template>
  <div class="min-h-screen bg-white dark:bg-black">
    <AppNavbar />
    <UserSidebar />
    <AdminSidebar />

    <main class="main-content container mx-auto px-4 py-12 pt-24">
      <div class="max-w-6xl mx-auto">
        <!-- Header -->
        <div class="flex items-center justify-between mb-8">
          <div>
            <h1 class="text-3xl font-bold mb-2">Trial Management</h1>
            <p class="text-neutral-600 dark:text-neutral-400">
              Manage trial API keys and usage limits
            </p>
          </div>
          <Button variant="outline" size="sm" @click="refreshAll">
            <RefreshCw class="size-4 mr-2" />
            Refresh All
          </Button>
        </div>

        <!-- Usage Summary Cards -->
        <div class="grid grid-cols-5 gap-4 mb-8">
          <Card>
            <CardContent class="p-6 text-center">
              <div v-if="isLoadingUsage" class="text-neutral-500">Loading...</div>
              <template v-else-if="usage">
                <div class="text-3xl font-bold">{{ usage.total_trial_keys }}</div>
                <div class="text-sm text-neutral-500 dark:text-neutral-400 mt-1">Total Trial Keys</div>
              </template>
            </CardContent>
          </Card>
          <Card>
            <CardContent class="p-6 text-center">
              <div v-if="isLoadingUsage" class="text-neutral-500">Loading...</div>
              <template v-else-if="usage">
                <div class="text-3xl font-bold">{{ usage.active_trial_keys }}</div>
                <div class="text-sm text-neutral-500 dark:text-neutral-400 mt-1">Active Trial Keys</div>
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

        <!-- Trial Limits Card -->
        <Card class="mb-8">
          <CardHeader>
            <div class="flex items-center justify-between">
              <div class="flex items-center gap-2">
                <Settings class="size-5" />
                <CardTitle>Trial Limits</CardTitle>
              </div>
              <Button variant="outline" size="sm" @click="openEditLimits" :disabled="isLoadingLimits || !limits">
                Edit Limits
              </Button>
            </div>
          </CardHeader>
          <CardContent>
            <div v-if="isLoadingLimits" class="text-neutral-500">Loading...</div>
            <div v-else-if="limits" class="grid grid-cols-4 gap-6">
              <div>
                <div class="text-sm text-neutral-500 dark:text-neutral-400">Max Duration</div>
                <div class="text-lg font-medium">{{ formatDuration(limits.max_duration_seconds) }}</div>
              </div>
              <div>
                <div class="text-sm text-neutral-500 dark:text-neutral-400">Max Sessions</div>
                <div class="text-lg font-medium">{{ limits.max_sessions }}</div>
              </div>
              <div>
                <div class="text-sm text-neutral-500 dark:text-neutral-400">Max Session Duration</div>
                <div class="text-lg font-medium">{{ formatDuration(limits.max_session_duration_seconds) }}</div>
              </div>
              <div>
                <div class="text-sm text-neutral-500 dark:text-neutral-400">Expiry Days</div>
                <div class="text-lg font-medium">{{ limits.expiry_days }} days</div>
              </div>
            </div>
          </CardContent>
        </Card>

        <!-- Trial Keys Table -->
        <Card>
          <CardHeader>
            <div class="flex items-center justify-between">
              <div class="flex items-center gap-2">
                <Sparkles class="size-5" />
                <CardTitle>Trial API Keys</CardTitle>
              </div>
              <div class="flex gap-2">
                <Button variant="outline" size="sm" @click="cleanupExpired" :disabled="isCleaningUp">
                  <Trash2 :class="['size-4 mr-2', isCleaningUp && 'animate-pulse']" />
                  {{ isCleaningUp ? 'Cleaning...' : 'Cleanup Expired' }}
                </Button>
                <Button variant="outline" size="sm" @click="fetchKeys" :disabled="isLoadingKeys">
                  <RefreshCw :class="['size-4 mr-2', isLoadingKeys && 'animate-spin']" />
                  Refresh
                </Button>
              </div>
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
                    <th class="p-4 font-medium">Key Prefix</th>
                    <th class="p-4 font-medium">Device Fingerprint</th>
                    <th class="p-4 font-medium">Created</th>
                    <th class="p-4 font-medium">Expires</th>
                    <th class="p-4 font-medium">Last Used</th>
                    <th class="p-4 font-medium">Sessions</th>
                    <th class="p-4 font-medium">Duration</th>
                    <th class="p-4 font-medium">Status</th>
                    <th class="p-4 font-medium">Actions</th>
                  </tr>
                </thead>
                <tbody class="divide-y divide-neutral-200 dark:divide-white/10">
                  <tr v-if="isLoadingKeys">
                    <td colspan="9" class="p-8 text-center text-neutral-500">
                      Loading...
                    </td>
                  </tr>
                  <tr v-else-if="keys.length === 0">
                    <td colspan="9" class="p-8 text-center text-neutral-500">
                      No trial keys found
                    </td>
                  </tr>
                  <tr v-else v-for="key in keys" :key="key.id" class="hover:bg-neutral-50 dark:hover:bg-white/5">
                    <td class="p-4">
                      <code class="text-sm bg-neutral-100 dark:bg-white/10 px-2 py-1 rounded">
                        {{ key.key_prefix }}...
                      </code>
                    </td>
                    <td class="p-4 text-sm">
                      <code class="text-xs bg-neutral-100 dark:bg-white/10 px-2 py-1 rounded">
                        {{ key.device_fingerprint.slice(0, 16) }}...
                      </code>
                    </td>
                    <td class="p-4 text-sm text-neutral-500">
                      {{ new Date(key.created_at).toLocaleDateString() }}
                    </td>
                    <td class="p-4 text-sm text-neutral-500">
                      {{ new Date(key.expires_at).toLocaleDateString() }}
                    </td>
                    <td class="p-4 text-sm text-neutral-500">
                      {{ key.last_used_at ? new Date(key.last_used_at).toLocaleDateString() : 'Never' }}
                    </td>
                    <td class="p-4 text-sm">{{ key.total_sessions }}</td>
                    <td class="p-4 text-sm">{{ formatDuration(key.total_duration_seconds) }}</td>
                    <td class="p-4">
                      <Badge :variant="getStatusVariant(getKeyStatus(key))">
                        {{ getKeyStatus(key) }}
                      </Badge>
                    </td>
                    <td class="p-4">
                      <Button
                        v-if="getKeyStatus(key) === 'active'"
                        variant="ghost"
                        size="sm"
                        @click="confirmRevoke(key)"
                        class="text-red-600 hover:text-red-700 hover:bg-red-50 dark:hover:bg-red-950"
                      >
                        <Trash2 class="size-4" />
                      </Button>
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

    <!-- Edit Limits Dialog -->
    <Dialog v-model:open="showEditLimitsDialog">
      <DialogContent>
        <DialogHeader>
          <DialogTitle>Edit Trial Limits</DialogTitle>
          <DialogDescription>Update the trial limits for all new trial keys.</DialogDescription>
        </DialogHeader>

        <form @submit.prevent="updateLimits" class="space-y-4">
          <Alert v-if="limitsError" variant="destructive">
            <AlertDescription>{{ limitsError }}</AlertDescription>
          </Alert>

          <div class="grid grid-cols-2 gap-4">
            <div class="space-y-2">
              <Label for="max_duration">Max Duration (seconds)</Label>
              <Input id="max_duration" type="number" v-model.number="editLimitsForm.max_duration_seconds" required min="1" />
            </div>
            <div class="space-y-2">
              <Label for="max_sessions">Max Sessions</Label>
              <Input id="max_sessions" type="number" v-model.number="editLimitsForm.max_sessions" required min="1" />
            </div>
          </div>

          <div class="grid grid-cols-2 gap-4">
            <div class="space-y-2">
              <Label for="max_session_duration">Max Session Duration (seconds)</Label>
              <Input id="max_session_duration" type="number" v-model.number="editLimitsForm.max_session_duration_seconds" required min="1" />
            </div>
            <div class="space-y-2">
              <Label for="expiry_days">Expiry Days</Label>
              <Input id="expiry_days" type="number" v-model.number="editLimitsForm.expiry_days" required min="1" />
            </div>
          </div>

          <DialogFooter>
            <Button type="button" variant="outline" @click="showEditLimitsDialog = false">
              Cancel
            </Button>
            <Button type="submit" :disabled="isUpdatingLimits">
              {{ isUpdatingLimits ? 'Saving...' : 'Save Changes' }}
            </Button>
          </DialogFooter>
        </form>
      </DialogContent>
    </Dialog>

    <!-- Revoke Confirmation Dialog -->
    <Dialog v-model:open="showRevokeDialog">
      <DialogContent>
        <DialogHeader>
          <DialogTitle>Revoke Trial Key</DialogTitle>
          <DialogDescription>
            Are you sure you want to revoke trial key <strong>{{ keyToRevoke?.key_prefix }}...</strong>? This action cannot be undone.
          </DialogDescription>
        </DialogHeader>
        <DialogFooter>
          <Button variant="outline" @click="showRevokeDialog = false">
            Cancel
          </Button>
          <Button variant="destructive" @click="revokeKey" :disabled="isRevoking">
            {{ isRevoking ? 'Revoking...' : 'Revoke' }}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  </div>
</template>

<style scoped>
@media (min-width: 768px) {
  .main-content {
    padding-left: 5rem;
    padding-right: 5rem;
  }
}
</style>
