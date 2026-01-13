<script setup lang="ts">
import { Plus, Trash2, ChevronLeft, ChevronRight, RefreshCw } from 'lucide-vue-next'
import type { User, ApiError } from '~/types/auth'

definePageMeta({
  middleware: 'admin'
})

useHead({
  title: 'Users - Admin - HyperWhisper'
})

interface PaginatedUsers {
  data: User[]
  total: number
  page: number
  per_page: number
  total_pages: number
}

const { getAuthHeaders } = useAuth()

// State
const users = ref<User[]>([])
const total = ref(0)
const page = ref(1)
const perPage = ref(20)
const totalPages = ref(0)
const isLoading = ref(false)
const error = ref<string | null>(null)

// Create user dialog
const showCreateDialog = ref(false)
const createForm = ref({
  username: '',
  email: '',
  password: '',
  first_name: '',
  last_name: '',
  user_type: 'user'
})
const createError = ref<string | null>(null)
const isCreating = ref(false)

// Delete confirmation dialog
const showDeleteDialog = ref(false)
const userToDelete = ref<User | null>(null)
const isDeleting = ref(false)

// Fetch users
async function fetchUsers() {
  isLoading.value = true
  error.value = null

  try {
    const response = await $fetch<PaginatedUsers>('/api/v1/admin/users', {
      headers: getAuthHeaders(),
      query: { page: page.value, per_page: perPage.value }
    })

    users.value = response.data
    total.value = response.total
    totalPages.value = response.total_pages
  } catch (e: any) {
    const apiError = e.data as ApiError
    error.value = apiError?.error || 'Failed to fetch users'
  } finally {
    isLoading.value = false
  }
}

// Create user
async function createUser() {
  isCreating.value = true
  createError.value = null

  try {
    await $fetch('/api/v1/admin/users', {
      method: 'POST',
      headers: getAuthHeaders(),
      body: createForm.value
    })

    showCreateDialog.value = false
    createForm.value = {
      username: '',
      email: '',
      password: '',
      first_name: '',
      last_name: '',
      user_type: 'user'
    }
    await fetchUsers()
  } catch (e: any) {
    const apiError = e.data as ApiError
    createError.value = apiError?.error || 'Failed to create user'
  } finally {
    isCreating.value = false
  }
}

// Delete user
async function deleteUser() {
  if (!userToDelete.value) return

  isDeleting.value = true

  try {
    await $fetch(`/api/v1/admin/users/${userToDelete.value.id}`, {
      method: 'DELETE',
      headers: getAuthHeaders()
    })

    showDeleteDialog.value = false
    userToDelete.value = null
    await fetchUsers()
  } catch (e: any) {
    const apiError = e.data as ApiError
    error.value = apiError?.error || 'Failed to delete user'
  } finally {
    isDeleting.value = false
  }
}

function confirmDelete(user: User) {
  userToDelete.value = user
  showDeleteDialog.value = true
}

function nextPage() {
  if (page.value < totalPages.value) {
    page.value++
    fetchUsers()
  }
}

function prevPage() {
  if (page.value > 1) {
    page.value--
    fetchUsers()
  }
}

onMounted(() => {
  fetchUsers()
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
            <h1 class="text-3xl font-bold mb-2">Users</h1>
            <p class="text-neutral-600 dark:text-neutral-400">
              Manage user accounts
            </p>
          </div>
          <div class="flex gap-2">
            <Button variant="outline" size="sm" @click="fetchUsers" :disabled="isLoading">
              <RefreshCw :class="['size-4 mr-2', isLoading && 'animate-spin']" />
              Refresh
            </Button>
            <Button size="sm" @click="showCreateDialog = true">
              <Plus class="size-4 mr-2" />
              Create User
            </Button>
          </div>
        </div>

        <!-- Error -->
        <Alert v-if="error" variant="destructive" class="mb-6">
          <AlertTitle>Error</AlertTitle>
          <AlertDescription>{{ error }}</AlertDescription>
        </Alert>

        <!-- Table -->
        <Card>
          <CardContent class="p-0">
            <div class="overflow-x-auto">
              <table class="w-full">
                <thead class="border-b border-neutral-200 dark:border-white/10">
                  <tr class="text-left text-sm text-neutral-500 dark:text-neutral-400">
                    <th class="p-4 font-medium">Username</th>
                    <th class="p-4 font-medium">Email</th>
                    <th class="p-4 font-medium">Name</th>
                    <th class="p-4 font-medium">Type</th>
                    <th class="p-4 font-medium">Created</th>
                    <th class="p-4 font-medium">Actions</th>
                  </tr>
                </thead>
                <tbody class="divide-y divide-neutral-200 dark:divide-white/10">
                  <tr v-if="isLoading">
                    <td colspan="6" class="p-8 text-center text-neutral-500">
                      Loading...
                    </td>
                  </tr>
                  <tr v-else-if="users.length === 0">
                    <td colspan="6" class="p-8 text-center text-neutral-500">
                      No users found
                    </td>
                  </tr>
                  <tr v-else v-for="user in users" :key="user.id" class="hover:bg-neutral-50 dark:hover:bg-white/5">
                    <td class="p-4">{{ user.username }}</td>
                    <td class="p-4">{{ user.email }}</td>
                    <td class="p-4">{{ user.first_name }} {{ user.last_name }}</td>
                    <td class="p-4">
                      <Badge :variant="user.user_type === 'admin' ? 'default' : 'secondary'">
                        {{ user.user_type }}
                      </Badge>
                    </td>
                    <td class="p-4 text-sm text-neutral-500">
                      {{ new Date(user.created_at).toLocaleDateString() }}
                    </td>
                    <td class="p-4">
                      <Button variant="ghost" size="sm" @click="confirmDelete(user)" class="text-red-600 hover:text-red-700 hover:bg-red-50 dark:hover:bg-red-950">
                        <Trash2 class="size-4" />
                      </Button>
                    </td>
                  </tr>
                </tbody>
              </table>
            </div>

            <!-- Pagination -->
            <div v-if="totalPages > 1" class="flex items-center justify-between p-4 border-t border-neutral-200 dark:border-white/10">
              <span class="text-sm text-neutral-500">
                Showing {{ (page - 1) * perPage + 1 }} to {{ Math.min(page * perPage, total) }} of {{ total }} users
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
      </div>
    </main>

    <!-- Create User Dialog -->
    <Dialog v-model:open="showCreateDialog">
      <DialogContent>
        <DialogHeader>
          <DialogTitle>Create User</DialogTitle>
          <DialogDescription>Add a new user to the system.</DialogDescription>
        </DialogHeader>

        <form @submit.prevent="createUser" class="space-y-4">
          <Alert v-if="createError" variant="destructive">
            <AlertDescription>{{ createError }}</AlertDescription>
          </Alert>

          <div class="grid grid-cols-2 gap-4">
            <div class="space-y-2">
              <Label for="first_name">First Name</Label>
              <Input id="first_name" v-model="createForm.first_name" />
            </div>
            <div class="space-y-2">
              <Label for="last_name">Last Name</Label>
              <Input id="last_name" v-model="createForm.last_name" />
            </div>
          </div>

          <div class="space-y-2">
            <Label for="username">Username *</Label>
            <Input id="username" v-model="createForm.username" required />
          </div>

          <div class="space-y-2">
            <Label for="email">Email *</Label>
            <Input id="email" type="email" v-model="createForm.email" required />
          </div>

          <div class="space-y-2">
            <Label for="password">Password *</Label>
            <Input id="password" type="password" v-model="createForm.password" required />
          </div>

          <div class="space-y-2">
            <Label for="user_type">User Type</Label>
            <Select v-model="createForm.user_type">
              <SelectTrigger>
                <SelectValue />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="user">User</SelectItem>
                <SelectItem value="admin">Admin</SelectItem>
              </SelectContent>
            </Select>
          </div>

          <DialogFooter>
            <Button type="button" variant="outline" @click="showCreateDialog = false">
              Cancel
            </Button>
            <button
              type="submit"
              class="inline-flex items-center justify-center gap-2 whitespace-nowrap rounded-md text-sm font-medium transition-colors focus-visible:outline-none focus-visible:ring-1 focus-visible:ring-ring disabled:pointer-events-none disabled:opacity-50 bg-primary text-primary-foreground shadow hover:bg-primary/90 h-9 px-4 py-2"
              :disabled="isCreating"
            >
              {{ isCreating ? 'Creating...' : 'Create' }}
            </button>
          </DialogFooter>
        </form>
      </DialogContent>
    </Dialog>

    <!-- Delete Confirmation Dialog -->
    <Dialog v-model:open="showDeleteDialog">
      <DialogContent>
        <DialogHeader>
          <DialogTitle>Delete User</DialogTitle>
          <DialogDescription>
            Are you sure you want to delete user <strong>{{ userToDelete?.username }}</strong>? This action cannot be undone.
          </DialogDescription>
        </DialogHeader>
        <DialogFooter>
          <Button variant="outline" @click="showDeleteDialog = false">
            Cancel
          </Button>
          <Button variant="destructive" @click="deleteUser" :disabled="isDeleting">
            {{ isDeleting ? 'Deleting...' : 'Delete' }}
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
