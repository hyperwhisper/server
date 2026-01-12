<script setup lang="ts">
import { Sun, Moon, Monitor } from 'lucide-vue-next'

const colorMode = useColorMode()

const themes = [
  { value: 'light', label: 'Light', icon: Sun },
  { value: 'dark', label: 'Dark', icon: Moon },
  { value: 'system', label: 'System', icon: Monitor },
] as const

function setTheme(theme: 'light' | 'dark' | 'system') {
  colorMode.preference = theme
}
</script>

<template>
  <DropdownMenu>
    <DropdownMenuTrigger as-child>
      <button
        class="inline-flex items-center justify-center size-9 rounded-md text-neutral-500 dark:text-neutral-400 hover:text-neutral-900 dark:hover:text-white hover:bg-neutral-100 dark:hover:bg-white/10 transition-colors"
      >
        <Sun class="size-5 rotate-0 scale-100 transition-all dark:-rotate-90 dark:scale-0" />
        <Moon class="absolute size-5 rotate-90 scale-0 transition-all dark:rotate-0 dark:scale-100" />
        <span class="sr-only">Toggle theme</span>
      </button>
    </DropdownMenuTrigger>
    <DropdownMenuContent align="end" class="bg-white dark:bg-neutral-900 border-neutral-200 dark:border-white/10">
      <DropdownMenuItem
        v-for="theme in themes"
        :key="theme.value"
        @click="setTheme(theme.value)"
        class="gap-2 text-neutral-700 dark:text-neutral-300 focus:text-neutral-900 dark:focus:text-white focus:bg-neutral-100 dark:focus:bg-white/10"
      >
        <component :is="theme.icon" class="size-4" />
        {{ theme.label }}
      </DropdownMenuItem>
    </DropdownMenuContent>
  </DropdownMenu>
</template>
