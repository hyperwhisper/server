<script setup lang="ts">
import { Apple, Monitor, Terminal } from 'lucide-vue-next'

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
  typeText(demoTexts[currentTextIndex.value], () => {
    setTimeout(() => {
      currentTextIndex.value = (currentTextIndex.value + 1) % demoTexts.length
      cycleTexts()
    }, 3000)
  })
}

onMounted(() => {
  cycleTexts()
})
</script>

<template>
  <div class="min-h-screen bg-white dark:bg-black text-neutral-900 dark:text-white overflow-hidden transition-colors duration-300">
    <!-- Subtle grain texture overlay -->
    <div class="fixed inset-0 opacity-[0.02] dark:opacity-[0.015] pointer-events-none bg-[url('data:image/svg+xml;base64,PHN2ZyB4bWxucz0iaHR0cDovL3d3dy53My5vcmcvMjAwMC9zdmciIHdpZHRoPSIzMDAiIGhlaWdodD0iMzAwIj48ZmlsdGVyIGlkPSJhIiB4PSIwIiB5PSIwIj48ZmVUdXJidWxlbmNlIGJhc2VGcmVxdWVuY3k9Ii43NSIgc3RpdGNoVGlsZXM9InN0aXRjaCIgdHlwZT0iZnJhY3RhbE5vaXNlIi8+PC9maWx0ZXI+PHJlY3Qgd2lkdGg9IjMwMCIgaGVpZ2h0PSIzMDAiIGZpbHRlcj0idXJsKCNhKSIgb3BhY2l0eT0iMSIvPjwvc3ZnPg==')]" />

    <AppNavbar />

    <main class="relative">
      <!-- Hero Section -->
      <section class="min-h-screen flex flex-col items-center justify-center px-4 pt-16">
        <!-- Abstract Logo/Icon -->
        <div class="relative mb-8 sm:mb-12">
          <div class="w-24 h-24 sm:w-32 sm:h-32 relative">
            <!-- Glowing orb effect -->
            <div class="absolute inset-0 rounded-full bg-gradient-to-br from-neutral-300 dark:from-white/20 to-transparent blur-2xl" />
            <svg viewBox="0 0 100 100" class="w-full h-full relative z-10">
              <!-- Sound wave / voice visualization -->
              <defs>
                <linearGradient id="waveGradient" x1="0%" y1="0%" x2="100%" y2="100%">
                  <stop offset="0%" class="[stop-color:theme(colors.neutral.800)] dark:[stop-color:#ffffff]" style="stop-opacity:0.9" />
                  <stop offset="100%" class="[stop-color:theme(colors.neutral.400)] dark:[stop-color:#666666]" style="stop-opacity:0.6" />
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
          <span class="bg-gradient-to-r from-neutral-900 via-neutral-900 to-neutral-500 dark:from-white dark:via-white dark:to-neutral-500 bg-clip-text text-transparent">
            Type 3x faster,
          </span>
          <br />
          <span class="bg-gradient-to-r from-neutral-600 to-neutral-400 dark:from-neutral-300 dark:to-neutral-600 bg-clip-text text-transparent">
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
          <div class="relative rounded-xl border border-neutral-200 dark:border-white/10 bg-neutral-50 dark:bg-white/[0.02] p-4 sm:p-6 backdrop-blur-sm">
            <p class="text-base sm:text-lg text-neutral-700 dark:text-neutral-300 min-h-[3rem] sm:min-h-[4rem]">
              {{ demoText }}<span v-if="isTyping" class="inline-block w-0.5 h-5 bg-neutral-900/70 dark:bg-white/70 ml-0.5 animate-pulse" />
            </p>
          </div>
        </div>

        <!-- Download Button -->
        <div class="flex flex-col sm:flex-row gap-3 sm:gap-4 mb-4 px-4 w-full sm:w-auto">
          <Button
            size="lg"
            class="bg-neutral-900 dark:bg-white text-white dark:text-black hover:bg-neutral-800 dark:hover:bg-neutral-200 gap-2 w-full sm:w-auto justify-center"
          >
            <Terminal class="size-4" />
            Download for Linux
          </Button>
        </div>

        <p class="text-xs text-neutral-400 dark:text-neutral-600 mb-6">
          requires Linux with PipeWire or PulseAudio
        </p>

        <!-- Mac & Windows Banner -->
        <div class="inline-flex items-center gap-2 px-4 py-2 rounded-full border border-neutral-200 dark:border-white/10 bg-neutral-50 dark:bg-white/[0.02]">
          <Apple class="size-4 text-neutral-400" />
          <Monitor class="size-4 text-neutral-400" />
          <span class="text-sm text-neutral-500 dark:text-neutral-400">Mac & Windows support coming soon</span>
        </div>
      </section>

      <!-- Features Section -->
      <section class="py-20 sm:py-32 border-t border-neutral-200 dark:border-white/5">
        <div class="container mx-auto px-4">
          <div class="grid sm:grid-cols-3 gap-8 sm:gap-12 max-w-4xl mx-auto">
            <div class="text-center sm:text-left">
              <h3 class="text-lg font-medium mb-2">100% Local</h3>
              <p class="text-sm text-neutral-500">
                All processing happens on your device. Your voice never leaves your machine.
              </p>
            </div>
            <div class="text-center sm:text-left">
              <h3 class="text-lg font-medium mb-2">Blazing Fast</h3>
              <p class="text-sm text-neutral-500">
                Real-time transcription as you speak. No waiting, no delays.
              </p>
            </div>
            <div class="text-center sm:text-left">
              <h3 class="text-lg font-medium mb-2">Open Source</h3>
              <p class="text-sm text-neutral-500">
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
