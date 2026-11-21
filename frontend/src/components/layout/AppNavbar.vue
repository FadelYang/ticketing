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
  <header class="sticky top-0 z-10 border-b border-slate-200 bg-white/90 backdrop-blur">
    <div class="mx-auto flex max-w-4xl items-center justify-between px-4 py-3.5 sm:px-6 lg:px-8">
      <div class="flex items-center gap-2.5">
        <div
          class="flex h-8 w-8 items-center justify-center rounded-lg bg-indigo-600 text-sm font-bold text-white shadow-sm"
        >
          EX
        </div>
        <div>
          <h1 class="text-base font-semibold tracking-tight text-slate-900">Monorepo Starter</h1>
          <p class="text-xs text-slate-500">Go Fiber + pgx/v5 · Vue 3 + Pinia</p>
        </div>
      </div>

      <button
        type="button"
        title="Click to re-check backend health"
        class="inline-flex cursor-pointer items-center gap-2 rounded-full border border-slate-200 bg-slate-50 px-3 py-1 text-xs font-medium text-slate-700 transition-colors hover:bg-slate-100"
        @click="refresh"
      >
        <span class="h-2 w-2 rounded-full" :class="dotClass" />
        <span>{{ statusLabel }}</span>
      </button>
    </div>
  </header>
</template>
