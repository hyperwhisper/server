<script setup lang="ts">
import { Apple, Monitor, Terminal, Settings, Copy, Mic } from 'lucide-vue-next'

useHead({
  title: 'HyperWhisper - Open Source Voice to Text',
  meta: [
    { name: 'description', content: 'Type 3x faster, without lifting a finger' }
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
  }, 10)
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
    <!-- Cracked wall texture background -->
    <svg class="fixed inset-0 w-full h-full pointer-events-none opacity-[0.25] dark:opacity-[0.35] z-50" xmlns="http://www.w3.org/2000/svg">
      <defs>
        <pattern id="cracks" patternUnits="userSpaceOnUse" width="200" height="200">
          <g fill="none" stroke="currentColor" stroke-width="0.25" stroke-linecap="round">
            <!-- Cluster 1 - top left -->
            <path d="M10,0 Q12,15 8,30 T12,50" />
            <path d="M8,30 Q0,35 -5,45" />
            <path d="M12,50 Q20,55 30,52" />
            <!-- Cluster 2 - top center -->
            <path d="M80,5 Q75,20 82,35 T78,55 Q72,70 80,85" />
            <path d="M82,35 Q95,40 110,38" />
            <path d="M78,55 Q65,60 55,58" />
            <!-- Cluster 3 - top right -->
            <path d="M160,0 Q155,18 162,35 T155,60" />
            <path d="M162,35 Q175,40 190,37" />
            <path d="M155,60 Q145,65 135,62" />
            <!-- Cluster 4 - middle left -->
            <path d="M0,100 Q15,105 30,100 T55,108" />
            <path d="M30,100 Q35,115 32,130" />
            <path d="M55,108 Q60,95 68,85" />
            <!-- Cluster 5 - center -->
            <path d="M95,90 Q90,105 98,120 T92,145 Q88,160 95,175" />
            <path d="M98,120 Q110,125 125,122" />
            <path d="M92,145 Q80,150 68,147" />
            <path d="M95,175 Q105,180 118,177" />
            <!-- Cluster 6 - middle right -->
            <path d="M170,95 Q165,110 172,125 T168,150" />
            <path d="M172,125 Q185,130 200,127" />
            <path d="M168,150 Q155,155 142,152" />
            <!-- Cluster 7 - bottom left -->
            <path d="M25,165 Q20,180 28,195 T22,200" />
            <path d="M28,195 Q40,200 55,197" />
            <path d="M25,165 Q12,160 0,163" />
            <!-- Cluster 8 - bottom center -->
            <path d="M110,170 Q105,185 112,200" />
            <path d="M112,185 Q125,190 140,187" />
            <path d="M110,170 Q98,165 85,168" />
            <!-- Cluster 9 - bottom right -->
            <path d="M175,175 Q170,190 178,200" />
            <path d="M175,175 Q188,170 200,173" />
            <path d="M178,188 Q165,192 152,189" />
            <!-- Small accent cracks -->
            <path d="M45,20 Q48,30 45,40" />
            <path d="M130,15 Q133,28 128,38" />
            <path d="M185,55 Q190,65 187,78" />
            <path d="M15,70 Q20,82 17,95" />
            <path d="M140,75 Q145,88 142,100" />
            <path d="M55,140 Q60,152 57,165" />
            <path d="M150,140 Q155,150 152,162" />
            <path d="M70,195 Q75,200 73,200" />
            <!-- Tiny hairline cracks -->
            <path d="M35,55 Q38,62 36,70" />
            <path d="M115,45 Q118,52 116,60" />
            <path d="M180,30 Q183,38 181,48" />
            <path d="M5,130 Q10,138 8,148" />
            <path d="M65,115 Q70,122 68,132" />
            <path d="M195,115 Q200,123 198,133" />
            <path d="M40,180 Q45,188 43,198" />
            <path d="M125,155 Q130,163 128,173" />
            <path d="M90,60 Q93,68 91,78" />
            <path d="M160,165 Q163,172 161,182" />
          </g>
        </pattern>
      </defs>
      <rect width="100%" height="100%" fill="url(#cracks)" />
    </svg>
    <!-- Subtle grain overlay -->
    <div class="fixed inset-0 opacity-[0.015] dark:opacity-[0.02] pointer-events-none bg-[url('data:image/svg+xml;base64,PHN2ZyB4bWxucz0iaHR0cDovL3d3dy53My5vcmcvMjAwMC9zdmciIHdpZHRoPSIyMDAiIGhlaWdodD0iMjAwIj48ZmlsdGVyIGlkPSJuIj48ZmVUdXJidWxlbmNlIHR5cGU9ImZyYWN0YWxOb2lzZSIgYmFzZUZyZXF1ZW5jeT0iMC44IiBudW1PY3RhdmVzPSI0IiBzdGl0Y2hUaWxlcz0ic3RpdGNoIi8+PC9maWx0ZXI+PHJlY3Qgd2lkdGg9IjIwMCIgaGVpZ2h0PSIyMDAiIGZpbHRlcj0idXJsKCNuKSIvPjwvc3ZnPg==')]" />

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
              <!-- Mic icon -->
              <div class="p-1.5 rounded text-neutral-500 dark:text-neutral-400">
                <Mic class="size-4" />
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
