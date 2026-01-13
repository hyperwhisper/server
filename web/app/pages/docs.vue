<script setup lang="ts">
import { Copy, Check, Mic, Key, BarChart3, FileText, Zap } from 'lucide-vue-next'

useHead({
  title: 'API Documentation - HyperWhisper'
})

const copiedCode = ref<string | null>(null)

async function copyCode(code: string, id: string) {
  await navigator.clipboard.writeText(code)
  copiedCode.value = id
  setTimeout(() => {
    copiedCode.value = null
  }, 2000)
}

const baseUrl = computed(() => {
  if (import.meta.client) {
    return window.location.origin
  }
  return 'https://hyperwhisper.dev'
})

const wsProtocol = computed(() => {
  if (import.meta.client) {
    return window.location.protocol === 'https:' ? 'wss:' : 'ws:'
  }
  return 'wss:'
})

// Code snippets for copy buttons
const codeSnippets = {
  response: `{
  "type": "Results",
  "channel_index": [0, 1],
  "duration": 1.5,
  "start": 0.0,
  "is_final": true,
  "channel": {
    "alternatives": [
      {
        "transcript": "Hello, world.",
        "confidence": 0.98
      }
    ]
  }
}`,
  close: '{"type": "CloseStream"}',
  javascript: `const apiKey = 'hw_live_your_api_key';
const ws = new WebSocket(
  \`wss://hyperwhisper.dev/api/v1/deepgram/listen?api_key=\${apiKey}&model=nova-3&smart_format=true&interim_results=true&encoding=linear16&sample_rate=48000\`
);

ws.onopen = () => {
  console.log('Connected');
  // Start sending audio data
};

ws.onmessage = (event) => {
  const data = JSON.parse(event.data);
  if (data.type === 'Results') {
    const transcript = data.channel?.alternatives?.[0]?.transcript;
    if (data.is_final) {
      console.log('Final:', transcript);
    } else {
      console.log('Interim:', transcript);
    }
  }
};

// Send audio as Int16Array
function sendAudio(float32Data) {
  const int16Data = new Int16Array(float32Data.length);
  for (let i = 0; i < float32Data.length; i++) {
    const s = Math.max(-1, Math.min(1, float32Data[i]));
    int16Data[i] = s < 0 ? s * 0x8000 : s * 0x7FFF;
  }
  ws.send(int16Data.buffer);
}

// Close gracefully
ws.send(JSON.stringify({ type: 'CloseStream' }));
ws.close();`,
  python: `import asyncio
import websockets
import json

API_KEY = "hw_live_your_api_key"
URL = f"wss://hyperwhisper.dev/api/v1/deepgram/listen?api_key={API_KEY}&model=nova-3&encoding=linear16&sample_rate=16000"

async def transcribe():
    async with websockets.connect(URL) as ws:
        # Send audio data (bytes)
        with open("audio.raw", "rb") as f:
            while chunk := f.read(4096):
                await ws.send(chunk)

        # Close stream
        await ws.send(json.dumps({"type": "CloseStream"}))

        # Receive results
        async for message in ws:
            data = json.loads(message)
            if data["type"] == "Results":
                transcript = data["channel"]["alternatives"][0]["transcript"]
                if data["is_final"]:
                    print(f"Final: {transcript}")

asyncio.run(transcribe())`
}
</script>

<template>
  <div class="min-h-screen bg-white dark:bg-black">
    <AppNavbar />

    <main class="container mx-auto px-4 py-12 pt-24">
      <div class="max-w-4xl mx-auto">
        <!-- Header -->
        <div class="docs-header">
          <h1 class="text-4xl font-bold mb-4">API Documentation</h1>
          <p class="text-lg text-neutral-600 dark:text-neutral-400">
            Integrate real-time speech-to-text transcription into your applications using the HyperWhisper API.
          </p>
        </div>

        <!-- Quick Start -->
        <section class="docs-section">
          <div class="docs-section-title flex items-center gap-3">
            <Zap class="size-5" />
            <h2 class="text-2xl font-bold">Quick Start</h2>
          </div>

          <div class="space-y-3">
            <div class="p-4 rounded-lg border border-neutral-200 dark:border-white/10 bg-neutral-50 dark:bg-white/5">
              <h3 class="font-medium mb-2">1. Create an API Key</h3>
              <p class="text-sm text-neutral-600 dark:text-neutral-400">
                Sign in to your <NuxtLink to="/dashboard" class="text-blue-600 dark:text-blue-400 hover:underline">dashboard</NuxtLink> and create an API key. Save it securely - you won't be able to see it again.
              </p>
            </div>

            <div class="p-4 rounded-lg border border-neutral-200 dark:border-white/10 bg-neutral-50 dark:bg-white/5">
              <h3 class="font-medium mb-2">2. Connect via WebSocket</h3>
              <p class="text-sm text-neutral-600 dark:text-neutral-400">
                Open a WebSocket connection to the transcription endpoint with your API key.
              </p>
            </div>

            <div class="p-4 rounded-lg border border-neutral-200 dark:border-white/10 bg-neutral-50 dark:bg-white/5">
              <h3 class="font-medium mb-2">3. Stream Audio</h3>
              <p class="text-sm text-neutral-600 dark:text-neutral-400">
                Send audio data as binary messages and receive real-time transcription results.
              </p>
            </div>
          </div>
        </section>

        <!-- WebSocket API -->
        <section class="docs-section">
          <div class="docs-section-title flex items-center gap-3">
            <Mic class="size-5" />
            <h2 class="text-2xl font-bold">WebSocket Transcription API</h2>
          </div>

          <Card class="mb-6">
            <CardHeader>
              <CardTitle class="text-lg">Endpoint</CardTitle>
            </CardHeader>
            <CardContent>
              <div class="flex items-center gap-2">
                <code class="flex-1 p-3 rounded bg-neutral-100 dark:bg-white/10 text-sm font-mono overflow-x-auto">
                  {{ wsProtocol }}//{{ baseUrl.replace(/^https?:\/\//, '') }}/api/v1/deepgram/listen
                </code>
                <Button
                  variant="ghost"
                  size="sm"
                  @click="copyCode(`${wsProtocol}//${baseUrl.replace(/^https?:\/\//, '')}/api/v1/deepgram/listen`, 'endpoint')"
                >
                  <Check v-if="copiedCode === 'endpoint'" class="size-4" />
                  <Copy v-else class="size-4" />
                </Button>
              </div>
            </CardContent>
          </Card>

          <Card class="mb-6">
            <CardHeader>
              <CardTitle class="text-lg">Authentication</CardTitle>
            </CardHeader>
            <CardContent class="space-y-4">
              <p class="text-sm text-neutral-600 dark:text-neutral-400">
                Pass your API key as a query parameter:
              </p>
              <div class="flex items-center gap-2">
                <code class="flex-1 p-3 rounded bg-neutral-100 dark:bg-white/10 text-sm font-mono overflow-x-auto">
                  ?api_key=hw_live_your_api_key_here
                </code>
              </div>
              <p class="text-sm text-neutral-500 dark:text-neutral-500">
                Or use the <code class="px-1 py-0.5 rounded bg-neutral-200 dark:bg-white/10">X-API-Key</code> header (for non-browser clients).
              </p>
            </CardContent>
          </Card>

          <Card class="mb-6">
            <CardHeader>
              <CardTitle class="text-lg">Query Parameters</CardTitle>
            </CardHeader>
            <CardContent>
              <div class="overflow-x-auto">
                <table class="w-full text-sm">
                  <thead class="border-b border-neutral-200 dark:border-white/10">
                    <tr class="text-left">
                      <th class="p-3 font-medium">Parameter</th>
                      <th class="p-3 font-medium">Type</th>
                      <th class="p-3 font-medium">Description</th>
                    </tr>
                  </thead>
                  <tbody class="divide-y divide-neutral-200 dark:divide-white/10">
                    <tr>
                      <td class="p-3"><code class="text-xs bg-neutral-100 dark:bg-white/10 px-1.5 py-0.5 rounded">model</code></td>
                      <td class="p-3 text-neutral-500">string</td>
                      <td class="p-3 text-neutral-600 dark:text-neutral-400">Model to use (e.g., <code class="text-xs bg-neutral-100 dark:bg-white/10 px-1 rounded">nova-3</code>, <code class="text-xs bg-neutral-100 dark:bg-white/10 px-1 rounded">nova-2</code>)</td>
                    </tr>
                    <tr>
                      <td class="p-3"><code class="text-xs bg-neutral-100 dark:bg-white/10 px-1.5 py-0.5 rounded">language</code></td>
                      <td class="p-3 text-neutral-500">string</td>
                      <td class="p-3 text-neutral-600 dark:text-neutral-400">Language code (e.g., <code class="text-xs bg-neutral-100 dark:bg-white/10 px-1 rounded">en</code>, <code class="text-xs bg-neutral-100 dark:bg-white/10 px-1 rounded">es</code>, <code class="text-xs bg-neutral-100 dark:bg-white/10 px-1 rounded">fr</code>)</td>
                    </tr>
                    <tr>
                      <td class="p-3"><code class="text-xs bg-neutral-100 dark:bg-white/10 px-1.5 py-0.5 rounded">encoding</code></td>
                      <td class="p-3 text-neutral-500">string</td>
                      <td class="p-3 text-neutral-600 dark:text-neutral-400">Audio encoding (<code class="text-xs bg-neutral-100 dark:bg-white/10 px-1 rounded">linear16</code>, <code class="text-xs bg-neutral-100 dark:bg-white/10 px-1 rounded">opus</code>, <code class="text-xs bg-neutral-100 dark:bg-white/10 px-1 rounded">flac</code>)</td>
                    </tr>
                    <tr>
                      <td class="p-3"><code class="text-xs bg-neutral-100 dark:bg-white/10 px-1.5 py-0.5 rounded">sample_rate</code></td>
                      <td class="p-3 text-neutral-500">integer</td>
                      <td class="p-3 text-neutral-600 dark:text-neutral-400">Audio sample rate in Hz (e.g., <code class="text-xs bg-neutral-100 dark:bg-white/10 px-1 rounded">16000</code>, <code class="text-xs bg-neutral-100 dark:bg-white/10 px-1 rounded">48000</code>)</td>
                    </tr>
                    <tr>
                      <td class="p-3"><code class="text-xs bg-neutral-100 dark:bg-white/10 px-1.5 py-0.5 rounded">channels</code></td>
                      <td class="p-3 text-neutral-500">integer</td>
                      <td class="p-3 text-neutral-600 dark:text-neutral-400">Number of audio channels (default: 1)</td>
                    </tr>
                    <tr>
                      <td class="p-3"><code class="text-xs bg-neutral-100 dark:bg-white/10 px-1.5 py-0.5 rounded">smart_format</code></td>
                      <td class="p-3 text-neutral-500">boolean</td>
                      <td class="p-3 text-neutral-600 dark:text-neutral-400">Apply smart formatting (punctuation, capitalization)</td>
                    </tr>
                    <tr>
                      <td class="p-3"><code class="text-xs bg-neutral-100 dark:bg-white/10 px-1.5 py-0.5 rounded">interim_results</code></td>
                      <td class="p-3 text-neutral-500">boolean</td>
                      <td class="p-3 text-neutral-600 dark:text-neutral-400">Receive interim (partial) transcription results</td>
                    </tr>
                    <tr>
                      <td class="p-3"><code class="text-xs bg-neutral-100 dark:bg-white/10 px-1.5 py-0.5 rounded">punctuate</code></td>
                      <td class="p-3 text-neutral-500">boolean</td>
                      <td class="p-3 text-neutral-600 dark:text-neutral-400">Add punctuation to transcription</td>
                    </tr>
                    <tr>
                      <td class="p-3"><code class="text-xs bg-neutral-100 dark:bg-white/10 px-1.5 py-0.5 rounded">diarize</code></td>
                      <td class="p-3 text-neutral-500">boolean</td>
                      <td class="p-3 text-neutral-600 dark:text-neutral-400">Enable speaker diarization</td>
                    </tr>
                    <tr>
                      <td class="p-3"><code class="text-xs bg-neutral-100 dark:bg-white/10 px-1.5 py-0.5 rounded">utterances</code></td>
                      <td class="p-3 text-neutral-500">boolean</td>
                      <td class="p-3 text-neutral-600 dark:text-neutral-400">Segment transcription into utterances</td>
                    </tr>
                  </tbody>
                </table>
              </div>
            </CardContent>
          </Card>

          <Card class="mb-6">
            <CardHeader>
              <CardTitle class="text-lg">Response Format</CardTitle>
            </CardHeader>
            <CardContent class="space-y-4">
              <p class="text-sm text-neutral-600 dark:text-neutral-400">
                Transcription results are sent as JSON messages:
              </p>
              <div class="relative">
                <pre class="p-4 rounded bg-neutral-100 dark:bg-white/10 text-sm font-mono overflow-x-auto"><code>{{ codeSnippets.response }}</code></pre>
                <Button
                  variant="ghost"
                  size="sm"
                  class="absolute top-2 right-2"
                  @click="copyCode(codeSnippets.response, 'response')"
                >
                  <Check v-if="copiedCode === 'response'" class="size-4" />
                  <Copy v-else class="size-4" />
                </Button>
              </div>
            </CardContent>
          </Card>

          <Card>
            <CardHeader>
              <CardTitle class="text-lg">Closing the Connection</CardTitle>
            </CardHeader>
            <CardContent class="space-y-4">
              <p class="text-sm text-neutral-600 dark:text-neutral-400">
                To gracefully close the connection and receive final metadata, send a CloseStream message:
              </p>
              <div class="relative">
                <pre class="p-4 rounded bg-neutral-100 dark:bg-white/10 text-sm font-mono overflow-x-auto"><code>{{ codeSnippets.close }}</code></pre>
                <Button
                  variant="ghost"
                  size="sm"
                  class="absolute top-2 right-2"
                  @click="copyCode(codeSnippets.close, 'close')"
                >
                  <Check v-if="copiedCode === 'close'" class="size-4" />
                  <Copy v-else class="size-4" />
                </Button>
              </div>
            </CardContent>
          </Card>
        </section>

        <!-- Code Examples -->
        <section class="docs-section">
          <div class="docs-section-title flex items-center gap-3">
            <FileText class="size-5" />
            <h2 class="text-2xl font-bold">Code Examples</h2>
          </div>

          <Card class="mb-6">
            <CardHeader>
              <CardTitle class="text-lg">JavaScript / Browser</CardTitle>
            </CardHeader>
            <CardContent>
              <div class="relative">
                <pre class="p-4 rounded bg-neutral-100 dark:bg-white/10 text-sm font-mono overflow-x-auto"><code>{{ codeSnippets.javascript }}</code></pre>
                <Button
                  variant="ghost"
                  size="sm"
                  class="absolute top-2 right-2"
                  @click="copyCode(codeSnippets.javascript, 'js-example')"
                >
                  <Check v-if="copiedCode === 'js-example'" class="size-4" />
                  <Copy v-else class="size-4" />
                </Button>
              </div>
            </CardContent>
          </Card>

          <Card>
            <CardHeader>
              <CardTitle class="text-lg">Python</CardTitle>
            </CardHeader>
            <CardContent>
              <div class="relative">
                <pre class="p-4 rounded bg-neutral-100 dark:bg-white/10 text-sm font-mono overflow-x-auto"><code>{{ codeSnippets.python }}</code></pre>
                <Button
                  variant="ghost"
                  size="sm"
                  class="absolute top-2 right-2"
                  @click="copyCode(codeSnippets.python, 'python-example')"
                >
                  <Check v-if="copiedCode === 'python-example'" class="size-4" />
                  <Copy v-else class="size-4" />
                </Button>
              </div>
            </CardContent>
          </Card>
        </section>

        <!-- REST API -->
        <section class="docs-section">
          <div class="docs-section-title flex items-center gap-3">
            <Key class="size-5" />
            <h2 class="text-2xl font-bold">REST API</h2>
          </div>

          <p class="text-neutral-600 dark:text-neutral-400 mb-6">
            Manage your API keys and view usage statistics via the REST API. These endpoints require JWT authentication.
          </p>

          <Card class="mb-4">
            <CardHeader class="pb-2">
              <div class="flex items-center gap-3">
                <Badge class="bg-green-100 text-green-800 dark:bg-green-900 dark:text-green-200">POST</Badge>
                <code class="text-sm font-mono">/api/v1/deepgram/keys</code>
              </div>
            </CardHeader>
            <CardContent>
              <p class="text-sm text-neutral-600 dark:text-neutral-400">Create a new API key</p>
            </CardContent>
          </Card>

          <Card class="mb-4">
            <CardHeader class="pb-2">
              <div class="flex items-center gap-3">
                <Badge class="bg-blue-100 text-blue-800 dark:bg-blue-900 dark:text-blue-200">GET</Badge>
                <code class="text-sm font-mono">/api/v1/deepgram/keys</code>
              </div>
            </CardHeader>
            <CardContent>
              <p class="text-sm text-neutral-600 dark:text-neutral-400">List all your API keys</p>
            </CardContent>
          </Card>

          <Card class="mb-4">
            <CardHeader class="pb-2">
              <div class="flex items-center gap-3">
                <Badge class="bg-red-100 text-red-800 dark:bg-red-900 dark:text-red-200">DELETE</Badge>
                <code class="text-sm font-mono">/api/v1/deepgram/keys/:id</code>
              </div>
            </CardHeader>
            <CardContent>
              <p class="text-sm text-neutral-600 dark:text-neutral-400">Revoke an API key</p>
            </CardContent>
          </Card>

          <Card class="mb-4">
            <CardHeader class="pb-2">
              <div class="flex items-center gap-3">
                <Badge class="bg-blue-100 text-blue-800 dark:bg-blue-900 dark:text-blue-200">GET</Badge>
                <code class="text-sm font-mono">/api/v1/deepgram/usage</code>
              </div>
            </CardHeader>
            <CardContent>
              <p class="text-sm text-neutral-600 dark:text-neutral-400">Get usage summary for current month</p>
            </CardContent>
          </Card>

          <Card>
            <CardHeader class="pb-2">
              <div class="flex items-center gap-3">
                <Badge class="bg-blue-100 text-blue-800 dark:bg-blue-900 dark:text-blue-200">GET</Badge>
                <code class="text-sm font-mono">/api/v1/deepgram/logs</code>
              </div>
            </CardHeader>
            <CardContent>
              <p class="text-sm text-neutral-600 dark:text-neutral-400">List transcription logs (paginated)</p>
            </CardContent>
          </Card>
        </section>

        <!-- Rate Limits -->
        <section class="docs-section">
          <div class="docs-section-title flex items-center gap-3">
            <BarChart3 class="size-5" />
            <h2 class="text-2xl font-bold">Usage & Limits</h2>
          </div>

          <div class="grid md:grid-cols-2 gap-6">
            <Card>
              <CardHeader>
                <CardTitle class="text-lg">Billing</CardTitle>
              </CardHeader>
              <CardContent>
                <p class="text-sm text-neutral-600 dark:text-neutral-400">
                  Usage is billed based on audio duration processed. You can monitor your usage in the
                  <NuxtLink to="/dashboard" class="text-blue-600 dark:text-blue-400 hover:underline">dashboard</NuxtLink>.
                </p>
              </CardContent>
            </Card>

            <Card>
              <CardHeader>
                <CardTitle class="text-lg">Concurrent Connections</CardTitle>
              </CardHeader>
              <CardContent>
                <p class="text-sm text-neutral-600 dark:text-neutral-400">
                  Each API key can have multiple concurrent WebSocket connections. There is no hard limit, but excessive usage may be rate limited.
                </p>
              </CardContent>
            </Card>
          </div>
        </section>

        <!-- Support -->
        <section>
          <Card class="bg-neutral-50 dark:bg-white/5">
            <CardContent class="py-8 text-center">
              <h3 class="text-xl font-bold mb-2">Need Help?</h3>
              <p class="text-neutral-600 dark:text-neutral-400 mb-4">
                Check out our GitHub repository for issues, discussions, and more examples.
              </p>
              <Button variant="outline" as="a" href="https://github.com/hyperwhisper" target="_blank">
                View on GitHub
              </Button>
            </CardContent>
          </Card>
        </section>
      </div>
    </main>

    <footer class="border-t border-neutral-200 dark:border-white/5 py-8 mt-16">
      <div class="container mx-auto px-4 flex flex-col sm:flex-row items-center justify-between gap-4 text-sm text-neutral-500 dark:text-neutral-600">
        <p>hyperwhisper is open source software</p>
        <div class="flex items-center gap-6">
          <a href="https://github.com/hyperwhisper" target="_blank" rel="noopener" class="hover:text-neutral-700 dark:hover:text-neutral-400 transition-colors">
            GitHub
          </a>
          <NuxtLink to="/health" class="hover:text-neutral-700 dark:hover:text-neutral-400 transition-colors">
            Status
          </NuxtLink>
        </div>
      </div>
    </footer>
  </div>
</template>

<style scoped>
.docs-header {
  margin-bottom: 3rem;
}

.docs-section {
  margin-bottom: 3rem;
}

.docs-section-title {
  margin-bottom: 1.5rem;
}
</style>
