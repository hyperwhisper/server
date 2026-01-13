<script setup lang="ts">
import { Plus, Trash2, Copy, Check, RefreshCw, Key, Activity, Mic, MicOff } from 'lucide-vue-next'
import type { APIKey, APIKeyCreated, UsageSummary, PaginatedResponse } from '~/types/deepgram'
import type { ApiError } from '~/types/auth'

definePageMeta({
  middleware: 'auth'
})

useHead({
  title: 'Dashboard - HyperWhisper'
})

const { user, getAuthHeaders } = useAuth()

// API Keys state
const apiKeys = ref<APIKey[]>([])
const isLoadingKeys = ref(false)
const keysError = ref<string | null>(null)

// Usage state
const usage = ref<UsageSummary | null>(null)
const isLoadingUsage = ref(false)

// Create key dialog
const showCreateDialog = ref(false)
const createForm = ref({ name: '' })
const createError = ref<string | null>(null)
const isCreating = ref(false)
const createdKey = ref<APIKeyCreated | null>(null)
const copiedKey = ref(false)

// Revoke dialog
const showRevokeDialog = ref(false)
const keyToRevoke = ref<APIKey | null>(null)
const isRevoking = ref(false)

// Microphone / Transcription state
const isRecording = ref(false)
const transcription = ref('')
const interimTranscription = ref('')
const micError = ref<string | null>(null)
let websocket: WebSocket | null = null
let audioContext: AudioContext | null = null
let mediaStream: MediaStream | null = null
let workletNode: AudioWorkletNode | null = null
let sourceNode: MediaStreamAudioSourceNode | null = null

// Get the first active API key for transcription
const activeApiKey = computed(() => apiKeys.value.find(k => !k.revoked_at))

// Toggle microphone
async function toggleMic() {
  if (isRecording.value) {
    stopRecording()
  } else {
    await startRecording()
  }
}

// Start recording
async function startRecording() {
  micError.value = null

  // Check for API key
  if (!activeApiKey.value) {
    micError.value = 'Please create an API key first'
    return
  }

  try {
    // Get microphone access - request higher sample rate, we'll resample
    mediaStream = await navigator.mediaDevices.getUserMedia({
      audio: {
        channelCount: 1,
        echoCancellation: true,
        noiseSuppression: true
      }
    })

    // For now, let's check if we have a recently created key in session
    const storedKey = sessionStorage.getItem('hyperwhisper_api_key')
    if (!storedKey) {
      micError.value = 'Please create a new API key and copy it, then try again. The full key is needed for streaming.'
      mediaStream.getTracks().forEach(track => track.stop())
      return
    }

    // Create AudioContext - use native sample rate and resample
    audioContext = new AudioContext()
    const nativeSampleRate = audioContext.sampleRate
    console.log('Native sample rate:', nativeSampleRate)

    // Build WebSocket URL - use native sample rate
    const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:'
    const fullWsUrl = `${protocol}//${window.location.host}/api/v1/deepgram/listen?api_key=${storedKey}&model=nova-3&smart_format=true&interim_results=true&encoding=linear16&sample_rate=${nativeSampleRate}`

    // Connect to WebSocket
    websocket = new WebSocket(fullWsUrl)

    websocket.onopen = () => {
      console.log('WebSocket connected')
      isRecording.value = true
      startAudioProcessing()
    }

    websocket.onmessage = (event) => {
      try {
        const data = JSON.parse(event.data)
        console.log('Received from server:', data.type)

        if (data.type === 'Results' && data.channel?.alternatives?.[0]) {
          const result = data.channel.alternatives[0]
          const text = result.transcript || ''

          if (data.is_final) {
            // Final result - append to transcription
            if (text) {
              transcription.value += (transcription.value ? ' ' : '') + text
            }
            interimTranscription.value = ''
          } else {
            // Interim result - show as preview
            interimTranscription.value = text
          }
        }
      } catch (e) {
        console.error('Failed to parse WebSocket message:', e)
      }
    }

    websocket.onerror = (error) => {
      console.error('WebSocket error:', error)
      micError.value = 'Connection error. Please try again.'
      stopRecording()
    }

    websocket.onclose = (event) => {
      console.log('WebSocket closed:', event.code, event.reason)
      if (isRecording.value) {
        stopRecording()
      }
    }

  } catch (e: any) {
    console.error('Failed to start recording:', e)
    if (e.name === 'NotAllowedError') {
      micError.value = 'Microphone access denied. Please allow microphone access and try again.'
    } else {
      micError.value = 'Failed to access microphone: ' + e.message
    }
  }
}

// Start audio processing and send to WebSocket
function startAudioProcessing() {
  if (!mediaStream || !websocket || !audioContext) return

  sourceNode = audioContext.createMediaStreamSource(mediaStream)

  // Use ScriptProcessorNode (deprecated but widely supported)
  // Buffer size of 4096 at 48kHz = ~85ms chunks
  const bufferSize = 4096
  const processor = audioContext.createScriptProcessor(bufferSize, 1, 1)

  let packetCount = 0

  processor.onaudioprocess = (event) => {
    if (websocket?.readyState === WebSocket.OPEN) {
      const inputData = event.inputBuffer.getChannelData(0)

      // Convert float32 [-1, 1] to int16 [-32768, 32767]
      const int16Data = new Int16Array(inputData.length)
      for (let i = 0; i < inputData.length; i++) {
        const s = Math.max(-1, Math.min(1, inputData[i]))
        int16Data[i] = s < 0 ? s * 0x8000 : s * 0x7FFF
      }

      websocket.send(int16Data.buffer)
      packetCount++

      if (packetCount % 10 === 0) {
        console.log(`Sent ${packetCount} audio packets (${int16Data.length * 2} bytes each)`)
      }
    }
  }

  sourceNode.connect(processor)
  // Connect to destination to keep the processor running
  processor.connect(audioContext.destination)

  console.log('Audio processing started with buffer size:', bufferSize)
}

// Stop recording
function stopRecording() {
  isRecording.value = false

  // Disconnect audio nodes
  if (sourceNode) {
    sourceNode.disconnect()
    sourceNode = null
  }

  // Send CloseStream message
  if (websocket?.readyState === WebSocket.OPEN) {
    websocket.send(JSON.stringify({ type: 'CloseStream' }))
  }

  // Close WebSocket
  if (websocket) {
    websocket.close()
    websocket = null
  }

  // Stop audio context
  if (audioContext) {
    audioContext.close()
    audioContext = null
  }

  // Stop media stream
  if (mediaStream) {
    mediaStream.getTracks().forEach(track => track.stop())
    mediaStream = null
  }

  // Clear interim transcription
  interimTranscription.value = ''
}

// Clear transcription
function clearTranscription() {
  transcription.value = ''
  interimTranscription.value = ''
}

// Copy transcription state
const copiedTranscription = ref(false)

// Copy transcription to clipboard
async function copyTranscription() {
  if (transcription.value) {
    await navigator.clipboard.writeText(transcription.value)
    copiedTranscription.value = true
    setTimeout(() => {
      copiedTranscription.value = false
    }, 2000)
  }
}

// Store API key in session when created
function storeApiKey(key: string) {
  sessionStorage.setItem('hyperwhisper_api_key', key)
}

// Fetch API keys
async function fetchAPIKeys() {
  isLoadingKeys.value = true
  keysError.value = null

  try {
    const response = await $fetch<PaginatedResponse<APIKey>>('/api/v1/deepgram/keys', {
      headers: getAuthHeaders(),
      credentials: 'include'
    })
    apiKeys.value = response.data
  } catch (e: any) {
    const apiError = e.data as ApiError
    keysError.value = apiError?.error || 'Failed to fetch API keys'
  } finally {
    isLoadingKeys.value = false
  }
}

// Fetch usage summary
async function fetchUsage() {
  isLoadingUsage.value = true

  try {
    const response = await $fetch<UsageSummary>('/api/v1/deepgram/usage', {
      headers: getAuthHeaders(),
      credentials: 'include'
    })
    usage.value = response
  } catch (e: any) {
    // Silently fail for usage - not critical
    console.error('Failed to fetch usage:', e)
  } finally {
    isLoadingUsage.value = false
  }
}

// Create API key
async function createAPIKey() {
  isCreating.value = true
  createError.value = null

  try {
    const response = await $fetch<APIKeyCreated>('/api/v1/deepgram/keys', {
      method: 'POST',
      headers: getAuthHeaders(),
      body: { name: createForm.value.name || 'Default Key' },
      credentials: 'include'
    })

    createdKey.value = response
    // Store the full key for transcription use
    storeApiKey(response.key)
    createForm.value = { name: '' }
    await fetchAPIKeys()
  } catch (e: any) {
    const apiError = e.data as ApiError
    createError.value = apiError?.error || 'Failed to create API key'
  } finally {
    isCreating.value = false
  }
}

// Copy key to clipboard
async function copyKey() {
  if (createdKey.value?.key) {
    await navigator.clipboard.writeText(createdKey.value.key)
    copiedKey.value = true
    setTimeout(() => {
      copiedKey.value = false
    }, 2000)
  }
}

// Close create dialog
function closeCreateDialog() {
  showCreateDialog.value = false
  createdKey.value = null
  createError.value = null
  copiedKey.value = false
}

// Confirm revoke
function confirmRevoke(key: APIKey) {
  keyToRevoke.value = key
  showRevokeDialog.value = true
}

// Revoke API key
async function revokeAPIKey() {
  if (!keyToRevoke.value) return

  isRevoking.value = true

  try {
    await $fetch(`/api/v1/deepgram/keys/${keyToRevoke.value.id}`, {
      method: 'DELETE',
      headers: getAuthHeaders(),
      credentials: 'include'
    })

    showRevokeDialog.value = false
    keyToRevoke.value = null
    await fetchAPIKeys()
  } catch (e: any) {
    const apiError = e.data as ApiError
    keysError.value = apiError?.error || 'Failed to revoke API key'
  } finally {
    isRevoking.value = false
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
function formatDuration(seconds: number | string | null | undefined): string {
  if (seconds == null) return '-'
  const num = typeof seconds === 'string' ? parseFloat(seconds) : seconds
  if (isNaN(num) || num === 0) return '-'
  if (num < 60) return `${num.toFixed(1)}s`
  const minutes = Math.floor(num / 60)
  const remainingSeconds = num % 60
  if (minutes < 60) return `${minutes}m ${remainingSeconds.toFixed(0)}s`
  const hours = Math.floor(minutes / 60)
  const remainingMinutes = minutes % 60
  return `${hours}h ${remainingMinutes}m`
}

// Cleanup on unmount
onUnmounted(() => {
  if (isRecording.value) {
    stopRecording()
  }
})

onMounted(() => {
  fetchAPIKeys()
  fetchUsage()
})
</script>

<template>
  <div class="min-h-screen bg-white dark:bg-black">
    <AppNavbar />
    <UserSidebar />
    <AdminSidebar />

    <main class="main-content container mx-auto px-4 py-12 pt-24">
      <div class="max-w-4xl mx-auto space-y-8">
        <!-- Welcome -->
        <div>
          <h1 class="text-3xl font-bold mb-2">Welcome, {{ user?.first_name || user?.username }}</h1>
          <p class="text-neutral-600 dark:text-neutral-400">
            Live transcription powered by Deepgram
          </p>
        </div>

        <!-- Live Transcription Card -->
        <Card>
          <CardHeader>
            <div class="flex items-center justify-between">
              <div class="flex items-center gap-2">
                <Mic class="size-5" />
                <CardTitle>Live Transcription</CardTitle>
              </div>
              <div class="flex items-center gap-2">
                <Badge v-if="isRecording" variant="destructive" class="animate-pulse">
                  Recording
                </Badge>
                <Button
                  v-if="transcription || interimTranscription"
                  variant="ghost"
                  size="sm"
                  @click="clearTranscription"
                >
                  Clear
                </Button>
              </div>
            </div>
          </CardHeader>
          <CardContent class="space-y-4">
            <!-- Error -->
            <Alert v-if="micError" variant="destructive">
              <AlertDescription>{{ micError }}</AlertDescription>
            </Alert>

            <!-- No API Key Warning -->
            <Alert v-if="!isLoadingKeys && apiKeys.length === 0">
              <AlertDescription>
                Create an API key below to start using live transcription.
              </AlertDescription>
            </Alert>

            <!-- Mic Button -->
            <div class="flex justify-center">
              <Button
                size="lg"
                :variant="isRecording ? 'destructive' : 'default'"
                @click="toggleMic"
                :disabled="isLoadingKeys || apiKeys.length === 0"
                class="w-20 h-20 rounded-full"
              >
                <MicOff v-if="isRecording" class="size-8" />
                <Mic v-else class="size-8" />
              </Button>
            </div>

            <p class="text-center text-sm text-neutral-500 dark:text-neutral-400">
              {{ isRecording ? 'Click to stop recording' : 'Click to start recording' }}
            </p>

            <!-- Transcription Output -->
            <div class="space-y-2">
              <div class="flex items-center justify-between">
                <Label>Transcription</Label>
                <Button
                  v-if="transcription"
                  variant="ghost"
                  size="sm"
                  @click="copyTranscription"
                  class="h-8"
                >
                  <Check v-if="copiedTranscription" class="size-4 mr-1" />
                  <Copy v-else class="size-4 mr-1" />
                  {{ copiedTranscription ? 'Copied!' : 'Copy' }}
                </Button>
              </div>
              <div
                class="min-h-[150px] max-h-[300px] overflow-y-auto p-4 rounded-lg border border-neutral-200 dark:border-white/10 bg-neutral-50 dark:bg-white/5 font-mono text-sm"
              >
                <span v-if="transcription || interimTranscription">
                  {{ transcription }}
                  <span v-if="interimTranscription" class="text-neutral-400 dark:text-neutral-500">
                    {{ interimTranscription }}
                  </span>
                </span>
                <span v-else class="text-neutral-400 dark:text-neutral-500 italic">
                  Transcription will appear here...
                </span>
              </div>
            </div>
          </CardContent>
        </Card>

        <!-- Usage Summary -->
        <Card>
          <CardHeader>
            <div class="flex items-center gap-2">
              <Activity class="size-5" />
              <CardTitle>Usage This Month</CardTitle>
            </div>
          </CardHeader>
          <CardContent>
            <div v-if="isLoadingUsage" class="text-center py-4 text-neutral-500">
              Loading usage data...
            </div>
            <div v-else-if="usage" class="grid grid-cols-3 gap-4">
              <div class="text-center p-4 bg-neutral-50 dark:bg-white/5 rounded-lg">
                <div class="text-2xl font-bold">{{ usage.total_sessions }}</div>
                <div class="text-sm text-neutral-500 dark:text-neutral-400">Sessions</div>
              </div>
              <div class="text-center p-4 bg-neutral-50 dark:bg-white/5 rounded-lg">
                <div class="text-2xl font-bold">{{ formatDuration(usage.total_duration_seconds) }}</div>
                <div class="text-sm text-neutral-500 dark:text-neutral-400">Audio Duration</div>
              </div>
              <div class="text-center p-4 bg-neutral-50 dark:bg-white/5 rounded-lg">
                <div class="text-2xl font-bold">{{ formatBytes(usage.total_bytes_sent) }}</div>
                <div class="text-sm text-neutral-500 dark:text-neutral-400">Data Sent</div>
              </div>
            </div>
            <div v-else class="text-center py-4 text-neutral-500">
              No usage data available
            </div>
          </CardContent>
        </Card>

        <!-- API Keys -->
        <Card>
          <CardHeader>
            <div class="flex items-center justify-between">
              <div class="flex items-center gap-2">
                <Key class="size-5" />
                <CardTitle>API Keys</CardTitle>
              </div>
              <div class="flex gap-2">
                <Button variant="outline" size="sm" @click="fetchAPIKeys" :disabled="isLoadingKeys">
                  <RefreshCw :class="['size-4 mr-2', isLoadingKeys && 'animate-spin']" />
                  Refresh
                </Button>
                <Button size="sm" @click="showCreateDialog = true">
                  <Plus class="size-4 mr-2" />
                  Create Key
                </Button>
              </div>
            </div>
          </CardHeader>
          <CardContent>
            <Alert v-if="keysError" variant="destructive" class="mb-4">
              <AlertDescription>{{ keysError }}</AlertDescription>
            </Alert>

            <div class="overflow-x-auto">
              <table class="w-full">
                <thead class="border-b border-neutral-200 dark:border-white/10">
                  <tr class="text-left text-sm text-neutral-500 dark:text-neutral-400">
                    <th class="p-3 font-medium">Name</th>
                    <th class="p-3 font-medium">Key Prefix</th>
                    <th class="p-3 font-medium">Created</th>
                    <th class="p-3 font-medium">Last Used</th>
                    <th class="p-3 font-medium">Actions</th>
                  </tr>
                </thead>
                <tbody class="divide-y divide-neutral-200 dark:divide-white/10">
                  <tr v-if="isLoadingKeys">
                    <td colspan="5" class="p-6 text-center text-neutral-500">
                      Loading...
                    </td>
                  </tr>
                  <tr v-else-if="apiKeys.length === 0">
                    <td colspan="5" class="p-6 text-center text-neutral-500">
                      No API keys yet. Create one to get started.
                    </td>
                  </tr>
                  <tr v-else v-for="key in apiKeys" :key="key.id" class="hover:bg-neutral-50 dark:hover:bg-white/5">
                    <td class="p-3 font-medium">{{ key.name }}</td>
                    <td class="p-3">
                      <code class="text-sm bg-neutral-100 dark:bg-white/10 px-2 py-1 rounded">
                        {{ key.key_prefix }}...
                      </code>
                    </td>
                    <td class="p-3 text-sm text-neutral-500">
                      {{ new Date(key.created_at).toLocaleDateString() }}
                    </td>
                    <td class="p-3 text-sm text-neutral-500">
                      {{ key.last_used_at ? new Date(key.last_used_at).toLocaleDateString() : 'Never' }}
                    </td>
                    <td class="p-3">
                      <Button
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
          </CardContent>
        </Card>
      </div>
    </main>

    <!-- Create API Key Dialog -->
    <Dialog :open="showCreateDialog" @update:open="(open) => !open && closeCreateDialog()">
      <DialogContent>
        <DialogHeader>
          <DialogTitle>{{ createdKey ? 'API Key Created' : 'Create API Key' }}</DialogTitle>
          <DialogDescription v-if="!createdKey">
            Create a new API key to access the Deepgram proxy.
          </DialogDescription>
          <DialogDescription v-else>
            Copy your API key now. You won't be able to see it again.
          </DialogDescription>
        </DialogHeader>

        <!-- Create Form -->
        <form v-if="!createdKey" @submit.prevent="createAPIKey" class="space-y-4">
          <Alert v-if="createError" variant="destructive">
            <AlertDescription>{{ createError }}</AlertDescription>
          </Alert>

          <div class="space-y-2">
            <Label for="key_name">Key Name (optional)</Label>
            <Input
              id="key_name"
              v-model="createForm.name"
              placeholder="My API Key"
            />
          </div>

          <DialogFooter>
            <Button type="button" variant="outline" @click="closeCreateDialog">
              Cancel
            </Button>
            <button
              type="submit"
              class="inline-flex items-center justify-center gap-2 whitespace-nowrap rounded-md text-sm font-medium transition-colors focus-visible:outline-none focus-visible:ring-1 focus-visible:ring-ring disabled:pointer-events-none disabled:opacity-50 bg-primary text-primary-foreground shadow hover:bg-primary/90 h-9 px-4 py-2"
              :disabled="isCreating"
            >
              {{ isCreating ? 'Creating...' : 'Create Key' }}
            </button>
          </DialogFooter>
        </form>

        <!-- Key Created -->
        <div v-else class="space-y-4">
          <Alert>
            <AlertDescription>
              Make sure to copy your API key now. You won't be able to see it again!
              This key has been stored for use with live transcription.
            </AlertDescription>
          </Alert>

          <div class="space-y-2">
            <Label>Your API Key</Label>
            <div class="flex gap-2">
              <Input
                :model-value="createdKey.key"
                readonly
                class="font-mono text-sm"
              />
              <Button variant="outline" @click="copyKey">
                <Check v-if="copiedKey" class="size-4" />
                <Copy v-else class="size-4" />
              </Button>
            </div>
          </div>

          <DialogFooter>
            <Button @click="closeCreateDialog">
              Done
            </Button>
          </DialogFooter>
        </div>
      </DialogContent>
    </Dialog>

    <!-- Revoke Confirmation Dialog -->
    <Dialog v-model:open="showRevokeDialog">
      <DialogContent>
        <DialogHeader>
          <DialogTitle>Revoke API Key</DialogTitle>
          <DialogDescription>
            Are you sure you want to revoke <strong>{{ keyToRevoke?.name }}</strong>?
            This action cannot be undone and any applications using this key will stop working.
          </DialogDescription>
        </DialogHeader>
        <DialogFooter>
          <Button variant="outline" @click="showRevokeDialog = false">
            Cancel
          </Button>
          <Button variant="destructive" @click="revokeAPIKey" :disabled="isRevoking">
            {{ isRevoking ? 'Revoking...' : 'Revoke Key' }}
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
