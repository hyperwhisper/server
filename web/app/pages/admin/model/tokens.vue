<script setup lang="ts">
import { Ban, Trash2, ChevronLeft, ChevronRight, RefreshCw } from 'lucide-vue-next'
import type { ApiError } from '~/types/auth'

definePageMeta({
  middleware: 'admin'
})

useHead({
  title: 'Tokens - Admin - HyperWhisper'
})

interface Token {
  id: string
  token_jti: string
  user_id: string
  issued_at: string
  expires_at: string
  revoked_at: string | null
  revoked_reason: string | null
}

interface PaginatedTokens {
  data: Token[]
  total: number
  page: number
  per_page: number
  total_pages: number
}

const { getAuthHeaders } = useAuth()

// State
const tokens = ref<Token[]>([])
const total = ref(0)
const page = ref(1)
const perPage = ref(20)
const totalPages = ref(0)
const isLoading = ref(false)
const error = ref<string | null>(null)
const successMessage = ref<string | null>(null)

// Revoke token dialog
const showRevokeDialog = ref(false)
const tokenToRevoke = ref<Token | null>(null)
const revokeReason = ref('')
const isRevoking = ref(false)

// Cleanup state
const isCleaningUp = ref(false)

// Fetch tokens
async function fetchTokens() {
  isLoading.value = true
  error.value = null
  successMessage.value = null

  try {
    const response = await $fetch<PaginatedTokens>('/api/v1/admin/tokens', {
      headers: getAuthHeaders(),
      query: { page: page.value, per_page: perPage.value }
    })

    tokens.value = response.data
    total.value = response.total
    totalPages.value = response.total_pages
  } catch (e: any) {
    const apiError = e.data as ApiError
    error.value = apiError?.error || 'Failed to fetch tokens'
  } finally {
    isLoading.value = false
  }
}

// Revoke a token
async function revokeToken() {
  if (!tokenToRevoke.value) return

  isRevoking.value = true
  error.value = null

  try {
    await $fetch('/api/v1/admin/tokens/revoke', {
      method: 'POST',
      headers: getAuthHeaders(),
      body: {
        token_jti: tokenToRevoke.value.token_jti,
        reason: revokeReason.value || 'admin'
      }
    })

    showRevokeDialog.value = false
    tokenToRevoke.value = null
    revokeReason.value = ''
    successMessage.value = 'Token revoked successfully'
    await fetchTokens()
  } catch (e: any) {
    const apiError = e.data as ApiError
    error.value = apiError?.error || 'Failed to revoke token'
  } finally {
    isRevoking.value = false
  }
}

// Cleanup expired tokens
async function cleanupTokens() {
  isCleaningUp.value = true
  error.value = null
  successMessage.value = null

  try {
    await $fetch('/api/v1/admin/tokens/cleanup', {
      method: 'POST',
      headers: getAuthHeaders()
    })

    successMessage.value = 'Expired tokens cleaned up successfully'
    await fetchTokens()
  } catch (e: any) {
    const apiError = e.data as ApiError
    error.value = apiError?.error || 'Failed to cleanup tokens'
  } finally {
    isCleaningUp.value = false
  }
}

function confirmRevoke(token: Token) {
  tokenToRevoke.value = token
  showRevokeDialog.value = true
}

function nextPage() {
  if (page.value < totalPages.value) {
    page.value++
    fetchTokens()
  }
}

function prevPage() {
  if (page.value > 1) {
    page.value--
    fetchTokens()
  }
}

function formatDate(dateStr: string) {
  return new Date(dateStr).toLocaleString()
}

function isExpired(expiresAt: string) {
  return new Date(expiresAt) < new Date()
}

function getTokenStatus(token: Token) {
  if (token.revoked_at) return 'revoked'
  if (isExpired(token.expires_at)) return 'expired'
  return 'active'
}

function getStatusVariant(status: string) {
  switch (status) {
    case 'active': return 'default'
    case 'revoked': return 'destructive'
    case 'expired': return 'secondary'
    default: return 'secondary'
  }
}

onMounted(() => {
  fetchTokens()
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
            <h1 class="text-3xl font-bold mb-2">Tokens</h1>
            <p class="text-neutral-600 dark:text-neutral-400">
              View and manage all issued JWT tokens
            </p>
          </div>
          <div class="flex gap-2">
            <Button variant="outline" size="sm" @click="fetchTokens" :disabled="isLoading">
              <RefreshCw :class="['size-4 mr-2', isLoading && 'animate-spin']" />
              Refresh
            </Button>
            <Button variant="outline" size="sm" @click="cleanupTokens" :disabled="isCleaningUp">
              <Trash2 :class="['size-4 mr-2', isCleaningUp && 'animate-pulse']" />
              {{ isCleaningUp ? 'Cleaning...' : 'Cleanup Expired' }}
            </Button>
          </div>
        </div>

        <!-- Error -->
        <Alert v-if="error" variant="destructive" class="mb-6">
          <AlertTitle>Error</AlertTitle>
          <AlertDescription>{{ error }}</AlertDescription>
        </Alert>

        <!-- Success -->
        <Alert v-if="successMessage" class="mb-6 border-green-500 text-green-700 dark:text-green-400">
          <AlertDescription>{{ successMessage }}</AlertDescription>
        </Alert>

        <!-- Table -->
        <Card>
          <CardContent class="p-0">
            <div class="overflow-x-auto">
              <table class="w-full">
                <thead class="border-b border-neutral-200 dark:border-white/10">
                  <tr class="text-left text-sm text-neutral-500 dark:text-neutral-400">
                    <th class="p-4 font-medium">Token JTI</th>
                    <th class="p-4 font-medium">User ID</th>
                    <th class="p-4 font-medium">Issued At</th>
                    <th class="p-4 font-medium">Expires At</th>
                    <th class="p-4 font-medium">Status</th>
                    <th class="p-4 font-medium">Actions</th>
                  </tr>
                </thead>
                <tbody class="divide-y divide-neutral-200 dark:divide-white/10">
                  <tr v-if="isLoading">
                    <td colspan="6" class="p-8 text-center text-neutral-500">
                      Loading...
                    </td>
                  </tr>
                  <tr v-else-if="tokens.length === 0">
                    <td colspan="6" class="p-8 text-center text-neutral-500">
                      No tokens found
                    </td>
                  </tr>
                  <tr v-else v-for="token in tokens" :key="token.id" class="hover:bg-neutral-50 dark:hover:bg-white/5">
                    <td class="p-4 font-mono text-sm">
                      {{ token.token_jti.slice(0, 8) }}...{{ token.token_jti.slice(-4) }}
                    </td>
                    <td class="p-4 font-mono text-sm">
                      {{ token.user_id.slice(0, 8) }}...
                    </td>
                    <td class="p-4 text-sm text-neutral-500">
                      {{ formatDate(token.issued_at) }}
                    </td>
                    <td class="p-4 text-sm text-neutral-500">
                      {{ formatDate(token.expires_at) }}
                    </td>
                    <td class="p-4">
                      <Badge :variant="getStatusVariant(getTokenStatus(token))">
                        {{ getTokenStatus(token) }}
                      </Badge>
                      <span v-if="token.revoked_reason" class="ml-2 text-xs text-neutral-500">
                        ({{ token.revoked_reason }})
                      </span>
                    </td>
                    <td class="p-4">
                      <Button
                        v-if="getTokenStatus(token) === 'active'"
                        variant="ghost"
                        size="sm"
                        @click="confirmRevoke(token)"
                        class="text-red-600 hover:text-red-700 hover:bg-red-50 dark:hover:bg-red-950"
                      >
                        <Ban class="size-4" />
                      </Button>
                      <span v-else class="text-neutral-400 text-sm">-</span>
                    </td>
                  </tr>
                </tbody>
              </table>
            </div>

            <!-- Pagination -->
            <div v-if="totalPages > 1" class="flex items-center justify-between p-4 border-t border-neutral-200 dark:border-white/10">
              <span class="text-sm text-neutral-500">
                Showing {{ (page - 1) * perPage + 1 }} to {{ Math.min(page * perPage, total) }} of {{ total }} tokens
              </span>
              <div class="flex gap-2">
                <Button variant="outline" size="sm" @click="prevPage" :disabled="page <= 1">
                  <ChevronLeft class="size-4" />
                </Button>
                <Button variant="outline" size="sm" @click="nextPage" :disabled="page >= totalPages">
                  <ChevronRight class="size-4" />
                </Button>
              </div>
            </div>
          </CardContent>
        </Card>

        <!-- Info Card -->
        <Card class="mt-6">
          <CardHeader>
            <CardTitle class="text-lg">About Refresh Token Management</CardTitle>
          </CardHeader>
          <CardContent class="text-sm text-neutral-600 dark:text-neutral-400 space-y-2">
            <p>
              This page shows all refresh tokens that have been issued. Access tokens are short-lived and stateless, so they are not tracked here.
            </p>
            <p>
              <strong>Revoke:</strong> Immediately invalidate a refresh token. The user will need to sign in again to get new tokens.
            </p>
            <p>
              <strong>Cleanup Expired:</strong> Removes expired refresh tokens from the database to keep it clean.
            </p>
          </CardContent>
        </Card>
      </div>
    </main>

    <!-- Revoke Token Dialog -->
    <Dialog v-model:open="showRevokeDialog">
      <DialogContent>
        <DialogHeader>
          <DialogTitle>Revoke Token</DialogTitle>
          <DialogDescription>
            Are you sure you want to revoke this token? The user will need to sign in again.
          </DialogDescription>
        </DialogHeader>

        <div class="space-y-4">
          <div class="space-y-2">
            <Label for="revoke_reason">Reason (optional)</Label>
            <Input
              id="revoke_reason"
              v-model="revokeReason"
              placeholder="e.g., suspicious activity, user request"
            />
          </div>
        </div>

        <DialogFooter>
          <Button variant="outline" @click="showRevokeDialog = false">
            Cancel
          </Button>
          <Button variant="destructive" @click="revokeToken" :disabled="isRevoking">
            {{ isRevoking ? 'Revoking...' : 'Revoke Token' }}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  </div>
</template>
