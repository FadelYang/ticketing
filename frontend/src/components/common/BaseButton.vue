<script setup lang="ts">
import { computed } from 'vue'
import BaseIcon from '@/components/common/BaseIcon.vue'

interface Props {
  variant?: 'primary' | 'secondary' | 'danger' | 'ghost' | 'success'
  size?: 'sm' | 'md' | 'lg'
  type?: 'button' | 'submit' | 'reset'
  disabled?: boolean
  loading?: boolean
  block?: boolean
}

const props = withDefaults(defineProps<Props>(), {
  variant: 'primary',
  size: 'md',
  type: 'button',
  disabled: false,
  loading: false,
  block: false,
})

const classes = computed(() => {
  const base =
    'inline-flex items-center justify-center gap-space-xs rounded-lg font-headline-sm text-headline-sm transition-all focus:outline-none focus:ring-2 focus:ring-primary focus:ring-offset-2 active:scale-[0.99] disabled:cursor-not-allowed disabled:opacity-60 cursor-pointer shadow-sm'

  const sizes: Record<NonNullable<Props['size']>, string> = {
    sm: 'px-3 py-1.5 text-xs',
    md: 'px-space-lg py-2.5 text-headline-sm',
    lg: 'h-11 px-space-xl text-body-md font-semibold',
  }

  const variants: Record<NonNullable<Props['variant']>, string> = {
    primary: 'bg-primary-container text-on-primary hover:bg-primary',
    secondary:
      'border border-outline-variant bg-surface-container-lowest text-on-surface hover:bg-surface-container-low',
    danger: 'bg-error text-on-error hover:bg-on-error-container',
    ghost: 'bg-transparent shadow-none text-on-surface-variant hover:bg-surface-container-low hover:text-on-surface',
    success: 'bg-secondary text-on-secondary hover:opacity-95',
  }

  return [base, sizes[props.size], variants[props.variant], props.block ? 'w-full' : '']
    .filter(Boolean)
    .join(' ')
})
</script>

<template>
  <button :type="type" :disabled="disabled || loading" :class="classes">
    <BaseIcon v-if="loading" name="progress_activity" class="text-base animate-spin" />
    <slot />
  </button>
</template>
