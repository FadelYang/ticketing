<script setup lang="ts">
import BaseIcon from '@/components/common/BaseIcon.vue'

interface Props {
  variant?: 'error' | 'success' | 'info'
  message: string
}

withDefaults(defineProps<Props>(), {
  variant: 'error',
})
</script>

<template>
  <div
    role="alert"
    class="flex items-start gap-space-sm rounded-lg px-space-md py-2.5 text-body-sm font-body-sm"
    :class="
      variant === 'error'
        ? 'bg-error-container text-on-error-container'
        : variant === 'success'
          ? 'bg-secondary-fixed text-on-secondary-fixed'
          : 'bg-surface-container text-on-surface'
    "
  >
    <BaseIcon
      :name="variant === 'error' ? 'error' : variant === 'success' ? 'check_circle' : 'info'"
      class="text-base shrink-0 mt-0.5"
    />
    <span>{{ message }}</span>
  </div>
</template>
