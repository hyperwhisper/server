<script setup lang="ts">
import { Apple, Monitor, Terminal, Settings, Copy, Mic } from 'lucide-vue-next'

useHead({
  title: 'HyperWhisper - Open Source Voice to Text',
  meta: [
    { name: 'description', content: 'Open source AI-powered voice to text. Fast, private, and runs locally.' }
  ]
})

const demoText = ref('')
const demoTexts = [
  "Hey, just finished the design review. The new dashboard looks incredible - clean, fast, and exactly what we needed. Let's ship it tomorrow.",
  "Note to self: refactor the authentication module before the sprint ends. Also need to update the API docs.",
  "Quick standup update - fixed the memory leak in production, PR is ready for review. Moving on to the billing integration next.",
  "Dear team, I wanted to follow up on yesterday's discussion about the migration strategy. I think we should proceed with option B.",
  "Remind me to call Mom at 5pm. Also pick up groceries - milk, eggs, bread, and that fancy cheese she likes."
]
const currentTextIndex = ref(0)
const isTyping = ref(true)
const isListening = ref(false)
const showCopied = ref(false)

async function copyText() {
  if (demoText.value) {
    await navigator.clipboard.writeText(demoText.value)
    showCopied.value = true
    setTimeout(() => {
      showCopied.value = false
    }, 2000)
  }
}

// Generate bar heights for the waveform (mimics audio visualization)
const barCount = 160
const barHeights = ref<number[]>([])
let animationFrame: number | null = null

function generateWaveformHeights() {
  const heights: number[] = []
  for (let i = 0; i < barCount; i++) {
    // Create a wave pattern - higher in the middle, lower at edges
    const centerDistance = Math.abs(i - barCount / 2) / (barCount / 2)
    const baseHeight = (1 - centerDistance * 0.7) * 100
    const randomVariation = Math.random() * 40 - 20
    heights.push(Math.max(8, Math.min(100, baseHeight + randomVariation)))
  }
  barHeights.value = heights
}

function animateWaveform() {
  generateWaveformHeights()
  animationFrame = requestAnimationFrame(() => {
    setTimeout(animateWaveform, 80)
  })
}

function stopWaveformAnimation() {
  if (animationFrame) {
    cancelAnimationFrame(animationFrame)
    animationFrame = null
  }
}

function typeText(text: string, onComplete: () => void) {
  let i = 0
  demoText.value = ''
  isTyping.value = true

  const interval = setInterval(() => {
    if (i < text.length) {
      demoText.value += text[i]
      i++
    } else {
      isTyping.value = false
      clearInterval(interval)
      onComplete()
    }
  }, 40)
}

function cycleTexts() {
  // First show listening animation
  isListening.value = true
  animateWaveform()

  // After 2 seconds, stop listening and start typing
  setTimeout(() => {
    isListening.value = false
    stopWaveformAnimation()

    typeText(demoTexts[currentTextIndex.value], () => {
      setTimeout(() => {
        currentTextIndex.value = (currentTextIndex.value + 1) % demoTexts.length
        cycleTexts()
      }, 3000)
    })
  }, 2000)
}

onMounted(() => {
  generateWaveformHeights()
  cycleTexts()
})

onUnmounted(() => {
  stopWaveformAnimation()
})
</script>

<template>
  <div class="min-h-screen bg-neutral-50 dark:bg-neutral-950 text-neutral-900 dark:text-white overflow-hidden transition-colors duration-300">
    <!-- Subtle gradient background -->
    <div class="fixed inset-0 overflow-hidden pointer-events-none">
      <div class="absolute -top-[50%] left-[50%] -translate-x-1/2 w-[100%] h-[100%] rounded-full bg-gradient-to-b from-neutral-200/50 dark:from-neutral-800/30 to-transparent blur-3xl" />
    </div>

    <!-- Subtle grain texture overlay -->
    <div class="fixed inset-0 opacity-[0.02] dark:opacity-[0.015] pointer-events-none bg-[url('data:image/svg+xml;base64,PHN2ZyB4bWxucz0iaHR0cDovL3d3dy53My5vcmcvMjAwMC9zdmciIHdpZHRoPSIzMDAiIGhlaWdodD0iMzAwIj48ZmlsdGVyIGlkPSJhIiB4PSIwIiB5PSIwIj48ZmVUdXJidWxlbmNlIGJhc2VGcmVxdWVuY3k9Ii43NSIgc3RpdGNoVGlsZXM9InN0aXRjaCIgdHlwZT0iZnJhY3RhbE5vaXNlIi8+PC9maWx0ZXI+PHJlY3Qgd2lkdGg9IjMwMCIgaGVpZ2h0PSIzMDAiIGZpbHRlcj0idXJsKCNhKSIgb3BhY2l0eT0iMSIvPjwvc3ZnPg==')]" />

    <AppNavbar />

    <main class="relative">
      <!-- Hero Section -->
      <section class="min-h-screen flex flex-col items-center justify-center px-4 pt-16 relative">
        <!-- Abstract Logo/Icon -->
        <div class="relative mb-8 sm:mb-12">
          <div class="w-24 h-24 sm:w-32 sm:h-32 relative">
            <!-- Subtle glow -->
            <div class="absolute inset-0 rounded-full bg-neutral-300/50 dark:bg-white/10 blur-2xl" />
            <svg viewBox="0 0 100 100" class="w-full h-full relative z-10">
              <!-- Sound wave / voice visualization -->
              <defs>
                <linearGradient id="waveGradient" x1="0%" y1="0%" x2="100%" y2="100%">
                  <stop offset="0%" class="[stop-color:theme(colors.neutral.700)] dark:[stop-color:#e5e5e5]" style="stop-opacity:1" />
                  <stop offset="100%" class="[stop-color:theme(colors.neutral.400)] dark:[stop-color:#737373]" style="stop-opacity:0.6" />
                </linearGradient>
              </defs>
              <g fill="none" stroke="url(#waveGradient)" stroke-width="2" stroke-linecap="round">
                <path d="M20 50 Q20 30, 30 30 Q40 30, 40 50 Q40 70, 30 70 Q20 70, 20 50" opacity="0.4" />
                <path d="M35 50 Q35 25, 50 25 Q65 25, 65 50 Q65 75, 50 75 Q35 75, 35 50" opacity="0.7" />
                <path d="M50 50 Q50 20, 70 20 Q90 20, 90 50 Q90 80, 70 80 Q50 80, 50 50" opacity="1" />
              </g>
            </svg>
          </div>
        </div>

        <!-- Main Headline -->
        <h1 class="text-4xl sm:text-5xl md:text-7xl font-bold tracking-tight text-center mb-4 sm:mb-6">
          <span class="text-neutral-900 dark:text-white">
            Type 3x faster,
          </span>
          <br />
          <span class="text-neutral-500 dark:text-neutral-400">
            without lifting a finger.
          </span>
        </h1>

        <!-- Subtitle -->
        <div class="text-center mb-8 sm:mb-12">
          <p class="text-lg sm:text-xl text-neutral-600 dark:text-neutral-400 mb-1">hyperwhisper</p>
          <p class="text-sm sm:text-base text-neutral-500">Open source AI-powered voice to text</p>
        </div>

        <!-- Live Demo Text -->
        <div class="max-w-2xl mx-auto mb-10 sm:mb-14 px-4 w-full">
          <div class="relative rounded-xl overflow-hidden border border-neutral-200 dark:border-neutral-800 bg-white dark:bg-neutral-900 shadow-sm">
            <!-- Main text area -->
            <div class="relative p-4 sm:p-6 min-h-[5rem] sm:min-h-[6rem] bg-neutral-100 dark:bg-neutral-800/50">
              <!-- Listening Waveform Animation -->
              <div v-if="isListening" class="flex items-center justify-center h-[3rem] sm:h-[4rem] gap-[1px] w-full">
                <div
                  v-for="(height, index) in barHeights"
                  :key="index"
                  class="w-[1.5px] sm:w-[2px] rounded-full bg-neutral-400 dark:bg-neutral-500 transition-all duration-75"
                  :style="{ height: `${height}%` }"
                />
              </div>
              <!-- Typed Text -->
              <p v-else class="text-base sm:text-lg text-neutral-700 dark:text-neutral-200 min-h-[3rem] sm:min-h-[4rem] text-center pb-6">
                {{ demoText }}<span v-if="isTyping" class="inline-block w-0.5 h-5 bg-neutral-900 dark:bg-white ml-0.5 animate-pulse" />
              </p>
              <!-- Copy button -->
              <button
                v-if="!isListening"
                class="absolute bottom-3 right-3 p-1.5 rounded hover:bg-neutral-200 dark:hover:bg-neutral-700 text-neutral-400 hover:text-neutral-600 dark:hover:text-neutral-300 transition-colors flex items-center gap-1.5"
                @click="copyText"
              >
                <span v-if="showCopied" class="text-xs text-green-500 font-medium">Copied!</span>
                <Copy class="size-4" />
              </button>
            </div>
            <!-- Bottom bar -->
            <div class="flex items-center justify-between px-4 py-2.5 bg-white dark:bg-neutral-900 border-t border-neutral-200 dark:border-neutral-700/50">
              <div class="flex items-center gap-2 text-neutral-500 dark:text-neutral-400">
                <Settings class="size-4" />
                <!-- Recording indicator with red dot -->
                <div v-if="isListening" class="flex items-center gap-2">
                  <span class="w-2 h-2 rounded-full bg-red-500 animate-pulse" />
                  <span class="text-sm">Recording</span>
                </div>
                <span v-else class="text-sm">Ready</span>
              </div>
              <div class="flex items-center gap-3">
                <!-- Toggle switch -->
                <button class="relative w-10 h-5 rounded-full bg-neutral-200 dark:bg-neutral-700 transition-colors">
                  <span class="absolute left-0.5 top-0.5 w-4 h-4 rounded-full bg-white dark:bg-neutral-400 shadow transition-transform" />
                </button>
                <!-- Mic icon -->
                <div class="p-1.5 rounded text-neutral-500 dark:text-neutral-400">
                  <Mic class="size-4" />
                </div>
              </div>
            </div>
          </div>
        </div>

        <!-- Download Button -->
        <div class="flex flex-col sm:flex-row gap-3 sm:gap-4 mb-4 px-4 w-full sm:w-auto">
          <Button
            size="lg"
            class="bg-neutral-900 dark:bg-white text-white dark:text-neutral-900 hover:bg-neutral-800 dark:hover:bg-neutral-100 gap-2 w-full sm:w-auto justify-center transition-colors"
          >
            <Terminal class="size-4" />
            Download for Linux
          </Button>
        </div>

        <p class="text-xs text-neutral-400 dark:text-neutral-600 mb-6">
          requires Linux with PipeWire or PulseAudio
        </p>

        <!-- Mac & Windows Banner -->
        <div class="inline-flex items-center gap-2 px-4 py-2 rounded-full border border-neutral-200 dark:border-neutral-800 bg-white dark:bg-neutral-900">
          <Apple class="size-4 text-neutral-400" />
          <Monitor class="size-4 text-neutral-400" />
          <span class="text-sm text-neutral-500 dark:text-neutral-400">Mac & Windows support coming soon</span>
        </div>
      </section>

      <!-- Features Section -->
      <section class="py-20 sm:py-32 border-t border-neutral-200 dark:border-neutral-800">
        <div class="container mx-auto px-4">
          <div class="grid sm:grid-cols-3 gap-8 sm:gap-12 max-w-4xl mx-auto">
            <div class="text-center sm:text-left">
              <div class="inline-flex items-center justify-center w-10 h-10 rounded-lg bg-neutral-100 dark:bg-neutral-800 mb-4">
                <svg class="w-5 h-5 text-neutral-600 dark:text-neutral-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z" />
                </svg>
              </div>
              <h3 class="text-base font-medium mb-2 text-neutral-900 dark:text-white">Hyper Accurate</h3>
              <p class="text-sm text-neutral-500 dark:text-neutral-400">
                Powered by state-of-the-art AI models. Understands accents, jargon, and context.
              </p>
            </div>
            <div class="text-center sm:text-left">
              <div class="inline-flex items-center justify-center w-10 h-10 rounded-lg bg-neutral-100 dark:bg-neutral-800 mb-4">
                <svg class="w-5 h-5 text-neutral-600 dark:text-neutral-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M13 10V3L4 14h7v7l9-11h-7z" />
                </svg>
              </div>
              <h3 class="text-base font-medium mb-2 text-neutral-900 dark:text-white">Blazing Fast</h3>
              <p class="text-sm text-neutral-500 dark:text-neutral-400">
                Real-time transcription as you speak. No waiting, no delays.
              </p>
            </div>
            <div class="text-center sm:text-left">
              <div class="inline-flex items-center justify-center w-10 h-10 rounded-lg bg-neutral-100 dark:bg-neutral-800 mb-4">
                <svg class="w-5 h-5 text-neutral-600 dark:text-neutral-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M10 20l4-16m4 4l4 4-4 4M6 16l-4-4 4-4" />
                </svg>
              </div>
              <h3 class="text-base font-medium mb-2 text-neutral-900 dark:text-white">Open Source</h3>
              <p class="text-sm text-neutral-500 dark:text-neutral-400">
                Fully transparent. Audit, contribute, or self-host.
              </p>
            </div>
          </div>
        </div>
      </section>
    </main>

    <footer class="border-t border-neutral-200 dark:border-white/5 py-8">
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
