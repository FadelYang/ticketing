<script setup lang="ts">
import type { TicketPriority } from '@/types/ticket'

interface Props {
  modelValue: TicketPriority
}

defineProps<Props>()

const emit = defineEmits<{
  'update:modelValue': [value: TicketPriority]
}>()

const options: {
  value: TicketPriority
  label: string
  activeClass: string
  inactiveClass: string
}[] = [
  {
    value: 'low',
    label: 'Low',
    activeClass: 'bg-surface-container-highest text-on-surface ring-2 ring-outline font-semibold',
    inactiveClass: 'bg-surface-container-low text-outline hover:bg-surface-container',
  },
  {
    value: 'medium',
    label: 'Medium',
    activeClass: 'bg-secondary-fixed text-on-secondary-fixed ring-2 ring-secondary font-semibold',
    inactiveClass: 'bg-surface-container-low text-on-surface-variant hover:bg-surface-container',
  },
  {
    value: 'high',
    label: 'High',
    activeClass: 'bg-primary-fixed text-on-primary-fixed-variant ring-2 ring-primary font-semibold',
    inactiveClass: 'bg-surface-container-low text-on-surface-variant hover:bg-surface-container',
  },
  {
    value: 'urgent',
    label: 'Urgent',
    activeClass: 'bg-error-container text-on-error-container ring-2 ring-error font-semibold',
    inactiveClass: 'bg-surface-container-low text-on-surface-variant hover:bg-surface-container',
  },
]
</script>

<template>
  <div class="flex flex-col gap-space-sm">
    <span class="font-label-md text-label-md text-on-surface font-semibold flex items-center gap-1">
      Priority <span class="text-error">*</span>
    </span>
    <div class="grid grid-cols-2 gap-2">
      <button
        v-for="opt in options"
        :key="opt.value"
        type="button"
        class="px-3 py-2 rounded-lg font-label-md text-label-md transition-all cursor-pointer text-center"
        :class="modelValue === opt.value ? opt.activeClass : opt.inactiveClass"
        @click="emit('update:modelValue', opt.value)"
      >
        {{ opt.label }}
      </button>
    </div>
  </div>
</template>
