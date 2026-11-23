<script setup lang="ts">
import { computed } from 'vue'
import { useHealth } from '@/composables/useHealth'

const { status, database, refresh } = useHealth()

const statusLabel = computed(() => {
  if (status.value === 'checking') return 'Checking API...'
  if (status.value === 'healthy') return `API Online · DB ${database.value}`
  return 'API Offline'
})

const dotClass = computed(() => {
  if (status.value === 'healthy') return 'bg-emerald-500'
  if (status.value === 'checking') return 'bg-amber-400 animate-pulse'
  return 'bg-rose-500'
})
</script>

<template>
  <header class="sticky top-0 z-10 border-b border-outline-variant/40 bg-surface-container-lowest/90 backdrop-blur">
    <div class="mx-auto flex max-w-4xl items-center justify-between px-4 py-3.5 sm:px-6 lg:px-8">
      <RouterLink to="/" class="flex items-center gap-2.5">
        <div
          class="flex h-8 w-8 items-center justify-center rounded-lg bg-primary-container text-sm font-bold text-on-primary shadow-sm"
        >
          ST
        </div>
        <div>
          <h1 class="text-base font-semibold tracking-tight text-on-surface">Simple Ticketing</h1>
          <p class="text-xs text-on-surface-variant">Go Fiber + pgx/v5 · Vue 3 + Pinia</p>
        </div>
      </RouterLink>

      <div class="flex items-center gap-3">
        <nav class="flex items-center gap-2 text-xs font-medium">
          <RouterLink
            to="/login"
            class="rounded-lg px-2.5 py-1.5 text-on-surface-variant hover:bg-surface-container-low hover:text-on-surface transition-colors"
          >
            Login
          </RouterLink>
          <RouterLink
            to="/register"
            class="rounded-lg bg-primary-container px-3 py-1.5 text-on-primary hover:bg-primary transition-colors"
          >
            Sign up
          </RouterLink>
        </nav>

        <button
          type="button"
          title="Click to re-check backend health"
          class="inline-flex cursor-pointer items-center gap-2 rounded-full border border-outline-variant/50 bg-surface-container-low px-3 py-1 text-xs font-medium text-on-surface-variant transition-colors hover:bg-surface-container"
          @click="refresh"
        >
          <span class="h-2 w-2 rounded-full" :class="dotClass" />
          <span>{{ statusLabel }}</span>
        </button>
      </div>
    </div>
  </header>
</template>
