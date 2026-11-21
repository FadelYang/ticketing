<script setup lang="ts">
import { computed } from 'vue'

interface Props {
  tone?: 'slate' | 'blue' | 'amber' | 'emerald' | 'rose'
}

const props = withDefaults(defineProps<Props>(), {
  tone: 'slate',
})

const toneClass = computed(() => {
  const map: Record<string, string> = {
    slate: 'bg-slate-100 text-slate-700 ring-slate-200',
    blue: 'bg-sky-50 text-sky-700 ring-sky-200',
    amber: 'bg-amber-50 text-amber-700 ring-amber-200',
    emerald: 'bg-emerald-50 text-emerald-700 ring-emerald-200',
    rose: 'bg-rose-50 text-rose-700 ring-rose-200',
  }
  return map[props.tone]
})
</script>

<template>
  <span
    class="inline-flex items-center rounded-full px-2.5 py-0.5 text-xs font-medium ring-1 ring-inset"
    :class="toneClass"
  >
    <slot />
  </span>
</template>
