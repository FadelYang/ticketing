<script setup lang="ts">
import { computed } from 'vue'

interface Props {
  variant?: 'primary' | 'secondary' | 'danger' | 'ghost'
  size?: 'sm' | 'md'
  type?: 'button' | 'submit' | 'reset'
  disabled?: boolean
}

const props = withDefaults(defineProps<Props>(), {
  variant: 'primary',
  size: 'md',
  type: 'button',
  disabled: false,
})

const classes = computed(() => {
  const base =
    'inline-flex items-center justify-center gap-1.5 rounded-lg font-medium transition-colors focus:outline-none focus:ring-2 focus:ring-offset-2 disabled:cursor-not-allowed disabled:opacity-50 cursor-pointer'

  const sizes: Record<string, string> = {
    sm: 'px-3 py-1.5 text-xs',
    md: 'px-4 py-2 text-sm',
  }

  const variants: Record<string, string> = {
    primary: 'bg-indigo-600 text-white hover:bg-indigo-700 focus:ring-indigo-500',
    secondary:
      'border border-slate-300 bg-white text-slate-700 hover:bg-slate-50 focus:ring-indigo-500',
    danger: 'bg-rose-600 text-white hover:bg-rose-700 focus:ring-rose-500',
    ghost: 'text-slate-600 hover:bg-slate-100 hover:text-slate-900 focus:ring-slate-400',
  }

  return [base, sizes[props.size], variants[props.variant]].join(' ')
})
</script>

<template>
  <button :type="type" :disabled="disabled" :class="classes">
    <slot />
  </button>
</template>
